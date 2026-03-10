// Package topic implements the application service layer for topic management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
//
//go:generate mockgen -package=topic -destination=mocks.go github.com/backforge-app/backforge/internal/service/topic Repository,Transactor
package topic

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// Repository defines data access operations for Topic entities.
type Repository interface {
	// Create persists a new topic and returns its generated ID.
	Create(ctx context.Context, t *domain.Topic) (uuid.UUID, error)

	// Update modifies an existing topic.
	Update(ctx context.Context, t *domain.Topic) error

	// GetByID retrieves a topic by its ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)

	// GetBySlug retrieves a topic by its unique slug.
	GetBySlug(ctx context.Context, slug string) (*domain.Topic, error)

	// ListRows retrieves all topics with question counts.
	ListRows(ctx context.Context) ([]*domain.TopicRow, error)
}

// Transactor provides transactional execution scope for database operations.
type Transactor interface {
	// WithinTx executes the given function inside a database transaction.
	// The transaction is committed on success or rolled back on error/panic.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
