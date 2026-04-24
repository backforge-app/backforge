package user

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// CreateWithPasswordInput holds the data required to register a user via email and password.
type CreateWithPasswordInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  *string
	Username  *string
	PhotoURL  *string
}

// CreateOAuthInput holds the data required to register a user via a third-party provider.
type CreateOAuthInput struct {
	Email           string
	FirstName       string
	LastName        *string
	Username        *string
	PhotoURL        *string
	IsEmailVerified bool
}

// UpdateInput holds data for modifying an existing user's profile.
// Only non-nil fields will be applied.
type UpdateInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Username  *string
	PhotoURL  *string
	Role      *domain.UserRole
}
