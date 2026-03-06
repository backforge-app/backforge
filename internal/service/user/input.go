// Package user implements the User application service.
//
// It contains the business logic for managing users, including
// service methods, input DTOs, service-level errors,
// repository interfaces and service-level tests.
// The package coordinates domain entities with repository
// implementations defined in the parent service layer.
package user

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// CreateInput holds data for creating a new user.
type CreateInput struct {
	TelegramID int64
	FirstName  string
	LastName   *string
	Username   *string
	PhotoURL   *string
	IsPro      bool
}

// UpdateInput holds data for updating an existing user.
type UpdateInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Username  *string
	PhotoURL  *string
	Role      *domain.UserRole
	IsPro     *bool
}
