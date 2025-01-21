package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor(t *testing.T) {
	jwtManager := NewJWTManager("test-secret", time.Hour)
	interceptor := NewAuthInterceptor(jwtManager, []string{"/test.service/RequiresAuth"})

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
			claims := ctx.Value(ContextKey("claims")).(*Claims)
			assert.Equal(t, uint64(123), claims.UserID)
			assert.Equal(t, uint64(456), claims.ClientID)
			return "ok", nil
		}

		// Call interceptor
		info := &grpc.UnaryServerInfo{FullMethod: "/test.service/RequiresAuth"}
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
		info := &grpc.UnaryServerInfo{FullMethod: "/test.service/RequiresAuth"}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}

		resp, err := interceptor.Unary()(ctx, "request", info, handler)
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("Unary interceptor with no auth required", func(t *testing.T) {
		// Call interceptor on non-protected method
		info := &grpc.UnaryServerInfo{FullMethod: "/test.service/NoAuthRequired"}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}

		resp, err := interceptor.Unary()(context.Background(), "request", info, handler)
		assert.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})
}
