// Package topic provides the repository layer for accessing topic entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, and list topics.
package topic

import "errors"

var (
	// ErrTopicNotFound is returned when a topic is not found in the database.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrTopicAlreadyExists is returned when attempting to create a topic
	// that violates unique constraints (e.g., slug already exists).
	ErrTopicAlreadyExists = errors.New("topic already exists")
)
