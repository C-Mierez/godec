package server

import (
	"github.com/c-mierez/godec/internal/handlers"
	"github.com/labstack/echo/v5"
)

func (s *Server) registerRoutes() {
	// General routes
	registerHandlers(s.echo.Group("/"), []handlers.Handler{
		&handlers.HealthHandlers{},
	})

	// API v1 routes
	registerHandlers(s.echo.Group("/v1"), []handlers.Handler{
		&handlers.MediaHandlers{},
	})

}

func registerHandlers(g *echo.Group, handlers []handlers.Handler) {
	for _, h := range handlers {
		h.RegisterHandlers(g)
	}
}
