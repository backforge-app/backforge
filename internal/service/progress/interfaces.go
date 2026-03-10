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
//
//go:generate mockgen -package=progress -destination=mocks.go github.com/backforge-app/backforge/internal/service/progress UserQuestionProgressRepository,UserTopicProgressRepository
package progress

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// UserQuestionProgressRepository defines data access operations for user question progress.
type UserQuestionProgressRepository interface {
	// GetByUserAndQuestion retrieves a user's progress for a specific question.
	GetByUserAndQuestion(ctx context.Context, userID, questionID uuid.UUID) (*domain.UserQuestionProgress, error)

	// ListByUserAndTopic returns all progress entries for a user filtered by topic.
	ListByUserAndTopic(ctx context.Context, userID, topicID uuid.UUID) ([]*domain.UserQuestionProgress, error)

	// SetStatus creates or updates the progress status for a specific question.
	SetStatus(ctx context.Context, userID, questionID uuid.UUID, status domain.ProgressStatus) error

	// ResetByTopic deletes all progress entries for a user in a specific topic.
	ResetByTopic(ctx context.Context, userID, topicID uuid.UUID) error
}

// UserTopicProgressRepository defines data access operations for user topic progress.
type UserTopicProgressRepository interface {
	// GetByUserAndTopic retrieves the progress for a given user and topic.
	GetByUserAndTopic(ctx context.Context, userID, topicID uuid.UUID) (*domain.UserTopicProgress, error)

	// SetPosition sets or updates the current position of a user in a topic.
	SetPosition(ctx context.Context, userID, topicID uuid.UUID, pos int) error

	// ResetByTopic resets the current position of a user in a specific topic to 0.
	ResetByTopic(ctx context.Context, userID, topicID uuid.UUID) error
}
