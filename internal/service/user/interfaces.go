// Package user implements the application service layer for user management.
//
// It contains business logic for user creation, updates, retrieval,
// service-level errors, input DTOs (in other files), and coordinates
// domain entities with the persistence layer.
//
//go:generate mockgen -package=user -destination=mocks.go github.com/backforge-app/backforge/internal/service/user Repository,Transactor
package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// Repository defines data access operations for User entities.
type Repository interface {
	// Create persists a new user and returns its generated ID.
	Create(ctx context.Context, user *domain.User) (uuid.UUID, error)

	// Update modifies an existing user's mutable fields.
	Update(ctx context.Context, user *domain.User) error

	// GetByTelegramID retrieves a user by their Telegram ID.
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)

	// GetByID retrieves a user by its unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// IsAdmin checks if a user has the admin role.
	// Returns true if user.Role == "admin".
	// Returns false if user exists but is not an admin.
	// Returns an error if the user is not found or DB query fails.
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

// Transactor provides transactional execution scope for database operations.
type Transactor interface {
	// WithinTx executes the given function inside a database transaction.
	// The transaction is committed on success or rolled back on error/panic.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
