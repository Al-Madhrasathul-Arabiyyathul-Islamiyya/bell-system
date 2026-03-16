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
		nil,
	)
	audioHandler := handlers.NewAudioHandler(
		&mocks.MockSystemAudioFileRepo{},
		nil,
	)
	systemHandler := handlers.NewSystemHandler(
		&mocks.MockSystemStateRepo{},
		&mocks.MockSchedulerService{},
		nil,
	)

	r := router.New(authHandler, userHandler, sessionHandler, scheduleHandler, audioHandler, systemHandler)
	require.NotNil(t, r)

	// Collect all registered routes
	routes := make(map[string]bool)
	err := chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes[method+" "+route] = true
		return nil
	})
	require.NoError(t, err)

	expectedRoutes := []string{
		"POST /api/v1/auth/login",
		"POST /api/v1/auth/change-password",
		"POST /api/v1/auth/logout",
		"GET /api/v1/users/",
		"POST /api/v1/users/",
		"GET /api/v1/users/{id}",
		"PUT /api/v1/users/{id}",
		"DELETE /api/v1/users/{id}",
		"GET /api/v1/sessions/",
		"POST /api/v1/sessions/",
		"GET /api/v1/sessions/current",
		"GET /api/v1/sessions/{id}",
		"PUT /api/v1/sessions/{id}",
		"DELETE /api/v1/sessions/{id}",
		"GET /api/v1/schedule-items/",
		"POST /api/v1/schedule-items/",
		"GET /api/v1/schedule-items/{id}",
		"PUT /api/v1/schedule-items/{id}",
		"DELETE /api/v1/schedule-items/{id}",
		"GET /api/v1/audio-files/",
		"GET /api/v1/audio-files/checksums",
		"POST /api/v1/audio-files/",
		"GET /api/v1/audio-files/{id}",
		"PUT /api/v1/audio-files/{id}",
		"DELETE /api/v1/audio-files/{id}",
		"GET /api/v1/system/state",
		"POST /api/v1/system/state",
		"POST /api/v1/system/cancel-next-bell",
	}

	for _, expected := range expectedRoutes {
		assert.True(t, routes[expected], "expected route %q to be registered", expected)
	}
}
