// Package topic implements the application service layer for topic management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package topic

import (
	"github.com/google/uuid"
)

// CreateInput holds data for creating a new topic.
type CreateInput struct {
	Title       string
	Slug        string
	Description string
	CreatedBy   *uuid.UUID
}

// UpdateInput holds data for updating an existing topic.
type UpdateInput struct {
	ID          uuid.UUID
	Title       *string
	Slug        *string
	Description *string
	UpdatedBy   *uuid.UUID
}
