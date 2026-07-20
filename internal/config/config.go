// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Well-known environment names.
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// ServerEnv holds HTTP server configuration from environment variables.
type ServerEnv struct {
	ServerAddress      string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	Env                string `env:"ENV" envDefault:"development"`
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://127.0.0.1:8080,http://localhost:8080"`
}

// DatabaseEnv holds the PostgreSQL connection string.
type DatabaseEnv struct {
	URL string `env:"DATABASE_URL"`
}

// Config is the top-level application configuration loaded from environment variables.
type Config struct {
	Server struct {
		ServerEnv
	}
	Database struct {
		DatabaseEnv
	}
}

// Load reads .env and parses environment variables into a Config.
func Load() (*Config, error) {
	cfg := &Config{}

	// Load .env file if present. When running in Docker or CI, environment
	// variables are injected directly and no .env file exists — that is fine.
	// Only fatal on parse errors, not on missing file.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("loading .env file: %w", err)
	}

	// Load server configuration from environment variables
	var serverEnv ServerEnv

	if err := env.ParseWithOptions(&serverEnv, env.Options{RequiredIfNoDef: true}); err != nil {
		return nil, err
	}
	cfg.Server.ServerEnv = serverEnv

	// Load database configuration from environment variables
	var databaseEnv DatabaseEnv
	if err := env.ParseWithOptions(&databaseEnv, env.Options{RequiredIfNoDef: true}); err != nil {
		return nil, err
	}
	cfg.Database.DatabaseEnv = databaseEnv

	return cfg, nil
}
