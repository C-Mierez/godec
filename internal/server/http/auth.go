package http

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type AuthHandlers struct {
	// AuthDb returns the active DB pool
	AuthDb *pgxpool.Pool
}

/**
 * Registers authentication-related routes
 * Auth for media routes will be handled with API Keys.
 * These are endpoints for tenants to manage their API Keys.
 * TODO: For now, this will be public and unprotected. In the future, this will be protected with a different auth mechanism specific to tenants
 */
func (h *AuthHandlers) RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc) {
	g.POST("/create_key", h.CreateAPIKey, middleware...)
}

// Creates a new API key for a tenant
func (h *AuthHandlers) CreateAPIKey(c *echo.Context) error {

	// Placeholder
	return c.JSON(http.StatusOK, map[string]string{
		"apiKey": "placeholder",
	})
}
