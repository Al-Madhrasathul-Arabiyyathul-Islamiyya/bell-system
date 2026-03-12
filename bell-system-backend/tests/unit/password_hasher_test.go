package unit_test

import (
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPasswordHasher(t *testing.T) {
	h := services.NewPasswordHasher()
	require.NotNil(t, h)
}

func TestPasswordHasher_Hash(t *testing.T) {
	h := services.NewPasswordHasher()

	hash, err := h.Hash("password123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "password123", hash)
}

func TestPasswordHasher_Hash_DifferentSalts(t *testing.T) {
	h := services.NewPasswordHasher()

	hash1, err := h.Hash("samepassword")
	require.NoError(t, err)

	hash2, err := h.Hash("samepassword")
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "bcrypt should produce different hashes due to random salts")
}

func TestPasswordHasher_Compare_Success(t *testing.T) {
	h := services.NewPasswordHasher()

	hash, err := h.Hash("correctpassword")
	require.NoError(t, err)

	err = h.Compare(hash, "correctpassword")
	assert.NoError(t, err)
}

func TestPasswordHasher_Compare_WrongPassword(t *testing.T) {
	h := services.NewPasswordHasher()

	hash, err := h.Hash("correctpassword")
	require.NoError(t, err)

	err = h.Compare(hash, "wrongpassword")
	assert.Error(t, err)
}

func TestPasswordHasher_Compare_InvalidHash(t *testing.T) {
	h := services.NewPasswordHasher()

	err := h.Compare("not-a-valid-bcrypt-hash", "password")
	assert.Error(t, err)
}
