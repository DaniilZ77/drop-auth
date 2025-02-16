package http

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	userv1 "github.com/MAXXXIMUS-tropical-milkshake/beatflow-protos/gen/go/user"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	httpServer *http.Server
	cert       string
	key        string
}

func New(ctx context.Context, cfg *config.Config) *App {
	// creds, err := credentials.NewClientTLSFromFile(cfg.Cert, "") nolint
	// if err != nil {
	// 	logger.Log().Fatal(ctx, "failed to create server TLS credentials: %v", err)
	// }

	conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log().Fatal(ctx, "failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	// Register user
	err = userv1.RegisterUserServiceHandler(ctx, gwmux, conn)
	if err != nil {
		logger.Log().Fatal(ctx, "failed to register gateway: %v", err)
	}

	// Cors
	withCors := cors.AllowAll().Handler(mux)

	// Server
	gwServer := &http.Server{
		Addr:              cfg.HTTPPort,
		Handler:           withCors,
		ReadHeaderTimeout: time.Duration(cfg.ReadTimeout) * time.Second,
	}

	return &App{
		httpServer: gwServer,
		cert:       cfg.Cert,
		key:        cfg.Key,
	}
}

func (app *App) MustRun(ctx context.Context) {
	if err := app.Run(ctx); err != nil {
		logger.Log().Fatal(ctx, "Failed to run http server: %v", err)
	}
}

func (app *App) Run(ctx context.Context) error {
	logger.Log().Info(ctx, "http server started on %v", app.httpServer.Addr)
	return app.httpServer.ListenAndServeTLS(app.cert, app.key)
}

func (app *App) Stop(ctx context.Context) {
	logger.Log().Info(ctx, "stopping http server")

	if err := app.httpServer.Shutdown(ctx); err != nil {
		logger.Log().Fatal(ctx, "failed to shutdown http server: %v", err)
	}
}
