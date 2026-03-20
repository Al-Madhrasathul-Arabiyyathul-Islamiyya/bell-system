package mocks

import (
	"context"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"

	"github.com/google/uuid"
)

// --- Mock Repository Implementations ---

// MockUserRepo is a function-field mock for handlers.UserRepository.
type MockUserRepo struct {
	CreateFunc        func(ctx context.Context, user *models.User) error
	GetByIDFunc       func(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsernameFunc func(ctx context.Context, username string) (*models.User, error)
	ListFunc          func(ctx context.Context, page, size int, filterRole, sortSQL string) ([]*models.User, int, error)
	UpdateFunc        func(ctx context.Context, user *models.User) error
	DeleteFunc        func(ctx context.Context, id uuid.UUID) error
}

var _ handlers.UserRepository = (*MockUserRepo)(nil)

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	if m.CreateFunc == nil {
		panic("MockUserRepo.CreateFunc not set")
	}
	return m.CreateFunc(ctx, user)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.GetByIDFunc == nil {
		panic("MockUserRepo.GetByIDFunc not set")
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.GetByUsernameFunc == nil {
		panic("MockUserRepo.GetByUsernameFunc not set")
	}
	return m.GetByUsernameFunc(ctx, username)
}

func (m *MockUserRepo) List(ctx context.Context, page, size int, filterRole, sortSQL string) ([]*models.User, int, error) {
	if m.ListFunc == nil {
		panic("MockUserRepo.ListFunc not set")
	}
	return m.ListFunc(ctx, page, size, filterRole, sortSQL)
}

func (m *MockUserRepo) Update(ctx context.Context, user *models.User) error {
	if m.UpdateFunc == nil {
		panic("MockUserRepo.UpdateFunc not set")
	}
	return m.UpdateFunc(ctx, user)
}

func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc == nil {
		panic("MockUserRepo.DeleteFunc not set")
	}
	return m.DeleteFunc(ctx, id)
}

// MockSessionRepo is a function-field mock for handlers.SessionRepository.
type MockSessionRepo struct {
	CreateFunc            func(ctx context.Context, session *models.Session) error
	GetByIDFunc           func(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetSessionsByIDsFunc  func(ctx context.Context, sessionIDs []uuid.UUID) (map[uuid.UUID]*models.Session, error)
	ListFunc              func(ctx context.Context, sortSQL string) ([]*models.Session, error)
	GetCurrentSessionFunc func(ctx context.Context) (*models.Session, error)
	UpdateFunc            func(ctx context.Context, session *models.Session) error
	DeleteFunc            func(ctx context.Context, id uuid.UUID) error
}

var _ handlers.SessionRepository = (*MockSessionRepo)(nil)

func (m *MockSessionRepo) Create(ctx context.Context, session *models.Session) error {
	if m.CreateFunc == nil {
		panic("MockSessionRepo.CreateFunc not set")
	}
	return m.CreateFunc(ctx, session)
}

func (m *MockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	if m.GetByIDFunc == nil {
		panic("MockSessionRepo.GetByIDFunc not set")
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockSessionRepo) GetSessionsByIDs(ctx context.Context, sessionIDs []uuid.UUID) (map[uuid.UUID]*models.Session, error) {
	if m.GetSessionsByIDsFunc == nil {
		panic("MockSessionRepo.GetSessionsByIDsFunc not set")
	}
	return m.GetSessionsByIDsFunc(ctx, sessionIDs)
}

func (m *MockSessionRepo) List(ctx context.Context, sortSQL string) ([]*models.Session, error) {
	if m.ListFunc == nil {
		panic("MockSessionRepo.ListFunc not set")
	}
	return m.ListFunc(ctx, sortSQL)
}

func (m *MockSessionRepo) GetCurrentSession(ctx context.Context) (*models.Session, error) {
	if m.GetCurrentSessionFunc == nil {
		panic("MockSessionRepo.GetCurrentSessionFunc not set")
	}
	return m.GetCurrentSessionFunc(ctx)
}

func (m *MockSessionRepo) Update(ctx context.Context, session *models.Session) error {
	if m.UpdateFunc == nil {
		panic("MockSessionRepo.UpdateFunc not set")
	}
	return m.UpdateFunc(ctx, session)
}

func (m *MockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc == nil {
		panic("MockSessionRepo.DeleteFunc not set")
	}
	return m.DeleteFunc(ctx, id)
}

// MockScheduleItemRepo is a function-field mock for handlers.ScheduleItemRepository.
type MockScheduleItemRepo struct {
	CreateFunc                     func(ctx context.Context, item *models.ScheduleItem) error
	GetByIDFunc                    func(ctx context.Context, id uuid.UUID) (*models.ScheduleItem, error)
	ListFunc                       func(ctx context.Context, filterSessionID string, filterDay int, sortSQL string) ([]*models.ScheduleItem, error)
	GetCurrentSessionSchedulesFunc func(ctx context.Context, sessionID uuid.UUID) ([]*models.ScheduleItem, error)
	UpdateFunc                     func(ctx context.Context, item *models.ScheduleItem) error
	DeleteFunc                     func(ctx context.Context, id uuid.UUID) error
}

var _ handlers.ScheduleItemRepository = (*MockScheduleItemRepo)(nil)

func (m *MockScheduleItemRepo) Create(ctx context.Context, item *models.ScheduleItem) error {
	if m.CreateFunc == nil {
		panic("MockScheduleItemRepo.CreateFunc not set")
	}
	return m.CreateFunc(ctx, item)
}

func (m *MockScheduleItemRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
	if m.GetByIDFunc == nil {
		panic("MockScheduleItemRepo.GetByIDFunc not set")
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockScheduleItemRepo) List(ctx context.Context, filterSessionID string, filterDay int, sortSQL string) ([]*models.ScheduleItem, error) {
	if m.ListFunc == nil {
		panic("MockScheduleItemRepo.ListFunc not set")
	}
	return m.ListFunc(ctx, filterSessionID, filterDay, sortSQL)
}

func (m *MockScheduleItemRepo) GetCurrentSessionSchedules(ctx context.Context, sessionID uuid.UUID) ([]*models.ScheduleItem, error) {
	if m.GetCurrentSessionSchedulesFunc == nil {
		panic("MockScheduleItemRepo.GetCurrentSessionSchedulesFunc not set")
	}
	return m.GetCurrentSessionSchedulesFunc(ctx, sessionID)
}

func (m *MockScheduleItemRepo) Update(ctx context.Context, item *models.ScheduleItem) error {
	if m.UpdateFunc == nil {
		panic("MockScheduleItemRepo.UpdateFunc not set")
	}
	return m.UpdateFunc(ctx, item)
}

func (m *MockScheduleItemRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc == nil {
		panic("MockScheduleItemRepo.DeleteFunc not set")
	}
	return m.DeleteFunc(ctx, id)
}

// MockScheduleDayRepo is a function-field mock for handlers.ScheduleDayRepository.
type MockScheduleDayRepo struct {
	CreateFunc                  func(ctx context.Context, scheduleDay *models.ScheduleDay) error
	GetDaysForScheduleItemsFunc func(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID][]int, error)
	UpdateFunc                  func(ctx context.Context, scheduleDay *models.ScheduleDay) error
	DeleteFunc                  func(ctx context.Context, id uuid.UUID) error
}

var _ handlers.ScheduleDayRepository = (*MockScheduleDayRepo)(nil)

func (m *MockScheduleDayRepo) Create(ctx context.Context, scheduleDay *models.ScheduleDay) error {
	if m.CreateFunc == nil {
		panic("MockScheduleDayRepo.CreateFunc not set")
	}
	return m.CreateFunc(ctx, scheduleDay)
}

func (m *MockScheduleDayRepo) GetDaysForScheduleItems(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID][]int, error) {
	if m.GetDaysForScheduleItemsFunc == nil {
		panic("MockScheduleDayRepo.GetDaysForScheduleItemsFunc not set")
	}
	return m.GetDaysForScheduleItemsFunc(ctx, itemIDs)
}

func (m *MockScheduleDayRepo) Update(ctx context.Context, scheduleDay *models.ScheduleDay) error {
	if m.UpdateFunc == nil {
		panic("MockScheduleDayRepo.UpdateFunc not set")
	}
	return m.UpdateFunc(ctx, scheduleDay)
}

func (m *MockScheduleDayRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc == nil {
		panic("MockScheduleDayRepo.DeleteFunc not set")
	}
	return m.DeleteFunc(ctx, id)
}

// MockSystemAudioFileRepo is a function-field mock for handlers.SystemAudioFileRepository.
type MockSystemAudioFileRepo struct {
	CreateFunc         func(ctx context.Context, audio *models.SystemAudioFile) error
	GetByIDFunc        func(ctx context.Context, id uuid.UUID) (*models.SystemAudioFile, error)
	GetSoundsByIDsFunc func(ctx context.Context, soundIDs []uuid.UUID) (map[uuid.UUID]*models.SystemAudioFile, error)
	ListFunc           func(ctx context.Context, page, size int, filterFileType, sortSQL string) ([]*models.SystemAudioFile, int, error)
	UpdateFunc         func(ctx context.Context, audio *models.SystemAudioFile) error
	DeleteFunc         func(ctx context.Context, id uuid.UUID) error
}

var _ handlers.SystemAudioFileRepository = (*MockSystemAudioFileRepo)(nil)

func (m *MockSystemAudioFileRepo) Create(ctx context.Context, audio *models.SystemAudioFile) error {
	if m.CreateFunc == nil {
		panic("MockSystemAudioFileRepo.CreateFunc not set")
	}
	return m.CreateFunc(ctx, audio)
}

func (m *MockSystemAudioFileRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
	if m.GetByIDFunc == nil {
		panic("MockSystemAudioFileRepo.GetByIDFunc not set")
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockSystemAudioFileRepo) GetSoundsByIDs(ctx context.Context, soundIDs []uuid.UUID) (map[uuid.UUID]*models.SystemAudioFile, error) {
	if m.GetSoundsByIDsFunc == nil {
		panic("MockSystemAudioFileRepo.GetSoundsByIDsFunc not set")
	}
	return m.GetSoundsByIDsFunc(ctx, soundIDs)
}

func (m *MockSystemAudioFileRepo) List(ctx context.Context, page, size int, filterFileType, sortSQL string) ([]*models.SystemAudioFile, int, error) {
	if m.ListFunc == nil {
		panic("MockSystemAudioFileRepo.ListFunc not set")
	}
	return m.ListFunc(ctx, page, size, filterFileType, sortSQL)
}

func (m *MockSystemAudioFileRepo) Update(ctx context.Context, audio *models.SystemAudioFile) error {
	if m.UpdateFunc == nil {
		panic("MockSystemAudioFileRepo.UpdateFunc not set")
	}
	return m.UpdateFunc(ctx, audio)
}

func (m *MockSystemAudioFileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc == nil {
		panic("MockSystemAudioFileRepo.DeleteFunc not set")
	}
	return m.DeleteFunc(ctx, id)
}

// MockSystemStateRepo is a function-field mock for handlers.SystemStateRepository.
type MockSystemStateRepo struct {
	GetStateFunc func(ctx context.Context) (string, time.Time, error)
	SetStateFunc func(ctx context.Context, state string) error
}

var _ handlers.SystemStateRepository = (*MockSystemStateRepo)(nil)

func (m *MockSystemStateRepo) GetState(ctx context.Context) (string, time.Time, error) {
	if m.GetStateFunc == nil {
		return "active", time.Now(), nil
	}
	return m.GetStateFunc(ctx)
}

func (m *MockSystemStateRepo) SetState(ctx context.Context, state string) error {
	if m.SetStateFunc == nil {
		return nil
	}
	return m.SetStateFunc(ctx, state)
}
