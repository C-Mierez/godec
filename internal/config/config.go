package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type ServerEnv struct {
	ServerAddress      string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	Env                string `env:"ENV" envDefault:"development"`
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://127.0.0.1:8080,http://localhost:8080"`
}

type DatabaseEnv struct {
	URL string `env:"DATABASE_URL"`
}

type InternalEnv struct {
	GOOSE_DRIVER        string `env:"GOOSE_DRIVER"`
	GOOSE_DBSTRING      string `env:"GOOSE_DBSTRING"`
	GOOSE_MIGRATION_DIR string `env:"GOOSE_MIGRATION_DIR"`
	GOOSE_TABLE         string `env:"GOOSE_TABLE"`
}

type Config struct {
	Server struct {
		ServerEnv
	}
	Database struct {
		DatabaseEnv
	}
	internal struct {
		InternalEnv
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
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
