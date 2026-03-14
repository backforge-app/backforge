// Package progress implements HTTP handlers for tracking user learning progress.
package progress

import "github.com/google/uuid"

// markRequest represents the payload to advance progress.
type markRequest struct {
	TopicID    uuid.UUID `json:"topic_id" validate:"required"`
	QuestionID uuid.UUID `json:"question_id" validate:"required"`
}

// aggregateResponse returns the stats for a topic.
type aggregateResponse struct {
	Known           int `json:"known"`
	Learned         int `json:"learned"`
	Skipped         int `json:"skipped"`
	New             int `json:"new"`
	CurrentPosition int `json:"current_position"`
}
