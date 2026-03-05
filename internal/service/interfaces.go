// Package service contains application business logic, service-level errors,
// repository interfaces, and input DTOs used to communicate between
// delivery layers (handlers) and the domain layer.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/backforge-app/backforge/internal/domain"
)

// UserRepository provides methods for managing user data in the repository.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (uuid.UUID, error)
	Update(ctx context.Context, user *domain.User) error
	GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// Transactor runs a function in a DB transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}
