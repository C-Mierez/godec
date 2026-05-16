package main

import (
	"log"

	"github.com/c-mierez/godec/internal/config"
	"github.com/c-mierez/godec/internal/server"
)

func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start the server with loaded configuration
	s := server.NewServer(cfg)
	s.Start()
}
