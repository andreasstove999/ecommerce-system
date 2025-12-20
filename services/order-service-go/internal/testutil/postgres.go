package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartPostgres launches a postgres container, creates a unique database,
// applies the schema, and returns a ready-to-use sql.DB plus a cleanup
// function. The cleanup function is also registered with t.Cleanup.
func StartPostgres(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(90 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	baseDSN := fmt.Sprintf("postgres://postgres:password@%s:%s/postgres?sslmode=disable", host, mappedPort.Port())

	baseDB, err := sql.Open("postgres", baseDSN)
	require.NoError(t, err)
	require.NoError(t, waitForDB(ctx, baseDB))

	dbName := "order_service_test_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err = baseDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbName)))
	require.NoError(t, err)
	require.NoError(t, baseDB.Close())

	dsn := fmt.Sprintf("postgres://postgres:password@%s:%s/%s?sslmode=disable", host, mappedPort.Port(), dbName)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, waitForDB(ctx, db))

	schemaPath, err := filepath.Abs(filepath.Join("internal", "db", "schema.sql"))
	require.NoError(t, err)

	schema, err := os.ReadFile(schemaPath)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, string(schema))
	require.NoError(t, err)

	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()

		_ = db.Close()
		_ = container.Terminate(cleanupCtx)
	}

	t.Cleanup(cleanup)

	return db, cleanup
}

func waitForDB(ctx context.Context, db *sql.DB) error {
	deadline := time.Now().Add(30 * time.Second)

	for {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err := db.PingContext(pingCtx)
		cancel()
		if err == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return err
		}

		time.Sleep(500 * time.Millisecond)
	}
}
