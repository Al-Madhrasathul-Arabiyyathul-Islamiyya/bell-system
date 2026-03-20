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
	pkgerrors "arabiyya.edu.mv/bell-system-backend/pkg/errors"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"
	"arabiyya.edu.mv/bell-system-backend/tests/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newUserRouter creates a user chi.Router with the given mocks.
func newUserRouter(userRepo handlers.UserRepository, hasher handlers.PasswordHasher) http.Handler {
	h := handlers.NewUserHandler(userRepo, hasher)
	return router.UserRoutes(h, testutil.PermissiveTokenService)
}

// authReq adds a valid auth header to the request.
func authReq(req *http.Request) *http.Request {
	testutil.SetAuthHeader(req)
	return req
}

// jsonapiBody builds a JSON:API request body.
func jsonapiBody(typ string, attrs any, rels ...map[string]any) string {
	data := map[string]any{"type": typ, "attributes": attrs}
	if len(rels) > 0 {
		data["relationships"] = rels[0]
	}
	body, _ := json.Marshal(map[string]any{"data": data})
	return string(body)
}

// --- GET / (list users) ---

func TestUserHandler_List_Success(t *testing.T) {
	now := time.Now()
	users := []*models.User{
		{ID: uuid.New(), Username: "admin", Role: models.RoleAdmin, CreatedAt: now},
		{ID: uuid.New(), Username: "morning", Role: models.RoleMorningUser, CreatedAt: now},
	}

	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context, page, size int, filterRole, sortSQL string) ([]*models.User, int, error) {
			return users, len(users), nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, jsonapi.ContentType, rr.Header().Get("Content-Type"))

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp.Meta["total"])
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "users", resp.Data[0].Type)
}

func TestUserHandler_List_Empty(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context, _, _ int, _, _ string) ([]*models.User, int, error) {
			return []*models.User{}, 0, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(0), resp.Meta["total"])
	assert.Empty(t, resp.Data)
}

func TestUserHandler_List_DBError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context, _, _ int, _, _ string) ([]*models.User, int, error) {
			return nil, 0, errors.New("database error")
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/", nil))
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

	req := authReq(httptest.NewRequest(http.MethodGet, "/"+userID.String(), nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, jsonapi.ContentType, rr.Header().Get("Content-Type"))

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "users", resp.Data.Type)
	assert.Equal(t, userID.String(), resp.Data.ID)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp jsonapi.ErrorDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "not_found", resp.Errors[0].Code)
}

func TestUserHandler_GetByID_InvalidUUID(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil))
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

	body := jsonapiBody("users", map[string]string{"username": "newuser", "password": "password123", "role": "admin"})
	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, jsonapi.ContentType, rr.Header().Get("Content-Type"))

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "users", resp.Data.Type)
	assert.NotEmpty(t, resp.Data.ID)
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{broken`)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("users", map[string]string{"username": "existing", "password": "password123", "role": "admin"})
	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp jsonapi.ErrorDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "conflict", resp.Errors[0].Code)
}

func TestUserHandler_Create_MissingFields(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"username": "onlyuser"})
	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Create_InvalidRole(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"username": "user", "password": "password123", "role": "superadmin"})
	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("users", map[string]string{"username": "newname"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("users", map[string]string{"username": "newname"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	req := authReq(httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(`{bad`)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	req := authReq(httptest.NewRequest(http.MethodDelete, "/"+userID.String(), nil))
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

	req := authReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
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

	req := authReq(httptest.NewRequest(http.MethodDelete, "/"+userID.String(), nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Error path tests ---

func TestUserHandler_GetByID_ErrNotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUserHandler_GetByID_DBError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, errors.New("database error")
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_Create_PasswordTooShort(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"username": "user", "password": "short", "role": "admin"})
	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Create_HashError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		HashFunc: func(_ string) (string, error) {
			return "", errors.New("hash failed")
		},
	}

	body := jsonapiBody("users", map[string]string{"username": "user", "password": "password123", "role": "admin"})
	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_Create_DBError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByUsernameFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, nil
		},
		CreateFunc: func(_ context.Context, _ *models.User) error {
			return errors.New("create failed")
		},
	}
	hasher := &mocks.MockPasswordHasher{
		HashFunc: func(_ string) (string, error) {
			return "hashed", nil
		},
	}

	body := jsonapiBody("users", map[string]string{"username": "user", "password": "password123", "role": "admin"})
	req := authReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_Update_InvalidUUID(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"username": "newname"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/not-a-uuid", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Update_ErrNotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"username": "newname"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUserHandler_Update_GetByIDDBError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, errors.New("database error")
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"username": "newname"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_Update_WithPassword(t *testing.T) {
	userID := uuid.New()
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{ID: userID, Username: "old"}, nil
		},
		UpdateFunc: func(_ context.Context, user *models.User) error {
			assert.Equal(t, "new_hash", user.PasswordHash)
			return nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		HashFunc: func(_ string) (string, error) {
			return "new_hash", nil
		},
	}

	body := jsonapiBody("users", map[string]string{"password": "newpassword1"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandler_Update_PasswordHashError(t *testing.T) {
	userID := uuid.New()
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{
		HashFunc: func(_ string) (string, error) {
			return "", errors.New("hash failed")
		},
	}

	body := jsonapiBody("users", map[string]string{"password": "newpassword1"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_Update_WithRole(t *testing.T) {
	userID := uuid.New()
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{ID: userID, Username: "user", Role: models.RoleAdmin}, nil
		},
		UpdateFunc: func(_ context.Context, user *models.User) error {
			assert.Equal(t, models.RoleMorningUser, user.Role)
			return nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"role": "morning_user"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandler_Update_DBError(t *testing.T) {
	userID := uuid.New()
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
		UpdateFunc: func(_ context.Context, _ *models.User) error {
			return errors.New("update failed")
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	body := jsonapiBody("users", map[string]string{"username": "newname"})
	req := authReq(httptest.NewRequest(http.MethodPut, "/"+userID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_Delete_InvalidUUID(t *testing.T) {
	userRepo := &mocks.MockUserRepo{}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Delete_ErrNotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUserHandler_Delete_GetByIDDBError(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, errors.New("database error")
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Pagination/filter/sort tests ---

func TestUserHandler_List_WithPagination(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context, page, size int, _, _ string) ([]*models.User, int, error) {
			assert.Equal(t, 2, page)
			assert.Equal(t, 5, size)
			return []*models.User{
				{ID: uuid.New(), Username: "user6", Role: models.RoleAdmin, CreatedAt: time.Now()},
			}, 11, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/?page[number]=2&page[size]=5", nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(11), resp.Meta["total"])
	assert.Equal(t, float64(2), resp.Meta["page"])
	assert.Equal(t, float64(5), resp.Meta["pageSize"])
	assert.Equal(t, float64(3), resp.Meta["totalPages"])
	assert.Contains(t, resp.Links["self"], "page[number]=2")
}

func TestUserHandler_List_WithFilter(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context, _, _ int, filterRole, _ string) ([]*models.User, int, error) {
			assert.Equal(t, "admin", filterRole)
			return []*models.User{
				{ID: uuid.New(), Username: "admin", Role: models.RoleAdmin, CreatedAt: time.Now()},
			}, 1, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/?filter[role]=admin", nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp.Meta["total"])
}

func TestUserHandler_List_WithSort(t *testing.T) {
	userRepo := &mocks.MockUserRepo{
		ListFunc: func(_ context.Context, _, _ int, _, sortSQL string) ([]*models.User, int, error) {
			assert.Equal(t, "ORDER BY CreatedAt DESC", sortSQL)
			return []*models.User{}, 0, nil
		},
	}
	hasher := &mocks.MockPasswordHasher{}

	req := authReq(httptest.NewRequest(http.MethodGet, "/?sort=-createdAt", nil))
	rr := httptest.NewRecorder()

	r := newUserRouter(userRepo, hasher)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}
