package main

import (
	"log"

	"github.com/bazilio91/sferra-cloud/pkg/grpc/server"
)

//go:generate protoc --go_out=./../.. --go-grpc_out=./../.. -I=../../pkg/pb ./../../pkg/pb/models.proto ./../../pkg/pb/health.proto ./../../pkg/pb/scheduler.proto

func main() {
	if err := server.RunGRPCServer(); err != nil {
		log.Fatalf("Failed to run gRPC server: %v", err)
	}
}
