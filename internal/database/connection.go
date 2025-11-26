package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DigitLock/expense-tracker/internal/config"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// DB wraps pgxpool.Pool and sqlc.Queries for database operations
type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// New creates a new database connection pool
func New(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Connection pool settings
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create sqlc queries instance
	queries := sqlc.New(pool)

	return &DB{
		Pool:    pool,
		Queries: queries,
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Health checks if database connection is alive
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
