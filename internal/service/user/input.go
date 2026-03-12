// Package user implements the application service layer for user management.
//
// It contains business logic for user creation, updates, retrieval,
// service-level errors, input DTOs (in other files), and coordinates
// domain entities with the persistence layer.
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
}

// UpdateInput holds data for updating an existing user.
type UpdateInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Username  *string
	PhotoURL  *string
	Role      *domain.UserRole
}
