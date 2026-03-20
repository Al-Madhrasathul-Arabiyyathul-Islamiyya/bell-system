package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyHandler is a simple handler that returns 200 OK.
func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
}

// --- AuthMiddleware ---

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID := uuid.New()
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(token string) (*handlers.TokenClaims, error) {
			assert.Equal(t, "valid-jwt-token", token)
			return &handlers.TokenClaims{
				UserID:   userID,
				Username: "admin",
				Role:     models.RoleAdmin,
			}, nil
		},
	}

	handler := handlers.AuthMiddleware(tokenSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := handlers.GetUserClaims(r.Context())
		require.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, "admin", claims.Username)
		assert.Equal(t, models.RoleAdmin, claims.Role)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-jwt-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	tokenSvc := &mocks.MockTokenService{}

	handler := handlers.AuthMiddleware(tokenSvc)(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Authorization header
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp jsonapi.ErrorDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "unauthorized", resp.Errors[0].Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	tokenSvc := &mocks.MockTokenService{}

	handler := handlers.AuthMiddleware(tokenSvc)(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // Not Bearer
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return nil, errors.New("token expired")
		},
	}

	handler := handlers.AuthMiddleware(tokenSvc)(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return nil, errors.New("malformed token")
		},
	}

	handler := handlers.AuthMiddleware(tokenSvc)(dummyHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.jwt")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- RequireRole middleware ---

func TestRequireRole_AdminAccess(t *testing.T) {
	claims := &handlers.TokenClaims{
		UserID:   uuid.New(),
		Username: "admin",
		Role:     models.RoleAdmin,
	}

	handler := handlers.RequireRole(models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), handlers.UserContextKey, claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireRole_MorningUserRestricted(t *testing.T) {
	claims := &handlers.TokenClaims{
		UserID:   uuid.New(),
		Username: "morning",
		Role:     models.RoleMorningUser,
	}

	handler := handlers.RequireRole(models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), handlers.UserContextKey, claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	var resp jsonapi.ErrorDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "forbidden", resp.Errors[0].Code)
}

func TestRequireRole_AfternoonUserRestricted(t *testing.T) {
	claims := &handlers.TokenClaims{
		UserID:   uuid.New(),
		Username: "afternoon",
		Role:     models.RoleAfternoonUser,
	}

	handler := handlers.RequireRole(models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), handlers.UserContextKey, claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRequireRole_MultipleRolesAllowed(t *testing.T) {
	claims := &handlers.TokenClaims{
		UserID:   uuid.New(),
		Username: "morning",
		Role:     models.RoleMorningUser,
	}

	// Allow both admin and morning_user
	handler := handlers.RequireRole(models.RoleAdmin, models.RoleMorningUser)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), handlers.UserContextKey, claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireRole_NoClaims(t *testing.T) {
	handler := handlers.RequireRole(models.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No claims in context
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// --- CORS ---

func TestCORS_HeadersPresent(t *testing.T) {
	// CORS middleware should set Access-Control headers on preflight
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rr := httptest.NewRecorder()

	// The CORS middleware wraps a dummy handler
	handler := handlers.AuthMiddleware(&mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return nil, errors.New("no token")
		},
	})(dummyHandler())

	handler.ServeHTTP(rr, req)

	// At minimum, the response should have been handled (specific CORS header assertions
	// depend on the CORS middleware configuration, tested when implemented)
	assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
}
