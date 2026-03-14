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
	"github.com/backforge-app/backforge/internal/repository/question"
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
	content := map[string]interface{}{"ops": []interface{}{}}
	slug := "sample-question"
	topicID := uuid.New()
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
				Slug:      slug,
				TopicID:   &topicID,
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
				Slug:      slug,
				TopicID:   &topicID,
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
				Slug:      slug,
				TopicID:   &topicID,
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
				Slug:      slug,
				TopicID:   &topicID,
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
					Return(uuid.Nil, question.ErrQuestionAlreadyExists)
			},
			expectedErr: ErrQuestionAlreadyExists,
		},
		{
			name: "Fail - repository error",
			input: CreateInput{
				Title:     "DB Fail",
				Slug:      slug,
				TopicID:   &topicID,
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
			expectedErr: errors.New("db timeout"),
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
					Return(nil, question.ErrQuestionNotFound)
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
					Return(nil, question.ErrQuestionNotFound)
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
		ID:     uuid.New(),
		Title:  "Q1",
		Slug:   "q1",
		Level:  domain.QuestionLevelBeginner,
		Tags:   []string{"go", "postgresql"},
		IsNew:  true,
		IsFree: true,
	}
	card2 := &domain.QuestionCard{
		ID:     uuid.New(),
		Title:  "Q2",
		Slug:   "q2",
		Level:  domain.QuestionLevelMedium,
		Tags:   []string{"go"},
		IsNew:  false,
		IsFree: false,
	}

	input := ListInput{
		Limit:  10,
		Offset: 0,
		Search: ptrString("Q"),
		Level:  ptrQuestionLevel(domain.QuestionLevelBeginner),
		Tags:   []string{"go"},
	}

	repo.EXPECT().
		ListCards(ctx, gomock.AssignableToTypeOf(question.ListOptions{})).
		DoAndReturn(func(ctx context.Context, opts question.ListOptions) ([]*domain.QuestionCard, error) {
			assert.Equal(t, 10, opts.Limit)
			assert.Equal(t, 0, opts.Offset)
			assert.NotNil(t, opts.Search)
			assert.Equal(t, "Q", *opts.Search)
			assert.NotNil(t, opts.Level)
			assert.Equal(t, domain.QuestionLevelBeginner, *opts.Level)
			assert.Equal(t, []string{"go"}, opts.Tags)
			return []*domain.QuestionCard{card1, card2}, nil
		})

	result, err := svc.ListCards(ctx, input)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Q1", result[0].Title)
	assert.Equal(t, "Q2", result[1].Title)
}

func ptrString(s string) *string                                    { return &s }
func ptrQuestionLevel(l domain.QuestionLevel) *domain.QuestionLevel { return &l }

func TestQuestion_ListByTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	tagRepo := NewMockTagRepository(ctrl)
	svc := NewService(repo, tagRepo, nil)

	ctx := context.Background()
	topicID := uuid.New()

	q1 := &domain.Question{ID: uuid.New(), Title: "Full Q1", Content: map[string]interface{}{"text": "Content 1"}}
	q2 := &domain.Question{ID: uuid.New(), Title: "Full Q2", Content: map[string]interface{}{"text": "Content 2"}}

	repo.EXPECT().
		ListByTopic(ctx, topicID).
		Return([]*domain.Question{q1, q2}, nil)

	result, err := svc.ListByTopic(ctx, topicID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Full Q1", result[0].Title)
	assert.Equal(t, "Full Q2", result[1].Title)
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
	newTitle := "New Title"
	tag1 := uuid.New()
	tag2 := uuid.New()

	tests := []struct {
		name      string
		input     UpdateInput
		mockSetup func()
		wantErr   bool
	}{
		{
			name:  "Success - update title",
			input: UpdateInput{ID: qID, Title: &newTitle},
			mockSetup: func() {
				transactor.EXPECT().WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error { return fn(ctx) })
				repo.EXPECT().GetByID(ctx, qID).
					DoAndReturn(func(context.Context, uuid.UUID) (*domain.Question, error) {
						return &domain.Question{
							ID:      qID,
							Content: map[string]interface{}{"ops": []interface{}{}},
						}, nil
					})
				repo.EXPECT().Update(ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "Success - replace tags",
			input: UpdateInput{ID: qID, TagIDs: &[]uuid.UUID{tag1, tag2}},
			mockSetup: func() {
				transactor.EXPECT().WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error { return fn(ctx) })
				repo.EXPECT().GetByID(ctx, qID).
					Return(&domain.Question{ID: qID}, nil)
				tagRepo.EXPECT().RemoveAllForQuestion(ctx, qID).Return(nil)
				tagRepo.EXPECT().AddTagsToQuestion(ctx, qID, []uuid.UUID{tag1, tag2}).Return(nil)
				repo.EXPECT().Update(ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "Fail - question not found",
			input: UpdateInput{ID: qID},
			mockSetup: func() {
				transactor.EXPECT().WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error { return fn(ctx) })
				repo.EXPECT().GetByID(ctx, qID).
					Return(nil, question.ErrQuestionNotFound)
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
