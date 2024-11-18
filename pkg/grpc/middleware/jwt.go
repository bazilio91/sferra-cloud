package middleware

import (
	"context"
	"strings"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is a gRPC server interceptor for authentication
type AuthInterceptor struct {
	jwtManager *auth.JWTManager
}

// NewAuthInterceptor returns a new AuthInterceptor
func NewAuthInterceptor(jwtManager *auth.JWTManager) *AuthInterceptor {
	return &AuthInterceptor{jwtManager}
}

// Unary returns a server interceptor function to authenticate unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return interceptor.authenticate(ctx, req, info, handler)
	}
}

// authenticate checks the validity of the JWT token
func (interceptor *AuthInterceptor) authenticate(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Skip authentication for health check
	if info.FullMethod == "/pb.Health/Check" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	authHeader := authHeaders[0]
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization token format")
	}

	userID, err := interceptor.jwtManager.ValidateToken(parts[1])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Store userID in context if needed
	ctx = context.WithValue(ctx, "userID", userID)

	return handler(ctx, req)
}
