// Package question implements the Question application service.
//
// It contains the business logic for managing questions, including
// service methods, input DTOs, service-level errors,
// repository interfaces and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
package question

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/question"
)

//go:generate mockgen -package=question -destination=mocks.go github.com/backforge-app/backforge/internal/service/question Repository,Transactor

// Repository provides methods for managing question data in the repository.
type Repository interface {
	Create(ctx context.Context, q *domain.Question) (uuid.UUID, error)
	Update(ctx context.Context, q *domain.Question) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Question, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Question, error)
	List(ctx context.Context, opts question.ListOptions) ([]*domain.Question, error)
}

// Transactor runs a function in a DB transaction.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
