// Package question implements the Question application service.
//
// It contains the business logic for managing questions, including
// service methods, input DTOs, service-level errors,
// repository interfaces and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
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
	CreatedBy *uuid.UUID
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
	UpdatedBy *uuid.UUID
}

// ListInput holds filters and pagination options for listing questions.
type ListInput struct {
	Limit   int
	Offset  int
	Level   *domain.QuestionLevel
	TopicID *uuid.UUID
	IsFree  *bool
}
