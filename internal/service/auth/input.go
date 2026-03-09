// Package auth implements authentication and session management logic.
//
// It supports Telegram-based authentication, JWT issuance, refresh token rotation,
// session persistence, and revocation.
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
