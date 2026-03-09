package topic

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/backforge-app/backforge/internal/domain"
)

func TestTopic_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	topicID := uuid.New()
	createdBy := uuid.New()

	tests := []struct {
		name        string
		input       CreateInput
		mockSetup   func()
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "Success",
			input: CreateInput{
				Title:       "Go",
				Slug:        "go",
				Description: "Go language",
				CreatedBy:   &createdBy,
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(topicID, nil)
			},
			expectedID: topicID,
		},
		{
			name: "Fail - empty title",
			input: CreateInput{
				Description: "No title",
				CreatedBy:   &createdBy,
			},
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: ErrTopicInvalidData,
		},
		{
			name: "Fail - already exists",
			input: CreateInput{
				Title:     "Duplicate",
				Slug:      "dup",
				CreatedBy: &createdBy,
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, ErrTopicAlreadyExists)
			},
			expectedErr: ErrTopicAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			id, err := svc.Create(ctx, tt.input)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Equal(t, tt.expectedID, id)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestTopic_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	topicID := uuid.New()
	existingTopic := &domain.Topic{
		ID:    topicID,
		Title: "Old Title",
	}

	newTitle := "New Title"
	newDesc := "Updated description"
	updatedBy := uuid.New()

	tests := []struct {
		name      string
		input     UpdateInput
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "Success",
			input: UpdateInput{
				ID:          topicID,
				Title:       &newTitle,
				Description: &newDesc,
				UpdatedBy:   &updatedBy,
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error { return fn(ctx) })

				repo.EXPECT().GetByID(ctx, topicID).Return(existingTopic, nil)
				repo.EXPECT().Update(ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Fail - topic not found",
			input: UpdateInput{
				ID: topicID,
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error { return fn(ctx) })
				repo.EXPECT().GetByID(ctx, topicID).Return(nil, ErrTopicNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := svc.Update(ctx, tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTopic_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	topicID := uuid.New()
	expectedTopic := &domain.Topic{
		ID:    topicID,
		Title: "Test Topic",
	}

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func()
		expectedRes *domain.Topic
		expectedErr error
	}{
		{
			name: "Success",
			id:   topicID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, topicID).Return(expectedTopic, nil)
			},
			expectedRes: expectedTopic,
		},
		{
			name: "Fail - not found",
			id:   topicID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, topicID).Return(nil, ErrTopicNotFound)
			},
			expectedErr: ErrTopicNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			res, err := svc.GetByID(ctx, tt.id)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Nil(t, res)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestTopic_GetBySlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	slug := "go"
	expectedTopic := &domain.Topic{
		ID:    uuid.New(),
		Title: "Go",
		Slug:  slug,
	}

	tests := []struct {
		name        string
		slug        string
		mockSetup   func()
		expectedRes *domain.Topic
		expectedErr error
	}{
		{
			name: "Success",
			slug: slug,
			mockSetup: func() {
				repo.EXPECT().GetBySlug(ctx, slug).Return(expectedTopic, nil)
			},
			expectedRes: expectedTopic,
		},
		{
			name: "Fail - not found",
			slug: slug,
			mockSetup: func() {
				repo.EXPECT().GetBySlug(ctx, slug).Return(nil, ErrTopicNotFound)
			},
			expectedErr: ErrTopicNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			res, err := svc.GetBySlug(ctx, tt.slug)
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Nil(t, res)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestTopic_ListRows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()

	topic1 := &domain.TopicRow{ID: uuid.New(), Title: "T1", QuestionCount: 3}
	topic2 := &domain.TopicRow{ID: uuid.New(), Title: "T2", QuestionCount: 0}

	repo.EXPECT().
		ListRows(ctx).
		Return([]*domain.TopicRow{topic1, topic2}, nil)

	rows, err := svc.ListRows(ctx)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "T1", rows[0].Title)
	assert.Equal(t, 3, rows[0].QuestionCount)
	assert.Equal(t, "T2", rows[1].Title)
	assert.Equal(t, 0, rows[1].QuestionCount)
}
