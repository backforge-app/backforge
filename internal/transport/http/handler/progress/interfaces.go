// Package progress provides the interface to the progress service for HTTP handlers.
//
//go:generate mockgen -package=progress -destination=mocks.go github.com/backforge-app/backforge/internal/transport/http/handler/progress Service
package progress

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	serviceprogress "github.com/backforge-app/backforge/internal/service/progress"
)

// Service defines the interface for managing user progress across topics and questions.
type Service interface {
	// MarkKnown marks a user's question as known and advances the topic position.
	MarkKnown(ctx context.Context, input serviceprogress.MarkQuestionInput) error

	// MarkLearned marks a user's question as learned and advances the topic position.
	MarkLearned(ctx context.Context, input serviceprogress.MarkQuestionInput) error

	// MarkSkipped marks a user's question as skipped and advances the topic position.
	MarkSkipped(ctx context.Context, input serviceprogress.MarkQuestionInput) error

	// GetByTopic returns aggregated question progress and the current topic position for a user.
	GetByTopic(ctx context.Context, userID, topicID uuid.UUID) (*domain.TopicProgressAggregate, error)

	// GetByUserAndQuestion retrieves a user's progress for a specific question.
	GetByUserAndQuestion(ctx context.Context, userID, questionID uuid.UUID) (*domain.UserQuestionProgress, error)

	// ResetTopicProgress resets all question progress and topic position for a specific topic.
	ResetTopicProgress(ctx context.Context, userID, topicID uuid.UUID) error
}
