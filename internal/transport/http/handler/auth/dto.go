package auth

// registerRequest represents the JSON payload for a new user registration.
type registerRequest struct {
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8,max=72"`
	FirstName string  `json:"first_name" validate:"required,min=2,max=64"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2,max=64"`
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=32"`
}

// loginRequest represents the JSON payload for authenticating a user.
type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// verifyEmailRequest represents the payload for email verification.
type verifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// requestResetRequest represents the payload to request a password reset email.
type requestResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// resetPasswordRequest represents the payload to set a new password using a token.
type resetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

// yandexCallbackRequest represents the payload sent by the frontend after Yandex redirects back.
type yandexCallbackRequest struct {
	Code string `json:"code" validate:"required"`
}

// refreshRequest represents the JSON payload to exchange an old refresh token for new ones.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// tokenResponse contains the access and refresh tokens returned after a successful login or refresh.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// resendVerificationRequest represents the payload to request a new verification email.
type resendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}
