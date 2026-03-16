package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"
)

type scheduledTimer struct {
	itemID  string
	soundID string
	name    string
	fireAt  time.Time
	timer   *time.Timer
}

// Scheduler runs as a background goroutine, scheduling bell triggers via time.AfterFunc.
type Scheduler struct {
	sessions handlers.SessionRepository
	items    handlers.ScheduleItemRepository
	state    handlers.SystemStateRepository
	notifier handlers.EventNotifier
	logger   *logger.Logger
	interval time.Duration

	mu         sync.Mutex
	timers     []*scheduledTimer
	reloadCh   chan struct{}
	doneClosed chan struct{}
	nowFunc    func() time.Time
}

// New creates a new Scheduler.
func New(
	sessions handlers.SessionRepository,
	items handlers.ScheduleItemRepository,
	state handlers.SystemStateRepository,
	notifier handlers.EventNotifier,
	log *logger.Logger,
	intervalSeconds int,
) *Scheduler {
	interval := time.Duration(intervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Scheduler{
		sessions:   sessions,
		items:      items,
		state:      state,
		notifier:   notifier,
		logger:     log,
		interval:   interval,
		reloadCh:   make(chan struct{}, 1),
		doneClosed: make(chan struct{}),
		nowFunc:    time.Now,
	}
}

// Run starts the scheduler loop. It blocks until done is closed.
func (s *Scheduler) Run(done <-chan struct{}) {
	defer close(s.doneClosed)

	s.loadAndSchedule()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			s.cancelAll()
			return
		case <-ticker.C:
			s.loadAndSchedule()
		case <-s.reloadCh:
			s.loadAndSchedule()
		}
	}
}

// Done returns a channel that is closed when the scheduler has stopped.
func (s *Scheduler) Done() <-chan struct{} {
	return s.doneClosed
}

// Reload triggers a non-blocking reload of the schedule.
func (s *Scheduler) Reload() {
	select {
	case s.reloadCh <- struct{}{}:
	default:
		// already pending
	}
}

// CancelNextBell cancels the earliest pending bell timer and returns its schedule item ID.
func (s *Scheduler) CancelNextBell() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.timers) == 0 {
		return "", fmt.Errorf("no pending bells to cancel")
	}

	// Find the earliest timer
	earliest := 0
	for i, t := range s.timers {
		if t.fireAt.Before(s.timers[earliest].fireAt) {
			earliest = i
		}
	}

	target := s.timers[earliest]
	target.timer.Stop()
	itemID := target.itemID

	// Remove from slice
	s.timers = append(s.timers[:earliest], s.timers[earliest+1:]...)

	if s.notifier != nil {
		s.notifier.NotifyBellCancelled(itemID)
		s.notifier.NotifySystemLog("info", fmt.Sprintf("bell cancelled for schedule item %s", itemID), "scheduler")
	}

	return itemID, nil
}

func (s *Scheduler) loadAndSchedule() {
	s.cancelAll()

	ctx := context.Background()

	stateVal, _, err := s.state.GetState(ctx)
	if err != nil {
		s.logger.Error("scheduler: failed to get system state", err)
		return
	}
	if stateVal == "paused" {
		s.logger.Info("scheduler: system is paused, skipping schedule load")
		return
	}

	session, err := s.sessions.GetCurrentSession(ctx)
	if err != nil {
		s.logger.Error("scheduler: failed to get current session", err)
		return
	}
	if session == nil {
		s.logger.Info("scheduler: no active session, skipping schedule load")
		return
	}

	scheduleItems, err := s.items.GetCurrentSessionSchedules(ctx, session.ID)
	if err != nil {
		s.logger.Error("scheduler: failed to get schedule items", err)
		return
	}

	now := s.nowFunc()
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range scheduleItems {
		fireAt := time.Date(now.Year(), now.Month(), now.Day(),
			item.Time.Hour(), item.Time.Minute(), item.Time.Second(), 0, now.Location())

		delay := fireAt.Sub(now)
		if delay <= 0 {
			continue // past, skip
		}

		st := &scheduledTimer{
			itemID:  item.ID.String(),
			soundID: item.SoundID.String(),
			name:    item.Name,
			fireAt:  fireAt,
		}
		st.timer = time.AfterFunc(delay, func() {
			s.fireBell(st)
		})
		s.timers = append(s.timers, st)
	}

	if len(s.timers) > 0 {
		s.logger.Info(fmt.Sprintf("scheduler: scheduled %d bell(s) for session %q", len(s.timers), session.Name))
	}
}

func (s *Scheduler) fireBell(st *scheduledTimer) {
	// Re-check state at fire time
	ctx := context.Background()
	stateVal, _, err := s.state.GetState(ctx)
	if err != nil {
		s.logger.Error("scheduler: failed to re-check state at fire time", err)
		return
	}
	if stateVal != "active" {
		s.logger.Info(fmt.Sprintf("scheduler: skipping bell %q — system is %s", st.name, stateVal))
		return
	}

	if s.notifier != nil {
		s.notifier.NotifyBellTriggered(models.BellTriggeredPayload{
			ScheduleItemID: st.itemID,
			SoundID:        st.soundID,
			Name:           st.name,
		})
		s.notifier.NotifySystemLog("info", fmt.Sprintf("bell triggered: %s", st.name), "scheduler")
	}

	// Remove self from timers
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.timers {
		if t == st {
			s.timers = append(s.timers[:i], s.timers[i+1:]...)
			break
		}
	}
}

func (s *Scheduler) cancelAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.timers {
		t.timer.Stop()
	}
	s.timers = nil
}
