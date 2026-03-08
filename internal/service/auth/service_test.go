package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/backforge-app/backforge/internal/config"
	"github.com/backforge-app/backforge/internal/domain"
)

// makeValidTelegramInput creates a valid TelegramLoginInput with correct HMAC signature.
// It includes all provided fields and sorts keys alphabetically.
func makeValidTelegramInput(
	botToken string,
	id int64,
	firstName string,
	lastName *string,
	username *string,
	photoURL *string,
) TelegramLoginInput {
	authDate := time.Now().Unix()

	var pairs []string
	pairs = append(pairs, fmt.Sprintf("auth_date=%d", authDate))
	pairs = append(pairs, fmt.Sprintf("id=%d", id))
	if firstName != "" {
		pairs = append(pairs, fmt.Sprintf("first_name=%s", firstName))
	}
	if lastName != nil {
		pairs = append(pairs, fmt.Sprintf("last_name=%s", *lastName))
	}
	if username != nil {
		pairs = append(pairs, fmt.Sprintf("username=%s", *username))
	}
	if photoURL != nil {
		pairs = append(pairs, fmt.Sprintf("photo_url=%s", *photoURL))
	}

	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secret := sha256.Sum256([]byte(botToken))
	h := hmac.New(sha256.New, secret[:])
	h.Write([]byte(dataCheckString))

	return TelegramLoginInput{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		PhotoURL:  photoURL,
		AuthDate:  authDate,
		Hash:      hex.EncodeToString(h.Sum(nil)),
	}
}

func TestAuth_LoginWithTelegram(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := NewMockUserProvider(ctrl)
	mockRefresh := NewMockSessionRepository(ctrl)

	authCfg := &config.Auth{
		Secret:          "secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}

	botToken := "bot-token"

	svc := NewService(mockUsers, mockRefresh, nil, authCfg, botToken)

	ctx := context.Background()
	tgID := int64(12345)
	firstName := "John"
	userID := uuid.New()

	domainUser := &domain.User{
		ID:         userID,
		TelegramID: tgID,
		FirstName:  firstName,
	}

	t.Run("Success - minimal input", func(t *testing.T) {
		input := makeValidTelegramInput(botToken, tgID, firstName, nil, nil, nil)

		mockUsers.EXPECT().
			GetOrCreateByTelegramID(ctx, gomock.Any()).
			Return(domainUser, nil)

		mockRefresh.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		access, refresh, err := svc.LoginWithTelegram(ctx, input)
		require.NoError(t, err)
		require.NotEmpty(t, access)
		require.NotEmpty(t, refresh)
	})

	t.Run("Success - with optional fields", func(t *testing.T) {
		lastName := "Doe"
		username := "john_doe"
		photoURL := "https://t.me/i/userpic/320/abc.jpg"

		input := makeValidTelegramInput(botToken, tgID, firstName, &lastName, &username, &photoURL)

		mockUsers.EXPECT().
			GetOrCreateByTelegramID(ctx, gomock.Any()).
			Return(domainUser, nil)

		mockRefresh.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil)

		access, refresh, err := svc.LoginWithTelegram(ctx, input)
		require.NoError(t, err)
		require.NotEmpty(t, access)
		require.NotEmpty(t, refresh)
	})

	t.Run("Fail - invalid hash", func(t *testing.T) {
		input := makeValidTelegramInput(botToken, tgID, firstName, nil, nil, nil)
		input.Hash = "wronghash123"

		access, refresh, err := svc.LoginWithTelegram(ctx, input)
		assert.ErrorIs(t, err, ErrInvalidTelegramAuth)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})

	t.Run("Fail - expired auth date", func(t *testing.T) {
		oldAuthDate := time.Now().Add(-30 * time.Hour).Unix()

		var pairs []string
		pairs = append(pairs, fmt.Sprintf("auth_date=%d", oldAuthDate))
		pairs = append(pairs, fmt.Sprintf("id=%d", tgID))
		if firstName != "" {
			pairs = append(pairs, fmt.Sprintf("first_name=%s", firstName))
		}
		sort.Strings(pairs)
		data := strings.Join(pairs, "\n")

		secret := sha256.Sum256([]byte(botToken))
		h := hmac.New(sha256.New, secret[:])
		h.Write([]byte(data))

		input := TelegramLoginInput{
			ID:        tgID,
			FirstName: firstName,
			AuthDate:  oldAuthDate,
			Hash:      hex.EncodeToString(h.Sum(nil)),
		}

		access, refresh, err := svc.LoginWithTelegram(ctx, input)
		require.ErrorIs(t, err, ErrTelegramAuthExpired)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})

	t.Run("Fail - user creation error", func(t *testing.T) {
		input := makeValidTelegramInput(botToken, tgID, firstName, nil, nil, nil)

		mockUsers.EXPECT().
			GetOrCreateByTelegramID(ctx, gomock.Any()).
			Return(nil, assert.AnError)

		access, refresh, err := svc.LoginWithTelegram(ctx, input)
		assert.Error(t, err)
		assert.Empty(t, access)
		assert.Empty(t, refresh)
	})
}

func TestAuth_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsers := NewMockUserProvider(ctrl)
	mockRefresh := NewMockSessionRepository(ctrl)
	mockTx := NewMockTransactor(ctrl)

	authCfg := &config.Auth{
		Secret:          "secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}

	svc := NewService(mockUsers, mockRefresh, mockTx, authCfg, "bot-token")

	ctx := context.Background()
	userID := uuid.New()
	oldToken := "old-token"
	rt := &domain.Session{
		Token:     oldToken,
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
		Revoked:   false,
	}
	domainUser := &domain.User{ID: userID}

	t.Run("Success", func(t *testing.T) {
		mockTx.EXPECT().WithinTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		)

		mockRefresh.EXPECT().GetByToken(ctx, oldToken).Return(rt, nil)
		mockUsers.EXPECT().GetByID(ctx, userID).Return(domainUser, nil)
		mockRefresh.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRefresh.EXPECT().Revoke(ctx, oldToken).Return(nil)

		access, refresh, err := svc.Refresh(ctx, oldToken)
		require.NoError(t, err)
		assert.NotEmpty(t, access)
		assert.NotEmpty(t, refresh)
	})

	t.Run("Fail_Revoked", func(t *testing.T) {
		rtRevoked := *rt
		rtRevoked.Revoked = true

		mockTx.EXPECT().WithinTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRefresh.EXPECT().GetByToken(ctx, oldToken).Return(&rtRevoked, nil)

		_, _, err := svc.Refresh(ctx, oldToken)
		assert.ErrorIs(t, err, ErrRefreshTokenRevoked)
	})

	t.Run("Fail_Expired", func(t *testing.T) {
		rtExpired := *rt
		rtExpired.ExpiresAt = time.Now().Add(-time.Hour)

		mockTx.EXPECT().WithinTx(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		)
		mockRefresh.EXPECT().GetByToken(ctx, oldToken).Return(&rtExpired, nil)

		_, _, err := svc.Refresh(ctx, oldToken)
		assert.ErrorIs(t, err, ErrRefreshTokenInvalid)
	})
}

func TestAuth_validateTelegramAuth(t *testing.T) {
	botToken := "test-bot-token" //nolint:gosec
	svc := &Service{botToken: botToken}

	t.Run("valid minimal input", func(t *testing.T) {
		input := makeValidTelegramInput(botToken, 12345, "Alice", nil, nil, nil)
		err := svc.validateTelegramAuth(input)
		require.NoError(t, err)
	})

	t.Run("valid with all optional fields", func(t *testing.T) {
		last := "Smith"
		user := "alice_smith"
		photo := "https://t.me/i/userpic/320/abc.jpg"
		input := makeValidTelegramInput(botToken, 12345, "Alice", &last, &user, &photo)
		err := svc.validateTelegramAuth(input)
		require.NoError(t, err)
	})

	t.Run("invalid hash", func(t *testing.T) {
		input := makeValidTelegramInput(botToken, 12345, "Bob", nil, nil, nil)
		input.Hash = "wronghash123"
		err := svc.validateTelegramAuth(input)
		require.ErrorIs(t, err, ErrInvalidTelegramAuth)
	})

	t.Run("expired auth_date", func(t *testing.T) {
		oldDate := time.Now().Add(-30 * time.Hour).Unix()

		// Correctly compute the hash for this old date
		var pairs []string
		pairs = append(pairs, fmt.Sprintf("auth_date=%d", oldDate))
		pairs = append(pairs, fmt.Sprintf("id=%d", 999))
		pairs = append(pairs, fmt.Sprintf("first_name=%s", "Test"))
		sort.Strings(pairs)
		data := strings.Join(pairs, "\n")

		secret := sha256.Sum256([]byte(botToken))
		h := hmac.New(sha256.New, secret[:])
		h.Write([]byte(data))

		input := TelegramLoginInput{
			ID:        999,
			FirstName: "Test",
			AuthDate:  oldDate,
			Hash:      hex.EncodeToString(h.Sum(nil)),
		}

		err := svc.validateTelegramAuth(input)
		require.ErrorIs(t, err, ErrTelegramAuthExpired)
	})

	t.Run("empty first_name is allowed (Telegram can send empty)", func(t *testing.T) {
		input := makeValidTelegramInput(botToken, 12345, "", nil, nil, nil)
		err := svc.validateTelegramAuth(input)
		require.NoError(t, err) // should pass, as field is omitted in data-check-string
	})
}
