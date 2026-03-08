// Package auth implements authentication logic for the application.
//
// It handles Telegram-based login, JWT access token generation,
// refresh token issuance and rotation, and validation of Telegram auth data.
package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/service/user"
)

//go:generate mockgen -package=auth -destination=mocks.go github.com/backforge-app/backforge/internal/service/auth UserProvider,SessionRepository,Transactor

// UserProvider defines methods for accessing and creating users.
type UserProvider interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetOrCreateByTelegramID(ctx context.Context, input user.CreateInput) (*domain.User, error)
}

// SessionRepository defines operations for managing sessions.
type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	Revoke(ctx context.Context, token string) error
}

// Transactor provides transaction boundary control.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
