package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	pkgerrors "arabiyya.edu.mv/bell-system-backend/pkg/errors"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ScheduleHandler handles schedule item management HTTP requests.
type ScheduleHandler struct {
	Items    ScheduleItemRepository
	Days     ScheduleDayRepository
	Sessions SessionRepository
	Notifier EventNotifier
	NowFunc  func() time.Time
}

// NewScheduleHandler creates a new ScheduleHandler.
func NewScheduleHandler(items ScheduleItemRepository, days ScheduleDayRepository, sessions SessionRepository, notifier EventNotifier) *ScheduleHandler {
	return &ScheduleHandler{Items: items, Days: days, Sessions: sessions, Notifier: notifier}
}

// GetCurrent handles GET /current — returns today's schedule for the active session.
func (h *ScheduleHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	session, err := h.Sessions.GetCurrentSession(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get current session")
		return
	}
	if session == nil {
		writeJSON(w, http.StatusOK, models.CurrentScheduleResponse{
			Session: nil,
			Items:   []models.CurrentScheduleItem{},
		})
		return
	}

	items, err := h.Items.GetCurrentSessionSchedules(r.Context(), session.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get schedule items")
		return
	}

	nowFn := h.NowFunc
	if nowFn == nil {
		nowFn = time.Now
	}
	now := nowFn()
	currentTime := now.Hour()*60 + now.Minute()

	result := make([]models.CurrentScheduleItem, 0, len(items))
	for _, item := range items {
		itemTime := item.Time.Hour()*60 + item.Time.Minute()
		status := "pending"
		if itemTime < currentTime {
			status = "completed"
		} else if itemTime == currentTime {
			status = "current"
		}

		result = append(result, models.CurrentScheduleItem{
			ID:     item.ID,
			Name:   item.Name,
			Time:   item.Time.Format("15:04"),
			Sound:  item.Sound,
			Status: status,
		})
	}

	writeJSON(w, http.StatusOK, models.CurrentScheduleResponse{
		Session: session,
		Items:   result,
	})
}

func validateDays(days []int) bool {
	for _, d := range days {
		if d < 1 || d > 7 {
			return false
		}
	}
	return true
}

// List handles GET /.
func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.Items.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list schedule items")
		return
	}
	writeJSON(w, http.StatusOK, models.ListResponse{Total: len(items), Items: items})
}

// GetByID handles GET /{id}.
func (h *ScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid schedule item ID")
		return
	}

	item, err := h.Items.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "schedule item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get schedule item")
		return
	}
	if item == nil {
		writeError(w, http.StatusNotFound, "not_found", "schedule item not found")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// Create handles POST /.
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.ScheduleItemCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Name == "" || req.Time == "" || req.SoundID == uuid.Nil || len(req.Days) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "name, time, soundId, and days are required")
		return
	}

	if !validateDays(req.Days) {
		writeError(w, http.StatusBadRequest, "invalid_request", "days must be between 1 (Sunday) and 7 (Saturday)")
		return
	}

	parsedTime, err := time.Parse("15:04", req.Time)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid time format, use HH:MM")
		return
	}

	item := &models.ScheduleItem{
		ID:        uuid.New(),
		SessionID: req.SessionID,
		Name:      req.Name,
		Time:      parsedTime,
		SoundID:   req.SoundID,
		Days:      req.Days,
		CreatedAt: time.Now(),
	}

	if err := h.Items.Create(r.Context(), item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "failed to create schedule item")
		return
	}

	created, err := h.Items.GetByID(r.Context(), item.ID)
	if err != nil || created == nil {
		writeJSON(w, http.StatusCreated, item)
		if h.Notifier != nil {
			h.Notifier.NotifySchedulesUpdated()
		}
		return
	}

	writeJSON(w, http.StatusCreated, created)
	if h.Notifier != nil {
		h.Notifier.NotifySchedulesUpdated()
	}
}

// Update handles PUT /{id}.
func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid schedule item ID")
		return
	}

	existing, err := h.Items.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "schedule item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get schedule item")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "schedule item not found")
		return
	}

	var req models.ScheduleItemUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if len(req.Days) > 0 && !validateDays(req.Days) {
		writeError(w, http.StatusBadRequest, "invalid_request", "days must be between 1 (Sunday) and 7 (Saturday)")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Time != "" {
		t, err := time.Parse("15:04", req.Time)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid time format")
			return
		}
		existing.Time = t
	}
	if req.SoundID != uuid.Nil {
		existing.SoundID = req.SoundID
	}
	if req.SessionID != nil {
		existing.SessionID = req.SessionID
	}
	if len(req.Days) > 0 {
		existing.Days = req.Days
	}

	if err := h.Items.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update schedule item")
		return
	}

	writeJSON(w, http.StatusOK, existing)
	if h.Notifier != nil {
		h.Notifier.NotifySchedulesUpdated()
	}
}

// Delete handles DELETE /{id}.
func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid schedule item ID")
		return
	}

	existing, err := h.Items.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "schedule item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get schedule item")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "schedule item not found")
		return
	}

	if err := h.Items.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete schedule item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	if h.Notifier != nil {
		h.Notifier.NotifySchedulesUpdated()
	}
}
