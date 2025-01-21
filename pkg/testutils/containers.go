package testutils

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Functions for managing other containers (e.g., RabbitMQ) can be added here

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

// StopContainer terminates a generic container
func StopContainer(ctx context.Context, container testcontainers.Container) error {
	return container.Terminate(ctx)
}
