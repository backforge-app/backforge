// Package progress provides the repository layer for accessing user question progress entities.
//
// It includes PostgreSQL operations, transaction handling, repository-level errors,
// and methods to create, read, update, and delete progress entries.
package progress

import "errors"

var (
	// ErrQuestionProgressNotFound is returned when a user's progress for a question is not found.
	ErrQuestionProgressNotFound = errors.New("user question progress not found")

	// ErrTopicProgressNotFound is returned when a user's topic progress is not found.
	ErrTopicProgressNotFound = errors.New("user topic progress not found")
)
