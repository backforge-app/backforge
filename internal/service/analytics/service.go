// Package analytics implements the application service layer for user progress analytics.
//
// It provides aggregated statistics used by the UI dashboard,
// such as overall progress cards and topic completion charts.
//
// The package coordinates domain entities with the persistence layer
// and exposes analytics-focused operations for the application.
package analytics

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// Service provides analytics operations for user learning progress.
type Service struct {
	repo         Repository
	questionRepo UserQuestionProgressRepository
	topicRepo    UserTopicProgressRepository
}

// NewService creates a new analytics service instance.
func NewService(
	repo Repository,
	qRepo UserQuestionProgressRepository,
	tRepo UserTopicProgressRepository,
) *Service {
	return &Service{
		repo:         repo,
		questionRepo: qRepo,
		topicRepo:    tRepo,
	}
}

// GetOverallProgress returns aggregated progress statistics for dashboard cards.
func (s *Service) GetOverallProgress(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.OverallProgress, error) {
	result, err := s.repo.GetOverallProgress(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get overall analytics progress: %w", err)
	}

	return result, nil
}

// GetProgressByTopicPercent returns completion percentages for each topic.
//
// The result is used to render a horizontal bar chart where each bar
// represents the percentage of completed questions in the topic.
func (s *Service) GetProgressByTopicPercent(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.TopicProgressPercent, error) {
	result, err := s.repo.GetTopicProgressPercent(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get topic analytics progress: %w", err)
	}

	return result, nil
}

// ResetAllProgress removes all stored progress for the user.
//
// It clears both:
//
//   - question progress statuses
//   - topic resume positions
//
// This operation is typically triggered by a "Reset Progress"
// action in the user interface.
func (s *Service) ResetAllProgress(
	ctx context.Context,
	userID uuid.UUID,
) error {
	if err := s.questionRepo.ResetAll(ctx, userID); err != nil {
		return fmt.Errorf("reset question progress: %w", err)
	}

	if err := s.topicRepo.ResetAll(ctx, userID); err != nil {
		return fmt.Errorf("reset topic progress: %w", err)
	}

	return nil
}
