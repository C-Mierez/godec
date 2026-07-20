// Package logging provides structured logging setup for the application.
package logging

import (
	"log/slog"
	"os"

	"github.com/c-mierez/godec/internal/config"
)

// Setup initializes the global slog logger with JSON output.
func Setup(env string) {
	level := slog.LevelInfo
	if env == config.EnvDevelopment {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))
}
