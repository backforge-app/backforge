// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a session entity.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSession creates a new session.
func NewSession(userID uuid.UUID, token string, expiresAt time.Time) *Session {
	return &Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}
