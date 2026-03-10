package handlers

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	Users     UserRepository
	Passwords PasswordHasher
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(users UserRepository, passwords PasswordHasher) *UserHandler {
	return &UserHandler{
		Users:     users,
		Passwords: passwords,
	}
}
