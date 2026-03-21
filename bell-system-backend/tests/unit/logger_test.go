package unit_test

import (
	"testing"

	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNew_DevelopmentLogger(t *testing.T) {
	log, err := logger.New("development")
	require.NoError(t, err)
	require.NotNil(t, log)
	defer log.Close()
}

func TestNew_ProductionLogger(t *testing.T) {
	log, err := logger.New("production")
	require.NoError(t, err)
	require.NotNil(t, log)
	defer log.Close()
}

func TestNew_UnknownEnvironment_DefaultsToDevelopment(t *testing.T) {
	log, err := logger.New("staging")
	require.NoError(t, err)
	require.NotNil(t, log)
	defer log.Close()
}

func TestLogger_Info_DoesNotPanic(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Info("test message", zap.String("key", "value"))
	})
}

func TestLogger_Error_DoesNotPanic(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Error("test error", assert.AnError, zap.Int("code", 500))
	})
}

func TestLogger_Close_DoesNotPanic(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)
	// Sync may return an error on some platforms (stderr not syncable),
	// so we just verify Close doesn't panic.
	assert.NotPanics(t, func() {
		log.Close()
	})
}

func TestLogger_Fatal_LogsAndExits(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)
	defer log.Close()

	var exitCode int
	log.SetExitFunc(func(code int) { exitCode = code })

	log.Fatal("fatal error", assert.AnError, zap.String("context", "test"))
	assert.Equal(t, 1, exitCode)
}

func TestLogger_Fatal_NilError(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)
	defer log.Close()

	var exitCode int
	log.SetExitFunc(func(code int) { exitCode = code })

	log.Fatal("fatal with nil", nil)
	assert.Equal(t, 1, exitCode)
}

func TestLogger_Info_WithNoFields(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Info("bare message")
	})
}

func TestLogger_Error_WithNilError(t *testing.T) {
	log, err := logger.New("test")
	require.NoError(t, err)
	defer log.Close()

	assert.NotPanics(t, func() {
		log.Error("nil error case", nil)
	})
}
