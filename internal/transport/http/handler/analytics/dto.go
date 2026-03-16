// Package analytics provides HTTP request and response DTOs for analytics handlers.
package analytics

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// overallProgressResponse represents the JSON payload for aggregated progress statistics.
type overallProgressResponse struct {
	Total   int `json:"total"`
	Known   int `json:"known"`
	Learned int `json:"learned"`
	Skipped int `json:"skipped"`
	New     int `json:"new"`
}

// topicProgressPercentResponse represents the JSON payload for topic completion percentages.
type topicProgressPercentResponse struct {
	TopicID   uuid.UUID `json:"topic_id"`
	Completed int       `json:"completed"`
	Total     int       `json:"total"`
	Percent   float64   `json:"percent"`
}

// toOverallProgressResponse converts a domain.OverallProgress to overallProgressResponse DTO.
func toOverallProgressResponse(p *domain.OverallProgress) overallProgressResponse {
	if p == nil {
		return overallProgressResponse{}
	}

	return overallProgressResponse{
		Total:   p.Total,
		Known:   p.Known,
		Learned: p.Learned,
		Skipped: p.Skipped,
		New:     p.New,
	}
}

// toTopicProgressPercentResponse converts a domain.TopicProgressPercent to topicProgressPercentResponse DTO.
func toTopicProgressPercentResponse(p *domain.TopicProgressPercent) topicProgressPercentResponse {
	if p == nil {
		return topicProgressPercentResponse{}
	}

	return topicProgressPercentResponse{
		TopicID:   p.TopicID,
		Completed: p.Completed,
		Total:     p.Total,
		Percent:   p.Percent,
	}
}
