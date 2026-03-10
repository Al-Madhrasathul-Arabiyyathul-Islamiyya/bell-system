package handlers

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	Users     UserRepository
	Tokens    TokenService
	Passwords PasswordHasher
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(users UserRepository, tokens TokenService, passwords PasswordHasher) *AuthHandler {
	return &AuthHandler{
		Users:     users,
		Tokens:    tokens,
		Passwords: passwords,
	}
}
