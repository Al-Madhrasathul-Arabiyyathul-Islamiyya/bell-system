package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAuthRouter creates an auth chi.Router with the given mocks.
func newAuthRouter(userRepo handlers.UserRepository, tokenSvc handlers.TokenService, hasher handlers.PasswordHasher) http.Handler {
	h := handlers.NewAuthHandler(userRepo, tokenSvc, hasher)
	return router.AuthRoutes(h)
}

// --- POST /login ---

func TestAuthHandler_Login_Success(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Username:     "admin",
		PasswordHash: "hashed_password",
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, username string) (*models.User, error) {
			assert.Equal(t, "admin", username)
			return user, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		CompareFunc: func(hash, password string) error {
			assert.Equal(t, "hashed_password", hash)
			assert.Equal(t, "password123", password)
			return nil
		},
	}
	tokenSvc := &mocks.MockTokenService{
		GenerateTokenFunc: func(u *models.User) (string, error) {
			assert.Equal(t, userID, u.ID)
			return "jwt-token-here", nil
		},
	}

	body := `{"username":"admin","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "jwt-token-here", resp.Token)
	assert.Equal(t, userID, resp.User.ID)
	assert.Equal(t, "admin", resp.User.Username)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}
	tokenSvc := &mocks.MockTokenService{}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp models.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_request", resp.Error.Code)
}

func TestAuthHandler_Login_MissingUsername(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}
	tokenSvc := &mocks.MockTokenService{}

	body := `{"password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_Login_MissingPassword(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}
	tokenSvc := &mocks.MockTokenService{}

	body := `{"username":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}
	tokenSvc := &mocks.MockTokenService{}

	body := `{"username":"nonexistent","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp models.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", resp.Error.Code)
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: "hashed_password",
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return user, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		CompareFunc: func(_, _ string) error {
			return errors.New("password mismatch")
		},
	}
	tokenSvc := &mocks.MockTokenService{}

	body := `{"username":"admin","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Login_TokenGenerationError(t *testing.T) {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: "hashed_password",
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return user, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		CompareFunc: func(_, _ string) error { return nil },
	}
	tokenSvc := &mocks.MockTokenService{
		GenerateTokenFunc: func(_ *models.User) (string, error) {
			return "", errors.New("token generation failed")
		},
	}

	body := `{"username":"admin","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAuthHandler_Login_DBError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, errors.New("database connection lost")
		},
	}
	hasher := &mocks.MockPasswordHasher{}
	tokenSvc := &mocks.MockTokenService{}

	body := `{"username":"admin","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- POST /change-password ---

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Username:     "admin",
		PasswordHash: "old_hash",
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			assert.Equal(t, userID, id)
			return user, nil
		},
		UpdateFunc: func(_ context.Context, u *models.User) error {
			assert.Equal(t, "new_hash", u.PasswordHash)
			return nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		CompareFunc: func(hash, password string) error {
			if hash == "old_hash" && password == "oldpassword" {
				return nil
			}
			return errors.New("mismatch")
		},
		HashFunc: func(password string) (string, error) {
			assert.Equal(t, "newpassword1", password)
			return "new_hash", nil
		},
	}
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return &handlers.TokenClaims{
				UserID:   userID,
				Username: "admin",
				Role:     models.RoleAdmin,
			}, nil
		},
	}

	body := `{"oldPassword":"oldpassword","newPassword":"newpassword1"}`
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthHandler_ChangePassword_Unauthorized(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return nil, errors.New("invalid token")
		},
	}

	body := `{"oldPassword":"old","newPassword":"newpassword1"}`
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_ChangePassword_WrongOldPassword(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Username:     "admin",
		PasswordHash: "old_hash",
		Role:         models.RoleAdmin,
	}

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return user, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		CompareFunc: func(_, _ string) error {
			return errors.New("password mismatch")
		},
	}
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return &handlers.TokenClaims{UserID: userID, Username: "admin", Role: models.RoleAdmin}, nil
		},
	}

	body := `{"oldPassword":"wrong","newPassword":"newpassword1"}`
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_ChangePassword_SamePassword(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Username:     "admin",
		PasswordHash: "old_hash",
		Role:         models.RoleAdmin,
	}

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return user, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		CompareFunc: func(_, _ string) error {
			return nil
		},
	}
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return &handlers.TokenClaims{UserID: userID, Username: "admin", Role: models.RoleAdmin}, nil
		},
	}

	body := `{"oldPassword":"samepass12","newPassword":"samepass12"}`
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_ChangePassword_HashError(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:           userID,
		Username:     "admin",
		PasswordHash: "old_hash",
		Role:         models.RoleAdmin,
	}

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return user, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		CompareFunc: func(_, _ string) error { return nil },
		HashFunc: func(_ string) (string, error) {
			return "", errors.New("hash error")
		},
	}
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return &handlers.TokenClaims{UserID: userID, Username: "admin", Role: models.RoleAdmin}, nil
		},
	}

	body := `{"oldPassword":"oldpassword","newPassword":"newpassword1"}`
	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- POST /logout ---

func TestAuthHandler_Logout_Success(t *testing.T) {
	userID := uuid.New()
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}
	tokenSvc := &mocks.MockTokenService{
		ValidateTokenFunc: func(_ string) (*handlers.TokenClaims, error) {
			return &handlers.TokenClaims{UserID: userID, Username: "admin", Role: models.RoleAdmin}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	r := newAuthRouter(userRepo, tokenSvc, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.SuccessResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}
