package mocks

import (
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// --- Mock Service Implementations ---

// MockTokenService is a function-field mock for handlers.TokenService.
type MockTokenService struct {
	GenerateTokenFunc func(user *models.User) (string, error)
	ValidateTokenFunc func(token string) (*handlers.TokenClaims, error)
}

var _ handlers.TokenService = (*MockTokenService)(nil)

func (m *MockTokenService) GenerateToken(user *models.User) (string, error) {
	if m.GenerateTokenFunc == nil {
		panic("MockTokenService.GenerateTokenFunc not set")
	}
	return m.GenerateTokenFunc(user)
}

func (m *MockTokenService) ValidateToken(token string) (*handlers.TokenClaims, error) {
	if m.ValidateTokenFunc == nil {
		panic("MockTokenService.ValidateTokenFunc not set")
	}
	return m.ValidateTokenFunc(token)
}

// MockPasswordHasher is a function-field mock for handlers.PasswordHasher.
type MockPasswordHasher struct {
	HashFunc    func(password string) (string, error)
	CompareFunc func(hash, password string) error
}

var _ handlers.PasswordHasher = (*MockPasswordHasher)(nil)

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	if m.HashFunc == nil {
		panic("MockPasswordHasher.HashFunc not set")
	}
	return m.HashFunc(password)
}

func (m *MockPasswordHasher) Compare(hash, password string) error {
	if m.CompareFunc == nil {
		panic("MockPasswordHasher.CompareFunc not set")
	}
	return m.CompareFunc(hash, password)
}
