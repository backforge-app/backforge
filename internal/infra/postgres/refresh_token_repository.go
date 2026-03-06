// Package postgres provides PostgreSQL infrastructure components.
// It includes connection pool setup, transaction handling, repository-level errors
// and repository implementations for accessing database entities like users,
// refresh tokens, questions etc.
package postgres

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

// RefreshTokenRepository is the repository for refresh token operations.
type RefreshTokenRepository struct {
	db transactor.DBTx
}

// NewRefreshTokenRepository creates a new RefreshToken repository.
func NewRefreshTokenRepository(db transactor.DBTx) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create inserts a new refresh token into the database.
func (r *RefreshTokenRepository) Create(ctx context.Context, rt *domain.RefreshToken) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO refresh_tokens (
		    user_id, token, expires_at
		) VALUES ($1, $2, $3)
	`

	_, err := db.Exec(ctx, q, rt.UserID, rt.Token, rt.ExpiresAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation &&
				pgErr.ConstraintName == "refresh_tokens_token_key" {
				return ErrRefreshTokenAlreadyExists
			}
		}
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByToken retrieves a refresh token by its token string value.
func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
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
		FROM refresh_tokens
		WHERE token = $1
	`

	var rt domain.RefreshToken

	err := db.QueryRow(ctx, q, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.Revoked,
		&rt.CreatedAt,
		&rt.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token by token: %w", err)
	}

	return &rt, nil
}

// Revoke marks the refresh token as revoked.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE refresh_tokens
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
		return ErrRefreshTokenNotFound
	}

	return nil
}
