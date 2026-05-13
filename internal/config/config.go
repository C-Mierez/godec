package config

import (
	"os"
)

type Config struct {
	Server struct {
		Port string
		Env  string
	}
	Database struct {
		URL string
	}
}

const (
	DefaultPort = "8080"
	DefaultEnv  = "development"
)

func Load() (*Config, error) {
	cfg := &Config{}

	// Server config
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	cfg.Server.Port = port

	env := os.Getenv("ENV")
	if env == "" {
		env = DefaultEnv
	}
	cfg.Server.Env = env

	// Database config
	cfg.Database.URL = os.Getenv("DATABASE_URL")

	return cfg, nil
}
