package server

import (
	"fmt"
	"log"
	"net"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/grpc/middleware"
	"github.com/bazilio91/sferra-cloud/pkg/pb"
	"google.golang.org/grpc"
)

func RunGRPCServer() error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Initialize the database
	if err := db.InitDB(cfg); err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Initialize AuthInterceptor
	authInterceptor := middleware.NewAuthInterceptor(jwtManager)

	// Get the server port from configuration
	port := cfg.GRPCServerPort

	// Set up the listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create gRPC server with middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	// Register services
	pb.RegisterHealthServer(grpcServer, &HealthServer{})

	log.Printf("Starting gRPC server on port %s...", port)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
