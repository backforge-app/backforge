// Package auth defines custom errors for authentication HTTP handlers.
package auth

import "errors"

var (
	// ErrInvalidCredentials indicates the provided credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInternalServer indicates an internal server error.
	ErrInternalServer = errors.New("internal server error")

	// ErrInvalidRequest indicates an bad request error.
	ErrInvalidRequest = errors.New("invalid request")
)
