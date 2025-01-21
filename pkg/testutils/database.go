// pkg/testutils/database.go
package testutils

import (
	"context"
	"fmt"
	"gorm.io/gorm"

	"github.com/bazilio91/sferra-cloud/pkg/config"
	"github.com/bazilio91/sferra-cloud/pkg/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDBContainer struct {
	Container testcontainers.Container
	Config    *config.Config
}

func StartTestDBContainer(ctx context.Context) (*TestDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
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
		postgresC.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		postgresC.Terminate(ctx)
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

	// Initialize the database using db.InitDB(cfg)
	if err := db.InitDB(cfg); err != nil {
		postgresC.Terminate(ctx)
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return &TestDBContainer{
		Container: postgresC,
		Config:    cfg,
	}, nil
}

func StartTestDB(ctx context.Context) (*TestDBContainer, *gorm.DB, error) {
	c, err := StartTestDBContainer(ctx)
	if err != nil {
		return c, nil, err
	}

	dsn := db.GetDSN(c.Config)
	if err := db.InitDBWithDSN(dsn); err != nil {
		c.Container.Terminate(ctx)
		return c, nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return c, db.DB, nil
}

func StopTestDBContainer(ctx context.Context, container *TestDBContainer) error {
	return container.Container.Terminate(ctx)
}

func ClearDatabase(DB *gorm.DB) {
	err := DB.Exec("DELETE FROM client_users").Error
	if err != nil {
		panic(err)
	}
	err = DB.Exec("DELETE FROM data_recognition_tasks").Error
	if err != nil {
		panic(err)
	}
	DB.Exec("DELETE FROM clients")
	DB.Exec("DELETE FROM admins")
}
