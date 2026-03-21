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
	p := jsonapi.ParsePagination(r)
	filters := jsonapi.ParseFilter(r, []string{"role"})

	users, total, err := h.Users.List(r.Context(), p.Page, p.Size, filters["role"], jsonapi.ParseSort(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list users")
		return
	}

	resources := make([]jsonapi.Resource, len(users))
	for i, u := range users {
		resources[i] = jsonapi.MarshalUser(u)
	}

	writeJSONAPI(w, http.StatusOK, jsonapi.CollectionDocument{
		Data:  resources,
		Meta:  jsonapi.PaginationMeta(total, p.Page, p.Size),
		Links: jsonapi.PaginationLinks("/api/v1/users", p, total),
	})
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

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalUser(user))})
}

type userCreateAttributes struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type userUpdateAttributes struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
}

// Create handles POST /.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	doc, err := jsonapi.ParseRequest(r, "users")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var attrs userCreateAttributes
	if err := doc.UnmarshalAttributes(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid attributes")
		return
	}

	if attrs.Username == "" || attrs.Password == "" || attrs.Role == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "username, password, and role are required")
		return
	}

	if len(attrs.Password) < 8 {
		writeError(w, http.StatusBadRequest, "invalid_request", "password must be at least 8 characters")
		return
	}

	role := models.Role(attrs.Role)
	if role != models.RoleAdmin && role != models.RoleMorningUser && role != models.RoleAfternoonUser {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid role")
		return
	}

	existing, _ := h.Users.GetByUsername(r.Context(), attrs.Username)
	if existing != nil {
		writeError(w, http.StatusConflict, "conflict", "username already exists")
		return
	}

	hash, err := h.Passwords.Hash(attrs.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to hash password")
		return
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     attrs.Username,
		PasswordHash: hash,
		Role:         role,
		CreatedAt:    time.Now(),
	}

	if err := h.Users.Create(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create user")
		return
	}

	writeJSONAPI(w, http.StatusCreated, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalUser(user))})
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

	doc, err := jsonapi.ParseRequest(r, "users")
	if err != nil {
		// Fall back to flat JSON for backwards compat during migration
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var attrs userUpdateAttributes
	if err := doc.UnmarshalAttributes(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid attributes")
		return
	}

	if attrs.Username != "" {
		existing.Username = attrs.Username
	}
	if attrs.Role != "" {
		existing.Role = models.Role(attrs.Role)
	}
	if attrs.Password != "" {
		hash, err := h.Passwords.Hash(attrs.Password)
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

	writeJSONAPI(w, http.StatusOK, jsonapi.Document{Data: resourcePtr(jsonapi.MarshalUser(existing))})
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

// resourcePtr returns a pointer to the given resource.
func resourcePtr(v jsonapi.Resource) *jsonapi.Resource {
	return &v
}
