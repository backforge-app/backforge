// Package postgres provides PostgreSQL infrastructure components.
// It includes connection pool setup, transaction handling, repository-level errors
// and repository implementations for accessing database entities like users.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/backforge-app/backforge/internal/domain"
)

// UserRepository is the repository for user-related operations.
type UserRepository struct {
	db DBTX
}

// NewUserRepository creates a new User repository.
func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database and returns its ID.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	const q = `
		INSERT INTO users (
			tg_user_id,
			tg_username,
			tg_first_name,
			tg_last_name,
			role,
			is_pro,
			pro_granted_at,
			pro_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id uuid.UUID

	err := r.db.QueryRow(ctx, q,
		user.TgUserID,
		user.TgUsername,
		user.TgFirstName,
		user.TgLastName,
		user.Role,
		user.IsPro,
		user.ProGrantedAt,
		user.ProType,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if pgErr.ConstraintName == "users_tg_user_id_key" {
					return uuid.Nil, ErrUserTgUserIDTaken
				}
			case pgerrcode.InvalidTextRepresentation, pgerrcode.InvalidParameterValue:
				return uuid.Nil, ErrUserInvalidRole
			}
		}
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}

// GetByTgUserID retrieves a user by their Telegram user ID.
func (r *UserRepository) GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.User, error) {
	const q = `
		SELECT 
			id,
			tg_user_id,
			tg_username,
			tg_first_name,
			tg_last_name,
			role,
			is_pro,
			pro_granted_at,
			pro_type,
			created_at,
			updated_at
		FROM users
		WHERE tg_user_id = $1
	`

	var user domain.User

	err := r.db.QueryRow(ctx, q, tgUserID).Scan(
		&user.ID,
		&user.TgUserID,
		&user.TgUsername,
		&user.TgFirstName,
		&user.TgLastName,
		&user.Role,
		&user.IsPro,
		&user.ProGrantedAt,
		&user.ProType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by tg_user_id: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by their UUID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT 
			id,
			tg_user_id,
			tg_username,
			tg_first_name,
			tg_last_name,
			role,
			is_pro,
			pro_granted_at,
			pro_type,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User

	err := r.db.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.TgUserID,
		&user.TgUsername,
		&user.TgFirstName,
		&user.TgLastName,
		&user.Role,
		&user.IsPro,
		&user.ProGrantedAt,
		&user.ProType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// Update modifies an existing user's details.
// Only updates fields that are typically changeable (username, names, role, pro-status).
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	const q = `
		UPDATE users
		SET
			tg_username     = $2,
			tg_first_name   = $3,
			tg_last_name    = $4,
			role            = $5,
			is_pro          = $6,
			pro_granted_at  = $7,
			pro_type        = $8,
			updated_at      = now()
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, q,
		user.ID,
		user.TgUsername,
		user.TgFirstName,
		user.TgLastName,
		user.Role,
		user.IsPro,
		user.ProGrantedAt,
		user.ProType,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.InvalidTextRepresentation, pgerrcode.InvalidParameterValue:
				return ErrUserInvalidRole
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
