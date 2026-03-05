// Package domain defines core business entities and types used across the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user.
type UserRole string

const (
	// UserRoleUser is the default role for regular users.
	UserRoleUser UserRole = "user"

	// UserRoleAdmin is the role for administrators.
	UserRoleAdmin UserRole = "admin"
)

// User represents a user entity in the system.
type User struct {
	ID uuid.UUID

	TgUserID    int64
	TgUsername  *string
	TgFirstName string
	TgLastName  *string

	Role UserRole

	IsPro        bool
	ProGrantedAt *time.Time
	ProType      *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new User instance with the provided details.
func NewUser(
	tgUserID int64,
	firstName string,
	lastName, username *string,
	isPro bool,
) *User {
	now := time.Now().UTC()
	proType := "channel"

	u := &User{
		TgUserID:    tgUserID,
		TgFirstName: firstName,
		TgLastName:  lastName,
		TgUsername:  username,
		IsPro:       isPro,
	}

	if isPro {
		u.ProGrantedAt = &now
		u.ProType = &proType
	}

	return u
}

// IsAdmin checks if the user has admin privileges.
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}
