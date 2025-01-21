package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthInterceptor(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)
	protectedMethods := []string{
		"/test.service/Protected",
	}
	interceptor := NewAuthInterceptor(jwtManager, protectedMethods)

	t.Run("Unary interceptor with valid token", func(t *testing.T) {
		// Generate valid token
		token, err := jwtManager.GenerateToken(123, 456)
		assert.NoError(t, err)

		// Create context with token
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs("authorization", "Bearer "+token),
		)

		// Mock handler that checks if claims are in context
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			claims := ctx.Value(auth.ContextKey("claims")).(*auth.Claims)
			assert.Equal(t, uint64(123), claims.UserID)
			assert.Equal(t, uint64(456), claims.ClientID)
			return "ok", nil
		}

		// Call interceptor on protected method
		info := &grpc.UnaryServerInfo{FullMethod: "/test.service/Protected"}
		resp, err := interceptor.Unary()(ctx, "request", info, handler)
		assert.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("Unary interceptor with invalid token", func(t *testing.T) {
		// Create context with invalid token
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs("authorization", "Bearer invalid-token"),
		)

		// Call interceptor
		info := &grpc.UnaryServerInfo{FullMethod: "/test.service/Protected"}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}

		resp, err := interceptor.Unary()(ctx, "request", info, handler)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Unary interceptor with no auth required", func(t *testing.T) {
		// Call interceptor on non-protected method
		info := &grpc.UnaryServerInfo{FullMethod: "/test.service/Public"}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}

		resp, err := interceptor.Unary()(context.Background(), "request", info, handler)
		assert.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	t.Run("Stream interceptor with valid token", func(t *testing.T) {
		// Generate valid token
		token, err := jwtManager.GenerateToken(123, 456)
		assert.NoError(t, err)

		// Create context with token
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs("authorization", "Bearer "+token),
		)

		// Mock stream
		stream := &mockServerStream{ctx: ctx}

		// Mock handler that checks if claims are in context
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			claims := stream.Context().Value(auth.ContextKey("claims")).(*auth.Claims)
			assert.Equal(t, uint64(123), claims.UserID)
			assert.Equal(t, uint64(456), claims.ClientID)
			return nil
		}

		// Call interceptor on protected method
		info := &grpc.StreamServerInfo{FullMethod: "/test.service/Protected"}
		err = interceptor.Stream()(nil, stream, info, handler)
		assert.NoError(t, err)
	})
}

// mockServerStream is a mock implementation of grpc.ServerStream
type mockServerStream struct {
	ctx context.Context
	grpc.ServerStream
}

func (s *mockServerStream) Context() context.Context {
	return s.ctx
}
