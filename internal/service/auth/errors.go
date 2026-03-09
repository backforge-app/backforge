// Package auth implements authentication and session management logic.
//
// It supports Telegram-based authentication, JWT issuance, refresh token rotation,
// session persistence, and revocation.
package auth

import "errors"

var (
	// ErrInvalidTelegramAuth is returned when Telegram authentication data
	// fails validation (invalid hash, expired data, etc.).
	ErrInvalidTelegramAuth = errors.New("invalid telegram auth data")

	// ErrTelegramAuthExpired is returned when the Telegram auth data is too old
	// (auth_date is older than 24 hours), to prevent replay attacks.
	ErrTelegramAuthExpired = errors.New("telegram auth data expired")

	// ErrRefreshTokenInvalid is returned when the refresh token is malformed,
	// expired, or otherwise unusable.
	ErrRefreshTokenInvalid = errors.New("invalid or expired refresh token")

	// ErrRefreshTokenRevoked is returned when attempting to use a refresh token
	// that has been explicitly revoked.
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")

	// ErrRefreshTokenAlreadyExists is returned when a newly generated refresh token
	// collides with an existing token (extremely rare, but possible).
	ErrRefreshTokenAlreadyExists = errors.New("refresh token already exists")
)
