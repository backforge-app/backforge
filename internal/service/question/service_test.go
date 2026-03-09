package question

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/backforge-app/backforge/internal/domain"
	repository "github.com/backforge-app/backforge/internal/repository/question"
)

func TestQuestion_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)
	transactor := NewMockTransactor(ctrl)

	svc := NewService(repo, tagRepo, transactor)

	ctx := context.Background()
	questionID := uuid.New()
	createdBy := uuid.New()

	content := map[string]interface{}{
		"ops": []interface{}{},
	}

	tagID := uuid.New()

	tests := []struct {
		name        string
		input       CreateInput
		mockSetup   func()
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "Success without tags",
			input: CreateInput{
				Title:     "Sample Question",
				Content:   content,
				Level:     domain.QuestionLevelBeginner,
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
					Return(questionID, nil)
			},
			expectedID: questionID,
		},
		{
			name: "Success with tags",
			input: CreateInput{
				Title:     "Sample Question",
				Content:   content,
				Level:     domain.QuestionLevelBeginner,
				CreatedBy: &createdBy,
				TagIDs:    []uuid.UUID{tagID},
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(questionID, nil)

				tagRepo.EXPECT().
					AddTagsToQuestion(ctx, questionID, []uuid.UUID{tagID}).
					Return(nil)
			},
			expectedID: questionID,
		},
		{
			name: "Fail - empty title",
			input: CreateInput{
				Content:   content,
				Level:     domain.QuestionLevelBeginner,
				CreatedBy: &createdBy,
			},
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: ErrQuestionInvalidData,
		},
		{
			name: "Fail - already exists",
			input: CreateInput{
				Title:     "Duplicate",
				Content:   content,
				Level:     domain.QuestionLevelBeginner,
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
					Return(uuid.Nil, repository.ErrQuestionAlreadyExists)
			},
			expectedErr: ErrQuestionAlreadyExists,
		},
		{
			name: "Fail - repository error",
			input: CreateInput{
				Title:     "DB Fail",
				Content:   content,
				Level:     domain.QuestionLevelBeginner,
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
					Return(uuid.Nil, errors.New("db timeout"))
			},
			expectedErr: errors.New("create question: db timeout"),
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

func TestQuestion_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)

	svc := NewService(repo, tagRepo, nil)

	ctx := context.Background()

	qID := uuid.New()

	expectedQ := &domain.Question{
		ID:    qID,
		Title: "Test Question",
	}

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func()
		expectedRes *domain.Question
		expectedErr error
	}{
		{
			name: "Success",
			id:   qID,
			mockSetup: func() {
				repo.EXPECT().
					GetByID(ctx, qID).
					Return(expectedQ, nil)
			},
			expectedRes: expectedQ,
		},
		{
			name: "Fail - not found",
			id:   qID,
			mockSetup: func() {
				repo.EXPECT().
					GetByID(ctx, qID).
					Return(nil, repository.ErrQuestionNotFound)
			},
			expectedErr: ErrQuestionNotFound,
		},
		{
			name: "Fail - general error",
			id:   qID,
			mockSetup: func() {
				repo.EXPECT().
					GetByID(ctx, qID).
					Return(nil, errors.New("db down"))
			},
			expectedErr: errors.New("get question by ID: db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			q, err := svc.GetByID(ctx, tt.id)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Nil(t, q)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, q)
			}
		})
	}
}

func TestQuestion_GetBySlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)

	svc := NewService(repo, tagRepo, nil)

	ctx := context.Background()

	slug := "test-question"

	expectedQ := &domain.Question{
		ID:    uuid.New(),
		Title: "Test Question",
		Slug:  slug,
	}

	tests := []struct {
		name        string
		slug        string
		mockSetup   func()
		expectedRes *domain.Question
		expectedErr error
	}{
		{
			name: "Success",
			slug: slug,
			mockSetup: func() {
				repo.EXPECT().
					GetBySlug(ctx, slug).
					Return(expectedQ, nil)
			},
			expectedRes: expectedQ,
		},
		{
			name: "Fail - not found",
			slug: slug,
			mockSetup: func() {
				repo.EXPECT().
					GetBySlug(ctx, slug).
					Return(nil, repository.ErrQuestionNotFound)
			},
			expectedErr: ErrQuestionNotFound,
		},
		{
			name: "Fail - general error",
			slug: slug,
			mockSetup: func() {
				repo.EXPECT().
					GetBySlug(ctx, slug).
					Return(nil, errors.New("db down"))
			},
			expectedErr: errors.New("get question by slug: db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			q, err := svc.GetBySlug(ctx, tt.slug)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Nil(t, q)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, q)
			}
		})
	}
}

func TestQuestion_ListCards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)
	svc := NewService(repo, tagRepo, nil)

	ctx := context.Background()

	card1 := &domain.QuestionCard{
		ID:    uuid.New(),
		Title: "Q1",
		Slug:  "q1",
		Level: domain.QuestionLevelBeginner,
	}
	card2 := &domain.QuestionCard{
		ID:    uuid.New(),
		Title: "Q2",
		Slug:  "q2",
		Level: domain.QuestionLevelMedium,
	}

	input := ListInput{
		Limit:  10,
		Offset: 0,
	}

	repo.EXPECT().
		ListCards(ctx, repository.ListOptions{
			Limit:   input.Limit,
			Offset:  input.Offset,
			Level:   input.Level,
			TopicID: input.TopicID,
			IsFree:  input.IsFree,
			TagIDs:  input.TagIDs,
		}).
		Return([]*domain.QuestionCard{card1, card2}, nil)

	result, err := svc.ListCards(ctx, input)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Q1", result[0].Title)
	assert.Equal(t, "Q2", result[1].Title)
}

func TestQuestion_ListByTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)
	svc := NewService(repo, tagRepo, nil)

	ctx := context.Background()
	topicID := uuid.New()

	q1 := &domain.Question{
		ID:    uuid.New(),
		Title: "Full Q1",
		Content: map[string]interface{}{
			"text": "Content 1",
		},
	}
	q2 := &domain.Question{
		ID:    uuid.New(),
		Title: "Full Q2",
		Content: map[string]interface{}{
			"text": "Content 2",
		},
	}

	repo.EXPECT().
		ListByTopic(ctx, topicID).
		Return([]*domain.Question{q1, q2}, nil)

	result, err := svc.ListByTopic(ctx, topicID)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Full Q1", result[0].Title)
	assert.Equal(t, "Full Q2", result[1].Title)
}

func TestQuestion_ListCards_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)
	svc := NewService(repo, tagRepo, nil)

	ctx := context.Background()
	input := ListInput{Limit: 10, Offset: 0}

	repo.EXPECT().
		ListCards(ctx, gomock.Any()).
		Return(nil, errors.New("db failure"))

	result, err := svc.ListCards(ctx, input)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "list question cards")
}

func TestQuestion_ListByTopic_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)
	svc := NewService(repo, tagRepo, nil)

	ctx := context.Background()
	topicID := uuid.New()

	repo.EXPECT().
		ListByTopic(ctx, topicID).
		Return(nil, errors.New("db failure"))

	result, err := svc.ListByTopic(ctx, topicID)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "list questions by topic")
}

func TestQuestion_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)
	transactor := NewMockTransactor(ctrl)

	svc := NewService(repo, tagRepo, transactor)

	ctx := context.Background()

	qID := uuid.New()

	existingQ := &domain.Question{
		ID:    qID,
		Title: "Old Title",
		Content: map[string]interface{}{
			"ops": []interface{}{},
		},
	}

	newTitle := "New Title"

	tests := []struct {
		name      string
		input     UpdateInput
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "Success - update title",
			input: UpdateInput{
				ID:    qID,
				Title: &newTitle,
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					GetByID(ctx, qID).
					Return(existingQ, nil)

				repo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Fail - question not found",
			input: UpdateInput{
				ID: qID,
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					GetByID(ctx, qID).
					Return(nil, repository.ErrQuestionNotFound)
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
