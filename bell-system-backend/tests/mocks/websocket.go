package mocks

import (
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// MockEventNotifier is a function-field mock for handlers.EventNotifier.
type MockEventNotifier struct {
	NotifySchedulesUpdatedFunc   func()
	NotifyAudioFilesUpdatedFunc  func()
	NotifyBellTriggeredFunc      func(payload models.BellTriggeredPayload)
	NotifyBellCancelledFunc      func(scheduleItemID string)
	NotifySystemStateChangedFunc func(state string)
	NotifySystemLogFunc          func(level, message, source string)
	SchedulesUpdatedCalled       bool
	AudioFilesUpdatedCalled      bool
	BellTriggeredCalled          bool
	BellCancelledCalled          bool
	SystemStateChangedCalled     bool
	SystemLogCalled              bool
}

var _ handlers.EventNotifier = (*MockEventNotifier)(nil)

func (m *MockEventNotifier) NotifySchedulesUpdated() {
	m.SchedulesUpdatedCalled = true
	if m.NotifySchedulesUpdatedFunc != nil {
		m.NotifySchedulesUpdatedFunc()
	}
}

func (m *MockEventNotifier) NotifyAudioFilesUpdated() {
	m.AudioFilesUpdatedCalled = true
	if m.NotifyAudioFilesUpdatedFunc != nil {
		m.NotifyAudioFilesUpdatedFunc()
	}
}

func (m *MockEventNotifier) NotifyBellTriggered(payload models.BellTriggeredPayload) {
	m.BellTriggeredCalled = true
	if m.NotifyBellTriggeredFunc != nil {
		m.NotifyBellTriggeredFunc(payload)
	}
}

func (m *MockEventNotifier) NotifyBellCancelled(scheduleItemID string) {
	m.BellCancelledCalled = true
	if m.NotifyBellCancelledFunc != nil {
		m.NotifyBellCancelledFunc(scheduleItemID)
	}
}

func (m *MockEventNotifier) NotifySystemStateChanged(state string) {
	m.SystemStateChangedCalled = true
	if m.NotifySystemStateChangedFunc != nil {
		m.NotifySystemStateChangedFunc(state)
	}
}

func (m *MockEventNotifier) NotifySystemLog(level, message, source string) {
	m.SystemLogCalled = true
	if m.NotifySystemLogFunc != nil {
		m.NotifySystemLogFunc(level, message, source)
	}
}
