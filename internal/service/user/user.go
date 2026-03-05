// Package user implements the User application service.
//
// It contains the business logic for managing users, including
// service methods, input DTOs, and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/backforge-app/backforge/internal/domain"
	"github.com/backforge-app/backforge/internal/infra/postgres"
	"github.com/backforge-app/backforge/internal/service"
)

// Service manages user business operations and coordinates with the repository layer.
type Service struct {
	userRepo   service.UserRepository
	transactor service.Transactor
}

// NewService creates a new service instance.
func NewService(userRepo service.UserRepository, transactor service.Transactor) *Service {
	return &Service{
		userRepo:   userRepo,
		transactor: transactor,
	}
}

// Create creates a new user based on the provided input.
//
// Returns the created user ID.
func (s *Service) Create(ctx context.Context, input CreateUserInput) (uuid.UUID, error) {
	if input.TgUserID <= 0 {
		return uuid.Nil, service.ErrUserInvalidData
	}

	user := domain.NewUser(
		input.TgUserID,
		input.FirstName,
		input.LastName,
		input.Username,
		input.IsPro,
	)

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrUserTgUserIDTaken):
			return uuid.Nil, service.ErrUserTgUserIDTaken
		case errors.Is(err, postgres.ErrUserInvalidRole):
			return uuid.Nil, service.ErrUserInvalidRole
		default:
			return uuid.Nil, fmt.Errorf("create user: %w", err)
		}
	}

	return id, nil
}

// Update modifies an existing user's details based on the provided input.
func (s *Service) Update(ctx context.Context, input UpdateUserInput) error {
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		repo := postgres.NewUserRepository(tx)
		user, err := repo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, postgres.ErrUserNotFound) {
				return service.ErrUserNotFound
			}
			return fmt.Errorf("get user: %w", err)
		}

		// Apply updates (only if provided)
		if input.FirstName != nil {
			user.TgFirstName = *input.FirstName
		}
		if input.LastName != nil {
			user.TgLastName = input.LastName
		}
		if input.Username != nil {
			user.TgUsername = input.Username
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

		if err := repo.Update(txCtx, user); err != nil {
			switch {
			case errors.Is(err, postgres.ErrUserNotFound):
				return service.ErrUserNotFound
			case errors.Is(err, postgres.ErrUserInvalidRole):
				return service.ErrUserInvalidRole
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
		if errors.Is(err, postgres.ErrUserNotFound) {
			return nil, service.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}
	return user, nil
}

// GetByTgUserID retrieves user details by Telegram user ID.
func (s *Service) GetByTgUserID(ctx context.Context, tgUserID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByTgUserID(ctx, tgUserID)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return nil, service.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by Telegram ID: %w", err)
	}
	return user, nil
}

// UpdateProStatus updates the Pro status for a user (separate method for common operation).
func (s *Service) UpdateProStatus(ctx context.Context, tgUserID int64, isPro bool) error {
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		repo := postgres.NewUserRepository(tx)
		user, err := repo.GetByTgUserID(txCtx, tgUserID)
		if err != nil {
			if errors.Is(err, postgres.ErrUserNotFound) {
				return service.ErrUserNotFound
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

		if err := repo.Update(txCtx, user); err != nil {
			return fmt.Errorf("update user Pro status: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
