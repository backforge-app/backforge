// Package postgres provides PostgreSQL infrastructure setup.
//
// It contains initialization and configuration of the pgx connection pool
// used by the application to interact with PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/backforge-app/backforge/internal/config"
)

// NewPool creates a new connection pool with the given DSN and config.
func NewPool(ctx context.Context, dsn string, cfg config.PoolConfig) (*pgxpool.Pool, error) {
	// Parse pool configuration for PostgreSQL connection.
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime

	// Initialize connection pool for PostgreSQL.
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	return pool, nil
}
