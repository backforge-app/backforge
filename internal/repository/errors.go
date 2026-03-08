// Package repository provides the repository layer for accessing database entities.
// It includes PostgreSQL transaction handling, repository-level and repository
// implementations for entities like users, sessions, questions, etc.
package repository

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

	// ErrSessionNotFound is returned when the provided session does not exist
	// in the database or has been deleted.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionAlreadyExists is returned when attempting to create a session
	// that violates the unique constraint on the token field.
	ErrSessionAlreadyExists = errors.New("session already exists")

	// ErrQuestionNotFound is returned when a question is not found in the database.
	ErrQuestionNotFound = errors.New("question not found")

	// ErrQuestionAlreadyExists is returned when attempting to create a question
	// that violates unique constraints (if any).
	ErrQuestionAlreadyExists = errors.New("question already exists")
)
