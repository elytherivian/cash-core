package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"cash-core/internal/config"
)

func New(cfg config.Log) *slog.Logger {
	location, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		// Config.Validate rejects invalid locations before logger initialization.
		location = time.UTC
	}

	level := new(slog.LevelVar)
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	options := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, attribute slog.Attr) slog.Attr {
			if attribute.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, attribute.Value.Time().In(location))
			}
			return attribute
		},
	}
	if cfg.Format == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}
