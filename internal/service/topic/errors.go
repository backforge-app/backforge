// Package topic implements the application service layer for topic management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package topic

import "errors"

var (
	// ErrTopicNotFound is returned when the requested topic does not exist.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrTopicAlreadyExists is returned when attempting to create a topic
	// that violates unique constraints.
	ErrTopicAlreadyExists = errors.New("topic already exists")

	// ErrTopicInvalidData is returned when the provided input data is invalid.
	ErrTopicInvalidData = errors.New("invalid topic data provided")
)
