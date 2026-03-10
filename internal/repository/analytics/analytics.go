// Package analytics provides the repository layer for retrieving
// aggregated analytics data about user learning progress.
//
// It contains optimized SQL queries that calculate statistics
// directly in the database to avoid loading large datasets into memory.
package analytics

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// Repository provides analytics queries for user progress.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new analytics repository.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// GetOverallProgress returns aggregated question progress statistics for a user.
func (r *Repository) GetOverallProgress(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.OverallProgress, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT
	    	COUNT(*) as total,
	    	COUNT(*) FILTER (WHERE up.status = 'known') as known,
	    	COUNT(*) FILTER (WHERE up.status = 'learned') as learned,
	    	COUNT(*) FILTER (WHERE up.status = 'skipped') as skipped,
	    	COUNT(*) FILTER (WHERE up.status = 'new') as new
		FROM questions q
		LEFT JOIN user_question_progress up
		    ON up.question_id = q.id AND up.user_id = $1
	`

	var result domain.OverallProgress

	err := db.QueryRow(ctx, q, userID).Scan(
		&result.Total,
		&result.Known,
		&result.Learned,
		&result.Skipped,
		&result.New,
	)
	if err != nil {
		return nil, fmt.Errorf("get overall progress: %w", err)
	}

	return &result, nil
}

// GetTopicProgressPercent returns completion percentage for each topic.
func (r *Repository) GetTopicProgressPercent(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.TopicProgressPercent, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT
		    q.topic_id,
		    COUNT(*) as total,
		    COUNT(*) FILTER (
		        WHERE up.status IN ('known','learned')
		    ) as completed
		FROM questions q
		LEFT JOIN user_question_progress up
		    ON up.question_id = q.id AND up.user_id = $1
		GROUP BY q.topic_id
	`

	rows, err := db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get topic progress percent: %w", err)
	}
	defer rows.Close()

	var result []*domain.TopicProgressPercent

	for rows.Next() {
		var (
			topicID   uuid.UUID
			total     int
			completed int
		)

		if err := rows.Scan(
			&topicID,
			&total,
			&completed,
		); err != nil {
			return nil, fmt.Errorf("scan topic progress: %w", err)
		}

		percent := 0.0
		if total > 0 {
			percent = float64(completed) / float64(total) * 100
		}

		result = append(result, &domain.TopicProgressPercent{
			TopicID:   topicID,
			Completed: completed,
			Total:     total,
			Percent:   percent,
		})
	}

	return result, nil
}
