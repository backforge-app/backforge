// Package question implements the application service layer for question management.
//
// It contains business logic, input DTOs, service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
//
//go:generate mockgen -package=question -destination=mocks.go github.com/backforge-app/backforge/internal/service/question Repository,TagRepository,Transactor
package question

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/question"
)

// Repository defines data access operations for Question entities.
type Repository interface {
	// Create persists a new question and returns its generated ID.
	Create(ctx context.Context, q *domain.Question) (uuid.UUID, error)

	// Update modifies an existing question.
	Update(ctx context.Context, q *domain.Question) error

	// GetByID retrieves a full question by its ID, including tags.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Question, error)

	// GetBySlug retrieves a full question by its unique slug, including tags.
	GetBySlug(ctx context.Context, slug string) (*domain.Question, error)

	// ListCards returns lightweight question representations suitable for listings/cards.
	ListCards(ctx context.Context, opts question.ListOptions) ([]*domain.QuestionCard, error)

	// ListByTopic returns all questions belonging to the specified topic.
	ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*domain.Question, error)
}

// TagRepository defines operations for managing tags associated with questions.
type TagRepository interface {
	// AddTagsToQuestion attaches one or more tags to a question.
	AddTagsToQuestion(ctx context.Context, questionID uuid.UUID, tagIDs []uuid.UUID) error

	// RemoveAllForQuestion removes all tag associations from the given question.
	RemoveAllForQuestion(ctx context.Context, questionID uuid.UUID) error

	// ListTagsForQuestion returns all tags currently attached to the question.
	ListTagsForQuestion(ctx context.Context, questionID uuid.UUID) ([]*domain.Tag, error)
}

// Transactor provides transactional execution scope for database operations.
type Transactor interface {
	// WithinTx executes the given function inside a database transaction.
	// The transaction is committed on success or rolled back on error/panic.
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
