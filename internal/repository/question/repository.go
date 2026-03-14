// Package question provides the repository layer for accessing question entities.
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, update, list, and manage questions.
package question

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

// Repository handles operations related to Question entities.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new Question repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// Create inserts a new Repository into the database and returns its ID.
func (r *Repository) Create(ctx context.Context, q *domain.Question) (uuid.UUID, error) {
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

// Update modifies an existing question. Returns ErrQuestionNotFound if not found
// or ErrQuestionAlreadyExists if slug conflict occurs.
func (r *Repository) Update(ctx context.Context, q *domain.Question) error {
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

// GetByID retrieves a complete question by UUID including all associated tags.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		SELECT
			q.id,
			q.title,
			q.slug,
			q.content,
			q.level,
			q.topic_id,
			q.is_free,
			q.created_by,
			q.updated_by,
			q.created_at,
			q.updated_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', t.id,
						'name', t.name
					)
				) FILTER (WHERE t.id IS NOT NULL),
				'[]'
			) as tags_json
		FROM questions q
		LEFT JOIN question_tags qt ON qt.question_id = q.id
		LEFT JOIN tags t ON t.id = qt.tag_id
		WHERE q.id = $1
		GROUP BY q.id
	`

	return scanQuestion(db.QueryRow(ctx, query, id))
}

// GetBySlug retrieves a complete question by slug including all associated tags.
// Question is the full business entity: always returns tags via JSON aggregation.
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*domain.Question, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
		SELECT
			q.id,
			q.title,
			q.slug,
			q.content,
			q.level,
			q.topic_id,
			q.is_free,
			q.created_by,
			q.updated_by,
			q.created_at,
			q.updated_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', t.id,
						'name', t.name
					)
				) FILTER (WHERE t.id IS NOT NULL),
				'[]'
			) as tags_json
		FROM questions q
		LEFT JOIN question_tags qt ON qt.question_id = q.id
		LEFT JOIN tags t ON t.id = qt.tag_id
		WHERE q.slug = $1
		GROUP BY q.id
	`

	return scanQuestion(db.QueryRow(ctx, query, slug))
}

func scanQuestion(row pgx.Row) (*domain.Question, error) {
	var q domain.Question
	var contentJSON []byte
	var tagsJSON []byte

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
		&tagsJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQuestionNotFound
		}

		return nil, err
	}

	if err := json.Unmarshal(contentJSON, &q.Content); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(tagsJSON, &q.Tags); err != nil {
		return nil, err
	}

	return &q, nil
}

// ListOptions contains filtering and pagination parameters for question listing.
type ListOptions struct {
	Limit  int
	Offset int
	Search *string

	Level *domain.QuestionLevel
	Tags  []string
}

// ListCards returns lightweight question representations (cards) according to the provided filters.
func (r *Repository) ListCards(ctx context.Context, opts ListOptions) ([]*domain.QuestionCard, error) {
	db := transactor.GetDB(ctx, r.db)

	query := `
	SELECT
		q.id,
		q.title,
		q.slug,
		q.level,
		q.is_free,
		q.created_at > now() - interval '3 days' as is_new,
		COALESCE(
			json_agg(
				t.name
			) FILTER (WHERE t.id IS NOT NULL),
			'[]'
		) as tags_json
	FROM questions q
	LEFT JOIN question_tags qt ON qt.question_id = q.id
	LEFT JOIN tags t ON t.id = qt.tag_id
	WHERE 1=1
	`

	args := make([]any, 0)
	arg := 1

	if opts.Search != nil {
		query += fmt.Sprintf(" AND q.title ILIKE '%%' || $%d || '%%'", arg)
		args = append(args, *opts.Search)
		arg++
	}

	if opts.Level != nil {
		query += fmt.Sprintf(" AND q.level = $%d", arg)
		args = append(args, *opts.Level)
		arg++
	}

	if len(opts.Tags) > 0 {
		query += fmt.Sprintf(`
		AND q.id IN (
			SELECT question_id
			FROM question_tags qt
			JOIN tags t2 ON t2.id = qt.tag_id
			WHERE t2.name = ANY($%d)
		)
		`, arg)
		args = append(args, opts.Tags)
		arg++
	}

	query += fmt.Sprintf(`
	GROUP BY q.id
	ORDER BY q.created_at DESC
	LIMIT $%d OFFSET $%d
	`, arg, arg+1)
	args = append(args, opts.Limit, opts.Offset)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list cards: %w", err)
	}
	defer rows.Close()

	var result []*domain.QuestionCard
	for rows.Next() {
		q, err := scanQuestionCard(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, q)
	}

	return result, nil
}

func scanQuestionCard(row pgx.Row) (*domain.QuestionCard, error) {
	var q domain.QuestionCard
	var tagsJSON []byte

	err := row.Scan(
		&q.ID,
		&q.Title,
		&q.Slug,
		&q.Level,
		&q.IsFree,
		&q.IsNew,
		&tagsJSON,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(tagsJSON, &q.Tags); err != nil {
		return nil, err
	}

	return &q, nil
}

// ListByTopic returns all full questions belonging to the specified topic, including tags.
func (r *Repository) ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*domain.Question, error) {
	db := transactor.GetDB(ctx, r.db)

	const query = `
        SELECT
            q.id,
            q.title,
            q.slug,
            q.level,
            q.content,
            q.topic_id,
            q.is_free,
            q.created_by,
            q.updated_by,
            q.created_at,
            q.updated_at,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', t.id,
                        'name', t.name,
                        'created_by', t.created_by,
                        'updated_by', t.updated_by,
                        'created_at', t.created_at,
                        'updated_at', t.updated_at
                    )
                ) FILTER (WHERE t.id IS NOT NULL),
                '[]'::json
            ) as tags_json
        FROM questions q
        LEFT JOIN question_tags qt ON qt.question_id = q.id
        LEFT JOIN tags t ON t.id = qt.tag_id
        WHERE q.topic_id = $1
        GROUP BY
            q.id, q.title, q.slug, q.level, q.content, q.topic_id, q.is_free,
            q.created_by, q.updated_by, q.created_at, q.updated_at
        ORDER BY q.created_at
    `

	rows, err := db.Query(ctx, query, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to list questions by topic: %w", err)
	}
	defer rows.Close()

	var questions []*domain.Question

	for rows.Next() {
		var q domain.Question
		var contentJSON []byte
		var tagsJSON []byte

		err := rows.Scan(
			&q.ID,
			&q.Title,
			&q.Slug,
			&q.Level,
			&contentJSON,
			&q.TopicID,
			&q.IsFree,
			&q.CreatedBy,
			&q.UpdatedBy,
			&q.CreatedAt,
			&q.UpdatedAt,
			&tagsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}

		if err := json.Unmarshal(contentJSON, &q.Content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &q.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		questions = append(questions, &q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return questions, nil
}
