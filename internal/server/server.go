package server

import (
	"context"
	"log"
	"time"

	"github.com/c-mierez/godec/internal/config"
	"github.com/c-mierez/godec/pkg/lib/graceful"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type Server struct {
	echo   *echo.Echo
	config *config.Config
	authDb *pgxpool.Pool
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		echo:   echo.New(),
		config: cfg,
		authDb: nil,
	}
}

func (s *Server) Start() {
	address := s.config.Server.ServerAddress
	sc := echo.StartConfig{
		Address:         address,
		GracefulTimeout: 20 * time.Second,
	}

	graceful.RunWithGracefulShutdown(
		func(executionContext context.Context) {
			// Establish DB connection pool using the execution context so it follows cancellation
			pool, err := pgxpool.New(executionContext, s.config.Database.URL)
			if err != nil {
				log.Printf("Failed to create DB pool: %v", err)
				return
			}
			s.authDb = pool

			s.registerMiddleware()
			s.registerRoutes()

			// Start server
			if err := sc.Start(executionContext, s.echo); err != nil {
				log.Printf("Server error: %v", err)
			}
		},
		func() {
			log.Println("Initiating graceful shutdown...")

			// Close DB pool if present
			if s.authDb != nil {
				s.authDb.Close()
			}

			log.Println("Graceful shutdown complete.")
		},
	)
}
