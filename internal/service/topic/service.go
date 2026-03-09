// Package topic implements the application service layer for topic management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with the persistence layer.
package topic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/topic"
)

// Service manages topic business operations and coordinates with the repository layer.
type Service struct {
	topicRepo  Repository
	transactor Transactor
}

// NewService creates a new topic service instance.
func NewService(topicRepo Repository, transactor Transactor) *Service {
	return &Service{
		topicRepo:  topicRepo,
		transactor: transactor,
	}
}

// Create creates a new topic based on the provided input.
//
// Returns the created topic ID.
func (s *Service) Create(ctx context.Context, input CreateInput) (uuid.UUID, error) {
	if input.Title == "" {
		return uuid.Nil, ErrTopicInvalidData
	}

	t := domain.NewTopic(
		input.Title,
		input.Slug,
		input.Description,
		input.CreatedBy,
	)

	var topicID uuid.UUID

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		id, err := s.topicRepo.Create(txCtx, t)
		if err != nil {
			if errors.Is(err, topic.ErrTopicAlreadyExists) {
				return ErrTopicAlreadyExists
			}
			return fmt.Errorf("create topic: %w", err)
		}

		topicID = id
		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return topicID, nil
}

// Update modifies an existing topic.
//
// Only admins should call this method.
func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		t, err := s.topicRepo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, topic.ErrTopicNotFound) {
				return ErrTopicNotFound
			}
			return fmt.Errorf("get topic: %w", err)
		}

		if input.Title != nil {
			t.Title = *input.Title
		}
		if input.Slug != nil {
			t.Slug = *input.Slug
		}
		if input.Description != nil {
			t.Description = *input.Description
		}
		if input.UpdatedBy != nil {
			t.UpdatedBy = input.UpdatedBy
		}
		t.UpdatedAt = time.Now().UTC()

		if err := s.topicRepo.Update(txCtx, t); err != nil {
			if errors.Is(err, topic.ErrTopicNotFound) {
				return ErrTopicNotFound
			}
			if errors.Is(err, topic.ErrTopicAlreadyExists) {
				return ErrTopicAlreadyExists
			}
			return fmt.Errorf("update topic: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a topic by its unique identifier.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error) {
	t, err := s.topicRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, topic.ErrTopicNotFound) {
			return nil, ErrTopicNotFound
		}
		return nil, fmt.Errorf("get topic by ID: %w", err)
	}
	return t, nil
}

// GetBySlug retrieves a topic by its slug.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.Topic, error) {
	t, err := s.topicRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, topic.ErrTopicNotFound) {
			return nil, ErrTopicNotFound
		}
		return nil, fmt.Errorf("get topic by slug: %w", err)
	}
	return t, nil
}

// ListRows retrieves topics with their question counts for displaying in a table.
func (s *Service) ListRows(ctx context.Context) ([]*domain.TopicRow, error) {
	topics, err := s.topicRepo.ListRows(ctx)
	if err != nil {
		return nil, err
	}
	return topics, nil
}
