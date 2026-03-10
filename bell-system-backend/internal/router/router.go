package router

import (
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"

	"github.com/go-chi/chi/v5"
)

// New creates the main router with all routes registered.
func New(
	auth *handlers.AuthHandler,
	users *handlers.UserHandler,
	sessions *handlers.SessionHandler,
	schedule *handlers.ScheduleHandler,
	audio *handlers.AudioHandler,
) chi.Router {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", AuthRoutes(auth))
		r.Mount("/users", UserRoutes(users))
		r.Mount("/sessions", SessionRoutes(sessions))
		r.Mount("/schedule-items", ScheduleRoutes(schedule))
		r.Mount("/audio-files", AudioRoutes(audio))
	})

	return r
}

// AuthRoutes returns a chi.Router with authentication routes.
func AuthRoutes(h *handlers.AuthHandler) chi.Router {
	r := chi.NewRouter()
	// TODO: implement route handlers
	// POST /login
	// POST /change-password (auth required)
	// POST /logout (auth required)
	return r
}

// UserRoutes returns a chi.Router with user management routes.
func UserRoutes(h *handlers.UserHandler) chi.Router {
	r := chi.NewRouter()
	// TODO: implement route handlers (admin only)
	// GET / (list)
	// POST / (create)
	// GET /{id}
	// PUT /{id}
	// DELETE /{id}
	return r
}

// SessionRoutes returns a chi.Router with session routes.
func SessionRoutes(h *handlers.SessionHandler) chi.Router {
	r := chi.NewRouter()
	// TODO: implement route handlers
	// GET / (list)
	// POST / (create)
	// GET /current
	// GET /{id}
	// PUT /{id}
	// DELETE /{id}
	return r
}

// ScheduleRoutes returns a chi.Router with schedule item routes.
func ScheduleRoutes(h *handlers.ScheduleHandler) chi.Router {
	r := chi.NewRouter()
	// TODO: implement route handlers
	// GET / (list)
	// POST / (create)
	// GET /{id}
	// PUT /{id}
	// DELETE /{id}
	return r
}

// AudioRoutes returns a chi.Router with audio file routes.
func AudioRoutes(h *handlers.AudioHandler) chi.Router {
	r := chi.NewRouter()
	// TODO: implement route handlers
	// GET / (list)
	// POST / (upload, multipart)
	// GET /{id}
	// PUT /{id}
	// DELETE /{id}
	return r
}
