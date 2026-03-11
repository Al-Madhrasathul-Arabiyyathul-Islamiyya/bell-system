package unit_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalFileStorage_NewCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "subdir", "audio")
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)
	require.NotNil(t, fs)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestLocalFileStorage_Save(t *testing.T) {
	dir := t.TempDir()
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)

	content := "hello audio data"
	path, checksum, err := fs.Save("test-id", ".wav", strings.NewReader(content))
	require.NoError(t, err)

	// Verify path
	assert.Equal(t, filepath.Join(dir, "test-id.wav"), path)

	// Verify file contents
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	// Verify checksum
	h := sha256.Sum256([]byte(content))
	expected := hex.EncodeToString(h[:])
	assert.Equal(t, expected, checksum)
}

func TestLocalFileStorage_Open(t *testing.T) {
	dir := t.TempDir()
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)

	content := "test content for open"
	path, _, err := fs.Save("open-test", ".mp3", strings.NewReader(content))
	require.NoError(t, err)

	reader, err := fs.Open(path)
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestLocalFileStorage_Open_NotFound(t *testing.T) {
	dir := t.TempDir()
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)

	_, err = fs.Open(filepath.Join(dir, "nonexistent.wav"))
	assert.Error(t, err)
}

func TestLocalFileStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)

	path, _, err := fs.Save("delete-test", ".wav", strings.NewReader("data"))
	require.NoError(t, err)

	err = fs.Delete(path)
	require.NoError(t, err)

	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestLocalFileStorage_Delete_NonexistentFile(t *testing.T) {
	dir := t.TempDir()
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)

	err = fs.Delete(filepath.Join(dir, "does-not-exist.wav"))
	assert.NoError(t, err)
}
