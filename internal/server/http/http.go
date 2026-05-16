package http

import "github.com/labstack/echo/v5"

type HttpHandler interface {
	RegisterHandlers(g *echo.Group, middleware ...echo.MiddlewareFunc)
}
