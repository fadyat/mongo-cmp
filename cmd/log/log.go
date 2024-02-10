package log

import (
	"log/slog"
	"os"
)

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func JsonLogger(level string) *slog.Logger {
	lvl, ok := levels[level]
	if !ok {
		lvl = slog.LevelInfo
	}

	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		}),
	)
}
