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
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSessionRouter creates a session chi.Router with the given mock.
func newSessionRouter(sessionRepo handlers.SessionRepository) http.Handler {
	h := handlers.NewSessionHandler(sessionRepo)
	return router.SessionRoutes(h)
}

// --- GET / (list sessions) ---

func TestSessionHandler_List_Success(t *testing.T) {
	now := time.Now()
	sessions := []*models.Session{
		{ID: uuid.New(), Name: "Morning", StartTime: now, EndTime: now.Add(4 * time.Hour)},
		{ID: uuid.New(), Name: "Afternoon", StartTime: now.Add(5 * time.Hour), EndTime: now.Add(9 * time.Hour)},
	}

	repo := &mocks.MockSessionRepo{
		ListFunc: func(_ context.Context) ([]*models.Session, error) {
			return sessions, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Total int              `json:"total"`
		Items []models.Session `json:"items"`
	}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Items, 2)
}

func TestSessionHandler_List_Empty(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		ListFunc: func(_ context.Context) ([]*models.Session, error) {
			return []*models.Session{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

// --- GET /{id} (get session by ID) ---

func TestSessionHandler_GetByID_Success(t *testing.T) {
	sessionID := uuid.New()
	session := &models.Session{
		ID:        sessionID,
		Name:      "Morning",
		StartTime: time.Date(2026, 1, 1, 7, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Session, error) {
			assert.Equal(t, sessionID, id)
			return session, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+sessionID.String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.Session
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, sessionID, resp.ID)
	assert.Equal(t, "Morning", resp.Name)
}

func TestSessionHandler_GetByID_NotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSessionHandler_GetByID_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	req := httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- GET /current (get current session) ---

func TestSessionHandler_GetCurrent_Success(t *testing.T) {
	session := &models.Session{
		ID:        uuid.New(),
		Name:      "Morning",
		StartTime: time.Date(2026, 1, 1, 7, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	repo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return session, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.Session
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Morning", resp.Name)
}

func TestSessionHandler_GetCurrent_NoActiveSession(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- POST / (create session) ---

func TestSessionHandler_Create_Success(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		CreateFunc: func(_ context.Context, session *models.Session) error {
			assert.Equal(t, "Evening", session.Name)
			return nil
		},
	}

	body := `{"name":"Evening","startTime":"17:00","endTime":"20:00"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestSessionHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Create_OverlappingTimes(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		CreateFunc: func(_ context.Context, _ *models.Session) error {
			return errors.New("overlapping session times")
		},
	}

	body := `{"name":"Overlap","startTime":"07:00","endTime":"13:00"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

// --- PUT /{id} (update session) ---

func TestSessionHandler_Update_Success(t *testing.T) {
	sessionID := uuid.New()

	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Session, error) {
			return &models.Session{ID: id, Name: "Old"}, nil
		},
		UpdateFunc: func(_ context.Context, session *models.Session) error {
			assert.Equal(t, "Updated", session.Name)
			return nil
		},
	}

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+sessionID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestSessionHandler_Update_NotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, nil
		},
	}

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- DELETE /{id} (delete session) ---

func TestSessionHandler_Delete_Success(t *testing.T) {
	sessionID := uuid.New()

	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Session, error) {
			return &models.Session{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, id uuid.UUID) error {
			assert.Equal(t, sessionID, id)
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+sessionID.String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestSessionHandler_Delete_NotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- Error path tests ---

func TestSessionHandler_List_DBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		ListFunc: func(_ context.Context) ([]*models.Session, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_GetCurrent_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSessionHandler_GetCurrent_DBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_GetByID_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSessionHandler_GetByID_DBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Update_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/not-a-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Update_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSessionHandler_Update_GetByIDDBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, errors.New("database error")
		},
	}

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return &models.Session{ID: uuid.New()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Update_InvalidStartTime(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return &models.Session{ID: uuid.New()}, nil
		},
	}

	body := `{"startTime":"not-a-time"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Update_InvalidEndTime(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return &models.Session{ID: uuid.New()}, nil
		},
	}

	body := `{"endTime":"not-a-time"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Update_DBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return &models.Session{ID: uuid.New()}, nil
		},
		UpdateFunc: func(_ context.Context, _ *models.Session) error {
			return errors.New("update failed")
		},
	}

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Delete_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	req := httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Delete_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSessionHandler_Delete_GetByIDDBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Delete_DBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Session, error) {
			return &models.Session{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("delete failed")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Create_InvalidStartTime(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := `{"name":"Test","startTime":"not-time","endTime":"18:00"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Create_InvalidEndTime(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := `{"name":"Test","startTime":"07:00","endTime":"not-time"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Create_MissingFields(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := `{"name":"OnlyName"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
