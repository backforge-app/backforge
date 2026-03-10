// Package progress implements the application service layer for managing
// user progress across questions and topics.
//
// It contains business logic, input DTOs, service-level errors,
// and repository interfaces that coordinate domain entities with
// the persistence layer.
//
// The package is responsible for:
//   - tracking user progress on individual questions (known, learned, skipped)
//   - maintaining the current position of a user within a topic
//   - aggregating progress data for UI consumption (e.g., progress bars)
//   - resetting progress for a topic when requested by the user
//
// It orchestrates operations between the UserQuestionProgress and
// UserTopicProgress repositories to keep question status and topic
// position consistent.
package progress

import "github.com/google/uuid"

// MarkQuestionInput holds data required to update a question progress
// and advance the user's position within a topic.
type MarkQuestionInput struct {
	UserID     uuid.UUID
	TopicID    uuid.UUID
	QuestionID uuid.UUID
}
