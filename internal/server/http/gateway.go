package http

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type GatewayHandlers struct {
	service GatewayService
}

type GatewayService interface {
	GenerateApiKey() (string, error)
	ValidateApiKey(apiKey string) (bool, error)
}

/**
 * Registers Gateway-related handlers.
 * These API Keys are used by tenants to authenticate when they send media for processing.
 * TODO: For now, this will be public and unprotected. In the future, this will be protected with a different auth mechanism specific to tenants
 */
func (h *GatewayHandlers) RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc) {
	g.POST("/create_key", h.CreateAPIKey, middleware...)
}

// Creates a new API key for a tenant
func (h *GatewayHandlers) CreateAPIKey(c *echo.Context) error {

	// TODO: Validation ?

	// Generate the API key
	apiKey, err := h.service.GenerateApiKey()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate API key",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"api_key": apiKey,
	})
}
