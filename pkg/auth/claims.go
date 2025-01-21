package auth

import "github.com/golang-jwt/jwt/v5"

// Claims represents the JWT claims
type Claims struct {
	UserID   uint64 `json:"user_id"`
	ClientID uint64 `json:"client_id"`
	jwt.RegisteredClaims
}
