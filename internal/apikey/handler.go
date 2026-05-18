package apikey

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type APIKeyService interface {
	GenerateApiKey(ctx context.Context, tenantID uuid.UUID, name string, scopes []string) (string, *ApiKey, error)
}

type Handler struct {
	service APIKeyService
}

type CreateAPIKeyRequest struct {
	TenantID string   `json:"tenant_id"`
	Name     string   `json:"name"`
	Scopes   []string `json:"scopes"`
}

type CreateAPIKeyResponse struct {
	APIKey   string   `json:"api_key"`
	ID       string   `json:"id"`
	TenantID string   `json:"tenant_id"`
	Name     string   `json:"name"`
	Scopes   []string `json:"scopes"`
}

func NewHandler(service APIKeyService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc) {
	g.POST("/create_key", h.CreateAPIKey, middleware...)
}

func (h *Handler) CreateAPIKey(c *echo.Context) error {
	var request CreateAPIKeyRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	tenantID, err := uuid.Parse(request.TenantID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant_id"})
	}

	plainKey, apiKey, err := h.service.GenerateApiKey(c.Request().Context(), tenantID, request.Name, request.Scopes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate api key"})
	}

	if apiKey == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate api key"})
	}

	return c.JSON(http.StatusOK, CreateAPIKeyResponse{
		APIKey:   plainKey,
		ID:       apiKey.ID.String(),
		TenantID: apiKey.TenantID.String(),
		Name:     apiKey.Name,
		Scopes:   append([]string(nil), apiKey.Scopes...),
	})
}
