package testhelpers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	DB *pgxpool.Pool
}

func CreatePostgresContainer(ctx context.Context, n *testcontainers.DockerNetwork) (*PostgresContainer, *string, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16.4-alpine",
		postgres.WithDatabase("drop-auth"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts(filepath.Join("..", "internal", "db", "migrations", "000001_initial.up.sql")),
		postgres.BasicWaitStrategies(),
		network.WithNetwork(nil, n),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	host, err := pgContainer.ContainerIP(ctx)
	if err != nil {
		return nil, nil, err
	}

	db, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		DB:                db,
	}, &host, nil
}

type RedisContainer struct {
	*redis.RedisContainer
}

func CreateRedisContainer(ctx context.Context, n *testcontainers.DockerNetwork) (*RedisContainer, *string, error) {
	redisContainer, err := redis.Run(ctx,
		"redis:alpine",
		redis.WithLogLevel(redis.LogLevelVerbose),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
		),
		network.WithNetwork(nil, n),
	)
	if err != nil {
		return nil, nil, err
	}

	host, err := redisContainer.ContainerIP(ctx)
	if err != nil {
		return nil, nil, err
	}

	return &RedisContainer{
		RedisContainer: redisContainer,
	}, &host, nil
}

type BackendContainer struct {
	testcontainers.Container
	baseURL string
}

func CreateBackendContainer(ctx context.Context,
	networkName string,
	databaseHost string,
	redisHost string) (*BackendContainer, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    filepath.Join(".."),
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"CONFIG_PATH":  filepath.Join("config", "local_tests.yaml"),
			"DATABASE_URL": fmt.Sprintf("postgres://postgres:postgres@%s:5432/drop-auth", databaseHost),
			"REDIS_URL":    fmt.Sprintf("redis://default:redis@%s:6379/0", redisHost),
		},
		Networks:   []string{"bridge", networkName},
		WaitingFor: wait.ForHTTP("/health"),
	}
	backendContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := backendContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := backendContainer.MappedPort(ctx, "8080")
	if err != nil {
		return nil, err
	}

	return &BackendContainer{
		Container: backendContainer,
		baseURL:   "http://" + ip + ":" + port.Port(),
	}, nil
}

const tmaSecret = "5768337691:AAH5YkoiEuPk8-FZa32hStHTqXiLPtAEhx8"

type Option func(req *http.Request)

func WithTmaToken(params map[string]string) Option {
	return func(req *http.Request) {
		authDate := time.Now().Add(time.Hour)
		payload := map[string]string{
			"auth_date": fmt.Sprintf("%d", authDate.Unix()),
			"user":      fmt.Sprintf(`{"username":%q,"first_name":%q,"last_name":%q}`, params["username"], params["first_name"], params["last_name"]),
		}
		hash := initdata.Sign(payload, tmaSecret, authDate)
		token := "auth_date=" + payload["auth_date"] + "&user=" + payload["user"] + "&hash=" + hash
		req.Header.Set("authorization", "tma "+token)
	}
}

func WithBearerToken(token string) Option {
	return func(req *http.Request) {
		req.Header.Set("authorization", "bearer "+token)
	}
}

func (b *BackendContainer) bodyRequest(method string, path string, body string, opts ...Option) (*http.Response, error) {
	url, err := url.JoinPath(b.baseURL, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (b *BackendContainer) PostRequest(path string, body string, opts ...Option) (*http.Response, error) {
	return b.bodyRequest(http.MethodPost, path, body, opts...)
}

func (b *BackendContainer) PatchRequest(path string, body string, opts ...Option) (*http.Response, error) {
	return b.bodyRequest(http.MethodPatch, path, body, opts...)
}

func (b *BackendContainer) urlvalsRequest(method, path string, vals url.Values, opts ...Option) (*http.Response, error) {
	if vals == nil {
		vals = url.Values{}
	}

	url, err := url.JoinPath(b.baseURL, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url+"?"+vals.Encode(), nil)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (b *BackendContainer) DeleteRequest(path string, vals url.Values, opts ...Option) (*http.Response, error) {
	return b.urlvalsRequest(http.MethodDelete, path, vals, opts...)
}

func (b *BackendContainer) GetRequest(path string, vals url.Values, opts ...Option) (*http.Response, error) {
	return b.urlvalsRequest(http.MethodGet, path, vals, opts...)
}

type Network struct {
	*testcontainers.DockerNetwork
}

func CreateNetwork(ctx context.Context) (*Network, error) {
	network, err := network.New(ctx, network.WithDriver("bridge"))
	if err != nil {
		return nil, err
	}

	return &Network{network}, nil
}
