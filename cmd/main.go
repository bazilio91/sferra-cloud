package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bazilio91/sferra-cloud/pkg/admin"
	"github.com/bazilio91/sferra-cloud/pkg/api/handlers"
	"github.com/bazilio91/sferra-cloud/pkg/api/middleware"
	"github.com/bazilio91/sferra-cloud/pkg/api/router"
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/grpc/server"
	"github.com/getsentry/sentry-go"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Sentry
	if cfg.SentryDSN != "" {
		err = sentry.Init(sentry.ClientOptions{
			Dsn: cfg.SentryDSN,
			// Set traces sample rate to capture errors
			TracesSampleRate: 0.2,
			// Set environment from config
			Environment: cfg.SentryEnv,
		})
		if err != nil {
			log.Fatalf("Failed to initialize Sentry: %v", err)
		}
		defer sentry.Flush(2 * time.Second)
		log.Printf("Sentry initialized with environment: %s", cfg.SentryEnv)
	} else {
		if cfg.SentryEnv == "dev" {
			log.Printf("Sentry init skipped: running in %s environment", cfg.SentryEnv)
		} else {
			log.Fatalf("Sentry DSN is required in non-dev environment")
		}
	}

	// Initialize the database
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize JWT manager for API server
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Hour*24)
	handlers.SetJWTManager(jwtManager)
	middleware.SetJWTManager(jwtManager)

	// Create a WaitGroup to manage our goroutines
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start API server
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := router.SetupRouter(jwtManager, cfg)
		srv := &http.Server{
			Addr:    ":" + cfg.APIServerPort,
			Handler: r,
		}
		go func() {
			<-ctx.Done()
			// Add a small timeout for graceful shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Printf("API server shutdown error: %v", err)
			}
		}()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.RunGRPCServer(); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Start Admin server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := admin.RunAdminServer(); err != nil {
			log.Printf("Admin server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cancel context to initiate shutdown
	cancel()

	// Wait for all servers to shut down
	wg.Wait()
	log.Println("All servers shut down successfully")
}
