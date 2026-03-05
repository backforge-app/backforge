// Package service contains application business logic, service-level errors,
// repository interfaces, and input DTOs used to communicate between
// delivery layers (handlers) and the domain layer.
package service

import "errors"

var (
	// ErrUserNotFound is returned when the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserTgUserIDTaken is returned when the Telegram user ID is already in use.
	ErrUserTgUserIDTaken = errors.New("telegram user ID already taken")

	// ErrUserInvalidRole is returned when an invalid user role is provided.
	ErrUserInvalidRole = errors.New("invalid user role")

	// ErrUserInvalidData is returned when invalid user data is provided.
	ErrUserInvalidData = errors.New("invalid user data provided")
)
