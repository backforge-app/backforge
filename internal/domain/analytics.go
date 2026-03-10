// Package domain defines core business entities and types used across the application.
package domain

import "github.com/google/uuid"

// OverallProgress represents aggregated user progress across all questions.
//
// It is used to populate the dashboard progress cards.
type OverallProgress struct {
	Total   int
	Known   int
	Learned int
	Skipped int
	New     int
}

// TopicProgressPercent represents completion percentage for a topic.
//
// A topic is considered completed when questions are marked as
// ProgressStatusKnown or ProgressStatusLearned.
type TopicProgressPercent struct {
	TopicID   uuid.UUID
	Completed int
	Total     int
	Percent   float64
}
