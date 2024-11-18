// pkg/config/config.go
package config

import (
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	APIServerPort  string
	GRPCServerPort string
}

func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Ignore error if .env file does not exist; environment variables can still be set
	}

	cfg := &Config{
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		APIServerPort:  os.Getenv("API_SERVER_PORT"),
		GRPCServerPort: os.Getenv("GRPC_SERVER_PORT"),
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) validate() error {
	if cfg.DBHost == "" {
		return errors.New("DB_HOST is not set")
	}
	if cfg.DBPort == "" {
		return errors.New("DB_PORT is not set")
	}
	if _, err := strconv.Atoi(cfg.DBPort); err != nil {
		return errors.New("DB_PORT must be a number")
	}
	if cfg.DBUser == "" {
		return errors.New("DB_USER is not set")
	}
	if cfg.DBPassword == "" {
		return errors.New("DB_PASSWORD is not set")
	}
	if cfg.DBName == "" {
		return errors.New("DB_NAME is not set")
	}
	if cfg.JWTSecret == "" {
		return errors.New("JWT_SECRET is not set")
	}
	if cfg.APIServerPort == "" {
		cfg.APIServerPort = "8080"
	}
	if _, err := strconv.Atoi(cfg.APIServerPort); err != nil {
		return errors.New("API_SERVER_PORT must be a number")
	}
	if cfg.GRPCServerPort == "" {
		cfg.GRPCServerPort = "50051"
	}
	if _, err := strconv.Atoi(cfg.GRPCServerPort); err != nil {
		return errors.New("GRPC_SERVER_PORT must be a number")
	}
	return nil
}
