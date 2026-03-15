// Package user implements the application service layer for user management.
//
// It contains business logic for user creation, updates, retrieval,
// service-level errors, input DTOs (in other files), and coordinates
// domain entities with the persistence layer.
package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/repository/user"
)

// Service manages user business operations and coordinates with the repository layer.
type Service struct {
	userRepo   Repository
	transactor Transactor
}

// NewService creates a new service instance.
func NewService(userRepo Repository, transactor Transactor) *Service {
	return &Service{
		userRepo:   userRepo,
		transactor: transactor,
	}
}

// Create creates a new user based on the provided input.
//
// Returns the created user ID.
func (s *Service) Create(ctx context.Context, input CreateInput) (uuid.UUID, error) {
	if input.TelegramID <= 0 {
		return uuid.Nil, ErrUserInvalidData
	}

	u := domain.NewUser(
		input.TelegramID,
		input.FirstName,
		input.LastName,
		input.Username,
		input.PhotoURL,
	)

	id, err := s.userRepo.Create(ctx, u)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrUserTelegramIDTaken):
			return uuid.Nil, ErrUserTelegramIDTaken
		case errors.Is(err, user.ErrUserInvalidRole):
			return uuid.Nil, ErrUserInvalidRole
		default:
			return uuid.Nil, fmt.Errorf("create user: %w", err)
		}
	}

	return id, nil
}

// Update modifies an existing user's details based on the provided input.
func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("get user: %w", err)
		}

		// Apply updates (only if provided)
		if input.FirstName != nil {
			u.FirstName = *input.FirstName
		}
		if input.LastName != nil {
			u.LastName = input.LastName
		}
		if input.Username != nil {
			u.Username = input.Username
		}
		if input.Role != nil {
			u.Role = *input.Role
		}

		if err := s.userRepo.Update(txCtx, u); err != nil {
			switch {
			case errors.Is(err, user.ErrUserNotFound):
				return ErrUserNotFound
			case errors.Is(err, user.ErrUserInvalidRole):
				return ErrUserInvalidRole
			default:
				return fmt.Errorf("update user: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves user details by unique identifier.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}
	return u, nil
}

// GetByTelegramID retrieves user details by Telegram user ID.
func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	u, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by Telegram ID: %w", err)
	}
	return u, nil
}

// GetOrCreateByTelegramID gets a user by TelegramID or creates one if not exists.
func (s *Service) GetOrCreateByTelegramID(ctx context.Context, input CreateInput) (*domain.User, error) {
	u, err := s.GetByTelegramID(ctx, input.TelegramID)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// Create if not found
	id, err := s.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}
