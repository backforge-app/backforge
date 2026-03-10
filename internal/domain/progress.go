// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProgressStatus represents the status of a user's progress on a question.
type ProgressStatus string

const (
	ProgressStatusNew     ProgressStatus = "new"     // User has not seen the question
	ProgressStatusSkipped ProgressStatus = "skipped" // User skipped the question
	ProgressStatusKnown   ProgressStatus = "known"   // User marked the question as known
	ProgressStatusLearned ProgressStatus = "learned" // User marked the question as learned
)

// IsValid checks whether the progress status is one of the supported values.
func (s ProgressStatus) IsValid() bool {
	switch s {
	case ProgressStatusNew, ProgressStatusSkipped, ProgressStatusKnown, ProgressStatusLearned:
		return true
	default:
		return false
	}
}

// UserQuestionProgress represents a user's progress on a single question.
// Each question has one progress entry per user.
type UserQuestionProgress struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	QuestionID uuid.UUID
	Status     ProgressStatus
	UpdatedAt  time.Time
}

// UserTopicProgress represents the progress of a user in a single topic.
// It stores the current position (index) of the user in the topic's question cards.
// Useful for resuming where the user left off and for optimizing topic page queries.
type UserTopicProgress struct {
	ID              uuid.UUID // Primary key
	UserID          uuid.UUID // User ID
	TopicID         uuid.UUID // Topic ID
	CurrentPosition int       // Current position in the topic's questions
	UpdatedAt       time.Time // Last update timestamp
}

// TopicProgressAggregate represents aggregated progress statistics for a topic.
//
// It is typically used by the service layer to provide summarized data
// for UI components such as progress bars or statistics cards.
// The struct counts how many questions in the topic fall into each
// progress status category for a specific user.
//
// CurrentPosition indicates the user's current position within the topic,
// allowing the UI to resume the learning flow from where the user left off.
type TopicProgressAggregate struct {
	Known           int
	Learned         int
	Skipped         int
	New             int
	CurrentPosition int
}
