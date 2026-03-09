// Package auth implements authentication logic for the application.
//
// It handles Telegram-based login, JWT access token generation,
// refresh token issuance and rotation, and validation of Telegram auth data.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/config"
	"github.com/backforge-app/backforge/internal/domain"
	repository "github.com/backforge-app/backforge/internal/repository/session"
	"github.com/backforge-app/backforge/internal/service/user"
)

// Service provides authentication operations: login via Telegram, token refresh,
// and related token generation/validation.
type Service struct {
	users       UserProvider
	sessionRepo SessionRepository
	transactor  Transactor
	authConfig  *config.Auth
	botToken    string
}

// NewService creates a new authentication service instance.
func NewService(
	users UserProvider,
	sessionRepo SessionRepository,
	transactor Transactor,
	authConfig *config.Auth,
	botToken string,
) *Service {
	return &Service{
		users:       users,
		sessionRepo: sessionRepo,
		transactor:  transactor,
		authConfig:  authConfig,
		botToken:    botToken,
	}
}

// LoginWithTelegram authenticates a user via Telegram Login data and returns
// a new access token (JWT) and refresh token.
//
// If the user does not exist, it will be created automatically.
func (s *Service) LoginWithTelegram(
	ctx context.Context,
	input TelegramLoginInput,
) (accessToken, refreshToken string, err error) {
	if err := s.validateTelegramAuth(input); err != nil {
		return "", "", err
	}

	userInput := user.CreateInput{
		TelegramID: input.ID,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Username:   input.Username,
		IsPro:      false,
	}

	u, err := s.users.GetOrCreateByTelegramID(ctx, userInput)
	if err != nil {
		return "", "", fmt.Errorf("get or create user: %w", err)
	}

	accessToken, err = s.generateAccessToken(u)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = s.generateRefreshToken(ctx, u.ID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// Refresh exchanges a valid refresh token for a new access token and a new refresh token
// (token rotation). The old refresh token is revoked.
func (s *Service) Refresh(ctx context.Context, oldToken string) (newAccessToken, newRefreshToken string, err error) {
	var accessToken string
	var refreshToken string

	err = s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		session, err := s.sessionRepo.GetByToken(txCtx, oldToken)
		if err != nil {
			if errors.Is(err, repository.ErrSessionNotFound) {
				return ErrRefreshTokenInvalid
			}
			return fmt.Errorf("get refresh token: %w", err)
		}

		if session.Revoked {
			return ErrRefreshTokenRevoked
		}

		if time.Now().After(session.ExpiresAt) {
			return ErrRefreshTokenInvalid
		}

		u, err := s.users.GetByID(txCtx, session.UserID)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		accessToken, err = s.generateAccessToken(u)
		if err != nil {
			return fmt.Errorf("generate access token: %w", err)
		}

		refreshToken, err = s.generateRefreshToken(txCtx, u.ID)
		if err != nil {
			if errors.Is(err, ErrRefreshTokenAlreadyExists) {
				return err
			}
			return fmt.Errorf("generate refresh token: %w", err)
		}

		if err := s.sessionRepo.Revoke(txCtx, oldToken); err != nil {
			return fmt.Errorf("revoke old refresh token: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken creates a new JWT access token for the given user.
func (s *Service) generateAccessToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":    user.ID.String(),
		"role":   user.Role,
		"is_pro": user.IsPro,
		"exp":    time.Now().Add(s.authConfig.AccessTokenTTL).Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(s.authConfig.Secret))
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signed, nil
}

// generateRefreshToken generates a new opaque refresh token, saves it to the database
// and returns the token string.
func (s *Service) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Using two UUIDs concatenated gives enough entropy and is URL-safe
	token := uuid.NewString() + uuid.NewString()

	expiresAt := time.Now().Add(s.authConfig.RefreshTokenTTL)

	session := domain.NewSession(userID, token, expiresAt)

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		if errors.Is(err, repository.ErrSessionAlreadyExists) {
			return "", ErrRefreshTokenAlreadyExists
		}
		return "", fmt.Errorf("create session: %w", err)
	}

	return token, nil
}

// validateTelegramAuth verifies the Telegram Login data signature using HMAC-SHA256
// with the bot token as secret key. It follows the exact specification from:
// https://core.telegram.org/widgets/login#checking-authorization
func (s *Service) validateTelegramAuth(input TelegramLoginInput) error {
	authTime := time.Unix(input.AuthDate, 0)
	if input.AuthDate <= 0 || time.Since(authTime) > 24*time.Hour {
		return ErrTelegramAuthExpired
	}

	var pairs []string

	// Required fields (always present according to Telegram docs)
	pairs = append(pairs, fmt.Sprintf("auth_date=%d", input.AuthDate))
	pairs = append(pairs, fmt.Sprintf("id=%d", input.ID))

	// Optional fields — include only if the pointer is non-nil (field was sent by Telegram)
	if input.FirstName != "" {
		pairs = append(pairs, fmt.Sprintf("first_name=%s", input.FirstName))
	}
	if input.LastName != nil {
		pairs = append(pairs, fmt.Sprintf("last_name=%s", *input.LastName))
	}
	if input.Username != nil {
		pairs = append(pairs, fmt.Sprintf("username=%s", *input.Username))
	}
	if input.PhotoURL != nil {
		pairs = append(pairs, fmt.Sprintf("photo_url=%s", *input.PhotoURL))
	}

	// Sort alphabetically by key (mandatory per Telegram specification!)
	sort.Strings(pairs)

	// Join with newline separator to form the data-check string
	dataCheckString := strings.Join(pairs, "\n")

	// Compute expected HMAC-SHA256 hash
	secret := sha256.Sum256([]byte(s.botToken))
	h := hmac.New(sha256.New, secret[:])
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// Constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(expectedHash), []byte(input.Hash)) {
		return ErrInvalidTelegramAuth
	}

	return nil
}
