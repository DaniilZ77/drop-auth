package app

import (
	"context"
	"log/slog"

	grpcapp "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/app/grpc"
	httpapp "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/app/http"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
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

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) *App {
	// Postgres connection
	pg, err := postgres.New(ctx, cfg.DatabaseURL, log)
	if err != nil {
		panic(err)
	}

	// Redis connection
	rdb, err := redis.New(ctx, cfg.RedisURL, log)
	if err != nil {
		panic(err)
	}

	// Auth config
	authConfig := model.AuthConfig{
		Secret:          cfg.Auth.JwtSecret,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
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
