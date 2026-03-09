// Package question provides the repository layer for accessing question entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, list, and manage questions.
package question

import "errors"

var (
	// ErrQuestionNotFound is returned when a question is not found in the database.
	ErrQuestionNotFound = errors.New("question not found")

	// ErrQuestionAlreadyExists is returned when attempting to create a question
	// that violates unique constraints (if any).
	ErrQuestionAlreadyExists = errors.New("question already exists")
)
