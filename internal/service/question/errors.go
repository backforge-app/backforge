// Package question implements the application service layer for question management.
//
// It contains business logic, input DTOs, service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package question

import "errors"

var (
	// ErrQuestionNotFound is returned when the requested question does not exist.
	ErrQuestionNotFound = errors.New("question not found")

	// ErrQuestionAlreadyExists is returned when attempting to create a question
	// that violates unique constraints.
	ErrQuestionAlreadyExists = errors.New("question already exists")

	// ErrQuestionInvalidData is returned when the provided input data is invalid.
	ErrQuestionInvalidData = errors.New("invalid question data provided")
)
