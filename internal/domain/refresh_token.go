// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token entity.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewRefreshToken creates a new refresh token.
func NewRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}
