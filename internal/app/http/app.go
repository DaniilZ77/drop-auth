package http

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
	sl "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
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
	log        *slog.Logger
}

func New(
	ctx context.Context,
	cfg *config.Config,
	log *slog.Logger,
) *App {
	// creds, err := credentials.NewClientTLSFromFile(cfg.Cert, "") nolint
	// if err != nil {
	// 	logger.Log().Fatal(ctx, "failed to create server TLS credentials: %v", err)
	// }

	conn, err := grpc.NewClient(cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Log(ctx, sl.LevelFatal, "failed to dial server", sl.Err(err))
		os.Exit(1)
	}

	gwmux := runtime.NewServeMux()
	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	// Register user
	err = userv1.RegisterUserServiceHandler(ctx, gwmux, conn)
	if err != nil {
		log.Log(ctx, sl.LevelFatal, "failed to register gateway", sl.Err(err))
		os.Exit(1)
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
		log:        log,
	}
}

func (app *App) MustRun(ctx context.Context) {
	if err := app.Run(); err != nil {
		app.log.Log(ctx, sl.LevelFatal, "failed to run http server", sl.Err(err))
		os.Exit(1)
	}
}

func (app *App) Run() error {
	app.log.Info("http server started", slog.String("port", app.httpServer.Addr))
	// return app.httpServer.ListenAndServeTLS(app.cert, app.key) nolint
	return app.httpServer.ListenAndServe()
}

func (app *App) Stop(ctx context.Context) {
	app.log.Info("stopping http server")

	if err := app.httpServer.Shutdown(ctx); err != nil {
		app.log.Log(ctx, sl.LevelFatal, "failed to shutdown http server", sl.Err(err))
		os.Exit(1)
	}
}
