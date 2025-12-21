package db

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // Register postgres driver
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

// openDB opens a database connection without pinging.
func openDB(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}
