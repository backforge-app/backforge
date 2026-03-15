// Package user defines custom errors for user HTTP handlers.
package user

import "errors"

var (
	// ErrUserNotFound indicates the user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUnauthorized indicates the user is not authenticated.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrInternalServer indicates an internal server error.
	ErrInternalServer = errors.New("internal server error")
)
