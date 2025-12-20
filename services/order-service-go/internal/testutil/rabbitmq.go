package testutil

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartRabbitMQ launches a RabbitMQ container and returns a ready AMQP connection
// plus a cleanup function. The cleanup function is registered with t.Cleanup.
func StartRabbitMQ(t *testing.T) (*amqp.Connection, func()) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.13-alpine",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForListeningPort("5672/tcp").WithStartupTimeout(90 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "5672")
	require.NoError(t, err)

	url := "amqp://" + host + ":" + mappedPort.Port() + "/"

	conn, err := amqp.DialConfig(url, amqp.Config{
		Dial: amqp.DefaultDial(10 * time.Second),
	})
	require.NoError(t, err)

	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()

		_ = conn.Close()
		_ = container.Terminate(cleanupCtx)
	}

	t.Cleanup(cleanup)

	return conn, cleanup
}
