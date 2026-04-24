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

	// GetByEmail retrieves a user by their unique email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByID retrieves a user by its unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// IsAdmin checks if a user has the admin role.
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

// Transactor provides transactional execution scope for database operations.
type Transactor interface {
	// WithinTx executes the given function inside a database transaction.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
