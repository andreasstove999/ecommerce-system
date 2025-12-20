package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// GetDSN returns the database DSN from environment.
func GetDSN() string {
	dsn := os.Getenv("ORDER_DB_DSN")
	if dsn == "" {
		log.Fatal("ORDER_DB_DSN not set")
	}
	return dsn
}

// openDB opens a database connection without pinging.
func openDB(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}

// MustOpen returns an open and verified database connection.
func MustOpen() *sql.DB {
	dsn := GetDSN()

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	return db
}
