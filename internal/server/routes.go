package server

import (
	handlers "github.com/c-mierez/godec/internal/server/http"
	"github.com/labstack/echo/v5"
)

type HttpHandler interface {
	RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc)
}

func (s *Server) registerRoutes() {
	// General routes
	registerHandlers(s.echo.Group(""), []HttpHandler{
		&handlers.HealthHandlers{},
	})

	// API v1 routes
	registerHandlers(s.echo.Group("/v1"), []HttpHandler{
		&handlers.MediaHandlers{},
		&handlers.GatewayHandlers{},
	})

}

func registerHandlers(g *echo.Group, handlers []HttpHandler, middleware ...echo.MiddlewareFunc) {
	for _, h := range handlers {
		h.RegisterHandlers(g, middleware...)
	}
}
