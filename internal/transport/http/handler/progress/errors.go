// Package progress defines custom errors for the progress HTTP handlers.
package progress

import "errors"

var (
	// ErrInvalidProgressStatus indicates the provided progress status is not supported.
	ErrInvalidProgressStatus = errors.New("invalid progress status")

	// ErrInternalServer indicates an unexpected server-side error occurred.
	ErrInternalServer = errors.New("internal server error")

	// ErrInvalidTopicID indicates the topic ID provided in the request is invalid or missing.
	ErrInvalidTopicID = errors.New("invalid topic id")

	// ErrInvalidQuestionID indicates the question ID provided in the request is invalid or missing.
	ErrInvalidQuestionID = errors.New("invalid question id")

	// ErrProgressNotFound indicates that the requested progress resource was not found.
	ErrProgressNotFound = errors.New("progress not found")

	// ErrUnauthorized indicates the request lacks valid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized: missing or invalid token")
)
