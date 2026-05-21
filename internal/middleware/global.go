package middleware

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func BuildGlobalMiddlewares(corsAllowedOrigins []string) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		middleware.RequestLogger(), middleware.Recover(), middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: corsAllowedOrigins,
			AllowMethods: []string{
				"GET",
				"POST",
				"PATCH",
				"DELETE",
				"OPTIONS",
			},
			AllowHeaders: []string{
				echo.HeaderAccept,
				echo.HeaderAuthorization,
				"X-API-Key",
				echo.HeaderContentType,
				echo.HeaderOrigin,
				echo.HeaderXRequestedWith,
			},
		}),
	}
}
