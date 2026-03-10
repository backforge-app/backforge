// Package progress implements the application service layer for managing
// user progress across questions and topics.
//
// It contains business logic, input DTOs, service-level errors,
// and repository interfaces that coordinate domain entities with
// the persistence layer.
//
// The package is responsible for:
//   - tracking user progress on individual questions (known, learned, skipped)
//   - maintaining the current position of a user within a topic
//   - aggregating progress data for UI consumption (e.g., progress bars)
//   - resetting progress for a topic when requested by the user
//
// It orchestrates operations between the UserQuestionProgress and
// UserTopicProgress repositories to keep question status and topic
// position consistent.
package progress

import "errors"

var (
	// ErrInvalidProgressStatus is returned when attempting to set a progress status
	// that is not defined in the progress_status enum.
	ErrInvalidProgressStatus = errors.New("invalid progress status")

	// ErrTopicProgressNotFound is returned when a topic progress entry for the user
	// does not exist in the database.
	ErrTopicProgressNotFound = errors.New("topic progress not found")

	// ErrQuestionProgressNotFound is returned when a question progress entry
	// for the user does not exist in the database.
	ErrQuestionProgressNotFound = errors.New("question progress not found")
)
