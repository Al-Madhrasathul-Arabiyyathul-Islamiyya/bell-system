package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger
type Logger struct {
	*zap.Logger
	exitFunc func(int) // injectable for testing Fatal; defaults to os.Exit
}

// New creates a new logger
func New(environment string) (*Logger, error) {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{Logger: logger, exitFunc: os.Exit}, nil
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	l.Logger.Error(msg, fields...)
}

// SetExitFunc overrides the function called by Fatal (default: os.Exit).
// Intended for testing.
func (l *Logger) SetExitFunc(f func(int)) {
	l.exitFunc = f
}

// Fatal logs the message at error level and terminates the process.
// The exit function is injectable for testing (defaults to os.Exit).
func (l *Logger) Fatal(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	l.Logger.Error(msg, fields...)
	l.exitFunc(1)
}

// Close flushes any buffered log entries
func (l *Logger) Close() error {
	return l.Logger.Sync()
}
