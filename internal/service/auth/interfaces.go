// Package auth implements authentication and session management logic.
//
// It supports Telegram-based authentication, JWT issuance, refresh token rotation,
// session persistence, and revocation.
//
//go:generate mockgen -package=auth -destination=mocks.go github.com/backforge-app/backforge/internal/service/auth UserProvider,SessionRepository,Transactor
package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/service/user"
)

// UserProvider defines access and creation operations for user entities.
type UserProvider interface {
	// GetByID retrieves a user by its unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetOrCreateByTelegramID finds a user by Telegram ID or creates a new one if not found.
	GetOrCreateByTelegramID(ctx context.Context, input user.CreateInput) (*domain.User, error)
}

// SessionRepository defines persistence operations for authentication sessions.
type SessionRepository interface {
	// Create stores a new session in the storage.
	Create(ctx context.Context, s *domain.Session) error

	// GetByToken retrieves an active session by its refresh or access token.
	GetByToken(ctx context.Context, token string) (*domain.Session, error)

	// Revoke invalidates a session by its token.
	Revoke(ctx context.Context, token string) error
}

// Transactor provides transactional execution scope for database operations.
type Transactor interface {
	// WithinTx executes the given function inside a database transaction.
	// The transaction is committed on success or rolled back on error/panic.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
