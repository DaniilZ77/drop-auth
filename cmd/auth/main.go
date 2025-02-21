package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/app"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/config"
	sl "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()
	log := sl.New(cfg.Env)

	application := app.New(ctx, cfg, log)

	// Closing DBs
	defer application.Pg.Close(ctx)
	defer application.Rdb.Close()

	go func() { application.GRPCServer.MustRun(ctx) }()

	go func() { application.HTTPServer.MustRun(ctx) }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	// Stopping server
	application.GRPCServer.Stop(ctx)
	application.HTTPServer.Stop(ctx)
}
