// Package postgres provides PostgreSQL infrastructure components.
// It includes connection pool setup, transaction handling, repository-level errors
// and repository implementations for accessing database entities like users.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Transactor handles database transactions using a connection pool.
type Transactor struct {
	pool *pgxpool.Pool
	lg   *zap.SugaredLogger
}

// NewTransactor creates a new Transactor instance with the given pool.
func NewTransactor(pool *pgxpool.Pool, lg *zap.SugaredLogger) *Transactor {
	return &Transactor{pool: pool, lg: lg}
}

// WithinTx executes the given function within a transaction.
func (t *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			t.lg.Errorw("tx rollback error", "error", err)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
