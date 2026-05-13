package server

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func (s *Server) registerMiddleware() {
	s.echo.Use(buildMiddlewareChain()...)
}

func buildMiddlewareChain() []echo.MiddlewareFunc {
	var middlewares []echo.MiddlewareFunc

	middlewares = append(middlewares,
		middleware.RequestLogger(),
		middleware.Recover(),
		// Other custom middlewares...
	)

	return middlewares
}
