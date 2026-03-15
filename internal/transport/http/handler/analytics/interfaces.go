// Package analytics provides the interface to the analytics service for HTTP handlers.
//
//go:generate mockgen -package=analytics -destination=mocks.go github.com/backforge-app/backforge/internal/transport/http/handler/analytics Service
package analytics

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// Service defines the interface that HTTP handlers use to perform analytics operations.
type Service interface {
	// GetOverallProgress returns aggregated progress statistics for dashboard cards.
	GetOverallProgress(ctx context.Context, userID uuid.UUID) (*domain.OverallProgress, error)

	// GetProgressByTopicPercent returns completion percentages for each topic.
	GetProgressByTopicPercent(ctx context.Context, userID uuid.UUID) ([]*domain.TopicProgressPercent, error)

	// ResetAllProgress removes all stored progress for the user.
	ResetAllProgress(ctx context.Context, userID uuid.UUID) error
}
