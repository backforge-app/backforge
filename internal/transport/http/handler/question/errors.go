// Package question defines custom errors for question HTTP handlers.
package question

import "errors"

var (
	// ErrQuestionNotFound indicates the requested question does not exist.
	ErrQuestionNotFound = errors.New("question not found")

	// ErrQuestionAlreadyExists indicates a question with the same slug or title already exists.
	ErrQuestionAlreadyExists = errors.New("question already exists")

	// ErrQuestionInvalidID indicates the provided question ID is invalid.
	ErrQuestionInvalidID = errors.New("invalid question ID")

	// ErrQuestionInvalidData indicates the provided data is invalid.
	ErrQuestionInvalidData = errors.New("invalid question data")

	// ErrInternalServer indicates an internal server error.
	ErrInternalServer = errors.New("internal server error")
)
