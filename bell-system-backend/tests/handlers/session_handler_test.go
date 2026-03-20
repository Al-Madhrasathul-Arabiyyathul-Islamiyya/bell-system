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

// newSessionRouter creates a session chi.Router with the given mock.
func newSessionRouter(sessionRepo handlers.SessionRepository) http.Handler {
	h := handlers.NewSessionHandler(sessionRepo)
	return router.SessionRoutes(h, testutil.PermissiveTokenService)
}

func authSessionReq(req *http.Request) *http.Request {
	testutil.SetAuthHeader(req)
	return req
}

// --- GET / (list sessions) ---

func TestSessionHandler_List_Success(t *testing.T) {
	now := time.Now()
	sessions := []*models.Session{
		{ID: uuid.New(), Name: "Morning", StartTime: now, EndTime: now.Add(4 * time.Hour)},
		{ID: uuid.New(), Name: "Afternoon", StartTime: now.Add(5 * time.Hour), EndTime: now.Add(9 * time.Hour)},
	}

	repo := &mocks.MockSessionRepo{
		ListFunc: func(_ context.Context, _ string) ([]*models.Session, error) {
			return sessions, nil
		},
	}

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, jsonapi.ContentType, rr.Header().Get("Content-Type"))

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp.Meta["total"])
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "sessions", resp.Data[0].Type)
}

func TestSessionHandler_List_WithSort(t *testing.T) {
	var capturedSort string
	sessions := []*models.Session{
		{ID: uuid.New(), Name: "Afternoon"},
		{ID: uuid.New(), Name: "Morning"},
	}

	repo := &mocks.MockSessionRepo{
		ListFunc: func(_ context.Context, sortSQL string) ([]*models.Session, error) {
			capturedSort = sortSQL
			return sessions, nil
		},
	}

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/?sort=-name", nil))
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ORDER BY Name DESC", capturedSort)
}

func TestSessionHandler_List_Empty(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		ListFunc: func(_ context.Context, _ string) ([]*models.Session, error) {
			return []*models.Session{}, nil
		},
	}

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/", nil))
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

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/"+sessionID.String(), nil))
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "sessions", resp.Data.Type)
	assert.Equal(t, sessionID.String(), resp.Data.ID)
}

func TestSessionHandler_GetByID_NotFound(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Session, error) {
			return nil, nil
		},
	}

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSessionHandler_GetByID_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil))
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

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "sessions", resp.Data.Type)
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

	body := jsonapiBody("sessions", map[string]string{"name": "Evening", "startTime": "17:00", "endTime": "20:00"})
	req := authSessionReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestSessionHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	req := authSessionReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad`)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"name": "Overlap", "startTime": "07:00", "endTime": "13:00"})
	req := authSessionReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"name": "Updated"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+sessionID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"name": "Updated"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	req := authSessionReq(httptest.NewRequest(http.MethodDelete, "/"+sessionID.String(), nil))
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

	req := authSessionReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- Error path tests ---

func TestSessionHandler_List_DBError(t *testing.T) {
	repo := &mocks.MockSessionRepo{
		ListFunc: func(_ context.Context, _ string) ([]*models.Session, error) {
			return nil, errors.New("database error")
		},
	}

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/", nil))
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

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
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

	req := authSessionReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Update_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := jsonapiBody("sessions", map[string]string{"name": "Updated"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/not-a-uuid", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"name": "Updated"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"name": "Updated"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(`{bad`)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"startTime": "not-a-time"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"endTime": "not-a-time"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
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

	body := jsonapiBody("sessions", map[string]string{"name": "Updated"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Delete_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	req := authSessionReq(httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil))
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

	req := authSessionReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
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

	req := authSessionReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
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

	req := authSessionReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSessionHandler_Create_InvalidStartTime(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := jsonapiBody("sessions", map[string]string{"name": "Test", "startTime": "not-time", "endTime": "18:00"})
	req := authSessionReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Create_InvalidEndTime(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := jsonapiBody("sessions", map[string]string{"name": "Test", "startTime": "07:00", "endTime": "not-time"})
	req := authSessionReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSessionHandler_Update_UpdateTimes(t *testing.T) {
	sessionID := uuid.New()

	repo := &mocks.MockSessionRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Session, error) {
			return &models.Session{
				ID:        id,
				Name:      "Morning",
				StartTime: time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC),
				EndTime:   time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
			}, nil
		},
		UpdateFunc: func(_ context.Context, session *models.Session) error {
			assert.Equal(t, "Morning Updated", session.Name)
			assert.Equal(t, 8, session.StartTime.Hour())
			assert.Equal(t, 13, session.EndTime.Hour())
			return nil
		},
	}

	body := jsonapiBody("sessions", map[string]string{"name": "Morning Updated", "startTime": "08:00", "endTime": "13:00"})
	req := authSessionReq(httptest.NewRequest(http.MethodPut, "/"+sessionID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestSessionHandler_Create_MissingFields(t *testing.T) {
	repo := &mocks.MockSessionRepo{}

	body := jsonapiBody("sessions", map[string]string{"name": "OnlyName"})
	req := authSessionReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newSessionRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
