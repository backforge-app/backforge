// Package tag defines custom errors for tag HTTP handlers.
package tag

import "errors"

var (
	// ErrTagInvalidData indicates the provided tag data is invalid.
	ErrTagInvalidData = errors.New("invalid tag data provided")

	// ErrTagNotFound indicates the requested tag does not exist.
	ErrTagNotFound = errors.New("tag not found")

	// ErrTagAlreadyExists indicates a tag with the same name already exists.
	ErrTagAlreadyExists = errors.New("tag already exists")

	// ErrTagInvalidID indicates the provided tag ID is not a valid UUID.
	ErrTagInvalidID = errors.New("invalid tag id format")

	// ErrInternalServer indicates an internal server error.
	ErrInternalServer = errors.New("internal server error")
)
