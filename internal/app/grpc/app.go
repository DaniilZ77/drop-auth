package gprc

import (
	"context"
	"log/slog"
	"net"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
	user "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/grpc"
	userservice "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type App struct {
	gRPCServer *grpc.Server
	port       string
	log        *slog.Logger
}

func New(
	ctx context.Context,
	cfg *config.Config,
	userService *userservice.UserService,
	log *slog.Logger,
) *App {
	// Methods that require authentication
	requireAuth := map[string]bool{
		"/user.UserService/UpdateUser":  true,
		"/user.UserService/DeleteUser":  true,
		"/user.UserService/Login":       true,
		"/user.UserService/AddAdmin":    true,
		"/user.UserService/DeleteAdmin": true,
		"/user.UserService/GetAdmins":   true,
	}

	requireAdmin := map[string]bool{
		"/user.UserService/AddAdmin":    true,
		"/user.UserService/DeleteAdmin": true,
		"/user.UserService/GetAdmins":   true,
	}

	var opts []grpc.ServerOption

	// Logger
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived,
			logging.PayloadSent,
		),
	}

	// Recovery
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Error("recovered from panic", slog.Any("error", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	secrets := map[string]string{
		"bearer": cfg.Auth.JwtSecret,
		"tma":    cfg.Auth.TmaSecret,
	}

	opts = append(opts, grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLogger(log), loggingOpts...),
		user.AuthMiddleware(secrets, requireAuth, requireAdmin),
	))

	// TLS nolint
	// creds, err := credentials.NewServerTLSFromFile(cfg.Cert, cfg.Key)
	// if err != nil {
	// 	logger.Log().Fatal(ctx, "failed to create server TLS credentials: %v", err)
	// }

	opts = append(opts, grpc.Creds(insecure.NewCredentials()))

	// Create gRPC server
	gRPCServer := grpc.NewServer(opts...)

	// Register services
	user.Register(gRPCServer, userService, userService, userService, log)

	return &App{
		gRPCServer: gRPCServer,
		port:       cfg.GrpcPort,
		log:        log,
	}
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			l.DebugContext(ctx, msg, fields...)
		case logging.LevelInfo:
			l.InfoContext(ctx, msg, fields...)
		case logging.LevelWarn:
			l.WarnContext(ctx, msg, fields...)
		case logging.LevelError:
			l.ErrorContext(ctx, msg, fields...)
		default:
			l.Debug("unknown level", slog.Any("level", lvl))

			panic("unknown level")
		}
	})
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		panic(err)
	}
}

func (a *App) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", a.port)
	if err != nil {
		return err
	}

	a.log.Info("grpc server started", slog.String("port", a.port))

	if err := a.gRPCServer.Serve(l); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info("stopping grpc server")

	a.gRPCServer.GracefulStop()
}
