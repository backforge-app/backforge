package auth

// RegisterInput holds the user-provided data for standard email registration.
type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  *string
	Username  *string
}

// LoginInput holds the credentials for standard email login.
type LoginInput struct {
	Email    string
	Password string
}

// OAuthProfile represents unified user data received from any OAuth provider.
type OAuthProfile struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}
