package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Sentinel error tests ---

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []error{
		ErrNotFound,
		ErrDuplicateEntry,
		ErrInvalidData,
		ErrForeignKeyViolation,
		ErrTransactionFailed,
	}
	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j {
				assert.False(t, errors.Is(a, b), "%v should not match %v", a, b)
			}
		}
	}
}

func TestSentinelErrors_HaveMessages(t *testing.T) {
	tests := []struct {
		err     error
		message string
	}{
		{ErrNotFound, "entity not found"},
		{ErrDuplicateEntry, "duplicate entry"},
		{ErrInvalidData, "invalid data"},
		{ErrForeignKeyViolation, "foreign key violation"},
		{ErrTransactionFailed, "transaction failed"},
	}
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			assert.Equal(t, tt.message, tt.err.Error())
		})
	}
}

// --- NotFoundError tests ---

func TestNewNotFoundError_SetsFields(t *testing.T) {
	err := NewNotFoundError("User", "abc-123")
	require.NotNil(t, err)
	assert.Equal(t, "User", err.EntityType)
	assert.Equal(t, "abc-123", err.ID)
}

func TestNotFoundError_ErrorMessage(t *testing.T) {
	err := NewNotFoundError("Session", "some-uuid")
	assert.Equal(t, "Session with ID some-uuid not found", err.Error())
}

func TestNotFoundError_IsErrNotFound(t *testing.T) {
	err := NewNotFoundError("User", "123")
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestNotFoundError_IsNotOtherSentinels(t *testing.T) {
	err := NewNotFoundError("User", "123")
	assert.False(t, errors.Is(err, ErrDuplicateEntry))
	assert.False(t, errors.Is(err, ErrInvalidData))
	assert.False(t, errors.Is(err, ErrForeignKeyViolation))
	assert.False(t, errors.Is(err, ErrTransactionFailed))
}

func TestNotFoundError_WrappedInFmtErrorw_StillMatchesErrNotFound(t *testing.T) {
	inner := NewNotFoundError("ScheduleItem", "xyz")
	wrapped := fmt.Errorf("operation failed: %w", inner)
	assert.True(t, errors.Is(wrapped, ErrNotFound))
}

// --- DatabaseError tests ---

func TestNewDatabaseError_SetsFields(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewDatabaseError("create", "User", cause)
	require.NotNil(t, err)
	assert.Equal(t, "create", err.Operation)
	assert.Equal(t, "User", err.Entity)
	assert.Equal(t, cause, err.Err)
}

func TestDatabaseError_ErrorMessage(t *testing.T) {
	cause := errors.New("timeout")
	err := NewDatabaseError("list", "Session", cause)
	assert.Equal(t, "database error during list operation on Session: timeout", err.Error())
}

func TestDatabaseError_Unwrap(t *testing.T) {
	cause := ErrDuplicateEntry
	err := NewDatabaseError("create", "User", cause)
	assert.True(t, errors.Is(err, ErrDuplicateEntry))
	assert.Equal(t, cause, errors.Unwrap(err))
}

func TestDatabaseError_UnwrapChain(t *testing.T) {
	root := errors.New("disk full")
	mid := fmt.Errorf("write failed: %w", root)
	dbErr := NewDatabaseError("create", "SystemAudioFile", mid)

	assert.True(t, errors.Is(dbErr, root))
}

func TestDatabaseError_AsType(t *testing.T) {
	err := NewDatabaseError("update", "User", errors.New("oops"))
	wrapped := fmt.Errorf("handler error: %w", err)

	var dbErr *DatabaseError
	assert.True(t, errors.As(wrapped, &dbErr))
	assert.Equal(t, "update", dbErr.Operation)
	assert.Equal(t, "User", dbErr.Entity)
}
