// Package tag provides HTTP request and response DTOs for tag handlers.
package tag

import "github.com/google/uuid"

// createRequest represents the JSON payload for creating a new tag.
type createRequest struct {
	Name      string     `json:"name" validate:"required"` // Tag name
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`     // Creator user ID
}

// createResponse contains the ID of the newly created tag.
type createResponse struct {
	ID uuid.UUID `json:"id"` // New tag ID
}

// tagResponse represents a tag payload for responses.
type tagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
