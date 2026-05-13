package handlers

import "github.com/labstack/echo/v5"

type Handler interface {
	RegisterHandlers(g *echo.Group)
}
