package mocks

import (
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/database"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/DATA-DOG/go-sqlmock"
)

// NewMockDB creates a sqlmock DB and returns both the mock and a *database.Repository.
// The DB is automatically closed when the test finishes.
func NewMockDB(t *testing.T) (*database.Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { log.Close() })

	repo := database.NewRepository(db, log)
	return repo, mock
}

// NewMockUserRepo creates a sqlmock-backed UserRepository for testing.
func NewMockUserRepo(t *testing.T) (*database.UserRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { log.Close() })

	return database.NewUserRepository(db, log), mock
}

// NewMockSessionRepo creates a sqlmock-backed SessionRepository for testing.
func NewMockSessionRepo(t *testing.T) (*database.SessionRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { log.Close() })

	return database.NewSessionRepository(db, log), mock
}

// NewMockScheduleDayRepo creates a sqlmock-backed ScheduleDayRepository for testing.
func NewMockScheduleDayRepo(t *testing.T) (*database.ScheduleDayRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { log.Close() })

	return database.NewScheduleDayRepository(db, log), mock
}

// NewMockSystemAudioFileRepo creates a sqlmock-backed SystemAudioFileRepository for testing.
func NewMockSystemAudioFileRepo(t *testing.T) (*database.SystemAudioFileRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { log.Close() })

	return database.NewSystemAudioFileRepository(db, log), mock
}

// NewMockSystemStateRepo creates a sqlmock-backed SystemStateRepository for testing.
func NewMockSystemStateRepo(t *testing.T) (*database.SystemStateRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { log.Close() })

	return database.NewSystemStateRepository(db, log), mock
}

// NewMockScheduleItemRepo creates a sqlmock-backed ScheduleItemRepository for testing.
// It shares a single mock DB across all sub-repositories.
func NewMockScheduleItemRepo(t *testing.T) (*database.ScheduleItemRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	t.Cleanup(func() { log.Close() })

	dayRepo := database.NewScheduleDayRepository(db, log)
	sessionRepo := database.NewSessionRepository(db, log)
	audioRepo := database.NewSystemAudioFileRepository(db, log)

	return database.NewScheduleItemRepository(db, log, dayRepo, sessionRepo, audioRepo), mock
}
