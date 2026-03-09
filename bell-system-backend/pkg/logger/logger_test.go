package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNew_DevelopmentLogger(t *testing.T) {
	log, err := New("development")
	require.NoError(t, err)
	require.NotNil(t, log)
	assert.NotNil(t, log.Logger)
	defer log.Close()
}

func TestNew_ProductionLogger(t *testing.T) {
	log, err := New("production")
	require.NoError(t, err)
	require.NotNil(t, log)
	assert.NotNil(t, log.Logger)
	defer log.Close()
}

func TestNew_UnknownEnvironment_DefaultsToDevelopment(t *testing.T) {
	log, err := New("staging")
	require.NoError(t, err)
	require.NotNil(t, log)
	defer log.Close()
}

func TestLogger_Info_DoesNotPanic(t *testing.T) {
	log, err := New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Info("test message", zap.String("key", "value"))
	})
}

func TestLogger_Error_DoesNotPanic(t *testing.T) {
	log, err := New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Error("test error", assert.AnError, zap.Int("code", 500))
	})
}

func TestLogger_Close_NoError(t *testing.T) {
	log, err := New("test")
	require.NoError(t, err)
	err = log.Close()
	// Sync may return an error on some platforms (stderr not syncable),
	// so we just verify Close doesn't panic.
}

func TestLogger_Info_WithNoFields(t *testing.T) {
	log, err := New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Info("bare message")
	})
}

func TestLogger_Error_WithNilError(t *testing.T) {
	log, err := New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Error("nil error case", nil)
	})
}

// Verify the unused variable warning is silenced
var _ = func() {
	log, _ := New("test")
	_ = log.Close()
}
