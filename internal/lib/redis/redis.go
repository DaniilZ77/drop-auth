package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sl "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/redis/go-redis/v9"
)

const (
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
	_defaultMaxPoolSize  = 10
)

type Redis struct {
	connAttempts int
	connTimeout  time.Duration
	maxPoolSize  int
	*redis.Client
}

type Config struct {
	Addr     string
	Password string
	DB       int
}

func New(ctx context.Context, config Config, log *slog.Logger, opts ...Option) (*Redis, error) {
	rdb := &Redis{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
		maxPoolSize:  _defaultMaxPoolSize,
	}

	// Custom options
	for _, opt := range opts {
		opt(rdb)
	}

	var db *redis.Client
	var err error

	for rdb.connAttempts > 0 {
		db = redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
			PoolSize: rdb.maxPoolSize,
		})

		_, err = db.Ping(ctx).Result()
		if err == nil {
			rdb.Client = db
			break
		}
		log.Debug("redis is trying to connect", slog.Any("attempts left", rdb.connAttempts))

		time.Sleep(rdb.connTimeout)

		rdb.connAttempts--
	}

	if err != nil {
		log.Error("failed to connect to redis", sl.Err(err))
		return nil, err
	}

	return rdb, nil
}

func (r *Redis) Close() error {
	if err := r.Client.Close(); err != nil {
		return fmt.Errorf("error closing redis client: %w", err)
	}

	return nil
}
