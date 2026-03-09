// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// QuestionLevel represents the difficulty of a question.
type QuestionLevel int16

const (
	// QuestionLevelBeginner represents an easy question.
	QuestionLevelBeginner QuestionLevel = 0
	// QuestionLevelMedium represents a medium difficulty question.
	QuestionLevelMedium QuestionLevel = 1
	// QuestionLevelAdvanced represents a hard question.
	QuestionLevelAdvanced QuestionLevel = 2
)

// QuestionCard is a lightweight representation of a question,
// typically used for listing, previews, search results and cards in UI.
type QuestionCard struct {
	ID    uuid.UUID
	Title string
	Slug  string
	Level QuestionLevel
	IsNew bool
	Tags  []Tag
}

// Question represents a question entity in the system.
// Questions belong to topics and can be free or restricted.
type Question struct {
	ID        uuid.UUID
	Title     string
	Slug      string
	Content   map[string]interface{} // JSONB content stored as map
	Level     QuestionLevel
	TopicID   *uuid.UUID
	IsFree    bool
	Tags      []*Tag
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewQuestion creates a new Question instance with required fields.
func NewQuestion(
	title string,
	slug string,
	content map[string]interface{},
	level QuestionLevel,
	topicID *uuid.UUID,
	isFree bool,
	createdBy *uuid.UUID,
) *Question {
	return &Question{
		Title:     title,
		Slug:      slug,
		Content:   content,
		Level:     level,
		TopicID:   topicID,
		IsFree:    isFree,
		CreatedBy: createdBy,
	}
}

// Update modifies an existing question's mutable fields.
// Only fields like Title, Slug, Content, Level, TopicID, IsFree, UpdatedBy can be updated.
func (q *Question) Update(
	title string,
	slug string,
	content map[string]interface{},
	level QuestionLevel,
	topicID *uuid.UUID,
	isFree bool,
	updatedBy *uuid.UUID,
) {
	q.Title = title
	q.Slug = slug
	q.Content = content
	q.Level = level
	q.TopicID = topicID
	q.IsFree = isFree
	q.UpdatedBy = updatedBy
	q.UpdatedAt = time.Now().UTC()
}
