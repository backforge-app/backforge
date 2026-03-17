// Package tag provides the repository layer for accessing tag entities.
//
// It includes PostgreSQL operations, transaction handling, repository-level errors
// and methods to create, read, list, and manage tags.
package tag

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

// Repository handles operations related to Tag entities.
type Repository struct {
	db transactor.DBTx
}

// NewRepository creates a new Tag repository instance.
func NewRepository(db transactor.DBTx) *Repository {
	return &Repository{db: db}
}

// Create inserts a new tag into the database and returns its ID.
func (r *Repository) Create(ctx context.Context, t *domain.Tag) (uuid.UUID, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		INSERT INTO tags (
		    name, 
		    created_by
		) VALUES ($1, $2)
		RETURNING id
	`

	var id uuid.UUID
	err := db.QueryRow(ctx, q,
		t.Name,
		t.CreatedBy,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == pgerrcode.UniqueViolation &&
			pgErr.ConstraintName == "tags_name_key" {
			return uuid.Nil, ErrTagAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("create tag: %w", err)
	}

	return id, nil
}

// GetByID retrieves a tag by its UUID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    id, 
		    name, 
		    created_by, 
		    updated_by, 
		    created_at, 
		    updated_at
		FROM tags
		WHERE id = $1
	`

	t, err := scanTag(db.QueryRow(ctx, q, id))
	if err != nil {
		return nil, fmt.Errorf("get tag by id: %w", err)
	}

	return t, nil
}

// GetByName retrieves a tag by its name.
func (r *Repository) GetByName(ctx context.Context, name string) (*domain.Tag, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    id, 
		    name, 
		    created_by, 
		    updated_by, 
		    created_at, 
		    updated_at
		FROM tags
		WHERE name = $1
	`

	t, err := scanTag(db.QueryRow(ctx, q, name))
	if err != nil {
		return nil, fmt.Errorf("get tag by name: %w", err)
	}

	return t, nil
}

func scanTag(row pgx.Row) (*domain.Tag, error) {
	var t domain.Tag
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.CreatedBy,
		&t.UpdatedBy,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	return &t, nil
}

// Delete removes a tag from the database by its UUID.
//
// Returns ErrTagNotFound if the tag does not exist.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	db := transactor.GetDB(ctx, r.db)

	const q = `DELETE FROM tags WHERE id = $1`

	cmdTag, err := db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrTagNotFound
	}

	return nil
}

// List retrieves all tags ordered by name.
func (r *Repository) List(ctx context.Context) ([]*domain.Tag, error) {
	db := transactor.GetDB(ctx, r.db)

	const q = `
		SELECT 
		    id, 
		    name, 
		    created_by, 
		    updated_by, 
		    created_at, 
		    updated_at
		FROM tags
		ORDER BY name
	`

	rows, err := db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
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
