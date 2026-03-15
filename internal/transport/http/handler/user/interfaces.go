// Package user provides the interface to the user service for HTTP handlers.
//
//go:generate mockgen -package=user -destination=mocks.go github.com/backforge-app/backforge/internal/transport/http/handler/user Service
package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// Service defines the interface that HTTP handlers use to perform user operations.
type Service interface {
	// GetByID retrieves user details by unique identifier.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
