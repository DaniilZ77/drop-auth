package app

import (
	"context"
	"os"

	grpcapp "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/app/grpc"
	httpapp "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/app/http"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
	sl "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/postgres"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/redis"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	userservice "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/service"
	userstore "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/store"
)

type App struct {
	GRPCServer *grpcapp.App
	HTTPServer *httpapp.App
	Pg         *postgres.Postgres
	Rdb        *redis.Redis
}

func New(ctx context.Context, cfg *config.Config) *App {
	log := sl.New(cfg.Env)

	// Postgres connection
	pg, err := postgres.New(ctx, cfg.DB.URL, log)
	if err != nil {
		log.Log(ctx, sl.LevelFatal, "error with connection to database", sl.Err(err))
		os.Exit(1)
	}

	// Redis connection
	rdb, err := redis.New(ctx, redis.Config{
		Addr:     cfg.DB.RedisAddr,
		Password: cfg.DB.RedisPassword,
		DB:       cfg.DB.RedisDB,
	}, log)
	if err != nil {
		log.Log(ctx, sl.LevelFatal, "error with connection to redis", sl.Err(err))
		os.Exit(1)
	}

	// Auth config
	authConfig := model.AuthConfig{
		Secret:          cfg.JWTSecret,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	}

	// Store
	userStore := userstore.NewUserStore(pg, log)
	refreshTokenStore := userstore.NewRefreshTokenStore(rdb)

	// Service
	userService := userservice.New(
		userStore,
		userStore,
		refreshTokenStore,
		refreshTokenStore,
		authConfig,
		log,
	)

	// gRPC server
	gRPCApp := grpcapp.New(ctx, cfg, userService, log)

	// HTTP server
	httpServer := httpapp.New(ctx, cfg, log)

	return &App{
		GRPCServer: gRPCApp,
		HTTPServer: httpServer,
		Pg:         pg,
		Rdb:        rdb,
	}
}
