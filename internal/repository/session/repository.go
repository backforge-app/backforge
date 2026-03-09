// Package session provides the repository layer for accessing session entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, and manage user sessions.
package session

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

// Repository handles operations related to Session entities.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new Session repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// Create inserts a new session into the database.
func (r *Repository) Create(ctx context.Context, s *domain.Session) error {
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
func (r *Repository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
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
func (r *Repository) Revoke(ctx context.Context, token string) error {
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
