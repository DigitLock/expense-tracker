// Package testutil provides shared helpers for database integration tests.
package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool returns a pgxpool connected to TEST_DATABASE_URL.
//
// If TEST_DATABASE_URL is empty, the calling test is skipped, so a plain
// `go test ./...` (without a database) stays green. The pool is closed
// automatically when the test finishes.
func Pool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database integration test")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// Truncate wipes the core tables for a clean test slate. CASCADE removes
// dependent rows (accounts, transactions, exchange_rates, audit_log) that
// reference families/users/categories via foreign keys.
func Truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		"TRUNCATE families, users, categories CASCADE")
	if err != nil {
		t.Fatalf("truncate test tables: %v", err)
	}
}
