// Package user provides HTTP request and response DTOs for user handlers.
package user

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// userResponse represents the JSON payload for user profile details.
type userResponse struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  string    `json:"first_name"`
	LastName   *string   `json:"last_name,omitempty"`
	Username   *string   `json:"username,omitempty"`
	PhotoURL   *string   `json:"photo_url,omitempty"`
	Role       string    `json:"role"`
}

// toUserResponse converts a domain.User entity to a userResponse DTO.
func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:         u.ID,
		TelegramID: u.TelegramID,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Username:   u.Username,
		PhotoURL:   u.PhotoURL,
		Role:       string(u.Role),
	}
}
