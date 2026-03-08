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

// Question represents a question entity in the system.
// Questions belong to topics and can be free or restricted.
type Question struct {
	ID        uuid.UUID
	Title     string
	Content   map[string]interface{} // JSONB content stored as map
	Level     QuestionLevel
	TopicID   *uuid.UUID
	IsFree    bool
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewQuestion creates a new Question instance with required fields.
func NewQuestion(title string, content map[string]interface{}, level QuestionLevel, topicID, createdBy *uuid.UUID, isFree bool) *Question {
	return &Question{
		Title:     title,
		Content:   content,
		Level:     level,
		TopicID:   topicID,
		CreatedBy: createdBy,
		IsFree:    isFree,
	}
}

// Update modifies an existing question's mutable fields.
// Only fields like Title, Content, Level, TopicID, IsFree, UpdatedBy can be updated.
func (q *Question) Update(title string, content map[string]interface{}, level QuestionLevel, topicID *uuid.UUID, isFree bool, updatedBy *uuid.UUID) {
	q.Title = title
	q.Content = content
	q.Level = level
	q.TopicID = topicID
	q.IsFree = isFree
	q.UpdatedBy = updatedBy
	q.UpdatedAt = time.Now().UTC()
}
