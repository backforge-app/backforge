// Package topic provides the repository layer for accessing topic entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, and list topics.
package topic

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// Repository handles operations related to Topic entities.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new Topic repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// Create inserts a new topic into the database and returns its ID.
func (r *Repository) Create(ctx context.Context, t *domain.Topic) (uuid.UUID, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		INSERT INTO topics (
		    title, 
		    slug, 
		    description, 
			created_by, 
		    updated_by
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id uuid.UUID
	err := db.QueryRow(ctx, query,
		t.Title,
		t.Slug,
		t.Description,
		t.CreatedBy,
		t.UpdatedBy,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == pgerrcode.UniqueViolation &&
			pgErr.ConstraintName == "topics_slug_key" {
			return uuid.Nil, ErrTopicAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("failed to create topic: %w", err)
	}

	return id, nil
}

// GetByID retrieves a topic by its UUID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		SELECT 
		    id, 
		    title, 
		    slug, 
		    description, 
		    created_by, 
		    updated_by, 
		    created_at, 
		    updated_at
		FROM topics
		WHERE id = $1
	`

	t, err := scanTopic(db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get topic by id: %w", err)
	}

	return t, nil
}

// GetBySlug retrieves a topic by its slug.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*domain.Topic, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		SELECT 
		    id, 
		    title, 
		    slug, 
		    description, 
		    created_by, 
		    updated_by, 
		    created_at, 
		    updated_at
		FROM topics
		WHERE slug = $1
	`

	t, err := scanTopic(db.QueryRow(ctx, query, slug))
	if err != nil {
		return nil, fmt.Errorf("failed to get topic by slug: %w", err)
	}

	return t, nil
}

func scanTopic(row pgx.Row) (*domain.Topic, error) {
	var t domain.Topic
	err := row.Scan(
		&t.ID,
		&t.Title,
		&t.Slug,
		&t.Description,
		&t.CreatedBy,
		&t.UpdatedBy,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTopicNotFound
		}
		return nil, err
	}

	return &t, nil
}

// Update modifies an existing topic.
func (r *Repository) Update(ctx context.Context, t *domain.Topic) error {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		UPDATE topics
		SET 
		    title = $2, 
		    slug = $3, 
		    description = $4, 
		    updated_by = $5, 
		    updated_at = now()
		WHERE id = $1
	`

	cmdTag, err := db.Exec(ctx, query,
		t.ID,
		t.Title,
		t.Slug,
		t.Description,
		t.UpdatedBy,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == pgerrcode.UniqueViolation &&
			pgErr.ConstraintName == "topics_slug_key" {
			return ErrTopicAlreadyExists
		}
		return fmt.Errorf("failed to update topic: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrTopicNotFound
	}

	return nil
}

// ListRows retrieves all topics with question counts.
// Optimized for displaying in a topic table in the UI.
func (r *Repository) ListRows(ctx context.Context) ([]*domain.TopicRow, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		SELECT t.id, t.title, t.slug, COALESCE(q_count.count, 0) AS question_count
		FROM topics t
		LEFT JOIN (
    		SELECT topic_id, COUNT(*) AS count
    		FROM questions
    		GROUP BY topic_id
		) q_count ON q_count.topic_id = t.id
		ORDER BY t.created_at DESC;
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list topic rows: %w", err)
	}
	defer rows.Close()

	var result []*domain.TopicRow
	for rows.Next() {
		var row domain.TopicRow
		if err := rows.Scan(
			&row.ID,
			&row.Title,
			&row.Slug,
			&row.QuestionCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan topic row: %w", err)
		}
		result = append(result, &row)
	}

	return result, nil
}
