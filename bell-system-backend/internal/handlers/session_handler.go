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

// SessionHandler handles session management HTTP requests.
type SessionHandler struct {
	Sessions SessionRepository
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessions SessionRepository) *SessionHandler {
	return &SessionHandler{Sessions: sessions}
}

type sessionCreateRequest struct {
	Name      string `json:"name"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type sessionUpdateRequest struct {
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
	writeJSON(w, http.StatusOK, models.ListResponse{Total: len(sessions), Items: sessions})
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

	writeJSON(w, http.StatusOK, session)
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

	writeJSON(w, http.StatusOK, session)
}

// Create handles POST /.
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req sessionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Name == "" || req.StartTime == "" || req.EndTime == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "name, startTime, and endTime are required")
		return
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid startTime format, use HH:MM")
		return
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid endTime format, use HH:MM")
		return
	}

	session := &models.Session{
		ID:        uuid.New(),
		Name:      req.Name,
		StartTime: startTime,
		EndTime:   endTime,
	}

	if err := h.Sessions.Create(r.Context(), session); err != nil {
		writeError(w, http.StatusConflict, "conflict", "session conflict")
		return
	}

	writeJSON(w, http.StatusCreated, session)
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

	var req sessionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.StartTime != "" {
		t, err := time.Parse("15:04", req.StartTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid startTime format")
			return
		}
		existing.StartTime = t
	}
	if req.EndTime != "" {
		t, err := time.Parse("15:04", req.EndTime)
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

	writeJSON(w, http.StatusOK, existing)
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
