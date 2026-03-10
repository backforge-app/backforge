// Package tag implements the application service layer for tag management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package tag

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

//go:generate mockgen -package=tag -destination=mocks.go github.com/backforge-app/backforge/internal/service/tag Repository

// Repository defines data access operations for Tag entities.
type Repository interface {
	// Create persists a new tag and returns its generated ID.
	//
	// Returns tag.ErrTagAlreadyExists if a tag with the same name already exists.
	Create(ctx context.Context, t *domain.Tag) (uuid.UUID, error)

	// GetByID retrieves a tag by its unique identifier.
	//
	// Returns tag.ErrTagNotFound if the tag does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error)

	// Delete removes a tag by its unique identifier.
	//
	// Returns tag.ErrTagNotFound if the tag does not exist.
	Delete(ctx context.Context, id uuid.UUID) error

	// List returns all tags ordered by name.
	List(ctx context.Context) ([]*domain.Tag, error)
}
