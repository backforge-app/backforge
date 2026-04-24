// Package verificationtoken provides the repository layer for managing secure,
// time-limited tokens used for out-of-band verification flows such as email
// confirmation and password resets.
package verificationtoken

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// Repository handles database operations related to verification tokens.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new verification token repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// Create inserts a new verification token hash into the database.
func (r *Repository) Create(ctx context.Context, token *domain.VerificationToken) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO verification_tokens (
			token_hash,
			user_id,
			purpose,
			expires_at
		) VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := db.QueryRow(ctx, q,
		token.TokenHash,
		token.UserID,
		token.Purpose,
		token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		return fmt.Errorf("create verification token: %w", err)
	}

	return nil
}

// GetByHash retrieves a token by its SHA-256 hash and intended purpose.
// For security reasons, it strictly requires the 'purpose' to match and
// automatically excludes tokens that have passed their expiration time.
func (r *Repository) GetByHash(
	ctx context.Context,
	tokenHash string,
	purpose domain.TokenPurpose,
) (*domain.VerificationToken, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			token_hash, user_id, purpose, expires_at, created_at
		FROM verification_tokens
		WHERE token_hash = $1 
		  AND purpose = $2 
		  AND expires_at > now()
	`

	var token domain.VerificationToken
	err := db.QueryRow(ctx, q, tokenHash, purpose).Scan(
		&token.TokenHash,
		&token.UserID,
		&token.Purpose,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("get verification token by hash: %w", err)
	}

	return &token, nil
}

// Delete removes a verification token by its hash.
// This must be called immediately after a token is successfully used
// to prevent replay attacks (one-time use enforcement).
func (r *Repository) Delete(ctx context.Context, tokenHash string) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		DELETE FROM verification_tokens
		WHERE token_hash = $1
	`

	cmdTag, err := db.Exec(ctx, q, tokenHash)
	if err != nil {
		return fmt.Errorf("delete verification token: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// DeleteAllForUser removes all existing tokens of a specific purpose for a given user.
// This is best practice when a user requests a new token (e.g., requesting a new password
// reset link should invalidate any previously sent, unused reset links).
func (r *Repository) DeleteAllForUser(
	ctx context.Context,
	userID uuid.UUID,
	purpose domain.TokenPurpose,
) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		DELETE FROM verification_tokens
		WHERE user_id = $1 AND purpose = $2
	`

	_, err := db.Exec(ctx, q, userID, purpose)
	if err != nil {
		return fmt.Errorf("delete all verification tokens for user: %w", err)
	}

	return nil
}
