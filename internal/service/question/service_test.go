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
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	questionID := uuid.New()
	createdBy := uuid.New()
	content := map[string]interface{}{"ops": []interface{}{}}

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
				Title:     "Sample Question",
				Content:   content,
				Level:     domain.QuestionLevelBeginner,
				CreatedBy: &createdBy,
			},
			mockSetup: func() {
				repo.EXPECT().Create(ctx, gomock.Any()).Return(questionID, nil)
			},
			expectedID:  questionID,
			expectedErr: nil,
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
				repo.EXPECT().Create(ctx, gomock.Any()).Return(uuid.Nil, repository.ErrQuestionAlreadyExists)
			},
			expectedID:  uuid.Nil,
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
				repo.EXPECT().Create(ctx, gomock.Any()).Return(uuid.Nil, errors.New("db timeout"))
			},
			expectedID:  uuid.Nil,
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
	svc := NewService(repo, nil)

	ctx := context.Background()
	qID := uuid.New()
	expectedQ := &domain.Question{ID: qID, Title: "Test Question"}

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
				repo.EXPECT().GetByID(ctx, qID).Return(expectedQ, nil)
			},
			expectedRes: expectedQ,
			expectedErr: nil,
		},
		{
			name: "Fail - not found",
			id:   qID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, qID).Return(nil, repository.ErrQuestionNotFound)
			},
			expectedRes: nil,
			expectedErr: ErrQuestionNotFound,
		},
		{
			name: "Fail - general error",
			id:   qID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, qID).Return(nil, errors.New("db down"))
			},
			expectedRes: nil,
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
	svc := NewService(repo, nil)

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
				repo.EXPECT().GetBySlug(ctx, slug).Return(expectedQ, nil)
			},
			expectedRes: expectedQ,
			expectedErr: nil,
		},
		{
			name: "Fail - not found",
			slug: slug,
			mockSetup: func() {
				repo.EXPECT().GetBySlug(ctx, slug).Return(nil, repository.ErrQuestionNotFound)
			},
			expectedRes: nil,
			expectedErr: ErrQuestionNotFound,
		},
		{
			name: "Fail - general error",
			slug: slug,
			mockSetup: func() {
				repo.EXPECT().GetBySlug(ctx, slug).Return(nil, errors.New("db down"))
			},
			expectedRes: nil,
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

func TestQuestion_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	q1 := &domain.Question{Title: "Q1"}
	q2 := &domain.Question{Title: "Q2"}

	input := ListInput{
		Limit:  10,
		Offset: 0,
	}

	repo.EXPECT().List(ctx, repository.ListOptions{
		Limit:  input.Limit,
		Offset: input.Offset,
	}).Return([]*domain.Question{q1, q2}, nil)

	result, err := svc.List(ctx, input)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Q1", result[0].Title)
	assert.Equal(t, "Q2", result[1].Title)
}

func TestQuestion_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

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
	newContent := map[string]interface{}{"ops": []interface{}{map[string]interface{}{"insert": "test"}}}

	tests := []struct {
		name      string
		input     UpdateInput
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "Success - update title and content",
			input: UpdateInput{
				ID:      qID,
				Title:   &newTitle,
				Content: newContent,
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
