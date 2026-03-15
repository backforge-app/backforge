// Package questiontag provides the repository layer for accessing question-tag many-to-many relationships.
// It includes PostgreSQL operations, transaction handling and methods to link/unlink tags to questions.
package questiontag

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// Repository handles operations related to question-tag associations.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new QuestionTag repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// AddTagsToQuestion inserts multiple tag links for a question.
// Uses ON CONFLICT DO NOTHING to avoid duplicates.
func (r *Repository) AddTagsToQuestion(ctx context.Context, questionID uuid.UUID, tagIDs []uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	for _, tagID := range tagIDs {
		const q = `
			INSERT INTO question_tags (question_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
		if _, err := db.Exec(ctx, q, questionID, tagID); err != nil {
			return fmt.Errorf("add tag to question: %w", err)
		}
	}

	return nil
}

// RemoveTagsFromQuestion deletes specific tag links for a question.
func (r *Repository) RemoveTagsFromQuestion(ctx context.Context, questionID uuid.UUID, tagIDs []uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	for _, tagID := range tagIDs {
		const q = `
			DELETE FROM question_tags
			WHERE question_id = $1 AND tag_id = $2
		`
		if _, err := db.Exec(ctx, q, questionID, tagID); err != nil {
			return fmt.Errorf("remove tag from question: %w", err)
		}
	}

	return nil
}

// RemoveAllForQuestion deletes all tag links for a specific question.
func (r *Repository) RemoveAllForQuestion(ctx context.Context, questionID uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		DELETE FROM question_tags
		WHERE question_id = $1
	`

	if _, err := db.Exec(ctx, q, questionID); err != nil {
		return fmt.Errorf("remove all tags from question: %w", err)
	}

	return nil
}

// ListTagsForQuestion returns all tags linked to a specific question, ordered by tag name.
func (r *Repository) ListTagsForQuestion(ctx context.Context, questionID uuid.UUID) ([]*domain.Tag, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    t.id, 
		    t.name, 
		    t.created_by, 
		    t.updated_by, 
		    t.created_at, 
		    t.updated_at
		FROM tags t
		JOIN question_tags qt ON qt.tag_id = t.id
		WHERE qt.question_id = $1
		ORDER BY t.name
	`

	rows, err := db.Query(ctx, q, questionID)
	if err != nil {
		return nil, fmt.Errorf("list tags for question: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.CreatedBy,
			&t.UpdatedBy,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, &t)
	}

	return tags, nil
}
