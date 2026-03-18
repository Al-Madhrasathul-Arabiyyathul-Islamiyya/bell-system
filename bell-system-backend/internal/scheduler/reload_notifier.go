package scheduler

import (
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// ReloadNotifier wraps an EventNotifier and triggers a scheduler reload
// whenever schedules or system state change.
type ReloadNotifier struct {
	Inner     handlers.EventNotifier
	Scheduler *Scheduler
}

var _ handlers.EventNotifier = (*ReloadNotifier)(nil)

func (rn *ReloadNotifier) NotifySchedulesUpdated() {
	rn.Inner.NotifySchedulesUpdated()
	if rn.Scheduler != nil {
		rn.Scheduler.Reload()
	}
}

func (rn *ReloadNotifier) NotifyAudioFilesUpdated() {
	rn.Inner.NotifyAudioFilesUpdated()
}

func (rn *ReloadNotifier) NotifyBellTriggered(payload models.BellTriggeredPayload) {
	rn.Inner.NotifyBellTriggered(payload)
}

func (rn *ReloadNotifier) NotifyBellCancelled(scheduleItemID string) {
	rn.Inner.NotifyBellCancelled(scheduleItemID)
}

func (rn *ReloadNotifier) NotifySystemStateChanged(state string) {
	rn.Inner.NotifySystemStateChanged(state)
	if rn.Scheduler != nil {
		rn.Scheduler.Reload()
	}
}

func (rn *ReloadNotifier) NotifySystemLog(level, message, source string) {
	rn.Inner.NotifySystemLog(level, message, source)
}
