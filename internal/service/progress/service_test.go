package progress

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	repoprogress "github.com/backforge-app/backforge/internal/repository/progress"
)

func newTestService(ctrl *gomock.Controller) (
	*Service,
	*MockUserQuestionProgressRepository,
	*MockUserTopicProgressRepository,
) {
	qRepo := NewMockUserQuestionProgressRepository(ctrl)
	tRepo := NewMockUserTopicProgressRepository(ctrl)

	service := NewService(qRepo, tRepo)

	return service, qRepo, tRepo
}

func TestService_MarkKnown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service, qRepo, tRepo := newTestService(ctrl)

	ctx := context.Background()
	userID := uuid.New()
	topicID := uuid.New()
	questionID := uuid.New()

	input := MarkQuestionInput{
		UserID:     userID,
		TopicID:    topicID,
		QuestionID: questionID,
	}

	t.Run("success existing topic progress", func(t *testing.T) {
		qRepo.EXPECT().
			SetStatus(ctx, userID, questionID, domain.ProgressStatusKnown).
			Return(nil)

		tRepo.EXPECT().
			GetByUserAndTopic(ctx, userID, topicID).
			Return(&domain.UserTopicProgress{
				CurrentPosition: 3,
			}, nil)

		tRepo.EXPECT().
			SetPosition(ctx, userID, topicID, 4).
			Return(nil)

		err := service.MarkKnown(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("topic progress not found", func(t *testing.T) {
		qRepo.EXPECT().
			SetStatus(ctx, userID, questionID, domain.ProgressStatusKnown).
			Return(nil)

		tRepo.EXPECT().
			GetByUserAndTopic(ctx, userID, topicID).
			Return(nil, repoprogress.ErrTopicProgressNotFound)

		tRepo.EXPECT().
			SetPosition(ctx, userID, topicID, 1).
			Return(nil)

		err := service.MarkKnown(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("set status error", func(t *testing.T) {
		qRepo.EXPECT().
			SetStatus(ctx, userID, questionID, domain.ProgressStatusKnown).
			Return(errors.New("db error"))

		err := service.MarkKnown(ctx, input)

		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestService_MarkSkipped(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service, qRepo, tRepo := newTestService(ctrl)

	ctx := context.Background()
	userID := uuid.New()
	topicID := uuid.New()
	questionID := uuid.New()

	input := MarkQuestionInput{
		UserID:     userID,
		TopicID:    topicID,
		QuestionID: questionID,
	}

	qRepo.EXPECT().
		SetStatus(ctx, userID, questionID, domain.ProgressStatusSkipped).
		Return(nil)

	tRepo.EXPECT().
		GetByUserAndTopic(ctx, userID, topicID).
		Return(nil, repoprogress.ErrTopicProgressNotFound)

	tRepo.EXPECT().
		SetPosition(ctx, userID, topicID, 1).
		Return(nil)

	err := service.MarkSkipped(ctx, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_GetByTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service, qRepo, tRepo := newTestService(ctrl)

	ctx := context.Background()
	userID := uuid.New()
	topicID := uuid.New()

	progress := []*domain.UserQuestionProgress{
		{Status: domain.ProgressStatusKnown},
		{Status: domain.ProgressStatusLearned},
		{Status: domain.ProgressStatusSkipped},
		{Status: domain.ProgressStatusNew},
	}

	t.Run("success", func(t *testing.T) {
		qRepo.EXPECT().
			ListByUserAndTopic(ctx, userID, topicID).
			Return(progress, nil)

		tRepo.EXPECT().
			GetByUserAndTopic(ctx, userID, topicID).
			Return(&domain.UserTopicProgress{
				CurrentPosition: 5,
			}, nil)

		result, err := service.GetByTopic(ctx, userID, topicID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Known != 1 ||
			result.Learned != 1 ||
			result.Skipped != 1 ||
			result.New != 1 {
			t.Fatal("incorrect aggregation")
		}

		if result.CurrentPosition != 5 {
			t.Fatal("incorrect topic position")
		}
	})

	t.Run("topic progress not found", func(t *testing.T) {
		qRepo.EXPECT().
			ListByUserAndTopic(ctx, userID, topicID).
			Return(progress, nil)

		tRepo.EXPECT().
			GetByUserAndTopic(ctx, userID, topicID).
			Return(nil, repoprogress.ErrTopicProgressNotFound)

		result, err := service.GetByTopic(ctx, userID, topicID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.CurrentPosition != 0 {
			t.Fatal("expected default position 0")
		}
	})
}

func TestService_GetByUserAndQuestion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service, qRepo, _ := newTestService(ctrl)

	ctx := context.Background()
	userID := uuid.New()
	questionID := uuid.New()

	t.Run("success", func(t *testing.T) {
		progress := &domain.UserQuestionProgress{
			UserID:     userID,
			QuestionID: questionID,
			Status:     domain.ProgressStatusKnown,
		}

		qRepo.EXPECT().
			GetByUserAndQuestion(ctx, userID, questionID).
			Return(progress, nil)

		result, err := service.GetByUserAndQuestion(ctx, userID, questionID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != domain.ProgressStatusKnown {
			t.Fatal("wrong status")
		}
	})

	t.Run("not found returns new", func(t *testing.T) {
		qRepo.EXPECT().
			GetByUserAndQuestion(ctx, userID, questionID).
			Return(nil, repoprogress.ErrQuestionProgressNotFound)

		result, err := service.GetByUserAndQuestion(ctx, userID, questionID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Status != domain.ProgressStatusNew {
			t.Fatal("expected new status")
		}
	})
}

func TestService_ResetTopicProgress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service, qRepo, tRepo := newTestService(ctrl)

	ctx := context.Background()
	userID := uuid.New()
	topicID := uuid.New()

	t.Run("success", func(t *testing.T) {
		qRepo.EXPECT().
			ResetByTopic(ctx, userID, topicID).
			Return(nil)

		tRepo.EXPECT().
			ResetByTopic(ctx, userID, topicID).
			Return(nil)

		err := service.ResetTopicProgress(ctx, userID, topicID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("question reset error", func(t *testing.T) {
		qRepo.EXPECT().
			ResetByTopic(ctx, userID, topicID).
			Return(errors.New("db error"))

		err := service.ResetTopicProgress(ctx, userID, topicID)

		if err == nil {
			t.Fatal("expected error")
		}
	})
}
