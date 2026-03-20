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
	"strings"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"
	"arabiyya.edu.mv/bell-system-backend/tests/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAudioRouterWithStorage(audioRepo handlers.SystemAudioFileRepository, fs handlers.FileStorage) http.Handler {
	h := handlers.NewAudioHandler(audioRepo, nil)
	h.FileStorage = fs
	return router.AudioRoutes(h, testutil.PermissiveTokenService)
}

// --- Upload with FileStorage ---

func TestAudioHandler_Upload_WithFileStorage(t *testing.T) {
	var savedID, savedExt string
	var savedData []byte

	repo := &mocks.MockSystemAudioFileRepo{
		CreateFunc: func(_ context.Context, audio *models.SystemAudioFile) error {
			assert.Equal(t, "/audio/"+savedID+savedExt, audio.FilePath)
			assert.Equal(t, "abc123checksum", audio.Checksum)
			return nil
		},
	}

	fs := &mocks.MockFileStorage{
		SaveFunc: func(id, ext string, data io.Reader) (string, string, error) {
			savedID = id
			savedExt = ext
			savedData, _ = io.ReadAll(data)
			return "/audio/" + id + ext, "abc123checksum", nil
		},
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test-bell")
	_ = writer.WriteField("type", "bell")
	part, _ := writer.CreateFormFile("file", "bell.wav")
	_, _ = part.Write([]byte("real audio data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, authAudioReq(req))

	require.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, ".wav", savedExt)
	assert.Equal(t, "real audio data", string(savedData))

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "audio-files", resp.Data.Type)

	attrs, ok := resp.Data.Attributes.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "abc123checksum", attrs["checksum"])
	assert.Contains(t, attrs["filePath"], ".wav")
}

func TestAudioHandler_Upload_FileStorageSaveError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{}

	fs := &mocks.MockFileStorage{
		SaveFunc: func(_, _ string, _ io.Reader) (string, string, error) {
			return "", "", io.ErrUnexpectedEOF
		},
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("name", "test")
	_ = writer.WriteField("type", "bell")
	part, _ := writer.CreateFormFile("file", "test.wav")
	_, _ = part.Write([]byte("data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, authAudioReq(req))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- GetContent with FileStorage (streams binary) ---

func TestAudioHandler_GetContent_StreamsFile(t *testing.T) {
	fileID := uuid.New()
	fileContent := "fake audio binary content"

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{
				ID:       fileID,
				Name:     "bell",
				FilePath: "/audio/" + fileID.String() + ".wav",
				FileType: models.FileTypeBell,
				Checksum: "abc123",
			}, nil
		},
	}

	fs := &mocks.MockFileStorage{
		OpenFunc: func(path string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(fileContent)), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, fileContent, rr.Body.String())
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "bell.wav")
}

func TestAudioHandler_GetContent_UnknownExtension_FallsBackToOctetStream(t *testing.T) {
	fileID := uuid.New()
	fileContent := "binary data"

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{
				ID:       fileID,
				Name:     "custom",
				FilePath: "/audio/" + fileID.String() + ".bellsound",
				FileType: models.FileTypeOther,
			}, nil
		},
	}

	fs := &mocks.MockFileStorage{
		OpenFunc: func(path string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(fileContent)), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/octet-stream", rr.Header().Get("Content-Type"))
	assert.Equal(t, fileContent, rr.Body.String())
}

func TestAudioHandler_GetContent_FileOpenError(t *testing.T) {
	fileID := uuid.New()

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{
				ID:       fileID,
				Name:     "bell",
				FilePath: "/audio/" + fileID.String() + ".wav",
				FileType: models.FileTypeBell,
			}, nil
		},
	}

	fs := &mocks.MockFileStorage{
		OpenFunc: func(path string) (io.ReadCloser, error) {
			return nil, errors.New("disk read error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_GetContent_NoFilePath_Returns404(t *testing.T) {
	fileID := uuid.New()

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{
				ID:       fileID,
				Name:     "bell",
				FilePath: "",
				FileType: models.FileTypeBell,
			}, nil
		},
	}

	fs := &mocks.MockFileStorage{}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String()+"/content", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAudioHandler_GetByID_ReturnsMetadata(t *testing.T) {
	fileID := uuid.New()

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{
				ID:       fileID,
				Name:     "bell",
				FilePath: "/audio/" + fileID.String() + ".wav",
				FileType: models.FileTypeBell,
			}, nil
		},
	}

	fs := &mocks.MockFileStorage{}

	req := httptest.NewRequest(http.MethodGet, "/"+fileID.String(), nil)
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), jsonapi.ContentType)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "audio-files", resp.Data.Type)
	assert.Equal(t, fileID.String(), resp.Data.ID)
}

// --- Delete with FileStorage ---

func TestAudioHandler_Delete_RemovesFile(t *testing.T) {
	fileID := uuid.New()
	var deletedPath string

	repo := &mocks.MockSystemAudioFileRepo{
		GetByIDFunc: func(_ context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
			return &models.SystemAudioFile{
				ID:       fileID,
				FilePath: "/audio/" + fileID.String() + ".wav",
			}, nil
		},
		DeleteFunc: func(_ context.Context, id uuid.UUID) error {
			return nil
		},
	}

	fs := &mocks.MockFileStorage{
		DeleteFunc: func(path string) error {
			deletedPath = path
			return nil
		},
	}

	req := authAudioReq(httptest.NewRequest(http.MethodDelete, "/"+fileID.String(), nil))
	rr := httptest.NewRecorder()

	r := newAudioRouterWithStorage(repo, fs)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "/audio/"+fileID.String()+".wav", deletedPath)
}

// --- ListChecksums ---

func TestAudioHandler_ListChecksums_Success(t *testing.T) {
	files := []*models.SystemAudioFile{
		{ID: uuid.New(), FileType: models.FileTypeBell, Checksum: "aaa"},
		{ID: uuid.New(), FileType: models.FileTypeAnthem, Checksum: "bbb"},
	}

	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context) ([]*models.SystemAudioFile, error) {
			return files, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/checksums", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.CollectionDocument
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "audio-files", resp.Data[0].Type)
}

func TestAudioHandler_ListChecksums_DBError(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context) ([]*models.SystemAudioFile, error) {
			return nil, errors.New("database error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/checksums", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAudioHandler_ListChecksums_Empty(t *testing.T) {
	repo := &mocks.MockSystemAudioFileRepo{
		ListFunc: func(_ context.Context) ([]*models.SystemAudioFile, error) {
			return []*models.SystemAudioFile{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/checksums", nil)
	rr := httptest.NewRecorder()

	r := newAudioRouter(repo)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}
