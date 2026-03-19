package unit_test

import (
	"strings"
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

func TestPasswordHasher_Hash_TooLong(t *testing.T) {
	h := services.NewPasswordHasher()

	// Go's bcrypt rejects passwords over 72 bytes
	longPassword := strings.Repeat("a", 73)
	hash, err := h.Hash(longPassword)
	assert.Error(t, err, "bcrypt should reject passwords over 72 bytes")
	assert.Empty(t, hash)
}

func TestPasswordHasher_Hash_ExactlyAtLimit(t *testing.T) {
	h := services.NewPasswordHasher()

	// 72 bytes is the max — should succeed
	password := strings.Repeat("a", 72)
	hash, err := h.Hash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = h.Compare(hash, password)
	assert.NoError(t, err)
}

func TestPasswordHasher_Compare_InvalidHash(t *testing.T) {
	h := services.NewPasswordHasher()

	err := h.Compare("not-a-valid-bcrypt-hash", "password")
	assert.Error(t, err)
}
