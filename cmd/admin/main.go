package main

import (
	"log"

	"github.com/bazilio91/sferra-cloud/pkg/admin"
)

//go:generate protoc --go_out=./../.. --go-grpc_out=./../.. -I=../../pkg/pb ./../../pkg/pb/models.proto
//go:generate protoc --go_out=./../.. --go-grpc_out=./../.. -I=../../pkg/pb ./../../pkg/pb/health.proto

func main() {
	if err := admin.RunAdminServer(); err != nil {
		log.Fatalf("Failed to run admin server: %v", err)
	}
}
