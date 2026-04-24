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

// Create inserts a new user into the database and returns its assigned UUID.
// It handles unique constraint violations for email and username.
func (r *Repository) Create(ctx context.Context, u *domain.User) (uuid.UUID, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO users (
			email,
			password_hash,
			username,
			first_name,
			last_name,
			photo_url,
			role,
			is_email_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id uuid.UUID

	err := db.QueryRow(ctx, q,
		u.Email,
		u.PasswordHash,
		u.Username,
		u.FirstName,
		u.LastName,
		u.PhotoURL,
		u.Role,
		u.IsEmailVerified,
	).Scan(&id)

	if err != nil {
		return uuid.Nil, handlePostgresError(err, "create user")
	}

	return id, nil
}

// GetByEmail retrieves a user by their unique email address.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id, email, password_hash, is_email_verified,
			username, first_name, last_name, photo_url,
			role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	u, err := scanUser(db.QueryRow(ctx, q, email))
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return u, nil
}

// GetByID retrieves a user by their UUID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
			id, email, password_hash, is_email_verified,
			username, first_name, last_name, photo_url,
			role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	u, err := scanUser(db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return u, nil
}

// Update modifies an existing user's details.
// It updates timestamps automatically and handles unique constraint violations.
func (r *Repository) Update(ctx context.Context, u *domain.User) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE users
		SET
			email             = $2,
			password_hash     = $3,
			is_email_verified = $4,
			username          = $5,
			first_name        = $6,
			last_name         = $7,
			photo_url         = $8,
			role              = $9,
			updated_at        = now()
		WHERE id = $1
	`

	cmdTag, err := db.Exec(ctx, q,
		u.ID,
		u.Email,
		u.PasswordHash,
		u.IsEmailVerified,
		u.Username,
		u.FirstName,
		u.LastName,
		u.PhotoURL,
		u.Role,
	)

	if err != nil {
		return handlePostgresError(err, "update user")
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// IsAdmin checks if a user has the admin role.
// Returns true if user.Role == "admin".
// Returns false if the user exists but is not an admin.
// Returns an error if the user is not found.
func (r *Repository) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT role
		FROM users
		WHERE id = $1
	`

	var role string
	err := db.QueryRow(ctx, q, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrUserNotFound
		}
		return false, fmt.Errorf("failed to query user role: %w", err)
	}

	return role == "admin", nil
}

// scanUser is a helper function to map a database row to the User domain entity.
func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.IsEmailVerified,
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
		return nil, fmt.Errorf("scan user row: %w", err)
	}

	return &u, nil
}

// handlePostgresError parses pgx errors and maps database constraint violations
// to domain-specific sentinel errors.
func handlePostgresError(err error, operation string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			// Match constraint names based on the table definition.
			if pgErr.ConstraintName == "users_email_key" {
				return ErrUserEmailTaken
			}
			if pgErr.ConstraintName == "users_username_key" {
				return ErrUserUsernameTaken
			}
		case pgerrcode.InvalidTextRepresentation, pgerrcode.InvalidParameterValue:
			// Triggered if the role ENUM is invalid.
			return ErrUserInvalidRole
		}
	}
	return fmt.Errorf("db %s failed: %w", operation, err)
}
