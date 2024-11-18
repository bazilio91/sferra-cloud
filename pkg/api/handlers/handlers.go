package handlers

import (
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
)

var (
	jwtManager *auth.JWTManager
	cfg        *config.Config
)

func SetJWTManager(manager *auth.JWTManager) {
	jwtManager = manager
}

func SetConfig(c *config.Config) {
	cfg = c
}
