package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	pkgerrors "arabiyya.edu.mv/bell-system-backend/pkg/errors"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"
	"arabiyya.edu.mv/bell-system-backend/tests/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAudioRouter creates an audio chi.Router with the given mock.
func newAudioRouter(audioRepo handlers.SystemAudioFileRepository) http.Handler {
	h := handlers.NewAudioHandler(audioRepo, nil)
	return router.AudioRoutes(h, testutil.PermissiveTokenService)
}

func authAudioReq(req *http.Request) *http.Request {
	testutil.SetAuthHeader(req)
	if req.Method == http.MethodPut {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}
	return req
}

// --- GET / (list audio files) ---

func TestAudioHandler_List_Success(t *testing.T) {
	files := []*models.SystemAudioFile{
		{ID: uuid.New(), Name: "bell.wav", FileType: models.FileTypeBell, Checksum: "abc123"},
		{ID: uuid.New(), Name: "anthem.mp3", FileType: models.FileTypeAnthem, Checksum: "def456"},
	}

	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context, _, _ int, _ string, _ []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
			return files, len(files), nil
		},
	}

	req := authAudioReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, jsonapi.ContentType, rr.Header().Get("Content-Type"))

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp.Meta["total"])
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "audio-files", resp.Data[0].Type)
}

func TestAudioHandler_List_Empty(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context, _, _ int, _ string, _ []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
			return []*models.SystemAudioFile{}, 0, nil
		},
	}

	req := authAudioReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_List_DBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context, _, _ int, _ string, _ []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
			return nil, 0, errors.New("database error")
		},
	}

	req := authAudioReq(httptest.NewRequest(http.MethodGet, "/", nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- GET /{id} (get audio file metadata) ---

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
	assert.Equal(t, jsonapi.ContentType, rr.Header().Get("Content-Type"))

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "audio-files", resp.Data.Type)
	assert.Equal(t, fileID.String(), resp.Data.ID)
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

// --- GET /{id}/content (stream audio content) ---

func TestAudioHandler_GetContent_Success(t *testing.T) {
	fileID := uuid.New()
	file := &models.SystemAudioFile{
		ID:       fileID,
		Name:     "bell",
		FilePath: "/audio/bell.mp3",
		FileType: models.FileTypeBell,
	}

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			assert.Equal(t, fileID, id)
			return file, nil
		},
	}
	fs := &mocks.MockFileStorage{
		OpenFunc: func(path string) (io.ReadCloser, error) {
			assert.Equal(t, "/audio/bell.mp3", path)
			return io.NopCloser(bytes.NewReader([]byte("fake audio data"))), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "audio/mpeg", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "bell.mp3")
	assert.Equal(t, "fake audio data", rr.Body.String())
}

func TestAudioHandler_GetContent_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	req := httptest.NewRequest(http.MethodGet, "/not-valid/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAudioHandler_GetContent_NotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_GetContent_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_GetContent_DBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, errors.New("db error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+uuid.New().String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_GetContent_NoFileStorage(t *testing.T) {
	fileID := uuid.New()
	file := &models.SystemAudioFile{
		ID:       fileID,
		Name:     "bell",
		FilePath: "/audio/bell.mp3",
		FileType: models.FileTypeBell,
	}

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return file, nil
		},
	}

	// No FileStorage set (nil) — handler created via newAudioRouter which doesn't set FileStorage
	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_GetContent_EmptyFilePath(t *testing.T) {
	fileID := uuid.New()
	file := &models.SystemAudioFile{
		ID:       fileID,
		Name:     "bell",
		FilePath: "",
		FileType: models.FileTypeBell,
	}

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return file, nil
		},
	}
	fs := &mocks.MockFileStorage{}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_GetContent_OpenError(t *testing.T) {
	fileID := uuid.New()
	file := &models.SystemAudioFile{
		ID:       fileID,
		Name:     "bell",
		FilePath: "/audio/bell.mp3",
		FileType: models.FileTypeBell,
	}

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return file, nil
		},
	}
	fs := &mocks.MockFileStorage{
		OpenFunc: func(_ string) (io.ReadCloser, error) {
			return nil, errors.New("file not found on disk")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_GetContent_UnknownExtension(t *testing.T) {
	fileID := uuid.New()
	file := &models.SystemAudioFile{
		ID:       fileID,
		Name:     "sound",
		FilePath: "/audio/sound.xyz",
		FileType: models.FileTypeBell,
	}

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return file, nil
		},
	}
	fs := &mocks.MockFileStorage{
		OpenFunc: func(_ string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("data"))), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"))
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
	r.ServeHTTP(rr, authAudioReq(req))

	require.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, jsonapi.ContentType, rr.Header().Get("Content-Type"))
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
	r.ServeHTTP(rr, authAudioReq(req))

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
	r.ServeHTTP(rr, authAudioReq(req))

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

	body := jsonapiBody("audio-files", map[string]string{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_Update_NotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, nil
		},
	}

	body := jsonapiBody("audio-files", map[string]string{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

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

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/"+fileID.String(), nil))
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

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
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

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/"+fileID.String(), nil))
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
	r.ServeHTTP(rr, authAudioReq(req))

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
	r.ServeHTTP(rr, authAudioReq(req))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_Update_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	body := jsonapiBody("audio-files", map[string]string{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/not-a-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAudioHandler_Update_ErrNotFound(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, pkgerrors.ErrNotFound
		},
	}

	body := jsonapiBody("audio-files", map[string]string{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_Update_GetByIDDBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return nil, errors.New("database error")
		},
	}

	body := jsonapiBody("audio-files", map[string]string{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: uuid.New()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

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

	body := jsonapiBody("audio-files", map[string]string{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

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

	body := jsonapiBody("audio-files", map[string]string{"fileType": "anthem"})
	req := httptest.NewRequest(http.MethodPut, "/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, authAudioReq(req))

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_Delete_InvalidUUID(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil))
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

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
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

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- Notifier coverage ---

func TestAudioHandler_Upload_NotifiesOnSuccess(t *testing.T) {
	notifier := &mocks.MockEventNotifier{}
	repo := &mocks.MockSystemAudioFileRepo{
		CreateFunc: func(_ context.Context, _ *models.SystemAudioFile) error { return nil },
	}

	h := handlers.NewAudioHandler(repo, notifier)
	r := router.AudioRoutes(h, testutil.PermissiveTokenService)

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
	r.ServeHTTP(rr, authAudioReq(req))

	require.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, notifier.AudioFilesUpdatedCalled)
}

func TestAudioHandler_Update_NotifiesOnSuccess(t *testing.T) {
	fileID := uuid.New()
	notifier := &mocks.MockEventNotifier{}
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: fileID, Name: "old", FileType: models.FileTypeBell}, nil
		},
		UpdateFunc: func(_ context.Context, _ *models.SystemAudioFile) error { return nil },
	}

	h := handlers.NewAudioHandler(repo, notifier)
	r := router.AudioRoutes(h, testutil.PermissiveTokenService)

	body := jsonapiBody("audio-files", map[string]string{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", jsonapi.ContentType)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, authAudioReq(req))

	require.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, notifier.AudioFilesUpdatedCalled)
}

func TestAudioHandler_Delete_NotifiesOnSuccess(t *testing.T) {
	fileID := uuid.New()
	notifier := &mocks.MockEventNotifier{}
	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{ID: id}, nil
		},
		DeleteFunc: func(_ context.Context, _ uuid.UUID) error { return nil },
	}

	h := handlers.NewAudioHandler(repo, notifier)
	r := router.AudioRoutes(h, testutil.PermissiveTokenService)

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/"+fileID.String(), nil))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.True(t, notifier.AudioFilesUpdatedCalled)
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
			r.ServeHTTP(rr, authAudioReq(req))

			require.Equal(t, http.StatusCreated, rr.Code)
		})
	}
}

// --- Pagination/filter/sort tests ---

func TestAudioHandler_List_WithPagination(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context, page, size int, _ string, _ []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
			assert.Equal(t, 2, page)
			assert.Equal(t, 5, size)
			return []*models.SystemAudioFile{
				{ID: uuid.New(), Name: "bell.mp3", FileType: models.FileTypeBell},
			}, 11, nil
		},
	}

	req := authReq(httptest.NewRequest(http.MethodGet, "/?page[number]=2&page[size]=5", nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(11), resp.Meta["total"])
	page := resp.Meta["page"].(map[string]any)
	assert.Equal(t, float64(2), page["number"])
}

func TestAudioHandler_List_WithFilter(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context, _, _ int, filterFileType string, _ []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
			assert.Equal(t, "bell", filterFileType)
			return []*models.SystemAudioFile{}, 0, nil
		},
	}

	req := authReq(httptest.NewRequest(http.MethodGet, "/?filter[fileType]=bell", nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestAudioHandler_List_WithSort(t *testing.T) {
	var capturedSorts []jsonapi.SortField
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context, _, _ int, _ string, sorts []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
			capturedSorts = sorts
			return []*models.SystemAudioFile{}, 0, nil
		},
	}

	req := authReq(httptest.NewRequest(http.MethodGet, "/?sort=-createdAt", nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, capturedSorts, 1)
	assert.Equal(t, "createdAt", capturedSorts[0].Field)
	assert.True(t, capturedSorts[0].Desc)
}

func TestAudioHandler_List_WithSparseFieldset(t *testing.T) {
	now := time.Now()
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context, _, _ int, _ string, _ []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
			return []*models.SystemAudioFile{
				{ID: uuid.New(), Name: "bell.wav", FileType: "bell", Checksum: "abc123", CreatedAt: now, UpdatedAt: now},
			}, 1, nil
		},
	}

	req := authReq(httptest.NewRequest(http.MethodGet, "/?fields[audio-files]=name,checksum", nil))
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)

	attrs := resp.Data[0].Attributes.(map[string]any)
	assert.Equal(t, "bell.wav", attrs["name"])
	assert.Equal(t, "abc123", attrs["checksum"])
	_, hasFileType := attrs["fileType"]
	assert.False(t, hasFileType, "fileType should be filtered out by sparse fieldset")
}
