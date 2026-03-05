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
	"github.com/backforge-app/backforge/internal/infra/postgres"
	"github.com/backforge-app/backforge/internal/service"
)

func TestUser_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := service.NewMockUserRepository(ctrl)
	transactor := service.NewMockTransactor(ctrl)
	svc := NewService(repo, transactor)

	ctx := context.Background()
	validID := uuid.New()

	firstName := "John"
	username := "johndoe"

	tests := []struct {
		name        string
		input       CreateUserInput
		mockSetup   func()
		expectedID  uuid.UUID
		expectedErr error
	}{
		{
			name: "Success",
			input: CreateUserInput{
				TgUserID:  12345,
				FirstName: "John",
				LastName:  &firstName,
				Username:  &username,
				IsPro:     false,
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
			name: "Fail_InvalidTgUserID",
			input: CreateUserInput{
				TgUserID: 0,
			},
			mockSetup:   func() {},
			expectedID:  uuid.Nil,
			expectedErr: service.ErrUserInvalidData,
		},
		{
			name: "Fail_TgUserIDTaken",
			input: CreateUserInput{
				TgUserID: 12345,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, postgres.ErrUserTgUserIDTaken)
			},
			expectedID:  uuid.Nil,
			expectedErr: service.ErrUserTgUserIDTaken,
		},
		{
			name: "Fail_InvalidRole",
			input: CreateUserInput{
				TgUserID: 12345,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, postgres.ErrUserInvalidRole)
			},
			expectedID:  uuid.Nil,
			expectedErr: service.ErrUserInvalidRole,
		},
		{
			name: "Fail_GeneralRepoError",
			input: CreateUserInput{
				TgUserID: 12345,
			},
			mockSetup: func() {
				repo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(uuid.Nil, errors.New("db connection down"))
			},
			expectedID:  uuid.Nil,
			expectedErr: errors.New("create user: db connection down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			id, err := svc.Create(ctx, tt.input)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
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

	repo := service.NewMockUserRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	userID := uuid.New()
	expectedUser := &domain.User{ID: userID}

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
			name: "Fail_UserNotFound",
			id:   userID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, userID).Return(nil, postgres.ErrUserNotFound)
			},
			expectedRes: nil,
			expectedErr: service.ErrUserNotFound,
		},
		{
			name: "Fail_GeneralError",
			id:   userID,
			mockSetup: func() {
				repo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("get user by ID: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := svc.GetByID(ctx, tt.id)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, user)
			}
		})
	}
}

func TestUser_GetByTgUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := service.NewMockUserRepository(ctrl)
	svc := NewService(repo, nil)

	ctx := context.Background()
	tgUserID := int64(12345)
	expectedUser := &domain.User{TgUserID: tgUserID}

	tests := []struct {
		name        string
		tgID        int64
		mockSetup   func()
		expectedRes *domain.User
		expectedErr error
	}{
		{
			name: "Success",
			tgID: tgUserID,
			mockSetup: func() {
				repo.EXPECT().GetByTgUserID(ctx, tgUserID).Return(expectedUser, nil)
			},
			expectedRes: expectedUser,
			expectedErr: nil,
		},
		{
			name: "Fail_UserNotFound",
			tgID: tgUserID,
			mockSetup: func() {
				repo.EXPECT().GetByTgUserID(ctx, tgUserID).Return(nil, postgres.ErrUserNotFound)
			},
			expectedRes: nil,
			expectedErr: service.ErrUserNotFound,
		},
		{
			name: "Fail_GeneralError",
			tgID: tgUserID,
			mockSetup: func() {
				repo.EXPECT().GetByTgUserID(ctx, tgUserID).Return(nil, errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: errors.New("get user by Telegram ID: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := svc.GetByTgUserID(ctx, tt.tgID)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, user)
			}
		})
	}
}

func TestUser_Update_TxHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transactor := service.NewMockTransactor(ctrl)
	svc := NewService(nil, transactor)

	ctx := context.Background()
	input := UpdateUserInput{ID: uuid.New()}

	t.Run("Success_Tx", func(t *testing.T) {
		transactor.EXPECT().
			WithinTx(ctx, gomock.Any()).
			Return(nil)

		err := svc.Update(ctx, input)
		require.NoError(t, err)
	})

	t.Run("Fail_Tx", func(t *testing.T) {
		expectedErr := errors.New("transaction failed")
		transactor.EXPECT().
			WithinTx(ctx, gomock.Any()).
			Return(expectedErr)

		err := svc.Update(ctx, input)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestUser_UpdateProStatus_TxHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	transactor := service.NewMockTransactor(ctrl)
	svc := NewService(nil, transactor)

	ctx := context.Background()

	t.Run("Success_Tx", func(t *testing.T) {
		transactor.EXPECT().
			WithinTx(ctx, gomock.Any()).
			Return(nil)

		err := svc.UpdateProStatus(ctx, 12345, true)
		require.NoError(t, err)
	})

	t.Run("Fail_Tx", func(t *testing.T) {
		expectedErr := errors.New("transaction failed")
		transactor.EXPECT().
			WithinTx(ctx, gomock.Any()).
			Return(expectedErr)

		err := svc.UpdateProStatus(ctx, 12345, true)
		require.ErrorIs(t, err, expectedErr)
	})
}
