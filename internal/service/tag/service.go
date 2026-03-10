// Package tag implements the application service layer for tag management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package tag

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/tag"
)

// Service manages tag business operations and coordinates with the repository layer.
type Service struct {
	tagRepo Repository
}

// NewService creates a new tag service instance.
func NewService(tagRepo Repository) *Service {
	return &Service{tagRepo: tagRepo}
}

// Create creates a new tag based on the provided input.
//
// Returns the created tag ID.
func (s *Service) Create(ctx context.Context, name string, createdBy *uuid.UUID) (uuid.UUID, error) {
	if name == "" {
		return uuid.Nil, ErrTagInvalidData
	}

	t := &domain.Tag{
		Name:      name,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	id, err := s.tagRepo.Create(ctx, t)
	if err != nil {
		if errors.Is(err, tag.ErrTagAlreadyExists) {
			return uuid.Nil, ErrTagAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("create tag: %w", err)
	}

	return id, nil
}

// Delete removes a tag by its ID.
//
// Returns ErrTagNotFound if the tag does not exist.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.tagRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, tag.ErrTagNotFound) {
			return ErrTagNotFound
		}
		return fmt.Errorf("delete tag: %w", err)
	}

	return nil
}

// GetByID retrieves a tag by its UUID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	t, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, tag.ErrTagNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, fmt.Errorf("get tag by ID: %w", err)
	}
	return t, nil
}

// List retrieves all tags ordered by name.
func (s *Service) List(ctx context.Context) ([]*domain.Tag, error) {
	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return tags, nil
}
