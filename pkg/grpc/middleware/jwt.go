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
	jwtManager      *auth.JWTManager
	allowedMethods  map[string]bool
}

// NewAuthInterceptor returns a new AuthInterceptor
func NewAuthInterceptor(jwtManager *auth.JWTManager, allowedMethods []string) *AuthInterceptor {
	allowedMap := make(map[string]bool)
	for _, method := range allowedMethods {
		allowedMap[method] = true
	}

	return &AuthInterceptor{
		jwtManager:     jwtManager,
		allowedMethods: allowedMap,
	}
}

// Unary returns a server interceptor function to authenticate unary RPC
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

		claims, err := interceptor.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		// Add claims to context
		ctx = context.WithValue(ctx, auth.ContextKey("claims"), claims)
		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function to authenticate stream RPC
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

		claims, err := interceptor.authenticate(stream.Context())
		if err != nil {
			return err
		}

		// Wrap the stream to inject authenticated context
		wrappedStream := &wrappedStream{
			ServerStream: stream,
			ctx:         context.WithValue(stream.Context(), auth.ContextKey("claims"), claims),
		}
		
		return handler(srv, wrappedStream)
	}
}

// authenticate checks the validity of the JWT token
func (interceptor *AuthInterceptor) authenticate(ctx context.Context) (*auth.Claims, error) {
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
