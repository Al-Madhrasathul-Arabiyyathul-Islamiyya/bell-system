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

func TestLocalFileStorage_Save_CreateFileFails(t *testing.T) {
	// Use a path where a file exists instead of a directory, so os.Create fails
	dir := t.TempDir()
	blocker := filepath.Join(dir, "subdir")
	err := os.WriteFile(blocker, []byte("I am a file"), 0644)
	require.NoError(t, err)

	// BaseDir points to a file, so filepath.Join(BaseDir, id+ext) tries to create inside a file
	fs := &services.LocalFileStorage{BaseDir: blocker}
	_, _, err = fs.Save("test-id", ".wav", strings.NewReader("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")
}

// errReader is a reader that always returns an error.
type errReader struct{}

func (e errReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestLocalFileStorage_Save_CopyFails(t *testing.T) {
	dir := t.TempDir()
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)

	path, checksum, err := fs.Save("bad-reader", ".wav", errReader{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write file")
	assert.Empty(t, path)
	assert.Empty(t, checksum)

	// Verify the partial file was cleaned up
	_, statErr := os.Stat(filepath.Join(dir, "bad-reader.wav"))
	assert.True(t, os.IsNotExist(statErr), "partial file should be removed after copy failure")
}

func TestLocalFileStorage_Delete_NonEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	fs, err := services.NewLocalFileStorage(dir)
	require.NoError(t, err)

	// Create a subdirectory with a file inside — os.Remove on a non-empty dir
	// returns an error that is not os.ErrNotExist, hitting the error-wrapping path.
	subdir := filepath.Join(dir, "nonempty")
	require.NoError(t, os.Mkdir(subdir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "child.txt"), []byte("x"), 0644))

	err = fs.Delete(subdir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete file")
}

func TestLocalFileStorage_NewCreatesDirectory_Fails(t *testing.T) {
	// Use an invalid path that cannot be created
	// On Windows, NUL is a reserved name; on Unix, /dev/null/subdir won't work
	invalidPath := filepath.Join(string([]byte{0}), "impossible")
	fs, err := services.NewLocalFileStorage(invalidPath)
	assert.Error(t, err)
	assert.Nil(t, fs)
	assert.Contains(t, err.Error(), "failed to create audio directory")
}
