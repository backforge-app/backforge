// Package question provides the interface to the question service for HTTP handlers.
//
//go:generate mockgen -package=question -destination=mocks.go github.com/backforge-app/backforge/internal/transport/http/handler/question Service
package question

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	servicequestion "github.com/backforge-app/backforge/internal/service/question"
)

// Service defines the interface that HTTP handlers use to perform question operations.
type Service interface {
	// Create creates a new question and returns its ID.
	Create(ctx context.Context, input servicequestion.CreateInput) (uuid.UUID, error)

	// Update modifies an existing question.
	Update(ctx context.Context, input servicequestion.UpdateInput) error

	// GetByID retrieves a question by its unique ID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Question, error)

	// GetBySlug retrieves a question by its slug.
	GetBySlug(ctx context.Context, slug string) (*domain.Question, error)

	// ListCards returns question cards with filters and pagination.
	ListCards(ctx context.Context, input servicequestion.ListInput) ([]*domain.QuestionCard, error)

	// ListByTopic returns all questions for a given topic.
	ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*domain.Question, error)
}
