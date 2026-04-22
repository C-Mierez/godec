package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/c-mierez/godec-mvp/internal/platform/graceful"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {

	e := echo.New()
	e.Use(middleware.RequestLogger())

	e.GET("/hello", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	sc := echo.StartConfig{
		Address:         ":8080",
		GracefulTimeout: 20 * time.Second,
	}

	graceful.RunWithGracefulShutdown(
		func(executionContext context.Context) {
			if err := sc.Start(executionContext, e); err != nil {
				log.Printf("Server error: %v", err)
			}
		},
		func() {
			log.Println("Initiating graceful shutdown...")

			// Close DB, Notify other services, etc.

			log.Println("Graceful shutdown complete.")

		},
	)
}
