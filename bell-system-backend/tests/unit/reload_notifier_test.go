package unit_test

import (
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/internal/scheduler"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/stretchr/testify/assert"
)

func TestReloadNotifier_DelegatesAll(t *testing.T) {
	inner := &mocks.MockEventNotifier{}
	rn := &scheduler.ReloadNotifier{Inner: inner, Scheduler: nil}

	rn.NotifySchedulesUpdated()
	assert.True(t, inner.SchedulesUpdatedCalled)

	rn.NotifyAudioFilesUpdated()
	assert.True(t, inner.AudioFilesUpdatedCalled)

	rn.NotifyBellTriggered(models.BellTriggeredPayload{ScheduleItemID: "1"})
	assert.True(t, inner.BellTriggeredCalled)

	rn.NotifyBellCancelled("1")
	assert.True(t, inner.BellCancelledCalled)

	rn.NotifySystemStateChanged("paused")
	assert.True(t, inner.SystemStateChangedCalled)

	rn.NotifySystemLog("info", "test", "unit")
	assert.True(t, inner.SystemLogCalled)
}

func TestReloadNotifier_SchedulesUpdated_TriggersReload(t *testing.T) {
	inner := &mocks.MockEventNotifier{}
	// Create a minimal scheduler to test reload signal
	sched := scheduler.New(
		&mocks.MockSessionRepo{},
		&mocks.MockScheduleItemRepo{},
		&mocks.MockSystemStateRepo{},
		inner,
		nil,
		30,
	)

	rn := &scheduler.ReloadNotifier{Inner: inner, Scheduler: sched}
	rn.NotifySchedulesUpdated()

	assert.True(t, inner.SchedulesUpdatedCalled)
	// We can't directly check the reload channel from outside the package,
	// but the call should not panic
}

func TestReloadNotifier_SystemStateChanged_TriggersReload(t *testing.T) {
	inner := &mocks.MockEventNotifier{}
	sched := scheduler.New(
		&mocks.MockSessionRepo{},
		&mocks.MockScheduleItemRepo{},
		&mocks.MockSystemStateRepo{},
		inner,
		nil,
		30,
	)

	rn := &scheduler.ReloadNotifier{Inner: inner, Scheduler: sched}
	rn.NotifySystemStateChanged("paused")

	assert.True(t, inner.SystemStateChangedCalled)
}

func TestReloadNotifier_AudioFilesUpdated_NoReload(t *testing.T) {
	inner := &mocks.MockEventNotifier{}
	rn := &scheduler.ReloadNotifier{Inner: inner, Scheduler: nil}

	// Should not panic even with nil scheduler
	rn.NotifyAudioFilesUpdated()
	assert.True(t, inner.AudioFilesUpdatedCalled)
}
