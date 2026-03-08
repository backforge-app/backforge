// Package repository provides the repository layer for accessing database entities.
// It includes PostgreSQL transaction handling, repository-level and repository
// implementations for entities like users, sessions, questions, etc.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// Session is the repository for session operations.
type Session struct {
	db transactor.DBTx
}

// NewSession creates a new Session repository.
func NewSession(db transactor.DBTx) *Session {
	return &Session{db: db}
}

// Create inserts a new session into the database.
func (r *Session) Create(ctx context.Context, s *domain.Session) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO sessions (
		    user_id, token, expires_at
		) VALUES ($1, $2, $3)
	`

	_, err := db.Exec(ctx, q, s.UserID, s.Token, s.ExpiresAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation &&
				pgErr.ConstraintName == "sessions_token_key" {
				return ErrSessionAlreadyExists
			}
		}
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByToken retrieves a session by its token string value.
func (r *Session) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id,
			user_id,
			token,
			expires_at,
			revoked,
			created_at,
			updated_at
		FROM sessions
		WHERE token = $1
	`

	var s domain.Session

	err := db.QueryRow(ctx, q, token).Scan(
		&s.ID,
		&s.UserID,
		&s.Token,
		&s.ExpiresAt,
		&s.Revoked,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return &s, nil
}

// Revoke marks the refresh token as revoked.
func (r *Session) Revoke(ctx context.Context, token string) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE sessions
		SET 
			revoked    = TRUE,
			updated_at = now()
		WHERE token = $1
	`

	cmdTag, err := db.Exec(ctx, q, token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}
