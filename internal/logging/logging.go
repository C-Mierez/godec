package logging

import (
	"log/slog"
	"os"
)

// Setup initializes the global slog logger with JSON output.
func Setup(env string) {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))
}
