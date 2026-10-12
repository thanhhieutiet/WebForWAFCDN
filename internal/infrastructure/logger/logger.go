package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New initializes a slog JSON logger with the specified level.
// It maps the standard timestamp attribute key to "ts" and sets the logger as default.
func New(levelStr string) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				a.Key = "ts"
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}
