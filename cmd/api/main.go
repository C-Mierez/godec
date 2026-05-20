package main

import (
	"context"
	"errors"
	"log"

	"github.com/c-mierez/godec/internal/api"
	apikeypkg "github.com/c-mierez/godec/internal/apikey"
	"github.com/c-mierez/godec/internal/config"
	"github.com/c-mierez/godec/internal/middleware"
	"github.com/c-mierez/godec/internal/middleware/echovalidator"
	postgres "github.com/c-mierez/godec/internal/postgres"
	db "github.com/c-mierez/godec/internal/postgres/db"
	tenantpkg "github.com/c-mierez/godec/internal/tenant"
	"github.com/c-mierez/godec/pkg/graceful"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
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
			e.Use(middleware.BuildGlobalMiddlewares()...)

				// Centralized error handler: format AuthError into AuthErrorResponse
			//
			// Auth errors are handled here, NOT through the generated response types
			// (GetMediaUploadURL401JSONResponse, GetMediaUploadURL403JSONResponse in gen.go).
			// Those types are intentionally bypassed because oapi-codegen only generates
			// typed response constructors per-operation, but the echovalidator middleware
			// intercepts auth failures before they reach the operation handler. All auth
			// errors flow through this centralized handler instead.
			//
			// If this pattern changes in the future (e.g. moving auth handling into each
			// operation handler), consider using the generated response types to keep
			// the spec and implementation in sync.
			e.HTTPErrorHandler = func(c *echo.Context, err error) {
				var ae *middleware.AuthError
				if errors.As(err, &ae) {
					c.JSON(ae.Status, map[string]string{"error": ae.Message, "code": ae.Code})
					return
				}
				echo.DefaultHTTPErrorHandler(false)(c, err)
			}

			// Wire apikey service into strict middleware.
			akValidator := middleware.NewAPIKeyValidator(apiKeyService)

			swagger, err := api.GetSwagger()
			if err != nil {
				log.Printf("failed to load openapi spec: %v", err)
				return
			}

			// Disable server host checks from the spec to avoid false negatives behind proxies.
			swagger.Servers = nil

			validatorOptions := &echovalidator.Options{
				Options: openapi3filter.Options{
					AuthenticationFunc: middleware.APIKeyAuthenticator(akValidator),
				},
			}
			e.Use(echovalidator.OapiRequestValidatorWithOptions(swagger, validatorOptions))

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
