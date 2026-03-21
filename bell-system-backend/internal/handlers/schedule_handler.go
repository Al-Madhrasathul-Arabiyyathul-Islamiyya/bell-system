package handlers

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	pkgerrors "arabiyya.edu.mv/bell-system-backend/pkg/errors"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"

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
// This endpoint uses plain JSON (documented exception to JSON:API).
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
	filters := jsonapi.ParseFilter(r, []string{"sessionId", "day"})

	filterDay := 0
	if d := filters["day"]; d != "" {
		if v, err := strconv.Atoi(d); err == nil && v >= 1 && v <= 7 {
			filterDay = v
		}
	}

	items, err := h.Items.List(r.Context(), filters["sessionId"], filterDay, jsonapi.ParseSort(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list schedule items")
		return
	}

	resources := make([]jsonapi.Resource, len(items))
	for i, item := range items {
		resources[i] = jsonapi.MarshalScheduleItem(item)
	}

	includes := jsonapi.ParseInclude(r)
	included := collectIncluded(items, includes)

	writeJSONAPI(w, http.StatusOK, jsonapi.CollectionDocument{
		Data:     resources,
		Included: included,
		Meta:     map[string]any{"total": len(items)},
		Links:    map[string]any{"self": "/api/v1/schedule"},
	})
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

	includes := jsonapi.ParseInclude(r)
	included := collectIncluded([]*models.ScheduleItem{item}, includes)

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{
		Data:     resourcePtr(jsonapi.MarshalScheduleItem(item)),
		Included: included,
	})
}

// collectIncluded builds a deduplicated list of included resources from schedule items.
func collectIncluded(items []*models.ScheduleItem, includes []string) []jsonapi.Resource {
	includeSession := slices.Contains(includes, "session")
	includeSound := slices.Contains(includes, "sound")

	if !includeSession && !includeSound {
		return nil
	}

	var included []jsonapi.Resource
	seenSessions := map[uuid.UUID]bool{}
	seenSounds := map[uuid.UUID]bool{}

	for _, item := range items {
		if includeSession && item.Session != nil && !seenSessions[item.Session.ID] {
			included = append(included, jsonapi.MarshalSession(item.Session))
			seenSessions[item.Session.ID] = true
		}
		if includeSound && item.Sound != nil && !seenSounds[item.Sound.ID] {
			included = append(included, jsonapi.MarshalAudioFile(item.Sound))
			seenSounds[item.Sound.ID] = true
		}
	}

	return included
}

type scheduleCreateAttributes struct {
	Name string `json:"name"`
	Time string `json:"time"`
	Days []int  `json:"days"`
}

// Create handles POST /.
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	doc, err := jsonapi.ParseRequest(r, "schedule-items")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var attrs scheduleCreateAttributes
	if err := doc.UnmarshalAttributes(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid attributes")
		return
	}

	soundIDStr := doc.RelationshipID("sound")
	soundID, err := uuid.Parse(soundIDStr)
	if err != nil || soundID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "name, time, soundId, and days are required")
		return
	}

	if attrs.Name == "" || attrs.Time == "" || len(attrs.Days) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "name, time, soundId, and days are required")
		return
	}

	if !validateDays(attrs.Days) {
		writeError(w, http.StatusBadRequest, "invalid_request", "days must be between 1 (Sunday) and 7 (Saturday)")
		return
	}

	parsedTime, err := time.Parse("15:04", attrs.Time)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid time format, use HH:MM")
		return
	}

	var sessionID *uuid.UUID
	sessionIDStr := doc.RelationshipID("session")
	if sessionIDStr != "" {
		sid, err := uuid.Parse(sessionIDStr)
		if err == nil {
			sessionID = &sid
		}
	}

	item := &models.ScheduleItem{
		ID:        uuid.New(),
		SessionID: sessionID,
		Name:      attrs.Name,
		Time:      parsedTime,
		SoundID:   soundID,
		Days:      attrs.Days,
		CreatedAt: time.Now(),
	}

	if err := h.Items.Create(r.Context(), item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "failed to create schedule item")
		return
	}

	created, err := h.Items.GetByID(r.Context(), item.ID)
	if err != nil || created == nil {
		writeJSONAPI(w, http.StatusCreated, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalScheduleItem(item))})
		if h.Notifier != nil {
			h.Notifier.NotifySchedulesUpdated()
		}
		return
	}

	writeJSONAPI(w, http.StatusCreated, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalScheduleItem(created))})
	if h.Notifier != nil {
		h.Notifier.NotifySchedulesUpdated()
	}
}

type scheduleUpdateAttributes struct {
	Name string `json:"name,omitempty"`
	Time string `json:"time,omitempty"`
	Days []int  `json:"days,omitempty"`
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

	doc, err := jsonapi.ParseRequest(r, "schedule-items")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var attrs scheduleUpdateAttributes
	if err := doc.UnmarshalAttributes(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid attributes")
		return
	}

	if len(attrs.Days) > 0 && !validateDays(attrs.Days) {
		writeError(w, http.StatusBadRequest, "invalid_request", "days must be between 1 (Sunday) and 7 (Saturday)")
		return
	}

	if attrs.Name != "" {
		existing.Name = attrs.Name
	}
	if attrs.Time != "" {
		t, err := time.Parse("15:04", attrs.Time)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid time format")
			return
		}
		existing.Time = t
	}

	soundIDStr := doc.RelationshipID("sound")
	if soundIDStr != "" {
		sid, err := uuid.Parse(soundIDStr)
		if err == nil && sid != uuid.Nil {
			existing.SoundID = sid
		}
	}

	sessionIDStr := doc.RelationshipID("session")
	if sessionIDStr != "" {
		sid, err := uuid.Parse(sessionIDStr)
		if err == nil {
			existing.SessionID = &sid
		}
	}

	if len(attrs.Days) > 0 {
		existing.Days = attrs.Days
	}

	if err := h.Items.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update schedule item")
		return
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalScheduleItem(existing))})
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
