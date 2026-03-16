package handlers

import (
	"encoding/json"
	"net/http"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// SystemHandler handles system state HTTP requests.
type SystemHandler struct {
	State     SystemStateRepository
	Scheduler SchedulerService
	Notifier  EventNotifier
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(state SystemStateRepository, scheduler SchedulerService, notifier EventNotifier) *SystemHandler {
	return &SystemHandler{State: state, Scheduler: scheduler, Notifier: notifier}
}

// GetState handles GET /api/system/state.
func (h *SystemHandler) GetState(w http.ResponseWriter, r *http.Request) {
	state, updatedAt, err := h.State.GetState(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get system state")
		return
	}

	writeJSON(w, http.StatusOK, models.SystemStateResponse{
		State:       state,
		LastUpdated: updatedAt,
	})
}

// SetState handles POST /api/system/state.
func (h *SystemHandler) SetState(w http.ResponseWriter, r *http.Request) {
	var req models.SystemStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.State != "active" && req.State != "paused" {
		writeError(w, http.StatusBadRequest, "invalid_request", "state must be 'active' or 'paused'")
		return
	}

	if err := h.State.SetState(r.Context(), req.State); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to set system state")
		return
	}

	if h.Notifier != nil {
		h.Notifier.NotifySystemStateChanged(req.State)
	}

	state, updatedAt, err := h.State.GetState(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get updated state")
		return
	}

	writeJSON(w, http.StatusOK, models.SystemStateResponse{
		State:       state,
		LastUpdated: updatedAt,
	})
}

// CancelNextBell handles POST /api/system/cancel-next-bell.
func (h *SystemHandler) CancelNextBell(w http.ResponseWriter, r *http.Request) {
	if h.Scheduler == nil {
		writeError(w, http.StatusNotFound, "not_found", "no pending bells to cancel")
		return
	}
	cancelledID, err := h.Scheduler.CancelNextBell()
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "no pending bells to cancel")
		return
	}

	writeJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "cancelled bell for schedule item " + cancelledID,
	})
}
