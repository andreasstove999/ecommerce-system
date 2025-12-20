package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	dbschema "github.com/andreasstove999/ecommerce-system/order-service-go/internal/db"
)

const (
	dbUser     = "order_user"
	dbPassword = "order_pass"
	dbName     = "orders"
)

// StartPostgres launches a temporary Postgres container, applies the schema, and
// returns a database handle plus a cleanup function.
func StartPostgres(ctx context.Context, t *testing.T) (*sql.DB, func()) {
	t.Helper()

	containerName := "order-int-" + uuid.NewString()
	runArgs := []string{
		"run", "--rm", "-d",
		"-e", fmt.Sprintf("POSTGRES_USER=%s", dbUser),
		"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPassword),
		"-e", fmt.Sprintf("POSTGRES_DB=%s", dbName),
		"-P",
		"--name", containerName,
		"postgres:16-alpine",
	}

	if err := exec.CommandContext(ctx, "docker", runArgs...).Run(); err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	var db *sql.DB
	cleanup := func() {
		if db != nil {
			_ = db.Close()
		}
		_ = exec.Command("docker", "stop", containerName).Run()
	}

	hostPort := waitForPort(ctx, t, containerName)
	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", dbUser, dbPassword, hostPort, dbName)

	db = connectAndMigrate(ctx, t, dsn)

	return db, cleanup
}

func waitForPort(ctx context.Context, t *testing.T, containerName string) string {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for postgres port")
		}

		out, err := exec.CommandContext(ctx, "docker", "port", containerName, "5432/tcp").Output()
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(out)), ":")
			if len(parts) == 2 {
				return parts[1]
			}
		}

		select {
		case <-ctx.Done():
			t.Fatalf("context cancelled waiting for postgres port: %v", ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func connectAndMigrate(ctx context.Context, t *testing.T, dsn string) *sql.DB {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for {
		conn, err := sql.Open("postgres", dsn)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err = conn.PingContext(pingCtx)
			cancel()
			if err == nil {
				if _, err := conn.ExecContext(ctx, dbschema.Schema); err != nil {
					_ = conn.Close()
				} else {
					return conn
				}
			}
			_ = conn.Close()
		}

		if time.Now().After(deadline) {
			t.Fatalf("timeout connecting to postgres: %v", err)
		}

		select {
		case <-ctx.Done():
			t.Fatalf("context cancelled connecting to postgres: %v", ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
}
