package media

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

type Handler struct{}

func (h *Handler) RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc) {
	g.POST("/upload-url", h.GetUploadURL, middleware...)
}

func (h *Handler) GetUploadURL(c *echo.Context) error {
	// TODO Implement
	return c.JSON(http.StatusOK, map[string]string{
		"uploadURL": "https://example.invalid/upload/" + time.Now().UTC().Format(time.RFC3339Nano),
	})
}
