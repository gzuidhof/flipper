package telemetry

import (
	"io"
	"log/slog"

	"github.com/gzuidhof/flipper/config/cfgmodel"
)

func getLogLevel(cfgLevel string) slog.Level {
	switch cfgLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		panic("unsupported log level " + cfgLevel)
	}
}

// SetupLogger sets up a logger with the given configuration and writer.
func SetupLogger(cfg cfgmodel.LoggingConfig, w io.Writer) *slog.Logger {
	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: getLogLevel(cfg.Level),
		})
	case "text":
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: getLogLevel(cfg.Level),
		})
	default:
		panic("unsupported log format " + cfg.Format)
	}

	return slog.New(handler)
}
