// Package repository provides the repository layer for accessing database entities.
// It includes PostgreSQL transaction handling, repository-level errors, and repository
// implementations for entities like users, sessions, questions, etc.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/pkg/transactor"
)

// Question repository handles operations related to Question entities.
type Question struct {
	db transactor.DBTx
}

// NewQuestion creates a new Question repository instance.
func NewQuestion(db transactor.DBTx) *Question {
	return &Question{db: db}
}

// Create inserts a new question into the database and returns its ID.
func (r *Question) Create(ctx context.Context, q *domain.Question) (uuid.UUID, error) {
	db := transactor.GetDB(ctx, r.db)

	contentJSON, err := json.Marshal(q.Content)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	const query = `
		INSERT INTO questions (
			title,
		    slug,
			content,
			level,
			topic_id,
			is_free,
			created_by,
			updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id uuid.UUID
	err = db.QueryRow(ctx, query,
		q.Title,
		q.Slug,
		contentJSON,
		q.Level,
		q.TopicID,
		q.IsFree,
		q.CreatedBy,
		q.UpdatedBy,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation &&
			pgErr.ConstraintName == "questions_slug_key" {
			return uuid.Nil, ErrQuestionAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("failed to create question: %w", err)
	}

	return id, nil
}

// GetByID retrieves a question by its UUID.
func (r *Question) GetByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		SELECT 
		    id, 
		    title, 
		    slug, 
		    content, 
		    level, 
		    topic_id, 
		    is_free, 
		    created_by, 
		    updated_by, 
		    created_at, 
		    updated_at
		FROM questions
		WHERE id = $1
	`

	q, err := scanQuestion(db.QueryRow(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get question by id: %w", err)
	}

	return q, nil
}

// GetBySlug retrieves a question by its slug.
func (r *Question) GetBySlug(ctx context.Context, slug string) (*domain.Question, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		SELECT 
			id,
			title,
			slug,
			content,
			level,
			topic_id,
			is_free,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM questions
		WHERE slug = $1
	`

	q, err := scanQuestion(db.QueryRow(ctx, query, slug))
	if err != nil {
		return nil, fmt.Errorf("failed to get question by slug: %w", err)
	}

	return q, nil
}

func scanQuestion(row pgx.Row) (*domain.Question, error) {
	var q domain.Question
	var contentJSON []byte

	err := row.Scan(
		&q.ID,
		&q.Title,
		&q.Slug,
		&contentJSON,
		&q.Level,
		&q.TopicID,
		&q.IsFree,
		&q.CreatedBy,
		&q.UpdatedBy,
		&q.CreatedAt,
		&q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQuestionNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(contentJSON, &q.Content); err != nil {
		return nil, fmt.Errorf("unmarshal content: %w", err)
	}

	return &q, nil
}

// Update modifies an existing question's mutable fields.
// Only admins should call this method.
// Updates fields: Title, Content, Level, TopicID, IsFree, UpdatedBy.
func (r *Question) Update(ctx context.Context, q *domain.Question) error {
	db := transactor.GetDB(ctx, r.db)

	contentJSON, err := json.Marshal(q.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	const query = `
		UPDATE questions
		SET
			title      = $2,
			slug       = $3,
			content    = $4,
			level      = $5,
			topic_id   = $6,
			is_free    = $7,
			updated_by = $8,
			updated_at = now()
		WHERE id = $1
	`

	cmdTag, err := db.Exec(ctx, query,
		q.ID,
		q.Title,
		q.Slug,
		contentJSON,
		q.Level,
		q.TopicID,
		q.IsFree,
		q.UpdatedBy,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation &&
				pgErr.ConstraintName == "questions_slug_key" {
				return ErrQuestionAlreadyExists
			}
		}
		return fmt.Errorf("failed to update question: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrQuestionNotFound
	}

	return nil
}

// ListOptions defines filters and pagination for listing questions.
type ListOptions struct {
	Limit   int
	Offset  int
	Level   *domain.QuestionLevel
	TopicID *uuid.UUID
	IsFree  *bool
}

// List retrieves a list of questions based on filters and pagination.
func (r *Question) List(ctx context.Context, opts ListOptions) ([]*domain.Question, error) {
	db := transactor.GetDB(ctx, r.db)

	query := `
		SELECT 
		    id, 
		    title,
		    slug,
		    content, 
		    level, 
		    topic_id, 
		    is_free, 
		    created_by, 
		    updated_by, 
		    created_at, 
		    updated_at 
		FROM questions 
		WHERE 1=1`
	var args []interface{}
	argID := 1

	if opts.Level != nil {
		query += fmt.Sprintf(" AND level = $%d", argID)
		args = append(args, *opts.Level)
		argID++
	}

	if opts.TopicID != nil {
		query += fmt.Sprintf(" AND topic_id = $%d", argID)
		args = append(args, *opts.TopicID)
		argID++
	}

	if opts.IsFree != nil {
		query += fmt.Sprintf(" AND is_free = $%d", argID)
		args = append(args, *opts.IsFree)
		argID++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list questions: %w", err)
	}
	defer rows.Close()

	var result []*domain.Question
	for rows.Next() {
		var q domain.Question
		var contentJSON []byte
		if err := rows.Scan(
			&q.ID,
			&q.Title,
			&q.Slug,
			&contentJSON,
			&q.Level,
			&q.TopicID,
			&q.IsFree,
			&q.CreatedBy,
			&q.UpdatedBy,
			&q.CreatedAt,
			&q.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		if err := json.Unmarshal(contentJSON, &q.Content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}

		result = append(result, &q)
	}

	return result, nil
}
