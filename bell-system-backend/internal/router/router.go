package router

import (
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"

	"github.com/go-chi/chi/v5"
)

// New creates the main router with all API routes registered.
func New(
	auth *handlers.AuthHandler,
	users *handlers.UserHandler,
	sessions *handlers.SessionHandler,
	schedule *handlers.ScheduleHandler,
	audio *handlers.AudioHandler,
	system *handlers.SystemHandler,
) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/auth", AuthRoutes(auth))
		r.Mount("/users", UserRoutes(users))
		r.Mount("/sessions", SessionRoutes(sessions))
		r.Mount("/schedule-items", ScheduleRoutes(schedule))
		r.Mount("/audio-files", AudioRoutes(audio))
		r.Mount("/system", SystemRoutes(system))
	})

	return r
}

// AuthRoutes returns a chi.Router with authentication routes.
func AuthRoutes(h *handlers.AuthHandler) chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware(h.Tokens))
		r.Post("/change-password", h.ChangePassword)
		r.Post("/logout", h.Logout)
	})
	return r
}

// UserRoutes returns a chi.Router with user management routes.
func UserRoutes(h *handlers.UserHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// SessionRoutes returns a chi.Router with session routes.
func SessionRoutes(h *handlers.SessionHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/current", h.GetCurrent)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// ScheduleRoutes returns a chi.Router with schedule item routes.
func ScheduleRoutes(h *handlers.ScheduleHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/current", h.GetCurrent)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// SystemRoutes returns a chi.Router with system management routes.
func SystemRoutes(h *handlers.SystemHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/state", h.GetState)
	r.Post("/state", h.SetState)
	r.Post("/cancel-next-bell", h.CancelNextBell)
	return r
}

// AudioRoutes returns a chi.Router with audio file routes.
func AudioRoutes(h *handlers.AudioHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/checksums", h.ListChecksums)
	r.Post("/", h.Upload)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}
