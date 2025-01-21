package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/grpc/middleware"
	"google.golang.org/grpc"
)

// Protected gRPC methods that require authentication
var protectedMethods = []string{
	"/proto.TaskService/ReserveTask",
	"/proto.TaskService/ReportTaskStatus",
	"/proto.TaskService/FinishTask",
	"/proto.TaskService/FailTask",
}

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
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Hour*24) // 24 hour token duration

	//// Initialize S3 client
	//s3Client, err := storage.NewS3Client(storage.S3Config{
	//	Endpoint:        cfg.S3Endpoint,
	//	Region:          cfg.S3Region,
	//	Bucket:          cfg.S3Bucket,
	//	AccessKeyID:     cfg.S3AccessKeyID,
	//	SecretAccessKey: cfg.S3SecretAccessKey,
	//})
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize S3 client: %v", err)
	// }

	// Initialize AuthInterceptor with protected methods
	authInterceptor := middleware.NewAuthInterceptor(jwtManager, protectedMethods)

	// GetTaskImage the server port from configuration
	port := cfg.GRPCServerPort

	// Set up the listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create gRPC server with unary and stream interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	log.Printf("Starting gRPC server on port %s...", port)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
