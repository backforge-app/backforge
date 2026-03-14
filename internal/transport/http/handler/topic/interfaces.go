// Package topic provides the interface to the topic service for HTTP handlers.
//
//go:generate mockgen -package=topic -destination=mocks.go github.com/backforge-app/backforge/internal/transport/http/handler/topic Service
package topic

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	servicetopic "github.com/backforge-app/backforge/internal/service/topic"
)

// Service defines the interface that HTTP handlers use to perform topic operations.
type Service interface {
	// Create creates a new topic.
	Create(ctx context.Context, input servicetopic.CreateInput) (uuid.UUID, error)

	// Update modifies an existing topic.
	Update(ctx context.Context, input servicetopic.UpdateInput) error

	// GetByID retrieves a topic by its unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)

	// GetBySlug retrieves a topic by its unique slug.
	GetBySlug(ctx context.Context, slug string) (*domain.Topic, error)

	// ListRows retrieves a list of topics formatted as rows, typically including question counts.
	ListRows(ctx context.Context) ([]*domain.TopicRow, error)
}
