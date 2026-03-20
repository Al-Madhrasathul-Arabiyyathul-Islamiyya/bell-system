package handlers

import (
	"errors"
	"net/http"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	pkgerrors "arabiyya.edu.mv/bell-system-backend/pkg/errors"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// SessionHandler handles session management HTTP requests.
type SessionHandler struct {
	Sessions SessionRepository
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessions SessionRepository) *SessionHandler {
	return &SessionHandler{Sessions: sessions}
}

type sessionCreateAttributes struct {
	Name      string `json:"name"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type sessionUpdateAttributes struct {
	Name      string `json:"name,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

// List handles GET /.
func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.Sessions.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list sessions")
		return
	}

	resources := make([]jsonapi.Resource, len(sessions))
	for i, s := range sessions {
		resources[i] = jsonapi.MarshalSession(s)
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.CollectionDocument{
		Data:  resources,
		Meta:  map[string]any{"total": len(sessions)},
		Links: map[string]any{"self": "/api/v1/sessions"},
	})
}

// GetByID handles GET /{id}.
func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid session ID")
		return
	}

	session, err := h.Sessions.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get session")
		return
	}
	if session == nil {
		writeError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalSession(session))})
}

// GetCurrent handles GET /current.
func (h *SessionHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	session, err := h.Sessions.GetCurrentSession(r.Context())
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "no active session")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get current session")
		return
	}
	if session == nil {
		writeError(w, http.StatusNotFound, "not_found", "no active session")
		return
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalSession(session))})
}

// Create handles POST /.
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	doc, err := jsonapi.ParseRequest(r, "sessions")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var attrs sessionCreateAttributes
	if err := doc.UnmarshalAttributes(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid attributes")
		return
	}

	if attrs.Name == "" || attrs.StartTime == "" || attrs.EndTime == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "name, startTime, and endTime are required")
		return
	}

	startTime, err := time.Parse("15:04", attrs.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid startTime format, use HH:MM")
		return
	}

	endTime, err := time.Parse("15:04", attrs.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid endTime format, use HH:MM")
		return
	}

	session := &models.Session{
		ID:        uuid.New(),
		Name:      attrs.Name,
		StartTime: startTime,
		EndTime:   endTime,
	}

	if err := h.Sessions.Create(r.Context(), session); err != nil {
		writeError(w, http.StatusConflict, "conflict", "session conflict")
		return
	}

	writeJSONAPI(w, http.StatusCreated, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalSession(session))})
}

// Update handles PUT /{id}.
func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid session ID")
		return
	}

	existing, err := h.Sessions.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get session")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}

	doc, err := jsonapi.ParseRequest(r, "sessions")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var attrs sessionUpdateAttributes
	if err := doc.UnmarshalAttributes(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid attributes")
		return
	}

	if attrs.Name != "" {
		existing.Name = attrs.Name
	}
	if attrs.StartTime != "" {
		t, err := time.Parse("15:04", attrs.StartTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid startTime format")
			return
		}
		existing.StartTime = t
	}
	if attrs.EndTime != "" {
		t, err := time.Parse("15:04", attrs.EndTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid endTime format")
			return
		}
		existing.EndTime = t
	}

	if err := h.Sessions.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update session")
		return
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalSession(existing))})
}

// Delete handles DELETE /{id}.
func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid session ID")
		return
	}

	existing, err := h.Sessions.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get session")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "session not found")
		return
	}

	if err := h.Sessions.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
