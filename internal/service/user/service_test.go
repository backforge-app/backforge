package user

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/backforge-app/backforge/internal/domain"
	repository "github.com/backforge-app/backforge/internal/repository/user"
)

func TestUser_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	validID := uuid.New()

	lastName := "Doe"
	username := "johndoe"
	photoURL := "https://example.com/photo.jpg"

	tests := []struct {
		name        string
		input       CreateInput
		mockSetup   func()
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "Success with all fields",
			input: CreateInput{
				TelegramID: 12345,
				FirstName:  "John",
				LastName:   &lastName,
				Username:   &username,
				PhotoURL:   &photoURL,
				IsPro:      false,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(validID, nil)
			},
			expectedID:  validID,
			expectedErr: nil,
		},
		{
			name: "Success minimal",
			input: CreateInput{
				TelegramID: 54321,
				FirstName:  "Alice",
				IsPro:      true,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(validID, nil)
			},
			expectedID:  validID,
			expectedErr: nil,
		},
		{
			name: "Fail - invalid TelegramID",
			input: CreateInput{
				TelegramID: 0,
			},
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: ErrUserInvalidData,
		},
		{
			name: "Fail - TelegramID already taken",
			input: CreateInput{
				TelegramID: 12345,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, repository.ErrUserTelegramIDTaken)
			},
			expectedID:  uuid.Nil,
			expectedErr: ErrUserTelegramIDTaken,
		},
		{
			name: "Fail - invalid role",
			input: CreateInput{
				TelegramID: 12345,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, repository.ErrUserInvalidRole)
			},
			expectedID:  uuid.Nil,
			expectedErr: ErrUserInvalidRole,
		},
		{
			name: "Fail - general repository error",
			input: CreateInput{
				TelegramID: 12345,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, errors.New("database unavailable"))
			},
			expectedID:  uuid.Nil,
			expectedErr: errors.New("create user: database unavailable"),
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

func TestUser_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	userID := uuid.New()
	expectedUser := &domain.User{ID: userID, TelegramID: 12345}

	tests := []struct {
		name        string
		id          uuid.UUID
		mockSetup   func()
		expectedRes *domain.User
		expectedErr error
	}{
		{
			name: "Success",
			id:   userID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, userID).Return(expectedUser, nil)
			},
			expectedRes: expectedUser,
			expectedErr: nil,
		},
		{
			name: "Fail - user not found",
			id:   userID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, userID).Return(nil, repository.ErrUserNotFound)
			},
			expectedRes: nil,
			expectedErr: ErrUserNotFound,
		},
		{
			name: "Fail - general error",
			id:   userID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("db timeout"))
			},
			expectedRes: nil,
			expectedErr: errors.New("get user by ID: db timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := svc.GetByID(ctx, tt.id)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, user)
			}
		})
	}
}

func TestUser_GetByTelegramID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	telegramID := int64(98765)
	expectedUser := &domain.User{TelegramID: telegramID}

	tests := []struct {
		name        string
		tgID        int64
		mockSetup   func()
		expectedRes *domain.User
		expectedErr error
	}{
		{
			name: "Success",
			tgID: telegramID,
			mockSetup: func() {
				repo.EXPECT().GetByTelegramID(ctx, telegramID).Return(expectedUser, nil)
			},
			expectedRes: expectedUser,
			expectedErr: nil,
		},
		{
			name: "Fail - user not found",
			tgID: telegramID,
			mockSetup: func() {
				repo.EXPECT().GetByTelegramID(ctx, telegramID).Return(nil, repository.ErrUserNotFound)
			},
			expectedRes: nil,
			expectedErr: ErrUserNotFound,
		},
		{
			name: "Fail - general error",
			tgID: telegramID,
			mockSetup: func() {
				repo.EXPECT().GetByTelegramID(ctx, telegramID).Return(nil, errors.New("connection lost"))
			},
			expectedRes: nil,
			expectedErr: errors.New("get user by Telegram ID: connection lost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := svc.GetByTelegramID(ctx, tt.tgID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, user)
			}
		})
	}
}

func TestUser_GetOrCreateByTelegramID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()

	telegramID := int64(77777)
	userID := uuid.New()

	input := CreateInput{
		TelegramID: telegramID,
		FirstName:  "TestUser",
	}

	existingUser := &domain.User{
		ID:         userID,
		TelegramID: telegramID,
		FirstName:  "TestUser",
	}

	tests := []struct {
		name        string
		mockSetup   func()
		expected    *domain.User
		expectedErr error
	}{
		{
			name: "User already exists",
			mockSetup: func() {
				repo.EXPECT().
					GetByTelegramID(ctx, telegramID).
					Return(existingUser, nil)
			},
			expected:    existingUser,
			expectedErr: nil,
		},
		{
			name: "Create new user",
			mockSetup: func() {
				repo.EXPECT().
					GetByTelegramID(ctx, telegramID).
					Return(nil, repository.ErrUserNotFound)

				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(userID, nil)

				repo.EXPECT().
					GetByID(ctx, userID).
					Return(existingUser, nil)
			},
			expected:    existingUser,
			expectedErr: nil,
		},
		{
			name: "Fail - error on GetByTelegramID",
			mockSetup: func() {
				repo.EXPECT().
					GetByTelegramID(ctx, telegramID).
					Return(nil, errors.New("query failed"))
			},
			expectedErr: errors.New("query failed"),
		},
		{
			name: "Fail - error on Create",
			mockSetup: func() {
				repo.EXPECT().
					GetByTelegramID(ctx, telegramID).
					Return(nil, repository.ErrUserNotFound)

				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, errors.New("insert failed"))
			},
			expectedErr: errors.New("insert failed"),
		},
		{
			name: "Fail - error on GetByID after create",
			mockSetup: func() {
				repo.EXPECT().
					GetByTelegramID(ctx, telegramID).
					Return(nil, repository.ErrUserNotFound)

				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(userID, nil)

				repo.EXPECT().
					GetByID(ctx, userID).
					Return(nil, errors.New("fetch after create failed"))
			},
			expectedErr: errors.New("fetch after create failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := svc.GetOrCreateByTelegramID(ctx, input)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, user)
			}
		})
	}
}

func TestUser_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	userID := uuid.New()
	existingUser := &domain.User{
		ID:         userID,
		TelegramID: 11111,
		FirstName:  "OldName",
	}

	tests := []struct {
		name      string
		input     UpdateInput
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "Success - update names and pro status",
			input: UpdateInput{
				ID:        userID,
				FirstName: ptr("NewFirst"),
				LastName:  ptr("NewLast"),
				Username:  ptr("new_username"),
				IsPro:     ptr(true),
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					GetByID(ctx, userID).
					Return(existingUser, nil)

				repo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Fail - user not found",
			input: UpdateInput{
				ID: userID,
			},
			mockSetup: func() {
				transactor.EXPECT().
					WithinTx(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					GetByID(ctx, userID).
					Return(nil, repository.ErrUserNotFound)
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

func TestUser_UpdateProStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	telegramID := int64(99999)
	existingUser := &domain.User{TelegramID: telegramID}

	t.Run("Success - grant pro", func(t *testing.T) {
		transactor.EXPECT().
			WithinTx(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.EXPECT().
			GetByTelegramID(ctx, telegramID).
			Return(existingUser, nil)

		repo.EXPECT().
			Update(ctx, gomock.Any()).
			Return(nil)

		err := svc.UpdateProStatus(ctx, telegramID, true)
		require.NoError(t, err)
	})

	t.Run("Success - revoke pro", func(t *testing.T) {
		transactor.EXPECT().
			WithinTx(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.EXPECT().
			GetByTelegramID(ctx, telegramID).
			Return(existingUser, nil)

		repo.EXPECT().
			Update(ctx, gomock.Any()).
			Return(nil)

		err := svc.UpdateProStatus(ctx, telegramID, false)
		require.NoError(t, err)
	})

	t.Run("Fail - user not found", func(t *testing.T) {
		transactor.EXPECT().
			WithinTx(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		repo.EXPECT().
			GetByTelegramID(ctx, telegramID).
			Return(nil, repository.ErrUserNotFound)

		err := svc.UpdateProStatus(ctx, telegramID, true)
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func ptr[T any](v T) *T {
	return &v
}
