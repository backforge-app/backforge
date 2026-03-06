// Package postgres provides PostgreSQL infrastructure components.
// It includes connection pool setup, transaction handling, repository-level errors
// and repository implementations for accessing database entities like users.
package postgres

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

	// ErrRefreshTokenNotFound is returned when the provided refresh token does not exist
	// in the database or has been deleted.
	ErrRefreshTokenNotFound = errors.New("refresh token not found")

	// ErrRefreshTokenAlreadyExists is returned when attempting to create a refresh token
	// that violates the unique constraint on the token field.
	ErrRefreshTokenAlreadyExists = errors.New("refresh token already exists")
)
