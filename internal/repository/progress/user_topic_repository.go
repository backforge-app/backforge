// Package progress provides the repository layer for accessing user question progress entities.
//
// It includes PostgreSQL operations, transaction handling, repository-level errors,
// and methods to create, read, update, and delete progress entries.
package progress

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// UserTopicRepository provides data access operations for user topic progress.
type UserTopicRepository struct {
	db transactor.DBTx
}

// NewUserTopicRepository creates a new repository instance.
func NewUserTopicRepository(db transactor.DBTx) *UserTopicRepository {
	return &UserTopicRepository{db: db}
}

// GetByUserAndTopic retrieves the progress for a given user and topic.
// Returns ErrTopicProgressNotFound if no progress exists.
func (r *UserTopicRepository) GetByUserAndTopic(
	ctx context.Context,
	userID, topicID uuid.UUID,
) (*domain.UserTopicProgress, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    id, 
		    user_id, 
		    topic_id, 
		    current_position, 
		    updated_at
		FROM user_topic_progress
		WHERE user_id = $1 AND topic_id = $2
	`

	var p domain.UserTopicProgress
	err := db.QueryRow(ctx, q,
		userID,
		topicID,
	).Scan(
		&p.ID,
		&p.UserID,
		&p.TopicID,
		&p.CurrentPosition,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTopicProgressNotFound
		}
		return nil, fmt.Errorf("get topic progress by user and topic: %w", err)
	}

	return &p, nil
}

// SetPosition sets or updates the current position of a user in a topic.
// Creates a new entry if it does not exist.
func (r *UserTopicRepository) SetPosition(ctx context.Context, userID, topicID uuid.UUID, pos int) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO user_topic_progress (
			user_id, 
			topic_id, 
			current_position
		) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, topic_id) DO UPDATE
		SET 
		    current_position = EXCLUDED.current_position, 
		    updated_at = now()
	`

	_, err := db.Exec(ctx, q, userID, topicID, pos)
	if err != nil {
		return fmt.Errorf("set topic position: %w", err)
	}

	return nil
}

// ListByUser returns all topic progress entries for a given user.
func (r *UserTopicRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserTopicProgress, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    id, 
		    user_id, 
		    topic_id, 
		    current_position, 
		    updated_at
		FROM user_topic_progress
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list topic progress by user: %w", err)
	}
	defer rows.Close()

	var result []*domain.UserTopicProgress
	for rows.Next() {
		var p domain.UserTopicProgress
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.TopicID,
			&p.CurrentPosition,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan topic progress row: %w", err)
		}
		result = append(result, &p)
	}

	return result, nil
}

// ResetByTopic resets the current position of a user in a specific topic to 0.
func (r *UserTopicRepository) ResetByTopic(ctx context.Context, userID, topicID uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE user_topic_progress
		SET 
		    current_position = 0, 
		    updated_at = now()
		WHERE user_id = $1 AND topic_id = $2
	`

	_, err := db.Exec(ctx, q, userID, topicID)
	if err != nil {
		return fmt.Errorf("reset topic progress: %w", err)
	}

	return nil
}

// ResetAll resets the current position for all topics for a user.
func (r *UserTopicRepository) ResetAll(ctx context.Context, userID uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		UPDATE user_topic_progress
		SET 
		    current_position = 0, 
		    updated_at = now()
		WHERE user_id = $1
	`

	_, err := db.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("reset all topic progress: %w", err)
	}

	return nil
}
