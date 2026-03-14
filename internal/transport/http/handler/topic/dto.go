// Package topic provides HTTP request and response DTOs for topic handlers.
package topic

import "github.com/google/uuid"

// createRequest represents the JSON payload for creating a new topic.
type createRequest struct {
	Title       string     `json:"title" validate:"required"` // Topic title
	Slug        string     `json:"slug" validate:"required"`  // Unique slug for the topic
	Description *string    `json:"description,omitempty"`     // Optional description
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`      // Creator user ID
}

// createResponse contains the ID of the newly created topic.
type createResponse struct {
	ID uuid.UUID `json:"id"` // New topic ID
}

// updateRequest represents the JSON payload for updating an existing topic.
type updateRequest struct {
	Title       *string    `json:"title,omitempty"`       // Optional new title
	Slug        *string    `json:"slug,omitempty"`        // Optional new slug
	Description *string    `json:"description,omitempty"` // Optional new description
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`  // Updater user ID
}

// topicResponse represents a full topic payload for GET endpoints.
type topicResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
}

// topicRowResponse represents a topic row for list views.
type topicRowResponse struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	QuestionCount int       `json:"question_count"`
}
