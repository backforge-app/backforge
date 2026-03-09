// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a topic or category label.
type Tag struct {
	ID        uuid.UUID
	Name      string
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTag creates a new Tag instance.
func NewTag(name string, createdBy *uuid.UUID) *Tag {
	return &Tag{
		Name:      name,
		CreatedBy: createdBy,
	}
}
