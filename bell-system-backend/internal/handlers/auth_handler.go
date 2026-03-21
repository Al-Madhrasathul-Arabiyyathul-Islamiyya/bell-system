package handlers

import (
	"encoding/json"
	"net/http"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	Users     UserRepository
	Tokens    TokenService
	Passwords PasswordHasher
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(users UserRepository, tokens TokenService, passwords PasswordHasher) *AuthHandler {
	return &AuthHandler{
		Users:     users,
		Tokens:    tokens,
		Passwords: passwords,
	}
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// loginResponse is the login endpoint response — plain JSON with a JSON:API user resource.
type loginResponse struct {
	Token string           `json:"token"`
	User  jsonapi.Resource `json:"user"`
}

// Login handles POST /login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "username and password are required")
		return
	}

	user, err := h.Users.GetByUsername(r.Context(), req.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to look up user")
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	if err := h.Passwords.Compare(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	token, err := h.Tokens.GenerateToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User:  jsonapi.MarshalUser(user),
	})
}

// ChangePassword handles POST /change-password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := GetUserClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "old and new passwords are required")
		return
	}

	if req.OldPassword == req.NewPassword {
		writeError(w, http.StatusBadRequest, "invalid_request", "new password must be different")
		return
	}

	user, err := h.Users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to look up user")
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	if err := h.Passwords.Compare(user.PasswordHash, req.OldPassword); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "unprocessable_entity", "incorrect old password")
		return
	}

	newHash, err := h.Passwords.Hash(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to hash password")
		return
	}

	user.PasswordHash = newHash
	if err := h.Users.Update(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Logout handles POST /logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
