package health

import (
	"net/http"

	"github.com/c-mierez/godec/pkg/utils"
	"github.com/labstack/echo/v5"
)

const (
	LiveEndpoint  = "/live"
	ReadyEndpoint = "/ready"
	StatusOK      = "ok"
)

type Handler struct{}

type Response struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

func (h *Handler) RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc) {
	g.GET(LiveEndpoint, h.Live, middleware...)
	g.GET(ReadyEndpoint, h.Ready, middleware...)
}

func (h *Handler) Live(c *echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Status:    StatusOK,
		Timestamp: utils.TimeNow(),
	})
}

func (h *Handler) Ready(c *echo.Context) error {
	// TODO: Implement
	return c.JSON(http.StatusOK, Response{
		Status:    StatusOK,
		Timestamp: utils.TimeNow(),
	})
}
