package config

import (
	"github.com/caarlos0/env/v11"
)

type ServerEnv struct {
	Port string `env:"PORT" envDefault:"8080"`
	Env  string `env:"ENV" envDefault:"development"`
}

type DatabaseEnv struct {
	URL string `env:"DATABASE_URL"`
}

type Config struct {
	Server struct {
		ServerEnv
	}
	Database struct {
		DatabaseEnv
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Load server configuration from environment variables
	var serverEnv ServerEnv
	err := env.ParseWithOptions(&serverEnv, env.Options{RequiredIfNoDef: true})
	if err != nil {
		return nil, err
	}
	cfg.Server.ServerEnv = serverEnv

	// Load database configuration from environment variables
	var databaseEnv DatabaseEnv
	err = env.ParseWithOptions(&databaseEnv, env.Options{RequiredIfNoDef: true})
	if err != nil {
		return nil, err
	}
	cfg.Database.DatabaseEnv = databaseEnv

	return cfg, nil
}
