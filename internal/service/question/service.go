// Package question implements the application service layer for question management.
//
// It contains business logic, input DTOs (in other files), service-level errors,
// repository interfaces, and coordinates domain entities with persistence layer.
package question

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/question"
)

// Service manages question business operations and coordinates with the repository layer.
type Service struct {
	questionRepo Repository
	tagRepo      TagRepository
	transactor   Transactor
}

// NewService creates a new service instance.
func NewService(
	questionRepo Repository,
	tagRepo TagRepository,
	transactor Transactor,
) *Service {
	return &Service{
		questionRepo: questionRepo,
		tagRepo:      tagRepo,
		transactor:   transactor,
	}
}

// Create creates a new question based on the provided input.
//
// Returns the created question ID.
func (s *Service) Create(ctx context.Context, input CreateInput) (uuid.UUID, error) {
	if input.Title == "" {
		return uuid.Nil, ErrQuestionInvalidData
	}

	q := domain.NewQuestion(
		input.Title,
		input.Slug,
		input.Content,
		input.Level,
		input.TopicID,
		input.IsFree,
		input.CreatedBy,
	)

	var questionID uuid.UUID

	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		id, err := s.questionRepo.Create(txCtx, q)
		if err != nil {
			if errors.Is(err, question.ErrQuestionAlreadyExists) {
				return ErrQuestionAlreadyExists
			}
			return fmt.Errorf("create question: %w", err)
		}

		questionID = id

		if len(input.TagIDs) > 0 {
			err := s.tagRepo.AddTagsToQuestion(txCtx, id, input.TagIDs)
			if err != nil {
				return fmt.Errorf("attach tags: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return questionID, nil
}

// Update modifies an existing question's details.
// Only admins should call this method.
func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		q, err := s.questionRepo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, question.ErrQuestionNotFound) {
				return ErrQuestionNotFound
			}
			return fmt.Errorf("get question: %w", err)
		}

		// Apply updates if provided
		if err := applyUpdates(q, input, s.tagRepo, txCtx); err != nil {
			return err
		}
		q.UpdatedAt = time.Now().UTC()

		if err := s.questionRepo.Update(txCtx, q); err != nil {
			if errors.Is(err, question.ErrQuestionNotFound) {
				return ErrQuestionNotFound
			}
			if errors.Is(err, question.ErrQuestionAlreadyExists) {
				return ErrQuestionAlreadyExists
			}
			return fmt.Errorf("update question: %w", err)
		}

		return nil
	})
}

func applyUpdates(q *domain.Question, input UpdateInput, tagRepo TagRepository, ctx context.Context) error {
	if input.Title != nil {
		q.Title = *input.Title
	}
	if input.Slug != nil {
		q.Slug = *input.Slug
	}
	if input.Content != nil {
		q.Content = input.Content
	}
	if input.Level != nil {
		q.Level = *input.Level
	}
	if input.TopicID != nil {
		q.TopicID = input.TopicID
	}
	if input.IsFree != nil {
		q.IsFree = *input.IsFree
	}
	if input.TagIDs != nil {
		if err := replaceQuestionTags(ctx, tagRepo, q.ID, *input.TagIDs); err != nil {
			return err
		}
	}
	if input.UpdatedBy != nil {
		q.UpdatedBy = input.UpdatedBy
	}

	return nil
}

func replaceQuestionTags(ctx context.Context, repo TagRepository, questionID uuid.UUID, newTagIDs []uuid.UUID) error {
	if err := repo.RemoveAllForQuestion(ctx, questionID); err != nil {
		return fmt.Errorf("remove existing tags: %w", err)
	}
	if len(newTagIDs) == 0 {
		return nil
	}
	if err := repo.AddTagsToQuestion(ctx, questionID, newTagIDs); err != nil {
		return fmt.Errorf("add tags: %w", err)
	}

	return nil
}

// GetByID retrieves a question by its unique identifier.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	q, err := s.questionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, question.ErrQuestionNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("get question by ID: %w", err)
	}
	return q, nil
}

// GetBySlug retrieves a question by its slug.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*domain.Question, error) {
	q, err := s.questionRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, question.ErrQuestionNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("get question by slug: %w", err)
	}
	return q, nil
}

// ListCards returns question cards with tags and "IsNew" flag, with pagination and filtering.
func (s *Service) ListCards(ctx context.Context, input ListInput) ([]*domain.QuestionCard, error) {
	opts := question.ListOptions{
		Limit:   input.Limit,
		Offset:  input.Offset,
		Level:   input.Level,
		TopicID: input.TopicID,
		IsFree:  input.IsFree,
		TagIDs:  input.TagIDs,
	}
	cards, err := s.questionRepo.ListCards(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list question cards: %w", err)
	}
	return cards, nil
}

// ListByTopic returns all questions of a given topic with full content and associated tags.
// This is used for the topic page under the question cards.
func (s *Service) ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*domain.Question, error) {
	questions, err := s.questionRepo.ListByTopic(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("list questions by topic: %w", err)
	}
	return questions, nil
}
