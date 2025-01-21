package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor is a server interceptor for authentication and authorization
type AuthInterceptor struct {
	jwtManager      *JWTManager
	allowedMethods  map[string]bool
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(jwtManager *JWTManager, allowedMethods []string) *AuthInterceptor {
	allowedMap := make(map[string]bool)
	for _, method := range allowedMethods {
		allowedMap[method] = true
	}

	return &AuthInterceptor{
		jwtManager:      jwtManager,
		allowedMethods:  allowedMap,
	}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !interceptor.allowedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		claims, err := interceptor.authorize(ctx)
		if err != nil {
			return nil, err
		}

		// Add claims to context
		ctx = context.WithValue(ctx, ContextKey("claims"), claims)
		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function to authenticate and authorize stream RPC
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if !interceptor.allowedMethods[info.FullMethod] {
			return handler(srv, stream)
		}

		claims, err := interceptor.authorize(stream.Context())
		if err != nil {
			return err
		}

		// Wrap the stream to inject authenticated context
		wrappedStream := &wrappedStream{
			ServerStream: stream,
			ctx:         context.WithValue(stream.Context(), ContextKey("claims"), claims),
		}
		
		return handler(srv, wrappedStream)
	}
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	if !strings.HasPrefix(accessToken, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token format")
	}

	tokenStr := strings.TrimPrefix(accessToken, "Bearer ")
	claims, err := interceptor.jwtManager.VerifyJWT(tokenStr)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	return claims, nil
}

// wrappedStream wraps grpc.ServerStream to modify its context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
