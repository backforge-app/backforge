// Package user provides the repository layer for accessing user entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, and manage users.
package user

import "errors"

var (
	// ErrUserNotFound is returned when the requested user does not exist in the database.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserTelegramIDTaken is returned when attempting to create a user with a Telegram user ID
	// that is already associated with another user account.
	ErrUserTelegramIDTaken = errors.New("telegram user ID already taken")

	// ErrUserInvalidRole is returned when a user role value does not match any of the allowed
	// values defined in the user_role enum type.
	ErrUserInvalidRole = errors.New("invalid user role")
)
