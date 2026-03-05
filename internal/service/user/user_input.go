// Package user implements the User application service.
//
// It contains the business logic for managing users, including
// service methods, input DTOs, and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
package user

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// CreateUserInput holds data for creating a new user.
type CreateUserInput struct {
	TgUserID  int64
	FirstName string
	LastName  *string
	Username  *string
	IsPro     bool
}

// UpdateUserInput holds data for updating an existing user.
type UpdateUserInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Username  *string
	Role      *domain.UserRole
	IsPro     *bool
}
