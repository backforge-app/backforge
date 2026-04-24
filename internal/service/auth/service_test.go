package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/backforge-app/backforge/internal/config"
	"github.com/backforge-app/backforge/internal/domain"
	oauthrepo "github.com/backforge-app/backforge/internal/repository/oauthconnection"
	tokenrepo "github.com/backforge-app/backforge/internal/repository/verificationtoken"
	"github.com/backforge-app/backforge/internal/service/user"
)

// setupTxMock is a helper to automatically execute the function passed to WithinTx.
func setupTxMock(transactor *MockTransactor, ctx context.Context) {
	transactor.EXPECT().
		WithinTx(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

// makeUserWithPassword is a helper to create a domain user with a valid bcrypt password hash.
func makeUserWithPassword(t *testing.T, id uuid.UUID, email, plainPassword string, isVerified bool) *domain.User {
	u := &domain.User{
		ID:              id,
		Email:           email,
		IsEmailVerified: isVerified,
	}
	err := u.SetPassword(plainPassword)
	require.NoError(t, err)
	return u
}

// ServiceMocks holds all mock dependencies for easy access in tests.
type ServiceMocks struct {
	User        *MockUserService
	Session     *MockSessionRepository
	OAuth       *MockOAuthConnectionRepository
	Token       *MockVerificationTokenRepository
	Tx          *MockTransactor
	Email       *MockEmailSender
	OAuthClient *MockOAuthClient // Заменили GitHub на универсальный OAuthClient
}

func setupService(ctrl *gomock.Controller) (*Service, ServiceMocks) {
	m := ServiceMocks{
		User:        NewMockUserService(ctrl),
		Session:     NewMockSessionRepository(ctrl),
		OAuth:       NewMockOAuthConnectionRepository(ctrl),
		Token:       NewMockVerificationTokenRepository(ctrl),
		Tx:          NewMockTransactor(ctrl),
		Email:       NewMockEmailSender(ctrl),
		OAuthClient: NewMockOAuthClient(ctrl), // Инициализация нового мока
	}

	cfg := &config.Auth{
		Secret:               "test-secret",
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      24 * time.Hour,
		EmailVerificationTTL: 24 * time.Hour,
		PasswordResetTTL:     1 * time.Hour,
	}

	svc := NewService(m.User, m.Session, m.OAuth, m.Token, m.Email, m.OAuthClient, m.Tx, cfg)
	return svc, m
}

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()

	input := RegisterInput{
		Email:     "test@example.com",
		Password:  "securepassword",
		FirstName: "John",
	}

	t.Run("Success", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)
		userID := uuid.New()

		m.User.EXPECT().CreateWithPassword(ctx, gomock.Any()).Return(userID, nil)
		m.Token.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		m.Email.EXPECT().SendVerificationEmail(ctx, input.Email, gomock.Any()).Return(nil)

		err := svc.Register(ctx, input)
		require.NoError(t, err)
	})

	t.Run("Fail - email taken", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		m.User.EXPECT().CreateWithPassword(ctx, gomock.Any()).Return(uuid.Nil, user.ErrUserEmailTaken)

		err := svc.Register(ctx, input)
		require.ErrorIs(t, err, user.ErrUserEmailTaken)
	})
}

func TestService_VerifyEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()
	rawToken := "somerawtoken"
	tokenHash := domain.HashVerificationToken(rawToken)

	validToken := &domain.VerificationToken{
		UserID: uuid.New(),
	}

	t.Run("Success", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		m.Token.EXPECT().GetByHash(ctx, tokenHash, domain.TokenPurposeEmailVerification).Return(validToken, nil)
		m.User.EXPECT().MarkEmailVerified(ctx, validToken.UserID).Return(nil)
		m.Token.EXPECT().Delete(ctx, tokenHash).Return(nil)

		err := svc.VerifyEmail(ctx, rawToken)
		require.NoError(t, err)
	})

	t.Run("Fail - token not found", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		m.Token.EXPECT().GetByHash(ctx, tokenHash, domain.TokenPurposeEmailVerification).Return(nil, tokenrepo.ErrTokenNotFound)

		err := svc.VerifyEmail(ctx, rawToken)
		require.ErrorIs(t, err, ErrInvalidVerificationToken)
	})
}

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()

	input := LoginInput{
		Email:    "test@example.com",
		Password: "correctpassword",
	}
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		u := makeUserWithPassword(t, userID, input.Email, input.Password, true)

		m.User.EXPECT().GetByEmail(ctx, input.Email).Return(u, nil)
		m.Session.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		access, refresh, err := svc.Login(ctx, input)
		require.NoError(t, err)
		assert.NotEmpty(t, access)
		assert.NotEmpty(t, refresh)
	})

	t.Run("Fail - incorrect password", func(t *testing.T) {
		u := makeUserWithPassword(t, userID, input.Email, "differentpassword", true)

		m.User.EXPECT().GetByEmail(ctx, input.Email).Return(u, nil)

		access, refresh, err := svc.Login(ctx, input)
		require.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})

	t.Run("Fail - email not verified", func(t *testing.T) {
		u := makeUserWithPassword(t, userID, input.Email, input.Password, false) // isVerified = false

		m.User.EXPECT().GetByEmail(ctx, input.Email).Return(u, nil)

		access, refresh, err := svc.Login(ctx, input)
		require.ErrorIs(t, err, ErrEmailNotVerified)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})
}

func TestService_RequestPasswordReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()

	email := "test@example.com"
	u := &domain.User{ID: uuid.New(), Email: email}

	t.Run("Success", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		m.User.EXPECT().GetByEmail(ctx, email).Return(u, nil)
		m.Token.EXPECT().DeleteAllForUser(ctx, u.ID, domain.TokenPurposePasswordReset).Return(nil)
		m.Token.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		m.Email.EXPECT().SendPasswordResetEmail(ctx, email, gomock.Any()).Return(nil)

		err := svc.RequestPasswordReset(ctx, email)
		require.NoError(t, err)
	})

	t.Run("Silent Success - user not found", func(t *testing.T) {
		m.User.EXPECT().GetByEmail(ctx, email).Return(nil, user.ErrUserNotFound)

		err := svc.RequestPasswordReset(ctx, email)
		require.NoError(t, err) // Should silently succeed to prevent enumeration
	})
}

func TestService_ResetPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()

	rawToken := "somerawtoken"
	tokenHash := domain.HashVerificationToken(rawToken)
	newPassword := "new_secure_password"

	validToken := &domain.VerificationToken{
		UserID: uuid.New(),
	}

	t.Run("Success", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		m.Token.EXPECT().GetByHash(ctx, tokenHash, domain.TokenPurposePasswordReset).Return(validToken, nil)
		m.User.EXPECT().SetNewPassword(ctx, validToken.UserID, newPassword).Return(nil)
		m.Token.EXPECT().Delete(ctx, tokenHash).Return(nil)

		err := svc.ResetPassword(ctx, rawToken, newPassword)
		require.NoError(t, err)
	})
}

// Заменили GitHub на Yandex
func TestService_LoginWithYandex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()

	code := "yandex_oauth_code"
	profile := &OAuthProfile{ // Используем новую структуру OAuthProfile
		ProviderID: "123456",
		Email:      "yandex@example.com",
		Name:       "Yandex User",
		AvatarURL:  "https://avatar.yandex.net/test",
	}

	t.Run("Success - existing linked user", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)
		userID := uuid.New()
		u := &domain.User{ID: userID, Role: domain.UserRoleUser}

		m.OAuthClient.EXPECT().ExchangeCode(ctx, code).Return(profile, nil)
		m.OAuth.EXPECT().GetByProviderUserID(ctx, domain.OAuthProviderYandex, profile.ProviderID).Return(&domain.OAuthConnection{UserID: userID}, nil)
		m.User.EXPECT().GetByID(ctx, userID).Return(u, nil)
		m.Session.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		access, refresh, err := svc.LoginWithYandex(ctx, code)
		require.NoError(t, err)
		assert.NotEmpty(t, access)
		assert.NotEmpty(t, refresh)
	})

	t.Run("Success - new user creation", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)
		newUserID := uuid.New()
		newUser := &domain.User{ID: newUserID, Role: domain.UserRoleUser}

		m.OAuthClient.EXPECT().ExchangeCode(ctx, code).Return(profile, nil)
		m.OAuth.EXPECT().GetByProviderUserID(ctx, domain.OAuthProviderYandex, profile.ProviderID).Return(nil, oauthrepo.ErrConnectionNotFound)
		m.User.EXPECT().GetByEmail(ctx, profile.Email).Return(nil, user.ErrUserNotFound)
		m.User.EXPECT().CreateOAuthUser(ctx, gomock.Any()).Return(newUserID, nil)
		m.User.EXPECT().GetByID(ctx, newUserID).Return(newUser, nil)
		m.OAuth.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		m.Session.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		access, refresh, err := svc.LoginWithYandex(ctx, code)
		require.NoError(t, err)
		assert.NotEmpty(t, access)
		assert.NotEmpty(t, refresh)
	})
}

func TestService_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()

	oldToken := "old_refresh_token"
	hashed := hashToken(oldToken)
	userID := uuid.New()
	u := &domain.User{ID: userID, Role: domain.UserRoleUser}

	t.Run("Success", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		validSession := &domain.Session{
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
			Revoked:   false,
		}

		m.Session.EXPECT().GetByTokenHash(ctx, hashed).Return(validSession, nil)
		m.User.EXPECT().GetByID(ctx, userID).Return(u, nil)
		m.Session.EXPECT().Revoke(ctx, hashed).Return(nil)
		m.Session.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		access, refresh, err := svc.Refresh(ctx, oldToken)
		require.NoError(t, err)
		assert.NotEmpty(t, access)
		assert.NotEmpty(t, refresh)
	})

	t.Run("Fail - revoked token", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		revokedSession := &domain.Session{
			UserID:    userID,
			ExpiresAt: time.Now().Add(time.Hour),
			Revoked:   true,
		}

		m.Session.EXPECT().GetByTokenHash(ctx, hashed).Return(revokedSession, nil)

		access, refresh, err := svc.Refresh(ctx, oldToken)
		require.ErrorIs(t, err, ErrRefreshTokenRevoked)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})
}

func TestService_ResendVerificationEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, m := setupService(ctrl)
	ctx := context.Background()

	email := "test@example.com"
	u := &domain.User{ID: uuid.New(), Email: email, IsEmailVerified: false}

	t.Run("Success", func(t *testing.T) {
		setupTxMock(m.Tx, ctx)

		m.User.EXPECT().GetByEmail(ctx, email).Return(u, nil)
		m.Token.EXPECT().DeleteAllForUser(ctx, u.ID, domain.TokenPurposeEmailVerification).Return(nil)
		m.Token.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		m.Email.EXPECT().SendVerificationEmail(ctx, email, gomock.Any()).Return(nil)

		err := svc.ResendVerificationEmail(ctx, email)
		require.NoError(t, err)
	})

	t.Run("Silent Success - user not found", func(t *testing.T) {
		m.User.EXPECT().GetByEmail(ctx, email).Return(nil, user.ErrUserNotFound)

		err := svc.ResendVerificationEmail(ctx, email)
		require.NoError(t, err) // Should silently succeed to prevent enumeration
	})

	t.Run("Fail - already verified", func(t *testing.T) {
		verifiedUser := &domain.User{ID: uuid.New(), Email: email, IsEmailVerified: true}
		m.User.EXPECT().GetByEmail(ctx, email).Return(verifiedUser, nil)

		err := svc.ResendVerificationEmail(ctx, email)
		require.ErrorIs(t, err, ErrEmailAlreadyVerified)
	})
}
