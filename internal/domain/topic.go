// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// TopicRow is a lightweight representation of a topic for listing in UI tables.
// Contains basic info and the number of questions in the topic.
type TopicRow struct {
	ID            uuid.UUID
	Title         string
	Slug          string
	QuestionCount int
}

// Topic represents a category/topic for questions.
type Topic struct {
	ID          uuid.UUID
	Title       string
	Slug        string
	Description string
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewTopic creates a new Topic instance.
func NewTopic(title, slug, description string, createdBy *uuid.UUID) *Topic {
	return &Topic{
		Title:       title,
		Slug:        slug,
		Description: description,
		CreatedBy:   createdBy,
	}
}

// Update modifies the mutable fields of the topic.
func (t *Topic) Update(title, slug, description string, updatedBy *uuid.UUID) {
	t.Title = title
	t.Slug = slug
	t.Description = description
	t.UpdatedBy = updatedBy
	t.UpdatedAt = time.Now().UTC()
}
