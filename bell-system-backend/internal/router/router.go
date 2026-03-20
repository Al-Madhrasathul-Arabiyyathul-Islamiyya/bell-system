package router

import (
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// New creates the main router with all API routes registered.
func New(
	tokenSvc handlers.TokenService,
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
		r.Mount("/users", UserRoutes(users, tokenSvc))
		r.Mount("/sessions", SessionRoutes(sessions, tokenSvc))
		r.Mount("/schedule", ScheduleRoutes(schedule, tokenSvc))
		r.Mount("/audio", AudioRoutes(audio, tokenSvc))
		r.Mount("/system", SystemRoutes(system, tokenSvc))
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
// All endpoints require admin role.
func UserRoutes(h *handlers.UserHandler, tokenSvc handlers.TokenService) chi.Router {
	r := chi.NewRouter()
	r.Use(handlers.AuthMiddleware(tokenSvc))
	r.Use(handlers.RequireRole(models.RoleAdmin))
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// SessionRoutes returns a chi.Router with session routes.
// GET /current is public; all other endpoints require authentication.
func SessionRoutes(h *handlers.SessionHandler, tokenSvc handlers.TokenService) chi.Router {
	r := chi.NewRouter()
	r.Get("/current", h.GetCurrent)
	r.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware(tokenSvc))
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
	return r
}

// ScheduleRoutes returns a chi.Router with schedule item routes.
// GET /current is public; all other endpoints require authentication.
func ScheduleRoutes(h *handlers.ScheduleHandler, tokenSvc handlers.TokenService) chi.Router {
	r := chi.NewRouter()
	r.Get("/current", h.GetCurrent)
	r.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware(tokenSvc))
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
	return r
}

// SystemRoutes returns a chi.Router with system management routes.
// GET /state is public; POST endpoints require authentication.
func SystemRoutes(h *handlers.SystemHandler, tokenSvc handlers.TokenService) chi.Router {
	r := chi.NewRouter()
	r.Get("/state", h.GetState)
	r.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware(tokenSvc))
		r.Post("/state", h.SetState)
		r.Post("/cancel-next-bell", h.CancelNextBell)
	})
	return r
}

// AudioRoutes returns a chi.Router with audio file routes.
// GET /checksums and GET /{id} are public; all other endpoints require authentication.
func AudioRoutes(h *handlers.AudioHandler, tokenSvc handlers.TokenService) chi.Router {
	r := chi.NewRouter()
	r.Get("/checksums", h.ListChecksums)
	r.Get("/{id}", h.GetByID)
	r.Get("/{id}/content", h.GetContent)
	r.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware(tokenSvc))
		r.Get("/", h.List)
		r.Post("/", h.Upload)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
	return r
}
