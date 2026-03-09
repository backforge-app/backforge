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
	"time"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	repository "github.com/backforge-app/backforge/internal/repository/user"
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

	user := domain.NewUser(
		input.TelegramID,
		input.FirstName,
		input.LastName,
		input.Username,
		input.PhotoURL,
		input.IsPro,
	)

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserTelegramIDTaken):
			return uuid.Nil, ErrUserTelegramIDTaken
		case errors.Is(err, repository.ErrUserInvalidRole):
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
		user, err := s.userRepo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("get user: %w", err)
		}

		// Apply updates (only if provided)
		if input.FirstName != nil {
			user.FirstName = *input.FirstName
		}
		if input.LastName != nil {
			user.LastName = input.LastName
		}
		if input.Username != nil {
			user.Username = input.Username
		}
		if input.IsPro != nil {
			user.IsPro = *input.IsPro
			if user.IsPro {
				now := time.Now().UTC()
				user.ProGrantedAt = &now
				proType := "channel"
				user.ProType = &proType
			} else {
				user.ProGrantedAt = nil
				user.ProType = nil
			}
		}
		if input.Role != nil {
			user.Role = *input.Role
		}

		if err := s.userRepo.Update(txCtx, user); err != nil {
			switch {
			case errors.Is(err, repository.ErrUserNotFound):
				return ErrUserNotFound
			case errors.Is(err, repository.ErrUserInvalidRole):
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
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}
	return user, nil
}

// GetByTelegramID retrieves user details by Telegram user ID.
func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by Telegram ID: %w", err)
	}
	return user, nil
}

// GetOrCreateByTelegramID gets a user by TelegramID or creates one if not exists.
func (s *Service) GetOrCreateByTelegramID(ctx context.Context, input CreateInput) (*domain.User, error) {
	user, err := s.GetByTelegramID(ctx, input.TelegramID)
	if err == nil {
		return user, nil
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

// UpdateProStatus updates the Pro status for a user (separate method for common operation).
func (s *Service) UpdateProStatus(ctx context.Context, telegramID int64, isPro bool) error {
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		user, err := s.userRepo.GetByTelegramID(txCtx, telegramID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("get user: %w", err)
		}

		user.IsPro = isPro
		if isPro {
			now := time.Now().UTC()
			user.ProGrantedAt = &now
			proType := "channel"
			user.ProType = &proType
		} else {
			user.ProGrantedAt = nil
			user.ProType = nil
		}

		if err := s.userRepo.Update(txCtx, user); err != nil {
			return fmt.Errorf("update user Pro status: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
