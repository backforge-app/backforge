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

// UserQuestionRepository handles operations related to user question progress.
type UserQuestionRepository struct {
	db transactor.DBTx
}

// NewUserQuestionRepository creates a new Repository instance.
func NewUserQuestionRepository(db transactor.DBTx) *UserQuestionRepository {
	return &UserQuestionRepository{db: db}
}

// GetByUserAndQuestion retrieves the progress for a given user and question.
// Returns ErrProgressNotFound if no progress exists.
func (r *UserQuestionRepository) GetByUserAndQuestion(
	ctx context.Context,
	userID, questionID uuid.UUID,
) (*domain.UserQuestionProgress, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    id, 
		    user_id, 
		    question_id, 
		    status, 
		    updated_at
		FROM user_question_progress
		WHERE user_id = $1 AND question_id = $2
	`

	var p domain.UserQuestionProgress
	err := db.QueryRow(ctx, q,
		userID,
		questionID,
	).Scan(
		&p.ID,
		&p.UserID,
		&p.QuestionID,
		&p.Status,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQuestionProgressNotFound
		}
		return nil, fmt.Errorf("get progress by user and question: %w", err)
	}

	return &p, nil
}

// ListByUser returns all progress entries for a given user.
func (r *UserQuestionRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.UserQuestionProgress, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    id, 
		    user_id, 
		    question_id, 
		    status, 
		    updated_at
		FROM user_question_progress
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list progress by user: %w", err)
	}
	defer rows.Close()

	var result []*domain.UserQuestionProgress
	for rows.Next() {
		var p domain.UserQuestionProgress
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.QuestionID,
			&p.Status,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan progress row: %w", err)
		}
		result = append(result, &p)
	}

	return result, nil
}

// ListByUserAndTopic returns all progress entries for a user filtered by topic.
func (r *UserQuestionRepository) ListByUserAndTopic(
	ctx context.Context,
	userID, topicID uuid.UUID,
) ([]*domain.UserQuestionProgress, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    up.id, 
		    up.user_id, 
		    up.question_id, 
		    up.status, 
		    up.updated_at
		FROM user_question_progress up
		JOIN questions q ON up.question_id = q.id
		WHERE up.user_id = $1 AND q.topic_id = $2
		ORDER BY up.updated_at DESC
	`

	rows, err := db.Query(ctx, q, userID, topicID)
	if err != nil {
		return nil, fmt.Errorf("list progress by user and topic: %w", err)
	}
	defer rows.Close()

	var result []*domain.UserQuestionProgress
	for rows.Next() {
		var p domain.UserQuestionProgress
		if err := rows.Scan(&p.ID, &p.UserID, &p.QuestionID, &p.Status, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan progress row: %w", err)
		}
		result = append(result, &p)
	}

	return result, nil
}

// SetStatus creates or updates the progress status for a specific user and question.
func (r *UserQuestionRepository) SetStatus(
	ctx context.Context,
	userID, questionID uuid.UUID,
	status domain.ProgressStatus,
) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO user_question_progress (
		    user_id, 
		    question_id, 
			status
		) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, question_id) DO UPDATE
		SET 
		    status = EXCLUDED.status, 
		    updated_at = now()
	`

	_, err := db.Exec(ctx, q, userID, questionID, status)
	if err != nil {
		return fmt.Errorf("set progress status: %w", err)
	}

	return nil
}

// ResetByTopic deletes all progress entries for a user in a specific topic.
// This should be executed within a transaction if combined with other updates.
func (r *UserQuestionRepository) ResetByTopic(ctx context.Context, userID, topicID uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		DELETE FROM user_question_progress
		WHERE user_id = $1
			AND question_id IN (
				SELECT id 
				FROM questions 
				WHERE topic_id = $2
			)
	`

	_, err := db.Exec(ctx, q, userID, topicID)
	if err != nil {
		return fmt.Errorf("reset progress by topic: %w", err)
	}

	return nil
}

// ResetAll deletes all progress entries for a user across all topics.
func (r *UserQuestionRepository) ResetAll(ctx context.Context, userID uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		DELETE FROM user_question_progress
		WHERE user_id = $1
	`

	_, err := db.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("reset all progress: %w", err)
	}

	return nil
}
