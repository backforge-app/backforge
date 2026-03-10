// Package analytics implements the application service layer for user progress analytics.
//
// It provides aggregated statistics used by the UI dashboard,
// such as overall progress cards and topic completion charts.
//
// The package coordinates domain entities with the persistence layer
// and exposes analytics-focused operations for the application.
//
//go:generate mockgen -package=analytics -destination=mocks.go github.com/backforge-app/backforge/internal/service/analytics Repository,UserQuestionProgressRepository,UserTopicProgressRepository
package analytics

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// Repository defines analytics aggregation queries.
type Repository interface {
	// GetOverallProgress returns aggregated progress statistics.
	GetOverallProgress(ctx context.Context, userID uuid.UUID) (*domain.OverallProgress, error)

	// GetTopicProgressPercent returns completion percentage per topic.
	GetTopicProgressPercent(ctx context.Context, userID uuid.UUID) ([]*domain.TopicProgressPercent, error)
}

// UserQuestionProgressRepository defines required question progress operations.
type UserQuestionProgressRepository interface {
	// ResetAll deletes all question progress entries for the user.
	ResetAll(ctx context.Context, userID uuid.UUID) error
}

// UserTopicProgressRepository defines required topic progress operations.
type UserTopicProgressRepository interface {
	// ResetAll resets topic positions for the user.
	ResetAll(ctx context.Context, userID uuid.UUID) error
}
