package mocks

import (
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
)

// MockSchedulerService is a function-field mock for handlers.SchedulerService.
type MockSchedulerService struct {
	CancelNextBellFunc func() (string, error)
}

var _ handlers.SchedulerService = (*MockSchedulerService)(nil)

func (m *MockSchedulerService) CancelNextBell() (string, error) {
	if m.CancelNextBellFunc == nil {
		return "", nil
	}
	return m.CancelNextBellFunc()
}
