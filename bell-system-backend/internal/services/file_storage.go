package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
)

// Compile-time interface check.
var _ handlers.FileStorage = (*LocalFileStorage)(nil)

// LocalFileStorage stores audio files on the local filesystem.
type LocalFileStorage struct {
	BaseDir string
}

// NewLocalFileStorage creates a new LocalFileStorage and ensures the directory exists.
func NewLocalFileStorage(baseDir string) (*LocalFileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio directory: %w", err)
	}
	return &LocalFileStorage{BaseDir: baseDir}, nil
}

// Save writes file data to disk and returns the path and SHA-256 checksum.
func (fs *LocalFileStorage) Save(id string, ext string, data io.Reader) (string, string, error) {
	path := filepath.Join(fs.BaseDir, id+ext)

	f, err := os.Create(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}

	h := sha256.New()
	if _, err := io.Copy(f, io.TeeReader(data, h)); err != nil {
		f.Close()
		os.Remove(path)
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", "", fmt.Errorf("failed to close file: %w", err)
	}

	return path, hex.EncodeToString(h.Sum(nil)), nil
}

// Open returns a reader for the file at the given path.
func (fs *LocalFileStorage) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// Delete removes the file at the given path. Returns nil if the file does not exist.
func (fs *LocalFileStorage) Delete(path string) error {
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
