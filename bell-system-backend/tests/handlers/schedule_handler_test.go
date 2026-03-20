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

// newScheduleRouter creates a schedule chi.Router with the given mocks.
func newScheduleRouter(itemRepo handlers.ScheduleItemRepository, dayRepo handlers.ScheduleDayRepository) http.Handler {
	h := handlers.NewScheduleHandler(itemRepo, dayRepo, nil, nil)
	return router.ScheduleRoutes(h, testutil.PermissiveTokenService)
}

// newScheduleRouterWithSessions creates a schedule chi.Router with session support.
func newScheduleRouterWithSessions(itemRepo handlers.ScheduleItemRepository, sessionRepo handlers.SessionRepository) http.Handler {
	h := handlers.NewScheduleHandler(itemRepo, nil, sessionRepo, nil)
	return router.ScheduleRoutes(h, testutil.PermissiveTokenService)
}

func authScheduleReq(req *http.Request) *http.Request {
	testutil.SetAuthHeader(req)
	return req
}

// scheduleCreateBody builds a JSON:API request body for creating schedule items.
func scheduleCreateBody(name, timeStr string, days []int, soundID uuid.UUID, sessionID *uuid.UUID) *bytes.Buffer {
	rels := map[string]any{
		"sound": map[string]any{
			"data": map[string]any{"type": "audio-files", "id": soundID.String()},
		},
	}
	if sessionID != nil {
		rels["session"] = map[string]any{
			"data": map[string]any{"type": "sessions", "id": sessionID.String()},
		}
	}
	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"type":          "schedule-items",
			"attributes":    map[string]any{"name": name, "time": timeStr, "days": days},
			"relationships": rels,
		},
	})
	return bytes.NewBuffer(body)
}

// scheduleUpdateBody builds a JSON:API request body for updating schedule items.
func scheduleUpdateBody(attrs map[string]any, rels map[string]any) *bytes.Buffer {
	data := map[string]any{
		"type":       "schedule-items",
		"attributes": attrs,
	}
	if rels != nil {
		data["relationships"] = rels
	}
	body, _ := json.Marshal(map[string]any{"data": data})
	return bytes.NewBuffer(body)
}

// --- GET / (list schedule items) ---

func TestScheduleHandler_List_Success(t *testing.T) {
	sessionID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	items := []*models.ScheduleItem{
		{
			ID:        uuid.New(),
			SessionID: &sessionID,
			Name:      "Morning Bell",
			Time:      now,
			SoundID:   soundID,
			Days:      []int{2, 3, 4, 5, 6},
		},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, _ string) ([]*models.ScheduleItem, error) {
			return items, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp.Meta["total"])
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "schedule-items", resp.Data[0].Type)
	assert.Equal(t, "Morning Bell", resp.Data[0].Attributes.(map[string]any)["name"])
}

func TestScheduleHandler_List_WithSort(t *testing.T) {
	var capturedSort string
	items := []*models.ScheduleItem{
		{ID: uuid.New(), Name: "Bell", Days: []int{2}},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, sortSQL string) ([]*models.ScheduleItem, error) {
			capturedSort = sortSQL
			return items, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/?sort=-name", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ORDER BY Name DESC", capturedSort)
}

func TestScheduleHandler_List_IncludeSession(t *testing.T) {
	sessionID := uuid.New()
	session := &models.Session{ID: sessionID, Name: "Morning"}
	items := []*models.ScheduleItem{
		{ID: uuid.New(), Name: "Bell 1", SessionID: &sessionID, Session: session, Days: []int{2}},
		{ID: uuid.New(), Name: "Bell 2", SessionID: &sessionID, Session: session, Days: []int{2}},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, _ string) ([]*models.ScheduleItem, error) {
			return items, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/?include=session", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Included, 1) // deduplicated
	assert.Equal(t, "sessions", resp.Included[0].Type)
	assert.Equal(t, sessionID.String(), resp.Included[0].ID)
}

func TestScheduleHandler_List_IncludeSound(t *testing.T) {
	soundID := uuid.New()
	sound := &models.SystemAudioFile{ID: soundID, Name: "bell.wav", FileType: models.FileTypeBell}
	items := []*models.ScheduleItem{
		{ID: uuid.New(), Name: "Bell 1", SoundID: soundID, Sound: sound, Days: []int{2}},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, _ string) ([]*models.ScheduleItem, error) {
			return items, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/?include=sound", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Included, 1)
	assert.Equal(t, "audio-files", resp.Included[0].Type)
}

func TestScheduleHandler_List_IncludeBoth(t *testing.T) {
	sessionID := uuid.New()
	soundID := uuid.New()
	session := &models.Session{ID: sessionID, Name: "Morning"}
	sound := &models.SystemAudioFile{ID: soundID, Name: "bell.wav", FileType: models.FileTypeBell}
	items := []*models.ScheduleItem{
		{ID: uuid.New(), Name: "Bell", SessionID: &sessionID, Session: session, SoundID: soundID, Sound: sound, Days: []int{2}},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, _ string) ([]*models.ScheduleItem, error) {
			return items, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/?include=session,sound", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Included, 2)
}

func TestScheduleHandler_List_NoInclude(t *testing.T) {
	items := []*models.ScheduleItem{
		{ID: uuid.New(), Name: "Bell", Days: []int{2}},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, _ string) ([]*models.ScheduleItem, error) {
			return items, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Included)
}

func TestScheduleHandler_GetByID_IncludeSession(t *testing.T) {
	itemID := uuid.New()
	sessionID := uuid.New()
	session := &models.Session{ID: sessionID, Name: "Morning"}
	item := &models.ScheduleItem{
		ID: itemID, Name: "Bell", SessionID: &sessionID, Session: session, Days: []int{2},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return item, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/"+itemID.String()+"?include=session", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Included, 1)
	assert.Equal(t, "sessions", resp.Included[0].Type)
}

func TestScheduleHandler_List_WithFilter(t *testing.T) {
	var capturedSessionID string
	var capturedDay int

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, filterSessionID string, filterDay int, _ string) ([]*models.ScheduleItem, error) {
			capturedSessionID = filterSessionID
			capturedDay = filterDay
			return []*models.ScheduleItem{}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	sid := uuid.New().String()
	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/?filter[sessionId]="+sid+"&filter[day]=3", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, sid, capturedSessionID)
	assert.Equal(t, 3, capturedDay)
}

func TestScheduleHandler_List_Empty(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, _ string) ([]*models.ScheduleItem, error) {
			return []*models.ScheduleItem{}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestScheduleHandler_List_DBError(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context, _ string, _ int, _ string) ([]*models.ScheduleItem, error) {
			return nil, errors.New("database error")
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- GET /{id} (get schedule item by ID) ---

func TestScheduleHandler_GetByID_Success(t *testing.T) {
	itemID := uuid.New()
	item := &models.ScheduleItem{
		ID:      itemID,
		Name:    "Lunch Bell",
		SoundID: uuid.New(),
		Days:    []int{2, 3, 4, 5, 6},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			assert.Equal(t, itemID, id)
			return item, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/"+itemID.String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "schedule-items", resp.Data.Type)
	assert.Equal(t, itemID.String(), resp.Data.ID)
	assert.Equal(t, "Lunch Bell", resp.Data.Attributes.(map[string]any)["name"])
}

func TestScheduleHandler_GetByID_NotFound(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestScheduleHandler_GetByID_InvalidUUID(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/not-valid", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- POST / (create schedule item) ---

func TestScheduleHandler_Create_WithSession(t *testing.T) {
	sessionID := uuid.New()
	soundID := uuid.New()

	itemRepo := &mocks.MockScheduleItemRepo{
		CreateFunc: func(_ context.Context, item *models.ScheduleItem) error {
			assert.Equal(t, "Period 1", item.Name)
			assert.Equal(t, &sessionID, item.SessionID)
			assert.Equal(t, soundID, item.SoundID)
			assert.Equal(t, []int{2, 3, 4, 5, 6}, item.Days)
			return nil
		},
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{
				ID:        id,
				SessionID: &sessionID,
				Name:      "Period 1",
				SoundID:   soundID,
				Days:      []int{2, 3, 4, 5, 6},
			}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("Period 1", "08:00", []int{2, 3, 4, 5, 6}, soundID, &sessionID)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "schedule-items", resp.Data.Type)
}

func TestScheduleHandler_Create_WithoutSession(t *testing.T) {
	soundID := uuid.New()

	itemRepo := &mocks.MockScheduleItemRepo{
		CreateFunc: func(_ context.Context, item *models.ScheduleItem) error {
			assert.Nil(t, item.SessionID)
			return nil
		},
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{
				ID:      id,
				Name:    "Anthem",
				SoundID: soundID,
				Days:    []int{2, 3, 4, 5, 6},
			}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("Anthem", "07:45", []int{2, 3, 4, 5, 6}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestScheduleHandler_Create_InvalidJSON(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad`)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_InvalidDays_Zero(t *testing.T) {
	soundID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("Bad", "08:00", []int{0, 2, 3}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_InvalidDays_Eight(t *testing.T) {
	soundID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("Bad", "08:00", []int{1, 8}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_MissingRequiredFields(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	// Empty attributes, no relationships
	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"type":       "schedule-items",
			"attributes": map[string]any{},
		},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- PUT /{id} (update schedule item) ---

func TestScheduleHandler_Update_Success(t *testing.T) {
	itemID := uuid.New()
	existing := &models.ScheduleItem{
		ID:      itemID,
		Name:    "Old Name",
		SoundID: uuid.New(),
		Days:    []int{2, 3},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return existing, nil
		},
		UpdateFunc: func(_ context.Context, item *models.ScheduleItem) error {
			assert.Equal(t, "New Name", item.Name)
			return nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(
		map[string]any{"name": "New Name", "days": []int{2, 3, 4, 5, 6}},
		nil,
	)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "schedule-items", resp.Data.Type)
}

func TestScheduleHandler_Update_NotFound(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(map[string]any{"name": "New Name"}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestScheduleHandler_Update_InvalidDays(t *testing.T) {
	itemID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: itemID}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(map[string]any{"days": []int{0, 8}}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- DELETE /{id} (delete schedule item) ---

func TestScheduleHandler_Delete_Success(t *testing.T) {
	itemID := uuid.New()

	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, id uuid.UUID) error {
			assert.Equal(t, itemID, id)
			return nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodDelete, "/"+itemID.String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestScheduleHandler_Delete_NotFound(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- Day validation spec tests ---

func TestScheduleHandler_DaysValidation_ValidRange(t *testing.T) {
	soundID := uuid.New()

	itemRepo := &mocks.MockScheduleItemRepo{
		CreateFunc: func(_ context.Context, item *models.ScheduleItem) error {
			assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, item.Days)
			return nil
		},
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: id, SoundID: soundID, Days: []int{1, 2, 3, 4, 5, 6, 7}}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("All Days", "08:00", []int{1, 2, 3, 4, 5, 6, 7}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestScheduleHandler_DaysValidation_SundayIsOne(t *testing.T) {
	soundID := uuid.New()

	itemRepo := &mocks.MockScheduleItemRepo{
		CreateFunc: func(_ context.Context, item *models.ScheduleItem) error {
			assert.Contains(t, item.Days, 1)
			return nil
		},
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: id, SoundID: soundID, Days: []int{1}}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("Sunday Only", "08:00", []int{1}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

// --- Error path tests ---

func TestScheduleHandler_GetByID_ErrNotFound(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestScheduleHandler_GetByID_DBError(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, errors.New("database error")
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestScheduleHandler_Create_InvalidTime(t *testing.T) {
	soundID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("Bad Time", "not-a-time", []int{2}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_DBError(t *testing.T) {
	soundID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{
		CreateFunc: func(_ context.Context, _ *models.ScheduleItem) error {
			return errors.New("create failed")
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("Bell", "08:00", []int{2}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_GetByIDFailsAfterCreate(t *testing.T) {
	soundID := uuid.New()
	notifier := &mocks.MockEventNotifier{}
	itemRepo := &mocks.MockScheduleItemRepo{
		CreateFunc: func(_ context.Context, _ *models.ScheduleItem) error {
			return nil
		},
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, errors.New("fetch failed")
		},
	}

	h := handlers.NewScheduleHandler(itemRepo, nil, nil, notifier)
	r := router.ScheduleRoutes(h, testutil.PermissiveTokenService)

	buf := scheduleCreateBody("Bell", "08:00", []int{2}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Should still return 201 with the original item when GetByID fails
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, notifier.SchedulesUpdatedCalled)
}

func TestScheduleHandler_Update_AllFields(t *testing.T) {
	itemID := uuid.New()
	newSoundID := uuid.New()
	newSessionID := uuid.New()
	existing := &models.ScheduleItem{
		ID:      itemID,
		Name:    "Old Name",
		SoundID: uuid.New(),
		Days:    []int{2, 3},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return existing, nil
		},
		UpdateFunc: func(_ context.Context, item *models.ScheduleItem) error {
			assert.Equal(t, "New Name", item.Name)
			assert.Equal(t, 9, item.Time.Hour())
			assert.Equal(t, 30, item.Time.Minute())
			assert.Equal(t, newSoundID, item.SoundID)
			assert.Equal(t, &newSessionID, item.SessionID)
			assert.Equal(t, []int{1, 2, 3, 4, 5}, item.Days)
			return nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(
		map[string]any{"name": "New Name", "time": "09:30", "days": []int{1, 2, 3, 4, 5}},
		map[string]any{
			"sound":   map[string]any{"data": map[string]any{"type": "audio-files", "id": newSoundID.String()}},
			"session": map[string]any{"data": map[string]any{"type": "sessions", "id": newSessionID.String()}},
		},
	)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestScheduleHandler_Update_InvalidUUID(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(map[string]any{"name": "Updated"}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/not-a-uuid", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Update_ErrNotFound(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(map[string]any{"name": "Updated"}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestScheduleHandler_Update_GetByIDDBError(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, errors.New("database error")
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(map[string]any{"name": "Updated"}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestScheduleHandler_Update_InvalidJSON(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: uuid.New()}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(`{bad`)))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Update_InvalidTime(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: uuid.New()}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(map[string]any{"time": "not-a-time"}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Update_DBError(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: uuid.New()}, nil
		},
		UpdateFunc: func(_ context.Context, _ *models.ScheduleItem) error {
			return errors.New("update failed")
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleUpdateBody(map[string]any{"name": "Updated"}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestScheduleHandler_Delete_InvalidUUID(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Delete_ErrNotFound(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestScheduleHandler_Delete_GetByIDDBError(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, errors.New("database error")
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestScheduleHandler_Delete_DBError(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("delete failed")
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Notifier coverage ---

func TestScheduleHandler_Create_NotifiesOnSuccess(t *testing.T) {
	soundID := uuid.New()
	notifier := &mocks.MockEventNotifier{}

	itemRepo := &mocks.MockScheduleItemRepo{
		CreateFunc: func(_ context.Context, _ *models.ScheduleItem) error { return nil },
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: id, Name: "Bell", SoundID: soundID, Days: []int{2}}, nil
		},
	}
	h := handlers.NewScheduleHandler(itemRepo, nil, nil, notifier)
	r := router.ScheduleRoutes(h, testutil.PermissiveTokenService)

	buf := scheduleCreateBody("Bell", "08:00", []int{2}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, notifier.SchedulesUpdatedCalled)
}

func TestScheduleHandler_Update_NotifiesOnSuccess(t *testing.T) {
	itemID := uuid.New()
	notifier := &mocks.MockEventNotifier{}

	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: itemID, Name: "Old", SoundID: uuid.New(), Days: []int{2}}, nil
		},
		UpdateFunc: func(_ context.Context, _ *models.ScheduleItem) error { return nil },
	}
	h := handlers.NewScheduleHandler(itemRepo, nil, nil, notifier)
	r := router.ScheduleRoutes(h, testutil.PermissiveTokenService)

	buf := scheduleUpdateBody(map[string]any{"name": "New"}, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, notifier.SchedulesUpdatedCalled)
}

func TestScheduleHandler_Delete_NotifiesOnSuccess(t *testing.T) {
	itemID := uuid.New()
	notifier := &mocks.MockEventNotifier{}

	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
			return &models.ScheduleItem{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	h := handlers.NewScheduleHandler(itemRepo, nil, nil, notifier)
	r := router.ScheduleRoutes(h, testutil.PermissiveTokenService)

	req := authScheduleReq(httptest.NewRequest(http.MethodDelete, "/"+itemID.String(), nil))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.True(t, notifier.SchedulesUpdatedCalled)
}

func TestScheduleHandler_DaysValidation_EmptyDays(t *testing.T) {
	soundID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	buf := scheduleCreateBody("No Days", "08:00", []int{}, soundID, nil)
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", buf))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- GET /current (current schedule) ---

func TestScheduleHandler_GetCurrent_Success(t *testing.T) {
	sessionID := uuid.New()
	soundID := uuid.New()
	session := &models.Session{
		ID:   sessionID,
		Name: "Morning",
	}

	// Use a fixed time at noon to avoid midnight wrapping in any timezone.
	fixedNow := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	pastTime := time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC)
	futureTime := time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC)

	items := []*models.ScheduleItem{
		{ID: uuid.New(), Name: "Past Bell", Time: pastTime, SoundID: soundID},
		{ID: uuid.New(), Name: "Future Bell", Time: futureTime, SoundID: soundID},
	}

	sessionRepo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return session, nil
		},
	}
	itemRepo := &mocks.MockScheduleItemRepo{
		GetCurrentSessionSchedulesFunc: func(_ context.Context, id uuid.UUID) ([]*models.ScheduleItem, error) {
			assert.Equal(t, sessionID, id)
			return items, nil
		},
	}

	h := handlers.NewScheduleHandler(itemRepo, nil, sessionRepo, nil)
	h.NowFunc = func() time.Time { return fixedNow }

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()
	router.ScheduleRoutes(h, testutil.PermissiveTokenService).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.CurrentScheduleResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Morning", resp.Session.Name)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "completed", resp.Items[0].Status)
	assert.Equal(t, "pending", resp.Items[1].Status)
}

func TestScheduleHandler_GetCurrent_NowFunc_AllStatuses(t *testing.T) {
	sessionID := uuid.New()
	soundID := uuid.New()
	session := &models.Session{ID: sessionID, Name: "Afternoon"}

	// Fix "now" at exactly 14:00 — test all three statuses.
	fixedNow := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	items := []*models.ScheduleItem{
		{ID: uuid.New(), Name: "Past", Time: time.Date(0, 1, 1, 13, 0, 0, 0, time.UTC), SoundID: soundID},
		{ID: uuid.New(), Name: "Current", Time: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC), SoundID: soundID},
		{ID: uuid.New(), Name: "Future", Time: time.Date(0, 1, 1, 15, 0, 0, 0, time.UTC), SoundID: soundID},
	}

	sessionRepo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) { return session, nil },
	}
	itemRepo := &mocks.MockScheduleItemRepo{
		GetCurrentSessionSchedulesFunc: func(_ context.Context, _ uuid.UUID) ([]*models.ScheduleItem, error) { return items, nil },
	}

	h := handlers.NewScheduleHandler(itemRepo, nil, sessionRepo, nil)
	h.NowFunc = func() time.Time { return fixedNow }

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()
	router.ScheduleRoutes(h, testutil.PermissiveTokenService).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.CurrentScheduleResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Len(t, resp.Items, 3)
	assert.Equal(t, "completed", resp.Items[0].Status)
	assert.Equal(t, "current", resp.Items[1].Status)
	assert.Equal(t, "pending", resp.Items[2].Status)
}

func TestScheduleHandler_GetCurrent_NowFunc_NilFallback(t *testing.T) {
	// When NowFunc is nil, handler should use time.Now without panicking.
	session := &models.Session{ID: uuid.New(), Name: "Morning"}

	sessionRepo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) { return session, nil },
	}
	itemRepo := &mocks.MockScheduleItemRepo{
		GetCurrentSessionSchedulesFunc: func(_ context.Context, _ uuid.UUID) ([]*models.ScheduleItem, error) {
			return []*models.ScheduleItem{}, nil
		},
	}

	h := handlers.NewScheduleHandler(itemRepo, nil, sessionRepo, nil)
	// NowFunc intentionally left nil

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()
	router.ScheduleRoutes(h, testutil.PermissiveTokenService).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestScheduleHandler_GetCurrent_NoSession(t *testing.T) {
	sessionRepo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()
	newScheduleRouterWithSessions(&mocks.MockScheduleItemRepo{}, sessionRepo).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp models.CurrentScheduleResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Nil(t, resp.Session)
	assert.Empty(t, resp.Items)
}

func TestScheduleHandler_GetCurrent_SessionError(t *testing.T) {
	sessionRepo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return nil, errors.New("db error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()
	newScheduleRouterWithSessions(&mocks.MockScheduleItemRepo{}, sessionRepo).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestScheduleHandler_GetCurrent_ItemsError(t *testing.T) {
	session := &models.Session{ID: uuid.New(), Name: "Morning"}

	sessionRepo := &mocks.MockSessionRepo{
		GetCurrentSessionFunc: func(_ context.Context) (*models.Session, error) {
			return session, nil
		},
	}
	itemRepo := &mocks.MockScheduleItemRepo{
		GetCurrentSessionSchedulesFunc: func(_ context.Context, _ uuid.UUID) ([]*models.ScheduleItem, error) {
			return nil, errors.New("db error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/current", nil)
	rr := httptest.NewRecorder()
	newScheduleRouterWithSessions(itemRepo, sessionRepo).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
