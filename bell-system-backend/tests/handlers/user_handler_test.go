package handlers_test

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

// newUserRouter creates a user chi.Router with the given mocks.
func newUserRouter(userRepo handlers.UserRepository, hasher handlers.PasswordHasher) http.Handler {
	h := handlers.NewUserHandler(userRepo, hasher)
	return router.UserRoutes(h)
}

// --- GET / (list users) ---

func TestUserHandler_List_Success(t *testing.T) {
	now := time.Now()
	users := []*models.User{
		{ID: uuid.New(), Username: "admin", Role: models.RoleAdmin, CreatedAt: now},
		{ID: uuid.New(), Username: "morning", Role: models.RoleMorningUser, CreatedAt: now},
	}

	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context) ([]*models.User, error) {
			return users, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Total int           `json:"total"`
		Items []models.User `json:"items"`
	}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Items, 2)
}

func TestUserHandler_List_Empty(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context) ([]*models.User, error) {
			return []*models.User{}, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Total int           `json:"total"`
		Items []models.User `json:"items"`
	}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Items)
}

func TestUserHandler_List_DBError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context) ([]*models.User, error) {
			return nil, errors.New("database error")
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- GET /{id} (get user by ID) ---

func TestUserHandler_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "admin",
		Role:     models.RoleAdmin,
	}

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			assert.Equal(t, userID, id)
			return user, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodGet, "/"+userID.String(), nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.User
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, userID, resp.ID)
	assert.Equal(t, "admin", resp.Username)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp models.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "not_found", resp.Error.Code)
}

func TestUserHandler_GetByID_InvalidUUID(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- POST / (create user) ---

func TestUserHandler_Create_Success(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, nil
		},
		CreateFunc: func(_ context.Context, user *models.User) error {
			assert.Equal(t, "newuser", user.Username)
			assert.Equal(t, models.RoleAdmin, user.Role)
			assert.NotEmpty(t, user.PasswordHash)
			return nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		HashFunc: func(password string) (string, error) {
			return "hashed_" + password, nil
		},
	}

	body := `{"username":"newuser","password":"password123","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp models.User
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "newuser", resp.Username)
	assert.Equal(t, models.RoleAdmin, resp.Role)
	assert.NotEqual(t, uuid.Nil, resp.ID)
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{broken`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Create_DuplicateUsername(t *testing.T) {
	existing := &models.User{
		ID:       uuid.New(),
		Username: "existing",
		Role:     models.RoleAdmin,
	}

	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return existing, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	body := `{"username":"existing","password":"password123","role":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp models.ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "conflict", resp.Error.Code)
}

func TestUserHandler_Create_MissingFields(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	body := `{"username":"onlyuser"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Create_InvalidRole(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	body := `{"username":"user","password":"password123","role":"superadmin"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- PUT /{id} (update user) ---

func TestUserHandler_Update_Success(t *testing.T) {
	userID := uuid.New()
	existing := &models.User{
		ID:           userID,
		Username:     "oldname",
		PasswordHash: "oldhash",
		Role:         models.RoleAdmin,
	}

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return existing, nil
		},
		UpdateFunc: func(_ context.Context, user *models.User) error {
			assert.Equal(t, "newname", user.Username)
			return nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	body := `{"username":"newname"}`
	req := httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandler_Update_NotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	body := `{"username":"newname"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUserHandler_Update_InvalidJSON(t *testing.T) {
	userID := uuid.New()
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- DELETE /{id} (delete user) ---

func TestUserHandler_Delete_Success(t *testing.T) {
	userID := uuid.New()

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, id uuid.UUID) error {
			assert.Equal(t, userID, id)
			return nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodDelete, "/"+userID.String(), nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUserHandler_Delete_DBError(t *testing.T) {
	userID := uuid.New()

	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
		DeleteFunc: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("foreign key constraint")
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := httptest.NewRequest(http.MethodDelete, "/"+userID.String(), nil)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
