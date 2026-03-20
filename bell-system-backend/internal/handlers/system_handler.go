package handlers

import (
	"net/http"

	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
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

type systemStateSetAttributes struct {
	State string `json:"state"`
}

// GetState handles GET /api/system/state.
func (h *SystemHandler) GetState(w http.ResponseWriter, r *http.Request) {
	state, updatedAt, err := h.State.GetState(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get system state")
		return
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{
		Data: resourcePtr(jsonapi.MarshalSystemState(state, updatedAt)),
	})
}

// SetState handles POST /api/system/state.
func (h *SystemHandler) SetState(w http.ResponseWriter, r *http.Request) {
	doc, err := jsonapi.ParseRequest(r, "system-state")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var attrs systemStateSetAttributes
	if err := doc.UnmarshalAttributes(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid attributes")
		return
	}

	if attrs.State != "active" && attrs.State != "paused" {
		writeError(w, http.StatusBadRequest, "invalid_request", "state must be 'active' or 'paused'")
		return
	}

	if err := h.State.SetState(r.Context(), attrs.State); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to set system state")
		return
	}

	if h.Notifier != nil {
		h.Notifier.NotifySystemStateChanged(attrs.State)
	}

	state, updatedAt, err := h.State.GetState(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get updated state")
		return
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{
		Data: resourcePtr(jsonapi.MarshalSystemState(state, updatedAt)),
	})
}

// CancelNextBell handles POST /api/system/cancel-next-bell.
func (h *SystemHandler) CancelNextBell(w http.ResponseWriter, r *http.Request) {
	if h.Scheduler == nil {
		writeError(w, http.StatusNotFound, "not_found", "no pending bells to cancel")
		return
	}
	_, err := h.Scheduler.CancelNextBell()
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "no pending bells to cancel")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
