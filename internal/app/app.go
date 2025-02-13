package app

import (
	"context"

	grpcapp "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/app/grpc"
	httpapp "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/app/http"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
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
	// Init logger
	logger.New(cfg.Log.Level)

	// Postgres connection
	pg, err := postgres.New(ctx, cfg.DB.URL)
	if err != nil {
		logger.Log().Fatal(ctx, "error with connection to database: %s", err.Error())
	}

	// Redis connection
	rdb, err := redis.New(ctx, redis.Config{
		Addr:     cfg.DB.RedisAddr,
		Password: cfg.DB.RedisPassword,
		DB:       cfg.DB.RedisDB,
	})

	// Auth config
	authConfig := model.AuthConfig{
		Secret:          cfg.JWTSecret,
		AccessTokenTTL:  cfg.AccessTokenTTL,
		RefreshTokenTTL: cfg.RefreshTokenTTL,
	}

	// Store
	userStore := userstore.NewUserStore(pg)
	refreshTokenStore := userstore.NewRefreshTokenStore(rdb)

	// Service
	userService := userservice.New(userStore, userStore, refreshTokenStore, refreshTokenStore, authConfig)

	// gRPC server
	gRPCApp := grpcapp.New(ctx, cfg, userService)

	// HTTP server
	httpServer := httpapp.New(ctx, cfg)

	return &App{
		GRPCServer: gRPCApp,
		HTTPServer: httpServer,
		Pg:         pg,
		Rdb:        rdb,
	}
}
