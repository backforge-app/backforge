// Package auth provides HTTP request and response DTOs for authentication handlers.
package auth

// loginRequest represents the JSON payload for Telegram login.
type loginRequest struct {
	ID        int64   `json:"id" validate:"required"`         // Telegram user ID
	FirstName string  `json:"first_name" validate:"required"` // First name
	LastName  *string `json:"last_name,omitempty"`            // Last name (optional)
	Username  *string `json:"username,omitempty"`             // Telegram username (optional)
	PhotoURL  *string `json:"photo_url,omitempty"`            // Avatar URL (optional)
	AuthDate  int64   `json:"auth_date" validate:"required"`  // Authorization timestamp
	Hash      string  `json:"hash" validate:"required"`       // Telegram login hash
}

// loginResponse contains the access and refresh tokens returned after login.
type loginResponse struct {
	AccessToken  string `json:"access_token"`  // JWT token
	RefreshToken string `json:"refresh_token"` // Refresh token
}

// refreshRequest represents the JSON payload to refresh tokens.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"` // Existing refresh token
}

// refreshResponse contains new access and refresh tokens.
type refreshResponse struct {
	AccessToken  string `json:"access_token"`  // New JWT token
	RefreshToken string `json:"refresh_token"` // New refresh token
}
