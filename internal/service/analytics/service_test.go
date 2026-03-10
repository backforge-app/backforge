package analytics

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

func TestService_GetOverallProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := uuid.New()

	mockRepo := NewMockRepository(ctrl)
	mockQRepo := NewMockUserQuestionProgressRepository(ctrl)
	mockTRepo := NewMockUserTopicProgressRepository(ctrl)

	service := NewService(mockRepo, mockQRepo, mockTRepo)

	expected := &domain.OverallProgress{
		Total:   100,
		Known:   40,
		Learned: 20,
		Skipped: 10,
		New:     30,
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetOverallProgress(ctx, userID).
			Return(expected, nil)

		result, err := service.GetOverallProgress(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Total != expected.Total {
			t.Fatalf("expected total %d got %d", expected.Total, result.Total)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetOverallProgress(ctx, userID).
			Return(nil, errors.New("db error"))

		_, err := service.GetOverallProgress(ctx, userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestService_GetProgressByTopicPercent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := uuid.New()

	mockRepo := NewMockRepository(ctrl)
	mockQRepo := NewMockUserQuestionProgressRepository(ctrl)
	mockTRepo := NewMockUserTopicProgressRepository(ctrl)

	service := NewService(mockRepo, mockQRepo, mockTRepo)

	topics := []*domain.TopicProgressPercent{
		{
			TopicID:   uuid.New(),
			Completed: 20,
			Total:     40,
			Percent:   50,
		},
		{
			TopicID:   uuid.New(),
			Completed: 10,
			Total:     20,
			Percent:   50,
		},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetTopicProgressPercent(ctx, userID).
			Return(topics, nil)

		result, err := service.GetProgressByTopicPercent(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != len(topics) {
			t.Fatalf("expected %d topics got %d", len(topics), len(result))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetTopicProgressPercent(ctx, userID).
			Return(nil, errors.New("db error"))

		_, err := service.GetProgressByTopicPercent(ctx, userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestService_ResetAllProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := uuid.New()

	mockRepo := NewMockRepository(ctrl)
	mockQRepo := NewMockUserQuestionProgressRepository(ctrl)
	mockTRepo := NewMockUserTopicProgressRepository(ctrl)

	service := NewService(mockRepo, mockQRepo, mockTRepo)

	t.Run("success", func(t *testing.T) {
		mockQRepo.
			EXPECT().
			ResetAll(ctx, userID).
			Return(nil)

		mockTRepo.
			EXPECT().
			ResetAll(ctx, userID).
			Return(nil)

		err := service.ResetAllProgress(ctx, userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("question repo error", func(t *testing.T) {
		mockQRepo.
			EXPECT().
			ResetAll(ctx, userID).
			Return(errors.New("db error"))

		err := service.ResetAllProgress(ctx, userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("topic repo error", func(t *testing.T) {
		mockQRepo.
			EXPECT().
			ResetAll(ctx, userID).
			Return(nil)

		mockTRepo.
			EXPECT().
			ResetAll(ctx, userID).
			Return(errors.New("db error"))

		err := service.ResetAllProgress(ctx, userID)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
