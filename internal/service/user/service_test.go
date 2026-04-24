package user

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/backforge-app/backforge/internal/domain"
	repouser "github.com/backforge-app/backforge/internal/repository/user"
)

// setupTxMock is a helper to automatically execute the function passed to WithinTx.
func setupTxMock(transactor *MockTransactor, ctx context.Context) {
	transactor.EXPECT().
		WithinTx(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

func TestService_CreateWithPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	validID := uuid.New()
	longPassword := strings.Repeat("a", 73)

	tests := []struct {
		name        string
		input       CreateWithPasswordInput
		mockSetup   func()
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "Success",
			input: CreateWithPasswordInput{
				Email:     "test@example.com",
				Password:  "securepassword",
				FirstName: "John",
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
			name: "Fail - empty email",
			input: CreateWithPasswordInput{
				Email:    "",
				Password: "password123",
			},
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: ErrUserInvalidData,
		},
		{
			name: "Fail - password too long",
			input: CreateWithPasswordInput{
				Email:    "test@example.com",
				Password: longPassword,
			},
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: ErrPasswordTooLong,
		},
		{
			name: "Fail - email taken",
			input: CreateWithPasswordInput{
				Email:    "taken@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, repouser.ErrUserEmailTaken)
			},
			expectedID:  uuid.Nil,
			expectedErr: ErrUserEmailTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			id, err := svc.CreateWithPassword(ctx, tt.input)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Equal(t, tt.expectedID, id)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestService_CreateOAuthUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	validID := uuid.New()

	tests := []struct {
		name        string
		input       CreateOAuthInput
		mockSetup   func()
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "Success",
			input: CreateOAuthInput{
				Email:           "oauth@example.com",
				FirstName:       "GitHub",
				IsEmailVerified: true,
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
			name: "Fail - empty email",
			input: CreateOAuthInput{
				Email: "   ",
			},
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: ErrUserInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			id, err := svc.CreateOAuthUser(ctx, tt.input)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	userID := uuid.New()
	existingUser := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "OldName",
	}

	tests := []struct {
		name      string
		input     UpdateInput
		mockSetup func()
		wantErr   error
	}{
		{
			name: "Success",
			input: UpdateInput{
				ID:        userID,
				FirstName: ptr("NewFirst"),
				Username:  ptr("new_username"),
			},
			mockSetup: func() {
				setupTxMock(transactor, ctx)
				repo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)
				repo.EXPECT().Update(ctx, gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Fail - user not found",
			input: UpdateInput{
				ID: userID,
			},
			mockSetup: func() {
				setupTxMock(transactor, ctx)
				repo.EXPECT().GetByID(ctx, userID).Return(nil, repouser.ErrUserNotFound)
			},
			wantErr: ErrUserNotFound,
		},
		{
			name: "Fail - username taken",
			input: UpdateInput{
				ID:       userID,
				Username: ptr("taken_username"),
			},
			mockSetup: func() {
				setupTxMock(transactor, ctx)
				repo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)
				repo.EXPECT().Update(ctx, gomock.Any()).Return(repouser.ErrUserUsernameTaken)
			},
			wantErr: ErrUserUsernameTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := svc.Update(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_SetNewPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name      string
		id        uuid.UUID
		password  string
		mockSetup func()
		wantErr   error
	}{
		{
			name:     "Success",
			id:       userID,
			password: "new_secure_password",
			mockSetup: func() {
				setupTxMock(transactor, ctx)
				repo.EXPECT().GetByID(ctx, userID).Return(&domain.User{ID: userID}, nil)
				repo.EXPECT().Update(ctx, gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:     "Fail - password too long",
			id:       userID,
			password: strings.Repeat("a", 73),
			mockSetup: func() {
				setupTxMock(transactor, ctx)
				repo.EXPECT().GetByID(ctx, userID).Return(&domain.User{ID: userID}, nil)
			},
			wantErr: ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := svc.SetNewPassword(ctx, tt.id, tt.password)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_MarkEmailVerified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	transactor := NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		setupTxMock(transactor, ctx)
		repo.EXPECT().GetByID(ctx, userID).Return(&domain.User{ID: userID}, nil)
		repo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		err := svc.MarkEmailVerified(ctx, userID)
		require.NoError(t, err)
	})
}

func TestService_GetByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	email := "test@example.com"
	expectedUser := &domain.User{Email: email}

	tests := []struct {
		name        string
		email       string
		mockSetup   func()
		expectedRes *domain.User
		expectedErr error
	}{
		{
			name:  "Success (trims and lowers email)",
			email: "  TeSt@ExAmPlE.CoM  ",
			mockSetup: func() {
				repo.EXPECT().GetByEmail(ctx, "test@example.com").Return(expectedUser, nil)
			},
			expectedRes: expectedUser,
			expectedErr: nil,
		},
		{
			name:  "Fail - user not found",
			email: "notfound@example.com",
			mockSetup: func() {
				repo.EXPECT().GetByEmail(ctx, "notfound@example.com").Return(nil, repouser.ErrUserNotFound)
			},
			expectedRes: nil,
			expectedErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			u, err := svc.GetByEmail(ctx, tt.email)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, u)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, u)
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	userID := uuid.New()
	expectedUser := &domain.User{ID: userID}

	t.Run("Success", func(t *testing.T) {
		repo.EXPECT().GetByID(ctx, userID).Return(expectedUser, nil)
		u, err := svc.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, u)
	})

	t.Run("Not Found", func(t *testing.T) {
		repo.EXPECT().GetByID(ctx, userID).Return(nil, repouser.ErrUserNotFound)
		u, err := svc.GetByID(ctx, userID)
		require.ErrorIs(t, err, ErrUserNotFound)
		assert.Nil(t, u)
	})
}

func TestService_IsAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("True", func(t *testing.T) {
		repo.EXPECT().IsAdmin(ctx, userID).Return(true, nil)
		isAdmin, err := svc.IsAdmin(ctx, userID)
		require.NoError(t, err)
		assert.True(t, isAdmin)
	})

	t.Run("False", func(t *testing.T) {
		repo.EXPECT().IsAdmin(ctx, userID).Return(false, nil)
		isAdmin, err := svc.IsAdmin(ctx, userID)
		require.NoError(t, err)
		assert.False(t, isAdmin)
	})
}

func ptr[T any](v T) *T {
	return &v
}
