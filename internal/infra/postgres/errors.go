// Package postgres provides PostgreSQL infrastructure components.
// It includes connection pool setup, transaction handling, repository-level errors
// and repository implementations for accessing database entities like users.
package postgres

import "errors"

var (
	// ErrUserNotFound is returned when the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserTgUserIDTaken is returned when the Telegram user ID is already in use.
	ErrUserTgUserIDTaken = errors.New("telegram user ID already taken")

	// ErrUserInvalidRole is returned when an invalid user role is provided.
	ErrUserInvalidRole = errors.New("invalid user role")
)
