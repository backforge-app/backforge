// Package user implements the application service layer for user management.
//
// It contains business logic for user creation, updates, retrieval,
// service-level errors, input DTOs, and coordinates domain entities
// with the PostgreSQL persistence layer.
package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
	repouser "github.com/backforge-app/backforge/internal/repository/user"
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

// CreateWithPassword creates a new user via standard email and password registration.
func (s *Service) CreateWithPassword(ctx context.Context, input CreateWithPasswordInput) (uuid.UUID, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" || input.Password == "" {
		return uuid.Nil, ErrUserInvalidData
	}

	u, err := domain.NewUserWithPassword(
		input.Email,
		input.Password,
		input.FirstName,
		input.LastName,
		input.Username,
		input.PhotoURL,
	)
	if err != nil {
		if errors.Is(err, domain.ErrPasswordTooLong) {
			return uuid.Nil, ErrPasswordTooLong
		}
		return uuid.Nil, fmt.Errorf("create domain user: %w", err)
	}

	return s.saveUser(ctx, u)
}

// CreateOAuthUser creates a new user profile originating from a third-party provider.
func (s *Service) CreateOAuthUser(ctx context.Context, input CreateOAuthInput) (uuid.UUID, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Email == "" {
		return uuid.Nil, ErrUserInvalidData
	}

	u := domain.NewUserFromOAuth(
		input.Email,
		input.FirstName,
		input.LastName,
		input.Username,
		input.PhotoURL,
		input.IsEmailVerified,
	)

	return s.saveUser(ctx, u)
}

// saveUser is an internal helper to persist a user entity and map repository errors.
func (s *Service) saveUser(ctx context.Context, u *domain.User) (uuid.UUID, error) {
	id, err := s.userRepo.Create(ctx, u)
	if err != nil {
		switch {
		case errors.Is(err, repouser.ErrUserEmailTaken):
			return uuid.Nil, ErrUserEmailTaken
		case errors.Is(err, repouser.ErrUserUsernameTaken):
			return uuid.Nil, ErrUserUsernameTaken
		case errors.Is(err, repouser.ErrUserInvalidRole):
			return uuid.Nil, ErrUserInvalidRole
		default:
			return uuid.Nil, fmt.Errorf("insert user: %w", err)
		}
	}
	return id, nil
}

// Update modifies an existing user's details based on the provided input.
func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByID(txCtx, input.ID)
		if err != nil {
			if errors.Is(err, repouser.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("get user for update: %w", err)
		}

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
			case errors.Is(err, repouser.ErrUserNotFound):
				return ErrUserNotFound
			case errors.Is(err, repouser.ErrUserUsernameTaken):
				return ErrUserUsernameTaken
			case errors.Is(err, repouser.ErrUserInvalidRole):
				return ErrUserInvalidRole
			default:
				return fmt.Errorf("update user: %w", err)
			}
		}

		return nil
	})

	return err
}

// SetNewPassword updates a user's password securely.
// This is used for password reset flows or profile settings updates.
func (s *Service) SetNewPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, repouser.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("get user: %w", err)
		}

		if err := u.SetPassword(newPassword); err != nil {
			if errors.Is(err, domain.ErrPasswordTooLong) {
				return ErrPasswordTooLong
			}
			return fmt.Errorf("set domain password: %w", err)
		}

		if err := s.userRepo.Update(txCtx, u); err != nil {
			return fmt.Errorf("update user password: %w", err)
		}

		return nil
	})

	return err
}

// MarkEmailVerified updates the user's status to indicate their email address is confirmed.
func (s *Service) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	err := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		u, err := s.userRepo.GetByID(txCtx, id)
		if err != nil {
			if errors.Is(err, repouser.ErrUserNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("get user: %w", err)
		}

		u.IsEmailVerified = true

		if err := s.userRepo.Update(txCtx, u); err != nil {
			return fmt.Errorf("update email verification status: %w", err)
		}

		return nil
	})

	return err
}

// GetByID retrieves user details by unique identifier.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repouser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by ID: %w", err)
	}
	return u, nil
}

// GetByEmail retrieves user details by their email address.
func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repouser.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// IsAdmin checks if a user has the admin role.
func (s *Service) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	isAdmin, err := s.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("check if user is admin: %w", err)
	}
	return isAdmin, nil
}
