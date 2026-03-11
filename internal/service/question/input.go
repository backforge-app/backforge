// Package question implements the application service layer for question management.
//
// It contains business logic, input DTOs, service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package question

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// CreateInput holds data for creating a new question.
type CreateInput struct {
	Title     string
	Slug      string
	Content   map[string]interface{}
	Level     domain.QuestionLevel
	TopicID   *uuid.UUID
	IsFree    bool
	TagIDs    []uuid.UUID
	CreatedBy *uuid.UUID
}

func (in CreateInput) Validate() error {
	if in.Title == "" {
		return ErrQuestionInvalidData
	}

	if in.Slug == "" {
		return ErrQuestionInvalidData
	}

	if len(in.Content) == 0 {
		return ErrQuestionInvalidData
	}

	if !in.Level.IsValid() {
		return ErrQuestionInvalidData
	}

	return nil
}

// UpdateInput holds data for updating an existing question.
type UpdateInput struct {
	ID        uuid.UUID
	Title     *string
	Slug      *string
	Content   map[string]interface{}
	Level     *domain.QuestionLevel
	TopicID   *uuid.UUID
	IsFree    *bool
	TagIDs    *[]uuid.UUID
	UpdatedBy *uuid.UUID
}

// ListInput holds filters and pagination options for listing questions.
type ListInput struct {
	Limit   int
	Offset  int
	Level   *domain.QuestionLevel
	TopicID *uuid.UUID
	IsFree  *bool
	TagIDs  []uuid.UUID
}
