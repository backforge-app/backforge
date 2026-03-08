// Package question implements the Question application service.
//
// It contains the business logic for managing questions, including
// service methods, input DTOs, service-level errors,
// repository interfaces and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
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
