package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func MustOpen() *sql.DB {
	dsn := os.Getenv("CART_DB_DSN")
	if dsn == "" {
		log.Fatal("CART_DB_DSN not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	return db
}
