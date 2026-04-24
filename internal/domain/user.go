// Package domain defines core business entities and types used across the application.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordTooLong is returned when a password exceeds the bcrypt 72-byte limit.
	ErrPasswordTooLong = errors.New("password exceeds maximum length of 72 bytes")

	// ErrPasswordHashingFailed is returned when the system fails to securely hash the password.
	ErrPasswordHashingFailed = errors.New("failed to hash password")
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

	Email           string
	PasswordHash    *string // nullable because OAuth users might not have a password
	IsEmailVerified bool

	FirstName string
	LastName  *string
	Username  *string
	PhotoURL  *string

	Role UserRole

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUserWithPassword creates a new User instance for standard email/password registration.
func NewUserWithPassword(
	email, password, firstName string,
	lastName, username, photoURL *string,
) (*User, error) {
	u := &User{
		Email:           email,
		FirstName:       firstName,
		LastName:        lastName,
		Username:        username,
		PhotoURL:        photoURL,
		Role:            UserRoleUser,
		IsEmailVerified: false,
	}

	if err := u.SetPassword(password); err != nil {
		return nil, err
	}

	return u, nil
}

// NewUserFromOAuth creates a new User instance for OAuth registration (e.g., GitHub).
func NewUserFromOAuth(
	email, firstName string,
	lastName, username, photoURL *string,
	isEmailVerified bool,
) *User {
	return &User{
		Email:           email,
		PasswordHash:    nil, // no password for OAuth by default
		FirstName:       firstName,
		LastName:        lastName,
		Username:        username,
		PhotoURL:        photoURL,
		Role:            UserRoleUser,
		IsEmailVerified: isEmailVerified,
	}
}

// SetPassword validates, hashes the provided password, and updates the user's PasswordHash.
// It enforces the 72-byte maximum length limit required by the bcrypt algorithm.
func (u *User) SetPassword(password string) error {
	passwordBytes := []byte(password)
	if len(passwordBytes) > 72 {
		return ErrPasswordTooLong
	}

	hash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return ErrPasswordHashingFailed
	}

	hashStr := string(hash)
	u.PasswordHash = &hashStr
	return nil
}

// CheckPassword verifies if the provided password matches the user's stored hash.
// Returns false if the user has no password (e.g., registered via OAuth only).
func (u *User) CheckPassword(password string) bool {
	if u.PasswordHash == nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password))
	return err == nil
}

// IsAdmin checks if the user has admin privileges.
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// HasPassword checks if the user has set up a password login.
func (u *User) HasPassword() bool {
	return u.PasswordHash != nil
}
