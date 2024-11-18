// pkg/auth/jwt.go
package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTManager struct {
	secretKey string
}

func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{secretKey: secretKey}
}

func (manager *JWTManager) GenerateJWT(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString([]byte(manager.secretKey))
}

func (manager *JWTManager) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(manager.secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["userID"].(float64)
		if !ok {
			return 0, errors.New("invalid token claims")
		}
		return uint(userIDFloat), nil
	} else {
		return 0, errors.New("invalid token")
	}
}
