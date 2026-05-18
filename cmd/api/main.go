package main

import (
	"context"
	"log"

	"github.com/c-mierez/godec/internal/api"
	apikeypkg "github.com/c-mierez/godec/internal/apikey"
	"github.com/c-mierez/godec/internal/config"
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
			apiKeyService := apikeypkg.NewService(apiKeyStore)
			tenantService := tenantpkg.NewService(tenantStore)

			e := echo.New()
			e.Use(middleware.RequestLogger(), middleware.Recover(), middleware.CORSWithConfig(middleware.CORSConfig{
				AllowOrigins: []string{
					"http://127.0.0.1:8080",
					"http://localhost:8080",
					// TODO - Actual production url
				},
				AllowMethods: []string{
					"GET",
					"POST",
					"PATCH",
					"DELETE",
					"OPTIONS",
				},
				AllowHeaders: []string{
					echo.HeaderAccept,
					echo.HeaderAuthorization,
					echo.HeaderContentType,
					echo.HeaderOrigin,
					echo.HeaderXRequestedWith,
				},
			}))

			// Create API server and register handlers
			apiServer := api.NewServer(tenantService, apiKeyService)
			strictHandler := api.NewStrictHandler(apiServer, nil)
			api.RegisterHandlers(e, strictHandler)

			// Print startup information
			log.Printf("godec API server starting on %s", cfg.Server.ServerAddress)
			log.Printf("📖 API Documentation: http://%s/docs/api", cfg.Server.ServerAddress)

			if err := e.Start(cfg.Server.ServerAddress); err != nil {
				log.Printf("server error: %v", err)
			}
		},
		func() {
			log.Println("initiating graceful shutdown...")
		},
	)
}
