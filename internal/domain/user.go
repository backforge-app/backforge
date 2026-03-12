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
	ID         uuid.UUID
	TelegramID int64

	FirstName string
	LastName  *string
	Username  *string
	PhotoURL  *string

	Role UserRole

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new User instance with the provided details.
func NewUser(
	telegramID int64,
	firstName string,
	lastName, username, photoURL *string,
) *User {
	u := &User{
		TelegramID: telegramID,
		FirstName:  firstName,
		LastName:   lastName,
		Username:   username,
		PhotoURL:   photoURL,
	}

	return u
}

// IsAdmin checks if the user has admin privileges.
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}
