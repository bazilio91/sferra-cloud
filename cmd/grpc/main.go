package main

import (
	"log"

	"github.com/bazilio91/sferra-cloud/pkg/grpc/server"
)

func main() {
	if err := server.RunGRPCServer(); err != nil {
		log.Fatalf("Failed to run gRPC server: %v", err)
	}
}
