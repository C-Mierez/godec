package tenant

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type TenantService interface {
	CreateTenant(ctx context.Context, name, email string) (*Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	ListTenants(ctx context.Context, limit, offset int32) ([]*Tenant, error)
	SetTenantStatus(ctx context.Context, id uuid.UUID, status TenantStatus) (*Tenant, error)
}

type Handler struct {
	service TenantService
}

type CreateTenantRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateTenantStatusRequest struct {
	Status TenantStatus `json:"status"`
}

type TenantResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Status    TenantStatus `json:"status"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
}

type TenantListResponse struct {
	Items []*TenantResponse `json:"items"`
}

func NewHandler(service TenantService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc) {
	g.POST("", h.CreateTenant)
	g.GET("", h.ListTenants)
	g.GET("/:id", h.GetTenantByID)
	g.PATCH("/:id/status", h.SetTenantStatus)
}

func (h *Handler) CreateTenant(c *echo.Context) error {
	var request CreateTenantRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	tenant, err := h.service.CreateTenant(c.Request().Context(), request.Name, request.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create tenant"})
	}

	return c.JSON(http.StatusCreated, tenantToResponse(tenant))
}

func (h *Handler) GetTenantByID(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant id"})
	}

	tenant, err := h.service.GetTenantByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "tenant not found"})
	}

	return c.JSON(http.StatusOK, tenantToResponse(tenant))
}

func (h *Handler) ListTenants(c *echo.Context) error {
	limit := int32(20)
	offset := int32(0)

	if rawLimit := c.QueryParam("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid limit"})
		}
		limit = int32(parsed)
	}

	if rawOffset := c.QueryParam("offset"); rawOffset != "" {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil || parsed < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid offset"})
		}
		offset = int32(parsed)
	}

	tenants, err := h.service.ListTenants(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list tenants"})
	}

	items := make([]*TenantResponse, len(tenants))
	for i, tenant := range tenants {
		items[i] = tenantToResponse(tenant)
	}

	return c.JSON(http.StatusOK, TenantListResponse{Items: items})
}

func (h *Handler) SetTenantStatus(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid tenant id"})
	}

	var request UpdateTenantStatusRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	tenant, err := h.service.SetTenantStatus(c.Request().Context(), id, request.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update tenant status"})
	}

	return c.JSON(http.StatusOK, tenantToResponse(tenant))
}

func tenantToResponse(t *Tenant) *TenantResponse {
	if t == nil {
		return nil
	}

	return &TenantResponse{
		ID:        t.ID.String(),
		Name:      t.Name,
		Email:     t.Email,
		Status:    t.Status,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}
