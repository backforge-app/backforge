// Package user implements the User application service.
//
// It contains the business logic for managing users, including
// service methods, input DTOs, service-level errors,
// repository interfaces and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
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
