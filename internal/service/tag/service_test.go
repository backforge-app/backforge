package tag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/service/tag"
)

func TestTagService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := tag.NewMockRepository(ctrl)
	svc := tag.NewService(mockRepo)

	ctx := context.Background()
	createdBy := uuid.New()
	tagName := "go"

	t.Run("success", func(t *testing.T) {
		expectedID := uuid.New()
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(expectedID, nil)

		id, err := svc.Create(ctx, tagName, &createdBy)
		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)
	})

	t.Run("invalid name", func(t *testing.T) {
		id, err := svc.Create(ctx, "", &createdBy)
		assert.Equal(t, uuid.Nil, id)
		assert.ErrorIs(t, err, tag.ErrTagInvalidData)
	})

	t.Run("already exists", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(uuid.Nil, tag.ErrTagAlreadyExists)

		id, err := svc.Create(ctx, tagName, &createdBy)
		assert.Equal(t, uuid.Nil, id)
		assert.ErrorIs(t, err, tag.ErrTagAlreadyExists)
	})
}

func TestTagService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := tag.NewMockRepository(ctrl)
	svc := tag.NewService(mockRepo)

	ctx := context.Background()
	tagID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(ctx, tagID).
			Return(nil)

		err := svc.Delete(ctx, tagID)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(ctx, tagID).
			Return(tag.ErrTagNotFound)

		err := svc.Delete(ctx, tagID)
		assert.ErrorIs(t, err, tag.ErrTagNotFound)
	})

	t.Run("other error", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(ctx, tagID).
			Return(errors.New("db error"))

		err := svc.Delete(ctx, tagID)
		assert.Error(t, err)
		assert.NotEqual(t, tag.ErrTagNotFound, err)
	})
}

func TestTagService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := tag.NewMockRepository(ctrl)
	svc := tag.NewService(mockRepo)

	ctx := context.Background()
	tagID := uuid.New()
	expectedTag := &domain.Tag{
		ID:   tagID,
		Name: "go",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(ctx, tagID).
			Return(expectedTag, nil)

		result, err := svc.GetByID(ctx, tagID)
		assert.NoError(t, err)
		assert.Equal(t, expectedTag, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(ctx, tagID).
			Return(nil, tag.ErrTagNotFound)

		result, err := svc.GetByID(ctx, tagID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, tag.ErrTagNotFound)
	})

	t.Run("other error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(ctx, tagID).
			Return(nil, errors.New("db error"))

		result, err := svc.GetByID(ctx, tagID)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestTagService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := tag.NewMockRepository(ctrl)
	svc := tag.NewService(mockRepo)

	ctx := context.Background()
	expectedTags := []*domain.Tag{
		{ID: uuid.New(), Name: "go"},
		{ID: uuid.New(), Name: "postgres"},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			List(ctx).
			Return(expectedTags, nil)

		tags, err := svc.List(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedTags, tags)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.EXPECT().
			List(ctx).
			Return(nil, errors.New("db error"))

		tags, err := svc.List(ctx)
		assert.Nil(t, tags)
		assert.Error(t, err)
	})
}
