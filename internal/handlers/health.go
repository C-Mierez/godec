package handlers

import (
	"net/http"

	"github.com/c-mierez/godec/internal/lib/utils"
	"github.com/labstack/echo/v5"
)

const (
	HealthLiveEndpoint  = "/live"
	HealthReadyEndpoint = "/ready"
	HealthStatusOK      = "ok"
)

type HealthHandlers struct {
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

func (h *HealthHandlers) RegisterHandlers(g *echo.Group) {
	g.GET(HealthLiveEndpoint, h.Live)
	g.GET(HealthReadyEndpoint, h.Ready)
}

func (h *HealthHandlers) Live(c *echo.Context) error {
	response := HealthResponse{
		Status:    HealthStatusOK,
		Timestamp: utils.TimeNow(),
	}

	return c.JSON(http.StatusOK, response)
}

func (h *HealthHandlers) Ready(c *echo.Context) error {
	// TODO: Add actual readiness checks (DB connection, external services, etc.)

	response := HealthResponse{
		Status:    HealthStatusOK,
		Timestamp: utils.TimeNow(),
	}

	return c.JSON(http.StatusOK, response)
}
