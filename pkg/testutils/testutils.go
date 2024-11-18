package testutils

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"os"
	"testing"

	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/bazilio91/sferra-cloud/pkg/models"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDBContainer holds the reference to the PostgreSQL container
type TestDBContainer struct {
	Container testcontainers.Container
	Config    *config.Config
}

// StartTestDBContainer starts a PostgreSQL container for testing
func StartTestDBContainer(ctx context.Context) (*TestDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	host, err := postgresC.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %v", err)
	}

	cfg := &config.Config{
		DBHost:         host,
		DBPort:         port.Port(),
		DBUser:         "testuser",
		DBPassword:     "testpassword",
		DBName:         "testdb",
		JWTSecret:      "testsecret",
		APIServerPort:  "8080",
		GRPCServerPort: "50051",
	}

	// Initialize the database
	dsn := db.GetDSN(cfg)
	dbInstance, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		postgresC.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	db.DB = dbInstance

	// Auto migrate models
	if err := db.DB.AutoMigrate(&models.User{}); err != nil {
		postgresC.Terminate(ctx)
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return &TestDBContainer{
		Container: postgresC,
		Config:    cfg,
	}, nil
}

// StopTestDBContainer terminates the PostgreSQL container
func StopTestDBContainer(ctx context.Context, container *TestDBContainer) error {
	return container.Container.Terminate(ctx)
}

// CreateTestUser creates a user in the test database
func CreateTestUser(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	user := models.User{
		Email:    email,
		Password: string(hashedPassword),
	}
	if err := db.DB.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	return nil
}

// ClearDatabase clears the data from the database tables
func ClearDatabase() {
	db.DB.Exec("DELETE FROM users")
}

// LoadEnv loads environment variables from the .env file for testing
func LoadEnv() error {
	// Ensure that the .env file is loaded
	if err := os.Setenv("ENV", "test"); err != nil {
		return fmt.Errorf("failed to set ENV variable: %v", err)
	}
	return nil
}

// SetupTest sets up the testing environment
func SetupTest(m *testing.M) {
	// Load environment variables
	if err := LoadEnv(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	code := m.Run()
	os.Exit(code)
}

// StartRabbitMQContainer starts a RabbitMQ container for testing
func StartRabbitMQContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-management",
		ExposedPorts: []string{"5672/tcp", "15672/tcp"},
		WaitingFor:   wait.ForListeningPort("5672/tcp"),
	}
	rabbitMQC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start RabbitMQ container: %v", err)
	}
	return rabbitMQC, nil
}
