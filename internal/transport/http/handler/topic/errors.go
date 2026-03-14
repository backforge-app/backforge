// Package topic defines custom errors for topic HTTP handlers.
package topic

import "errors"

var (
	// ErrTopicInvalidData indicates the provided topic data is invalid.
	ErrTopicInvalidData = errors.New("invalid topic data provided")

	// ErrTopicNotFound indicates the requested topic does not exist.
	ErrTopicNotFound = errors.New("topic not found")

	// ErrTopicAlreadyExists indicates a topic with the given unique constraints (e.g., slug) already exists.
	ErrTopicAlreadyExists = errors.New("topic already exists")

	// ErrTopicInvalidID indicates the provided topic ID is not a valid UUID.
	ErrTopicInvalidID = errors.New("invalid topic id format")

	// ErrInternalServer indicates an unexpected internal server error.
	ErrInternalServer = errors.New("internal server error")
)
