package http

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
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

	conn, err := grpc.NewClient(cfg.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	gwmux := runtime.NewServeMux()
	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	// Register user
	err = userv1.RegisterUserServiceHandler(ctx, gwmux, conn)
	if err != nil {
		panic(err)
	}

	// Cors
	withCors := cors.AllowAll().Handler(mux)

	// Server
	gwServer := &http.Server{
		Addr:    cfg.HttpPort,
		Handler: withCors,
	}

	return &App{
		httpServer: gwServer,
		cert:       cfg.Tls.Cert,
		key:        cfg.Tls.Key,
		log:        log,
	}
}

func (app *App) MustRun(ctx context.Context) {
	if err := app.Run(); err != nil {
		panic(err)
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
		panic(err)
	}
}
