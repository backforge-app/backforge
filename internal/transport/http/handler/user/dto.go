package user

import (
	"github.com/google/uuid"

	"github.com/backforge-app/backforge/internal/domain"
)

// userResponse represents the JSON payload for user profile details.
type userResponse struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"is_email_verified"`
	HasPassword     bool      `json:"has_password"`
	FirstName       string    `json:"first_name"`
	LastName        *string   `json:"last_name,omitempty"`
	Username        *string   `json:"username,omitempty"`
	PhotoURL        *string   `json:"photo_url,omitempty"`
	Role            string    `json:"role"`
}

// updateProfileRequest represents the JSON payload for updating a user's profile.
type updateProfileRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=64"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=64"`
	Username  *string `json:"username" validate:"omitempty,min=3,max=32"`
	PhotoURL  *string `json:"photo_url" validate:"omitempty,url"`
}

// toUserResponse converts a domain.User entity to a userResponse DTO.
// It securely maps internal domain state (like password existence) to frontend-friendly booleans.
func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:              u.ID,
		Email:           u.Email,
		IsEmailVerified: u.IsEmailVerified,
		HasPassword:     u.HasPassword(), // used by frontend to determine UI states (e.g., Change vs Set Password)
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		Username:        u.Username,
		PhotoURL:        u.PhotoURL,
		Role:            string(u.Role),
	}
}
