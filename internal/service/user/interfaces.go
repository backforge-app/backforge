// Package user implements the User application service.
//
// It contains the business logic for managing users, including
// service methods, input DTOs, service-level errors,
// repository interfaces and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

//go:generate mockgen -package=user -destination=mocks.go github.com/backforge-app/backforge/internal/service/user Repository,Transactor

// Repository provides methods for managing user data in the repository.
type Repository interface {
	Create(ctx context.Context, user *domain.User) (uuid.UUID, error)
	Update(ctx context.Context, user *domain.User) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// Transactor runs a function in a DB transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
