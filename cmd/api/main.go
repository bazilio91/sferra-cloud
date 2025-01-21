// cmd/api/main.go
package main

import (
	"log"
	"time"

	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/api/middleware"
	"github.com/bazilio91/sferra-cloud/pkg/api/router"
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/db"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize the database
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Hour*24) // 24 hour token duration
	handlers.SetJWTManager(jwtManager)
	middleware.SetJWTManager(jwtManager)

	// Start the server
	r := router.SetupRouter(jwtManager, cfg)
	if err := r.Run(":" + cfg.APIServerPort); err != nil {
		log.Fatalf("Failed to run API server: %v", err)
	}
}
