package server

import (
	"context"
	"github.com/bazilio91/sferra-cloud/pkg/pb"
)

type HealthServer struct {
	pb.UnimplementedHealthServer
}

func (s *HealthServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: "SERVING"}, nil
}
