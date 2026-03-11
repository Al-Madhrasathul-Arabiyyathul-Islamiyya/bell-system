package mocks

import "io"

// MockFileStorage is a mock implementation of handlers.FileStorage.
type MockFileStorage struct {
	SaveFunc   func(id, ext string, data io.Reader) (string, string, error)
	OpenFunc   func(path string) (io.ReadCloser, error)
	DeleteFunc func(path string) error
}

func (m *MockFileStorage) Save(id, ext string, data io.Reader) (string, string, error) {
	return m.SaveFunc(id, ext, data)
}

func (m *MockFileStorage) Open(path string) (io.ReadCloser, error) {
	return m.OpenFunc(path)
}

func (m *MockFileStorage) Delete(path string) error {
	return m.DeleteFunc(path)
}
