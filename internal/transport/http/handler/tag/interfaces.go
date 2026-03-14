// Package tag provides the interface to the tag service for HTTP handlers.
//
//go:generate mockgen -package=tag -destination=mocks.go github.com/backforge-app/backforge/internal/transport/http/handler/tag Service
package tag

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// Service defines the interface that HTTP handlers use to perform tag operations.
type Service interface {
	// Create creates a new tag.
	Create(ctx context.Context, name string, createdBy *uuid.UUID) (uuid.UUID, error)

	// Delete removes a tag by its ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID retrieves a tag by its unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error)

	// List retrieves all tags.
	List(ctx context.Context) ([]*domain.Tag, error)
}
