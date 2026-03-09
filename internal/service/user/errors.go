// Package user implements the application service layer for user management.
//
// It contains business logic for user creation, updates, retrieval,
// service-level errors, input DTOs (in other files), and coordinates
// domain entities with the persistence layer.
package user

import "errors"

var (
	// ErrUserNotFound is returned when the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserTelegramIDTaken is returned when the Telegram user ID is already in use.
	ErrUserTelegramIDTaken = errors.New("telegram user ID already taken")

	// ErrUserInvalidRole is returned when an invalid user role is provided.
	ErrUserInvalidRole = errors.New("invalid user role")

	// ErrUserInvalidData is returned when invalid user data is provided.
	ErrUserInvalidData = errors.New("invalid user data provided")
)
