// Package oauthconnection provides the repository layer for managing third-party
// identity provider connections (e.g., GitHub, Google) linked to local user accounts.
package oauthconnection

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// Repository handles database operations related to OAuth connections.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new OAuth connection repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// Create inserts a new OAuth connection linking a local user to a third-party provider.
// It returns ErrDuplicateConnection if the external account is already linked to someone else.
func (r *Repository) Create(ctx context.Context, conn *domain.OAuthConnection) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO oauth_connections (
			user_id,
			provider,
			provider_user_id
		) VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := db.QueryRow(ctx, q,
		conn.UserID,
		conn.Provider,
		conn.ProviderUserID,
	).Scan(&conn.ID, &conn.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				// According to migration, the constraint is on (provider, provider_user_id).
				return ErrDuplicateConnection
			}
		}
		return fmt.Errorf("create oauth connection: %w", err)
	}

	return nil
}

// GetByProviderUserID retrieves an OAuth connection using the provider's name and the user's ID on that provider.
// This is primarily used during the OAuth login flow to find which local user owns the external account.
func (r *Repository) GetByProviderUserID(
	ctx context.Context,
	provider domain.OAuthProvider,
	providerUserID string,
) (*domain.OAuthConnection, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id, user_id, provider, provider_user_id, created_at
		FROM oauth_connections
		WHERE provider = $1 AND provider_user_id = $2
	`

	var conn domain.OAuthConnection
	err := db.QueryRow(ctx, q, provider, providerUserID).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.Provider,
		&conn.ProviderUserID,
		&conn.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}
		return nil, fmt.Errorf("get oauth connection by provider user id: %w", err)
	}

	return &conn, nil
}

// GetByUserID retrieves all OAuth connections linked to a specific local user.
// Useful for displaying connected accounts in the user's profile settings.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.OAuthConnection, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id, user_id, provider, provider_user_id, created_at
		FROM oauth_connections
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("query oauth connections by user id: %w", err)
	}
	defer rows.Close()

	var connections []*domain.OAuthConnection
	for rows.Next() {
		var conn domain.OAuthConnection
		if err := rows.Scan(
			&conn.ID,
			&conn.UserID,
			&conn.Provider,
			&conn.ProviderUserID,
			&conn.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan oauth connection: %w", err)
		}
		connections = append(connections, &conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate oauth connections: %w", err)
	}

	return connections, nil
}

// Delete removes a specific OAuth connection, effectively unlinking the third-party account.
func (r *Repository) Delete(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		DELETE FROM oauth_connections
		WHERE user_id = $1 AND provider = $2
	`

	cmdTag, err := db.Exec(ctx, q, userID, provider)
	if err != nil {
		return fmt.Errorf("delete oauth connection: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrConnectionNotFound
	}

	return nil
}
