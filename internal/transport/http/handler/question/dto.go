// Package question provides HTTP request and response DTOs for question handlers.
package question

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// createRequest represents the JSON payload for creating a new question.
type createRequest struct {
	Title     string                 `json:"title" validate:"required"`   // Question title
	Slug      string                 `json:"slug" validate:"required"`    // Unique slug
	Content   map[string]interface{} `json:"content" validate:"required"` // Question content (structured)
	Level     string                 `json:"level" validate:"required"`   // Difficulty level
	TopicID   *uuid.UUID             `json:"topic_id,omitempty"`          // Optional topic
	IsFree    bool                   `json:"is_free"`                     // Access flag
	TagIDs    []uuid.UUID            `json:"tag_ids,omitempty"`           // Tags
	CreatedBy *uuid.UUID             `json:"created_by,omitempty"`        // Creator user ID
}

// createResponse contains the ID of the newly created question.
type createResponse struct {
	ID uuid.UUID `json:"id"` // New question ID
}

// updateRequest represents the JSON payload for updating an existing question.
type updateRequest struct {
	Title     *string                `json:"title,omitempty"`      // Optional new title
	Slug      *string                `json:"slug,omitempty"`       // Optional new slug
	Content   map[string]interface{} `json:"content,omitempty"`    // Optional new content
	Level     *string                `json:"level,omitempty"`      // Optional new level
	TopicID   *uuid.UUID             `json:"topic_id,omitempty"`   // Optional topic
	IsFree    *bool                  `json:"is_free,omitempty"`    // Optional access flag
	TagIDs    *[]uuid.UUID           `json:"tag_ids,omitempty"`    // Optional new tags
	UpdatedBy *uuid.UUID             `json:"updated_by,omitempty"` // Updater user ID
}

// questionResponse represents a full question payload for GET endpoints.
type questionResponse struct {
	ID        uuid.UUID              `json:"id"`
	Title     string                 `json:"title"`
	Slug      string                 `json:"slug"`
	Content   map[string]interface{} `json:"content"`
	Level     string                 `json:"level"`
	TopicID   *uuid.UUID             `json:"topic_id,omitempty"`
	IsFree    bool                   `json:"is_free"`
	TagIDs    []uuid.UUID            `json:"tag_ids"`
	CreatedBy *uuid.UUID             `json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID             `json:"updated_by,omitempty"`
}

// listCardResponse represents a question card for the list.
type listCardResponse struct {
	ID     uuid.UUID `json:"id"`
	Title  string    `json:"title"`
	Slug   string    `json:"slug"`
	Level  string    `json:"level"`
	Tags   []string  `json:"tags"`
	IsNew  bool      `json:"is_new"`
	IsFree bool      `json:"is_free"`
}

// toQuestionResponse converts a domain.Question to questionResponse DTO.
func toQuestionResponse(q *domain.Question) questionResponse {
	tagIDs := make([]uuid.UUID, len(q.Tags))
	for i, t := range q.Tags {
		tagIDs[i] = t.ID
	}

	return questionResponse{
		ID:        q.ID,
		Title:     q.Title,
		Slug:      q.Slug,
		Content:   q.Content,
		Level:     q.Level.String(),
		TopicID:   q.TopicID,
		IsFree:    q.IsFree,
		TagIDs:    tagIDs,
		CreatedBy: q.CreatedBy,
		UpdatedBy: q.UpdatedBy,
	}
}
