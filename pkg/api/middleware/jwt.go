package middleware

import (
	"net/http"
	"strings"

	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/gin-gonic/gin"
)

var jwtManager *auth.JWTManager

func SetJWTManager(manager *auth.JWTManager) {
	jwtManager = manager
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		userID, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store userID in context
		c.Set("userID", userID)
		c.Next()
	}
}
