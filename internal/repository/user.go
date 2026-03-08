// Package repository provides the repository layer for accessing database entities.
// It includes PostgreSQL transaction handling, repository-level and repository
// implementations for entities like users, sessions, questions, etc.
package repository

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

// User is the repository for user-related operations.
type User struct {
	db transactor.DBTx
}

// NewUserRepository creates a new User repository.
func NewUserRepository(db transactor.DBTx) *User {
	return &User{db: db}
}

// Create inserts a new user into the database and returns its ID.
func (r *User) Create(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO users (
			telegram_id,
			username,
			first_name,
			last_name,
		    photo_url,
			role,
			is_pro,
			pro_granted_at,
			pro_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id uuid.UUID

	err := db.QueryRow(ctx, q,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.PhotoURL,
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
					return uuid.Nil, ErrUserTelegramIDTaken
				}
			case pgerrcode.InvalidTextRepresentation, pgerrcode.InvalidParameterValue:
				return uuid.Nil, ErrUserInvalidRole
			}
		}
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}

// GetByTelegramID retrieves a user by their Telegram user ID.
func (r *User) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id,
			telegram_id,
			username,
			first_name,
			last_name,
			photo_url,
			role,
			is_pro,
			pro_granted_at,
			pro_type,
			created_at,
			updated_at
		FROM users
		WHERE telegram_id = $1
	`

	var user domain.User

	err := db.QueryRow(ctx, q, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.PhotoURL,
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
func (r *User) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id,
			telegram_id,
			username,
			first_name,
			last_name,
			photo_url,
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

	err := db.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.PhotoURL,
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
func (r *User) Update(ctx context.Context, user *domain.User) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE users
		SET
			username     	= $2,
			first_name   	= $3,
			last_name    	= $4,
			photo_url    	= $5,
			role            = $6,
			is_pro          = $7,
			pro_granted_at  = $8,
			pro_type        = $9,
			updated_at      = now()
		WHERE id = $1
	`

	cmdTag, err := db.Exec(ctx, q,
		user.ID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.PhotoURL,
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
