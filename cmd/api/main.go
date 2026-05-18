package main

import (
	"context"
	"log"
	"time"

	apikey "github.com/c-mierez/godec/internal/apikey"
	"github.com/c-mierez/godec/internal/config"
	"github.com/c-mierez/godec/internal/health"
	"github.com/c-mierez/godec/internal/media"
	postgres "github.com/c-mierez/godec/internal/postgres"
	db "github.com/c-mierez/godec/internal/postgres/db"
	tenantpkg "github.com/c-mierez/godec/internal/tenant"
	"github.com/c-mierez/godec/pkg/graceful"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	graceful.RunWithGracefulShutdown(
		func(executionContext context.Context) {
			pool, err := pgxpool.New(executionContext, cfg.Database.URL)
			if err != nil {
				log.Printf("failed to create db pool: %v", err)
				return
			}
			defer pool.Close()

			queries := db.New(pool)
			apiKeyStore := postgres.NewApiKeyStore(queries)
			tenantStore := postgres.NewTenantStore(queries)
			service := apikey.NewService(apiKeyStore)
			tenantService := tenantpkg.NewService(tenantStore)

			e := echo.New()
			e.Use(middleware.RequestLogger(), middleware.Recover())

			health := &health.Handler{}
			mediaHandler := &media.Handler{}
			apikeyHandler := apikey.NewHandler(service)
			tenantHandler := tenantpkg.NewHandler(tenantService)

			health.RegisterHandlers(e.Group(""))
			v1 := e.Group("/v1")
			mediaHandler.RegisterHandlers(v1.Group("/media"))
			apikeyHandler.RegisterHandlers(v1.Group("/apikey"))
			tenantHandler.RegisterHandlers(v1.Group("/tenants"))

			sc := echo.StartConfig{
				Address:         cfg.Server.ServerAddress,
				GracefulTimeout: 20 * time.Second,
			}

			if err := sc.Start(executionContext, e); err != nil {
				log.Printf("server error: %v", err)
			}
		},
		func() {
			log.Println("initiating graceful shutdown...")
		},
	)
}
