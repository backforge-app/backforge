// Package analytics provides HTTP request and response DTOs for analytics handlers.
package analytics

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// OverallProgressResponse represents the JSON payload for aggregated progress statistics.
type OverallProgressResponse struct {
	Total   int `json:"total"`
	Known   int `json:"known"`
	Learned int `json:"learned"`
	Skipped int `json:"skipped"`
	New     int `json:"new"`
}

// TopicProgressPercentResponse represents the JSON payload for topic completion percentages.
type TopicProgressPercentResponse struct {
	TopicID   uuid.UUID `json:"topic_id"`
	Completed int       `json:"completed"`
	Total     int       `json:"total"`
	Percent   float64   `json:"percent"`
}

// toOverallProgressResponse converts a domain.OverallProgress to OverallProgressResponse DTO.
func toOverallProgressResponse(p *domain.OverallProgress) OverallProgressResponse {
	if p == nil {
		return OverallProgressResponse{}
	}

	return OverallProgressResponse{
		Total:   p.Total,
		Known:   p.Known,
		Learned: p.Learned,
		Skipped: p.Skipped,
		New:     p.New,
	}
}

// toTopicProgressPercentResponse converts a domain.TopicProgressPercent to TopicProgressPercentResponse DTO.
func toTopicProgressPercentResponse(p *domain.TopicProgressPercent) TopicProgressPercentResponse {
	if p == nil {
		return TopicProgressPercentResponse{}
	}

	return TopicProgressPercentResponse{
		TopicID:   p.TopicID,
		Completed: p.Completed,
		Total:     p.Total,
		Percent:   p.Percent,
	}
}
