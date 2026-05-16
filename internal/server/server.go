package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/c-mierez/godec/internal/config"
	"github.com/c-mierez/godec/internal/lib/graceful"
	"github.com/labstack/echo/v5"
)

type Server struct {
	echo   *echo.Echo
	config *config.Config
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		echo:   echo.New(),
		config: cfg,
	}
}

func (s *Server) Start() {
	s.registerMiddleware()
	s.registerRoutes()

	address := fmt.Sprintf(":%s", s.config.Server.Port)
	sc := echo.StartConfig{
		Address:         address,
		GracefulTimeout: 20 * time.Second,
	}

	graceful.RunWithGracefulShutdown(
		func(executionContext context.Context) {
			if err := sc.Start(executionContext, s.echo); err != nil {
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
