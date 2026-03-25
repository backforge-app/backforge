// Package auth provides the interface to the authentication service for HTTP handlers.
//
//go:generate mockgen -package=auth -destination=mocks.go github.com/backforge-app/backforge/internal/transport/http/handler/auth Service
package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/service/auth"
)

// Service defines the interface that HTTP handlers use to perform authentication operations.
type Service interface {
	// LoginWithTelegram authenticates the user via Telegram login.
	LoginWithTelegram(ctx context.Context, input auth.TelegramLoginInput) (accessToken, refreshToken string, err error)

	// Refresh exchanges a valid refresh token for new access and refresh tokens.
	Refresh(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)

	// DevLogin issues tokens for an existing user without Telegram verification.
	// Intended for development only.
	DevLogin(ctx context.Context, userID uuid.UUID) (accessToken, refreshToken string, err error)
}
