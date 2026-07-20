// Package main is the composition root for the godec API server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/c-mierez/godec/internal/api"
	apikeypkg "github.com/c-mierez/godec/internal/apikey"
	"github.com/c-mierez/godec/internal/config"
	"github.com/c-mierez/godec/internal/logging"
	"github.com/c-mierez/godec/internal/middleware"
	"github.com/c-mierez/godec/internal/middleware/echovalidator"
	postgres "github.com/c-mierez/godec/internal/postgres"
	db "github.com/c-mierez/godec/internal/postgres/db"
	tenantpkg "github.com/c-mierez/godec/internal/tenant"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}

	logging.Setup(cfg.Server.Env)

	serverCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		slog.Error("failed to create db pool", "error", err)
		return
	}
	defer pool.Close()

	queries := db.New(pool)
	apiKeyStore := postgres.NewAPIKeyStore(queries)
	tenantStore := postgres.NewTenantStore(queries)
	apiKeyService := apikeypkg.NewService(apiKeyStore)
	tenantService := tenantpkg.NewService(tenantStore)

	e := echo.New()
	corsOrigins := strings.Split(cfg.Server.CORSAllowedOrigins, ",")
	e.Use(middleware.BuildGlobalMiddlewares(corsOrigins)...)

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
		// Prevent double-write when RequestLogger with HandleError=true calls
		// this handler preemptively and then Echo's serveHTTP calls it again.
		// See Echo v5 middleware/request_logger.go:381-388
		if r, _ := echo.UnwrapResponse(c.Response()); r != nil && r.Committed {
			return
		}

		// Auth errors are intercepted by the echovalidator middleware before
		// the operation handler runs, so the generated per-operation response
		// types (GetMediaUploadURL401JSONResponse, etc.) cannot be used here.
		// We serialize the generated AuthErrorResponse struct directly to
		// guarantee the JSON shape matches the OpenAPI spec schema.
		var ae *middleware.AuthError
		if errors.As(err, &ae) {
			_ = c.JSON(ae.Status, api.AuthErrorResponse{
				Code:  api.AuthErrorResponseCode(ae.Code),
				Error: ae.Message,
			})
			return
		}
		echo.DefaultHTTPErrorHandler(false)(c, err)
	}

	// Wire apikey service into strict middleware.
	akValidator := middleware.NewAPIKeyValidator(apiKeyService)

	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("failed to load openapi spec", "error", err)
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
	healthCheckers := map[string]api.Pinger{
		api.CheckerDatabase: pool,
	}
	apiServer := api.NewServer(tenantService, apiKeyService, healthCheckers)
	strictHandler := api.NewStrictHandler(apiServer, nil)
	api.RegisterHandlers(e, strictHandler)

	// Print startup information
	slog.Info("server starting", "address", cfg.Server.ServerAddress)
	slog.Info("api documentation available", "url", "http://"+cfg.Server.ServerAddress+"/docs/api")

	sc := echo.StartConfig{
		Address:         cfg.Server.ServerAddress,
		GracefulTimeout: 10 * time.Second,
	}

	if err := sc.Start(serverCtx, e); err != nil {
		slog.Error("server error", "error", err)
	}
}
