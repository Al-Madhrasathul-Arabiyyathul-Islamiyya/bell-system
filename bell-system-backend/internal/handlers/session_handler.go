package handlers

// SessionHandler handles session management HTTP requests.
type SessionHandler struct {
	Sessions SessionRepository
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessions SessionRepository) *SessionHandler {
	return &SessionHandler{Sessions: sessions}
}
