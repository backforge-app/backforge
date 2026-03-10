// Package tag implements the application service layer for tag management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package tag

import "errors"

var (
	// ErrTagNotFound is returned when the requested tag does not exist.
	ErrTagNotFound = errors.New("tag not found")

	// ErrTagAlreadyExists is returned when attempting to create a tag
	// that violates unique constraints.
	ErrTagAlreadyExists = errors.New("tag already exists")

	// ErrTagInvalidData is returned when the provided input data is invalid.
	ErrTagInvalidData = errors.New("invalid tag data provided")
)
