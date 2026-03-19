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
			DayInfo: []models.ScheduleDayInfo{
				{DayNumber: 2, DayName: "Monday"},
				{DayNumber: 3, DayName: "Tuesday"},
				{DayNumber: 4, DayName: "Wednesday"},
				{DayNumber: 5, DayName: "Thursday"},
				{DayNumber: 6, DayName: "Friday"},
			},
			Sound:   &models.SystemAudioFile{ID: soundID, Name: "bell.wav", FileType: models.FileTypeBell},
			Session: &models.Session{ID: sessionID, Name: "Morning"},
		},
	}

	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context) ([]*models.ScheduleItem, error) {
			return items, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Total int                   `json:"total"`
		Items []models.ScheduleItem `json:"items"`
	}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "Morning Bell", resp.Items[0].Name)
	assert.NotNil(t, resp.Items[0].Sound)
	assert.NotNil(t, resp.Items[0].Session)
}

func TestScheduleHandler_List_Empty(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		ListFunc: func(_ context.Context) ([]*models.ScheduleItem, error) {
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
		ListFunc: func(_ context.Context) ([]*models.ScheduleItem, error) {
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

	var resp models.ScheduleItem
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, itemID, resp.ID)
	assert.Equal(t, "Lunch Bell", resp.Name)
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

	body, _ := json.Marshal(map[string]interface{}{
		"sessionId": sessionID.String(),
		"name":      "Period 1",
		"time":      "08:00",
		"soundId":   soundID.String(),
		"days":      []int{2, 3, 4, 5, 6},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
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

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "Anthem",
		"time":    "07:45",
		"soundId": soundID.String(),
		"days":    []int{2, 3, 4, 5, 6},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestScheduleHandler_Create_InvalidJSON(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_InvalidDays_Zero(t *testing.T) {
	soundID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "Bad",
		"time":    "08:00",
		"soundId": soundID.String(),
		"days":    []int{0, 2, 3},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_InvalidDays_Eight(t *testing.T) {
	soundID := uuid.New()
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "Bad",
		"time":    "08:00",
		"soundId": soundID.String(),
		"days":    []int{1, 8},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestScheduleHandler_Create_MissingRequiredFields(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	body := `{}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body := `{"name":"New Name","days":[2,3,4,5,6]}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestScheduleHandler_Update_NotFound(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.ScheduleItem, error) {
			return nil, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	body := `{"name":"New Name"}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body := `{"days":[0,8]}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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
			return &models.ScheduleItem{ID: id, Days: []int{1, 2, 3, 4, 5, 6, 7}}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "All Days",
		"time":    "08:00",
		"soundId": soundID.String(),
		"days":    []int{1, 2, 3, 4, 5, 6, 7},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
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
			return &models.ScheduleItem{ID: id, Days: []int{1}}, nil
		},
	}
	dayRepo := &mocks.MockScheduleDayRepo{}

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "Sunday Only",
		"time":    "08:00",
		"soundId": soundID.String(),
		"days":    []int{1},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "Bad Time",
		"time":    "not-a-time",
		"soundId": soundID.String(),
		"days":    []int{2},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "Bell",
		"time":    "08:00",
		"soundId": soundID.String(),
		"days":    []int{2},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "Bell",
		"time":    "08:00",
		"soundId": soundID.String(),
		"days":    []int{2},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body, _ := json.Marshal(map[string]interface{}{
		"name":      "New Name",
		"time":      "09:30",
		"soundId":   newSoundID.String(),
		"sessionId": newSessionID.String(),
		"days":      []int{1, 2, 3, 4, 5},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestScheduleHandler_Update_InvalidUUID(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	body := `{"name":"Updated"}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/not-a-uuid", bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body := `{"name":"Updated"}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body := `{"name":"Updated"}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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
	req.Header.Set("Content-Type", "application/json")
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

	body := `{"time":"not-a-time"}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body := `{"name":"Updated"}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body, _ := json.Marshal(map[string]interface{}{
		"name": "Bell", "time": "08:00", "soundId": soundID.String(), "days": []int{2},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body := `{"name":"New"}`
	req := authScheduleReq(httptest.NewRequest(http.MethodPut, "/"+itemID.String(), bytes.NewBufferString(body)))
	req.Header.Set("Content-Type", "application/json")
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

	body, _ := json.Marshal(map[string]interface{}{
		"name":    "No Days",
		"time":    "08:00",
		"soundId": soundID.String(),
		"days":    []int{},
	})
	req := authScheduleReq(httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body)))
	req.Header.Set("Content-Type", "application/json")
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
