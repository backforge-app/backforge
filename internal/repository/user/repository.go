// Package user provides the repository layer for accessing user entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, and manage users.
package user

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

// Repository handles operations related to User entities.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new User repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// Create inserts a new user into the database and returns its ID.
func (r *Repository) Create(ctx context.Context, u *domain.User) (uuid.UUID, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO users (
			telegram_id,
			username,
			first_name,
			last_name,
		    photo_url,
			role
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id uuid.UUID

	err := db.QueryRow(ctx, q,
		u.TelegramID,
		u.Username,
		u.FirstName,
		u.LastName,
		u.PhotoURL,
		u.Role,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if pgErr.ConstraintName == "users_telegram_id_key" {
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
func (r *Repository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
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
			created_at,
			updated_at
		FROM users
		WHERE telegram_id = $1
	`

	u, err := scanUser(db.QueryRow(ctx, q, telegramID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram id: %w", err)
	}

	return u, nil
}

// GetByID retrieves a user by their UUID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
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
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	u, err := scanUser(db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return u, nil
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User

	err := row.Scan(
		&u.ID,
		&u.TelegramID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.PhotoURL,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &u, nil
}

// Update modifies an existing user's details.
// Only updates fields that are typically changeable (username, names, role, pro-status).
func (r *Repository) Update(ctx context.Context, u *domain.User) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE users
		SET
			username     	= $2,
			first_name   	= $3,
			last_name    	= $4,
			photo_url    	= $5,
			role            = $6,
			updated_at      = now()
		WHERE id = $1
	`

	cmdTag, err := db.Exec(ctx, q,
		u.ID,
		u.Username,
		u.FirstName,
		u.LastName,
		u.PhotoURL,
		u.Role,
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
