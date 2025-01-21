// pkg/api/middleware/jwt.go
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

		claims, err := jwtManager.VerifyJWT(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Set("userID", claims.UserID)
		c.Set("clientID", claims.ClientID)
		c.Next()
	}
}
