// Package auth implements authentication logic for the application.
//
// It handles Telegram-based login, JWT access token generation,
// refresh token issuance and rotation, and validation of Telegram auth data.
package auth

// TelegramLoginInput represents the data received from Telegram Login Widget
// or Telegram Mini App during authentication.
type TelegramLoginInput struct {
	ID        int64
	FirstName string
	LastName  *string
	Username  *string
	PhotoURL  *string
	AuthDate  int64
	Hash      string
}
