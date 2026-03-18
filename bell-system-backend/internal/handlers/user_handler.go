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

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	Users     UserRepository
	Passwords PasswordHasher
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(users UserRepository, passwords PasswordHasher) *UserHandler {
	return &UserHandler{
		Users:     users,
		Passwords: passwords,
	}
}

// List handles GET /.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.Users.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}
	writeJSON(w, http.StatusOK, models.ListResponse{Total: len(users), Items: users})
}

// GetByID handles GET /{id}.
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid user ID")
		return
	}

	user, err := h.Users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// Create handles POST /.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Username == "" || req.Password == "" || req.Role == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "username, password, and role are required")
		return
	}

	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "invalid_request", "password must be at least 8 characters")
		return
	}

	if req.Role != models.RoleAdmin && req.Role != models.RoleMorningUser && req.Role != models.RoleAfternoonUser {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid role")
		return
	}

	existing, _ := h.Users.GetByUsername(r.Context(), req.Username)
	if existing != nil {
		writeError(w, http.StatusConflict, "conflict", "username already exists")
		return
	}

	hash, err := h.Passwords.Hash(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to hash password")
		return
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: hash,
		Role:         req.Role,
		CreatedAt:    time.Now(),
	}

	if err := h.Users.Create(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// Update handles PUT /{id}.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid user ID")
		return
	}

	existing, err := h.Users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	var req models.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.Role != "" {
		existing.Role = req.Role
	}
	if req.Password != "" {
		hash, err := h.Passwords.Hash(req.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to hash password")
			return
		}
		existing.PasswordHash = hash
	}

	if err := h.Users.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// Delete handles DELETE /{id}.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid user ID")
		return
	}

	existing, err := h.Users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	if err := h.Users.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
