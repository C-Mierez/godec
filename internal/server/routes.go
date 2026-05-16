package server

import (
	handlers "github.com/c-mierez/godec/internal/server/http"
	"github.com/labstack/echo/v5"
)

func (s *Server) registerRoutes() {
	// General routes
	registerHandlers(s.echo.Group(""), []handlers.HttpHandler{
		&handlers.HealthHandlers{},
	})

	// API v1 routes
	registerHandlers(s.echo.Group("/v1"), []handlers.HttpHandler{
		&handlers.MediaHandlers{},
		&handlers.AuthHandlers{
			AuthDb: s.db,
		},
	})

}

func registerHandlers(g *echo.Group, handlers []handlers.HttpHandler, middleware ...echo.MiddlewareFunc) {
	for _, h := range handlers {
		h.RegisterHandlers(g, middleware...)
	}
}
