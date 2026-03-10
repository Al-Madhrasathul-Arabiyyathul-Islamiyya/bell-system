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

// newScheduleRouter creates a schedule chi.Router with the given mocks.
func newScheduleRouter(itemRepo handlers.ScheduleItemRepository, dayRepo handlers.ScheduleDayRepository) http.Handler {
	h := handlers.NewScheduleHandler(itemRepo, dayRepo)
	return router.ScheduleRoutes(h)
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

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Total int                    `json:"total"`
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

	req := httptest.NewRequest(http.MethodGet, "/", nil)
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

	req := httptest.NewRequest(http.MethodGet, "/", nil)
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

	req := httptest.NewRequest(http.MethodGet, "/"+itemID.String(), nil)
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

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestScheduleHandler_GetByID_InvalidUUID(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := httptest.NewRequest(http.MethodGet, "/not-valid", nil)
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestScheduleHandler_Create_InvalidJSON(t *testing.T) {
	itemRepo := &mocks.MockScheduleItemRepo{}
	dayRepo := &mocks.MockScheduleDayRepo{}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad`))
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
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
	req := httptest.NewRequest(http.MethodPut, "/"+itemID.String(), bytes.NewBufferString(body))
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
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
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
	req := httptest.NewRequest(http.MethodPut, "/"+itemID.String(), bytes.NewBufferString(body))
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

	req := httptest.NewRequest(http.MethodDelete, "/"+itemID.String(), nil)
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

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
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
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newScheduleRouter(itemRepo, dayRepo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
