package handlers

import (
	"context"
	"io"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	List(ctx context.Context, page, size int, filterRole string, sorts []jsonapi.SortField) ([]*models.User, int, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SessionRepository defines the interface for session data access.
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetSessionsByIDs(ctx context.Context, sessionIDs []uuid.UUID) (map[uuid.UUID]*models.Session, error)
	List(ctx context.Context, sorts []jsonapi.SortField) ([]*models.Session, error)
	GetCurrentSession(ctx context.Context) (*models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ScheduleItemRepository defines the interface for schedule item data access.
type ScheduleItemRepository interface {
	Create(ctx context.Context, item *models.ScheduleItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.ScheduleItem, error)
	List(ctx context.Context, filterSessionID string, filterDay int, sortSQL string) ([]*models.ScheduleItem, error)
	GetCurrentSessionSchedules(ctx context.Context, sessionID uuid.UUID) ([]*models.ScheduleItem, error)
	Update(ctx context.Context, item *models.ScheduleItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ScheduleDayRepository defines the interface for schedule day data access.
type ScheduleDayRepository interface {
	Create(ctx context.Context, scheduleDay *models.ScheduleDay) error
	GetDaysForScheduleItems(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID][]int, error)
	Update(ctx context.Context, scheduleDay *models.ScheduleDay) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SystemAudioFileRepository defines the interface for system audio file data access.
type SystemAudioFileRepository interface {
	Create(ctx context.Context, audio *models.SystemAudioFile) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.SystemAudioFile, error)
	GetSoundsByIDs(ctx context.Context, soundIDs []uuid.UUID) (map[uuid.UUID]*models.SystemAudioFile, error)
	List(ctx context.Context, page, size int, filterFileType, sortSQL string) ([]*models.SystemAudioFile, int, error)
	Update(ctx context.Context, audio *models.SystemAudioFile) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SystemStateRepository defines the interface for system state data access.
type SystemStateRepository interface {
	GetState(ctx context.Context) (string, time.Time, error)
	SetState(ctx context.Context, state string) error
}

// SchedulerService defines the interface for scheduler operations.
type SchedulerService interface {
	CancelNextBell() (cancelledItemID string, err error)
}

// TokenService defines the interface for JWT token operations.
type TokenService interface {
	GenerateToken(user *models.User) (string, error)
	ValidateToken(token string) (*TokenClaims, error)
}

// TokenClaims represents the claims extracted from a validated JWT token.
type TokenClaims struct {
	UserID   uuid.UUID
	Username string
	Role     models.Role
}

// PasswordHasher defines the interface for password hashing operations.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// FileStorage defines the interface for audio file disk operations.
type FileStorage interface {
	Save(id string, ext string, data io.Reader) (path string, checksum string, err error)
	Open(path string) (io.ReadCloser, error)
	Delete(path string) error
}

// EventNotifier allows REST handlers to notify WebSocket clients of data changes.
type EventNotifier interface {
	NotifySchedulesUpdated()
	NotifyAudioFilesUpdated()
	NotifyBellTriggered(payload models.BellTriggeredPayload)
	NotifyBellCancelled(scheduleItemID string)
	NotifySystemStateChanged(state string)
	NotifySystemLog(level, message, source string)
}
