// Package question implements the Question application service.
//
// It contains the business logic for managing questions, including
// service methods, input DTOs, service-level errors,
// repository interfaces and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
package question

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository"
)

// Service manages question business operations and coordinates with the repository layer.
type Service struct {
	questionRepo Repository
	transactor   Transactor
}

// NewService creates a new service instance.
func NewService(questionRepo Repository, transactor Transactor) *Service {
	return &Service{
		questionRepo: questionRepo,
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
		input.Content,
		input.Level,
		input.TopicID,
		input.CreatedBy,
		input.IsFree,
	)

	id, err := s.questionRepo.Create(ctx, q)
	if err != nil {
		if errors.Is(err, repository.ErrQuestionAlreadyExists) {
			return uuid.Nil, ErrQuestionAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("create question: %w", err)
	}

	return id, nil
}

// Update modifies an existing question's details.
// Only admins should call this method.
func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	return s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		q, err := s.questionRepo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, repository.ErrQuestionNotFound) {
				return ErrQuestionNotFound
			}
			return fmt.Errorf("get question: %w", err)
		}

		// Apply updates if provided
		if input.Title != nil {
			q.Title = *input.Title
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
		if input.UpdatedBy != nil {
			q.UpdatedBy = input.UpdatedBy
		}
		q.UpdatedAt = time.Now().UTC()

		if err := s.questionRepo.Update(txCtx, q); err != nil {
			if errors.Is(err, repository.ErrQuestionNotFound) {
				return ErrQuestionNotFound
			}
			if errors.Is(err, repository.ErrQuestionAlreadyExists) {
				return ErrQuestionAlreadyExists
			}
			return fmt.Errorf("update question: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a question by its unique identifier.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	q, err := s.questionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrQuestionNotFound) {
			return nil, ErrQuestionNotFound
		}
		return nil, fmt.Errorf("get question by ID: %w", err)
	}
	return q, nil
}

// List retrieves a list of questions based on filters and pagination.
func (s *Service) List(ctx context.Context, input ListInput) ([]*domain.Question, error) {
	opts := repository.ListOptions{
		Limit:   input.Limit,
		Offset:  input.Offset,
		Level:   input.Level,
		TopicID: input.TopicID,
		IsFree:  input.IsFree,
	}
	questions, err := s.questionRepo.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	return questions, nil
}
