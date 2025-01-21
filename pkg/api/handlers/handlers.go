package handlers

import (
	"github.com/bazilio91/sferra-cloud/pkg/auth"
	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/services/storage"
)

var (
	jwtManager *auth.JWTManager
	cfg        *config.Config
	s3Client   *storage.S3Client
)

func SetJWTManager(manager *auth.JWTManager) {
	jwtManager = manager
}

func SetConfig(c *config.Config) {
	cfg = c
}

func SetS3Client(client *storage.S3Client) {
	s3Client = client
}

func GetS3Client() *storage.S3Client {
	return s3Client
}
