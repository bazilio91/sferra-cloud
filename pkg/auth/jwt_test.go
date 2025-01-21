package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTManager(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)

	t.Run("Generate and verify token", func(t *testing.T) {
		// Generate token
		token, err := manager.GenerateToken(123, 456)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token
		claims, err := manager.VerifyJWT(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, uint64(123), claims.UserID)
		assert.Equal(t, uint64(456), claims.ClientID)
	})

	t.Run("Invalid token", func(t *testing.T) {
		// Try to verify invalid token
		claims, err := manager.VerifyJWT("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Expired token", func(t *testing.T) {
		// Create manager with very short duration
		shortManager := NewJWTManager("test-secret", time.Millisecond)
		token, err := shortManager.GenerateToken(123, 456)
		assert.NoError(t, err)

		// Wait for token to expire
		time.Sleep(time.Millisecond * 2)

		// Try to verify expired token
		claims, err := shortManager.VerifyJWT(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
