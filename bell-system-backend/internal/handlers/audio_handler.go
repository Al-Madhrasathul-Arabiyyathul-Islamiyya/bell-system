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

// AudioHandler handles audio file management HTTP requests.
type AudioHandler struct {
	AudioFiles  SystemAudioFileRepository
	FileStorage FileStorage // nil = metadata-only mode (for tests)
}

// NewAudioHandler creates a new AudioHandler.
func NewAudioHandler(audioFiles SystemAudioFileRepository) *AudioHandler {
	return &AudioHandler{AudioFiles: audioFiles}
}

var validFileTypes = map[models.FileType]bool{
	models.FileTypeBell:       true,
	models.FileTypeAnthem:     true,
	models.FileTypeSchoolSong: true,
	models.FileTypeOther:      true,
}

// List handles GET /.
func (h *AudioHandler) List(w http.ResponseWriter, r *http.Request) {
	files, err := h.AudioFiles.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list audio files")
		return
	}
	writeJSON(w, http.StatusOK, models.ListResponse{Total: len(files), Items: files})
}

// GetByID handles GET /{id}.
func (h *AudioHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid audio file ID")
		return
	}

	file, err := h.AudioFiles.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "audio file not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get audio file")
		return
	}
	if file == nil {
		writeError(w, http.StatusNotFound, "not_found", "audio file not found")
		return
	}

	writeJSON(w, http.StatusOK, file)
}

// Upload handles POST / (multipart file upload).
func (h *AudioHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid multipart form")
		return
	}

	name := r.FormValue("name")
	fileType := models.FileType(r.FormValue("type"))

	if !validFileTypes[fileType] {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid file type")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "file is required")
		return
	}
	defer file.Close()

	audioFile := &models.SystemAudioFile{
		ID:        uuid.New(),
		Name:      name,
		FilePath:  header.Filename,
		FileType:  fileType,
		CreatedAt: time.Now(),
	}

	if err := h.AudioFiles.Create(r.Context(), audioFile); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create audio file")
		return
	}

	writeJSON(w, http.StatusCreated, audioFile)
}

type audioUpdateRequest struct {
	Name     string          `json:"name,omitempty"`
	FileType models.FileType `json:"fileType,omitempty"`
}

// Update handles PUT /{id}.
func (h *AudioHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid audio file ID")
		return
	}

	existing, err := h.AudioFiles.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "audio file not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get audio file")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "audio file not found")
		return
	}

	var req audioUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.FileType != "" {
		existing.FileType = req.FileType
	}

	if err := h.AudioFiles.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to update audio file")
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// Delete handles DELETE /{id}.
func (h *AudioHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid audio file ID")
		return
	}

	existing, err := h.AudioFiles.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "audio file not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get audio file")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "not_found", "audio file not found")
		return
	}

	if err := h.AudioFiles.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete audio file")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
