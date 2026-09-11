package db

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

var Pool *pgxpool.Pool

// InitDB initializes the database connection pool.
// It expects the DATABASE_URL environment variable to be set.
func InitDB(ctx context.Context) error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	var err error
	Pool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	// Read and execute schema.sql automatically to ensure tables exist
	_, err = Pool.Exec(ctx, schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to execute schema.sql: %w", err)
	}
	log.Println("Database schema verified/created successfully")

	return nil
}

// CloseDB closes the database connection pool.
func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}
