package logger

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"

	LevelTrace = slog.Level(8)
	LevelFatal = slog.Level(-12)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

func New(env string) *slog.Logger {
	var log *slog.Logger

	opts := &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level, ok := a.Value.Any().(slog.Level)
				if !ok {
					return a
				}

				levelLabel, ok := levelNames[level]
				if !ok {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
		AddSource: true,
	}

	switch env {
	case envLocal:
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	case envDev:
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	case envProd:
		opts.Level = slog.LevelInfo
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	default:
		panic("unknown env")
	}

	return log
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
