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

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/progress"
)

// Service manages user progress across questions and topics.
// It unifies question-level progress and topic position logic into a single service.
type Service struct {
	questionRepo UserQuestionProgressRepository
	topicRepo    UserTopicProgressRepository
}

// NewService creates a new ProgressService instance.
func NewService(
	questionRepo UserQuestionProgressRepository,
	topicRepo UserTopicProgressRepository,
) *Service {
	return &Service{
		questionRepo: questionRepo,
		topicRepo:    topicRepo,
	}
}

// MarkKnown marks a user's question as known and advances the topic position.
func (s *Service) MarkKnown(ctx context.Context, input MarkQuestionInput) error {
	return s.markAndAdvance(ctx, input, domain.ProgressStatusKnown)
}

// MarkLearned marks a user's question as learned and advances the topic position.
func (s *Service) MarkLearned(ctx context.Context, input MarkQuestionInput) error {
	return s.markAndAdvance(ctx, input, domain.ProgressStatusLearned)
}

// MarkSkipped marks a user's question as skipped and advances the topic position.
func (s *Service) MarkSkipped(ctx context.Context, input MarkQuestionInput) error {
	return s.markAndAdvance(ctx, input, domain.ProgressStatusSkipped)
}

// markAndAdvance sets the question progress status and advances
// the current topic position for the user.
func (s *Service) markAndAdvance(
	ctx context.Context,
	input MarkQuestionInput,
	status domain.ProgressStatus,
) error {
	if !status.IsValid() {
		return ErrInvalidProgressStatus
	}

	// Update question progress status.
	if err := s.questionRepo.SetStatus(ctx, input.UserID, input.QuestionID, status); err != nil {
		return fmt.Errorf("set question progress status: %w", err)
	}

	// Get current topic position.
	pos := 0
	topicProgress, err := s.topicRepo.GetByUserAndTopic(ctx, input.UserID, input.TopicID)
	if err != nil && !errors.Is(err, progress.ErrTopicProgressNotFound) {
		return fmt.Errorf("get topic progress: %w", err)
	}

	if topicProgress != nil {
		pos = topicProgress.CurrentPosition
	}

	// Advance position.
	pos++

	if err := s.topicRepo.SetPosition(ctx, input.UserID, input.TopicID, pos); err != nil {
		return fmt.Errorf("update topic position: %w", err)
	}

	return nil
}

// GetByTopic returns aggregated question progress and the current topic position.
func (s *Service) GetByTopic(
	ctx context.Context,
	userID, topicID uuid.UUID,
) (*domain.TopicProgressAggregate, error) {
	progressList, err := s.questionRepo.ListByUserAndTopic(ctx, userID, topicID)
	if err != nil {
		return nil, fmt.Errorf("list question progress by topic: %w", err)
	}

	var aggregate domain.TopicProgressAggregate
	for _, p := range progressList {
		switch p.Status {
		case domain.ProgressStatusKnown:
			aggregate.Known++
		case domain.ProgressStatusLearned:
			aggregate.Learned++
		case domain.ProgressStatusSkipped:
			aggregate.Skipped++
		case domain.ProgressStatusNew:
			aggregate.New++
		}
	}

	topicProgress, err := s.topicRepo.GetByUserAndTopic(ctx, userID, topicID)
	if err != nil {
		if errors.Is(err, progress.ErrTopicProgressNotFound) {
			aggregate.CurrentPosition = 0
		} else {
			return nil, fmt.Errorf("get topic progress: %w", err)
		}
	} else {
		aggregate.CurrentPosition = topicProgress.CurrentPosition
	}

	return &aggregate, nil
}

// GetByUserAndQuestion retrieves a user's progress for a specific question.
func (s *Service) GetByUserAndQuestion(
	ctx context.Context,
	userID, questionID uuid.UUID,
) (*domain.UserQuestionProgress, error) {
	p, err := s.questionRepo.GetByUserAndQuestion(ctx, userID, questionID)
	if err != nil {
		if errors.Is(err, progress.ErrQuestionProgressNotFound) {
			// Return a "default" progress (status 'new') if no progress exists.
			return &domain.UserQuestionProgress{
				UserID:     userID,
				QuestionID: questionID,
				Status:     domain.ProgressStatusNew,
			}, nil
		}
		return nil, fmt.Errorf("get progress by user and question: %w", err)
	}
	return p, nil
}

// ResetTopicProgress resets all question progress and topic position for a specific topic.
func (s *Service) ResetTopicProgress(ctx context.Context, userID, topicID uuid.UUID) error {
	if err := s.questionRepo.ResetByTopic(ctx, userID, topicID); err != nil {
		return fmt.Errorf("reset questions progress: %w", err)
	}
	if err := s.topicRepo.ResetByTopic(ctx, userID, topicID); err != nil {
		return fmt.Errorf("reset topic position: %w", err)
	}
	return nil
}
