package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	pkgerrors "arabiyya.edu.mv/bell-system-backend/pkg/errors"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAudioRouter creates an audio chi.Router with the given mock.
func newAudioRouter(audioRepo handlers.SystemAudioFileRepository) http.Handler {
	h := handlers.NewAudioHandler(audioRepo)
	return router.AudioRoutes(h)
}

// --- GET / (list audio files) ---

func TestAudioHandler_List_Success(t *testing.T) {
	files := []*models.SystemAudioFile{
		{ID: uuid.New(), Name: "bell.wav", FileType: models.FileTypeBell, Checksum: "abc123"},
		{ID: uuid.New(), Name: "anthem.mp3", FileType: models.FileTypeAnthem, Checksum: "def456"},
	}

	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context) ([]*models.SystemAudioFile, error) {
			return files, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Total int                      `json:"total"`
		Items []models.SystemAudioFile `json:"items"`
	}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Items, 2)
}

func TestAudioHandler_List_Empty(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context) ([]*models.SystemAudioFile, error) {
			return []*models.SystemAudioFile{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_List_DBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context) ([]*models.SystemAudioFile, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- GET /{id} (get audio file by ID) ---

func TestAudioHandler_GetByID_Success(t *testing.T) {
	fileID := uuid.New()
	file := &models.SystemAudioFile{
		ID:       fileID,
		Name:     "bell.wav",
		FilePath: "/audio/bell.wav",
		FileType: models.FileTypeBell,
		Checksum: "abc123",
	}

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			assert.Equal(t, fileID, id)
			return file, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_GetByID_NotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_GetByID_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	req := httptest.NewRequest(http.MethodGet, "/not-valid", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- POST / (upload audio file, multipart) ---

func TestAudioHandler_Upload_Success(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		CreateFunc: func(_ context.Context, audio *models.SystemAudioFile) error {
			assert.Equal(t, "test-bell", audio.Name)
			assert.Equal(t, models.FileTypeBell, audio.FileType)
			return nil
		},
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test-bell")
	_ = writer.WriteField("type", "bell")
	part, _ := writer.CreateFormFile("file", "bell.wav")
	_, _ = part.Write([]byte("fake audio data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestAudioHandler_Upload_MissingFile(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test-bell")
	_ = writer.WriteField("type", "bell")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAudioHandler_Upload_InvalidType(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test")
	_ = writer.WriteField("type", "invalid_type")
	part, _ := writer.CreateFormFile("file", "test.wav")
	_, _ = part.Write([]byte("fake audio data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- PUT /{id} (update audio file metadata) ---

func TestAudioHandler_Update_Success(t *testing.T) {
	fileID := uuid.New()

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: id, Name: "old", FileType: models.FileTypeBell}, nil
		},
		UpdateFunc: func(_ context.Context, audio *models.SystemAudioFile) error {
			assert.Equal(t, "updated", audio.Name)
			return nil
		},
	}

	body := `{"name":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_Update_NotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, nil
		},
	}

	body := `{"name":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- DELETE /{id} (delete audio file) ---

func TestAudioHandler_Delete_Success(t *testing.T) {
	fileID := uuid.New()

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, id uuid.UUID) error {
			assert.Equal(t, fileID, id)
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+fileID.String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestAudioHandler_Delete_NotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_Delete_DBError(t *testing.T) {
	fileID := uuid.New()

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: fileID}, nil
		},
		DeleteFunc: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("delete failed")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+fileID.String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Error path tests ---

func TestAudioHandler_GetByID_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_GetByID_DBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_Upload_InvalidMultipart(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not multipart"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAudioHandler_Upload_DBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		CreateFunc: func(_ context.Context, _ *models.SystemAudioFile) error {
			return errors.New("create failed")
		},
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test")
	_ = writer.WriteField("type", "bell")
	part, _ := writer.CreateFormFile("file", "test.wav")
	_, _ = part.Write([]byte("fake"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_Update_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	body := `{"name":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/not-a-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAudioHandler_Update_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	body := `{"name":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_Update_GetByIDDBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, errors.New("database error")
		},
	}

	body := `{"name":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: uuid.New()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAudioHandler_Update_DBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: uuid.New()}, nil
		},
		UpdateFunc: func(_ context.Context, _ *models.SystemAudioFile) error {
			return errors.New("update failed")
		},
	}

	body := `{"name":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_Update_FileTypeOnly(t *testing.T) {
	fileID := uuid.New()
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: fileID, Name: "old", FileType: models.FileTypeBell}, nil
		},
		UpdateFunc: func(_ context.Context, audio *models.SystemAudioFile) error {
			assert.Equal(t, models.FileTypeAnthem, audio.FileType)
			assert.Equal(t, "old", audio.Name) // name unchanged
			return nil
		},
	}

	body := `{"fileType":"anthem"}`
	req := httptest.NewRequest(http.MethodPut, "/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_Delete_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	req := httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAudioHandler_Delete_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_Delete_GetByIDDBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- File type validation ---

func TestAudioHandler_FileTypes_AllValid(t *testing.T) {
	validTypes := []models.FileType{
		models.FileTypeBell,
		models.FileTypeAnthem,
		models.FileTypeSchoolSong,
		models.FileTypeOther,
	}

	for _, ft := range validTypes {
		t.Run(string(ft), func(t *testing.T) {
			repo := &mocks.MockSystemAudioFileRepo{
				CreateFunc: func(_ context.Context, audio *models.SystemAudioFile) error {
					assert.Equal(t, ft, audio.FileType)
					return nil
				},
			}

			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			_ = writer.WriteField("name", "test")
			_ = writer.WriteField("type", string(ft))
			part, _ := writer.CreateFormFile("file", "test.wav")
			_, _ = part.Write([]byte("fake"))
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			rr := httptest.NewRecorder()

			r := newAudioRouter(repo)
			r.ServeHTTP(rr, req)

			require.Equal(t, http.StatusCreated, rr.Code)
		})
	}
}
