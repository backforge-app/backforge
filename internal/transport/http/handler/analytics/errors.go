// Package analytics defines custom errors for analytics HTTP handlers.
package analytics

import "errors"

var (
	// ErrInvalidUserID indicates the provided user ID in the request or context is invalid.
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrInternalServer indicates an unexpected internal server error occurred.
	ErrInternalServer = errors.New("internal server error")

	// ErrUnauthorized indicates the request lacks valid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized: missing or invalid token")
)
