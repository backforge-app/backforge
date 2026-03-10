// Package tag provides the repository layer for accessing tag entities.
//
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, list, and manage tags.
package tag

import "errors"

var (
	// ErrTagNotFound is returned when a tag is not found in the database.
	ErrTagNotFound = errors.New("tag not found")

	// ErrTagAlreadyExists is returned when attempting to create a tag
	// that violates unique constraints (for example, duplicate name).
	ErrTagAlreadyExists = errors.New("tag already exists")
)
