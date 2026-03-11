package unit_test

import (
	"net/http"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterNew_AllRoutesRegistered(t *testing.T) {
	// Build handlers with no-op mocks (routes just need to be registered, not called)
	authHandler := handlers.NewAuthHandler(
		&mocks.MockUserRepo{},
		&mocks.MockTokenService{},
		&mocks.MockPasswordHasher{},
	)
	userHandler := handlers.NewUserHandler(
		&mocks.MockUserRepo{},
		&mocks.MockPasswordHasher{},
	)
	sessionHandler := handlers.NewSessionHandler(
		&mocks.MockSessionRepo{},
	)
	scheduleHandler := handlers.NewScheduleHandler(
		&mocks.MockScheduleItemRepo{},
		&mocks.MockScheduleDayRepo{},
	)
	audioHandler := handlers.NewAudioHandler(
		&mocks.MockSystemAudioFileRepo{},
	)

	r := router.New(authHandler, userHandler, sessionHandler, scheduleHandler, audioHandler)
	require.NotNil(t, r)

	// Collect all registered routes
	routes := make(map[string]bool)
	err := chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes[method+" "+route] = true
		return nil
	})
	require.NoError(t, err)

	expectedRoutes := []string{
		"POST /api/auth/login",
		"POST /api/auth/change-password",
		"POST /api/auth/logout",
		"GET /api/users/",
		"POST /api/users/",
		"GET /api/users/{id}",
		"PUT /api/users/{id}",
		"DELETE /api/users/{id}",
		"GET /api/sessions/",
		"POST /api/sessions/",
		"GET /api/sessions/current",
		"GET /api/sessions/{id}",
		"PUT /api/sessions/{id}",
		"DELETE /api/sessions/{id}",
		"GET /api/schedule-items/",
		"POST /api/schedule-items/",
		"GET /api/schedule-items/{id}",
		"PUT /api/schedule-items/{id}",
		"DELETE /api/schedule-items/{id}",
		"GET /api/audio-files/",
		"GET /api/audio-files/checksums",
		"POST /api/audio-files/",
		"GET /api/audio-files/{id}",
		"PUT /api/audio-files/{id}",
		"DELETE /api/audio-files/{id}",
	}

	for _, expected := range expectedRoutes {
		assert.True(t, routes[expected], "expected route %q to be registered", expected)
	}
}
