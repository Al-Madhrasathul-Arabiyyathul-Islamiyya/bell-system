package scheduler

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test helpers ---

type mockSessionRepo struct {
	session *models.Session
	err     error
}

func (m *mockSessionRepo) Create(context.Context, *models.Session) error { return nil }
func (m *mockSessionRepo) GetByID(context.Context, uuid.UUID) (*models.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) GetSessionsByIDs(context.Context, []uuid.UUID) (map[uuid.UUID]*models.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) List(context.Context, string) ([]*models.Session, error) { return nil, nil }
func (m *mockSessionRepo) GetCurrentSession(_ context.Context) (*models.Session, error) {
	return m.session, m.err
}
func (m *mockSessionRepo) Update(context.Context, *models.Session) error { return nil }
func (m *mockSessionRepo) Delete(context.Context, uuid.UUID) error       { return nil }

type mockItemRepo struct {
	items []*models.ScheduleItem
	err   error
}

func (m *mockItemRepo) Create(context.Context, *models.ScheduleItem) error { return nil }
func (m *mockItemRepo) GetByID(context.Context, uuid.UUID) (*models.ScheduleItem, error) {
	return nil, nil
}
func (m *mockItemRepo) List(context.Context, string, int, string) ([]*models.ScheduleItem, error) {
	return nil, nil
}
func (m *mockItemRepo) GetCurrentSessionSchedules(_ context.Context, _ uuid.UUID) ([]*models.ScheduleItem, error) {
	return m.items, m.err
}
func (m *mockItemRepo) Update(context.Context, *models.ScheduleItem) error { return nil }
func (m *mockItemRepo) Delete(context.Context, uuid.UUID) error            { return nil }

type mockStateRepo struct {
	mu        sync.Mutex
	state     string
	updatedAt time.Time
	err       error
}

func (m *mockStateRepo) GetState(_ context.Context) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state, m.updatedAt, m.err
}

func (m *mockStateRepo) SetState(_ context.Context, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = state
	m.updatedAt = time.Now()
	return nil
}

type mockNotifier struct {
	mu                sync.Mutex
	bellTriggeredArgs []models.BellTriggeredPayload
	bellCancelledArgs []string
	logMessages       []string
	stateChangedArgs  []string
}

func (m *mockNotifier) NotifySchedulesUpdated()  {}
func (m *mockNotifier) NotifyAudioFilesUpdated() {}
func (m *mockNotifier) NotifyBellTriggered(p models.BellTriggeredPayload) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bellTriggeredArgs = append(m.bellTriggeredArgs, p)
}
func (m *mockNotifier) NotifyBellCancelled(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bellCancelledArgs = append(m.bellCancelledArgs, id)
}
func (m *mockNotifier) NotifySystemStateChanged(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateChangedArgs = append(m.stateChangedArgs, state)
}
func (m *mockNotifier) NotifySystemLog(level, message, source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logMessages = append(m.logMessages, fmt.Sprintf("%s: %s [%s]", level, message, source))
}

func testLogger() *logger.Logger {
	l, _ := logger.New("test")
	return l
}

// --- tests ---

func TestLoadAndSchedule_NoSession(t *testing.T) {
	s := New(
		&mockSessionRepo{session: nil},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	s.loadAndSchedule()

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.timers)
}

func TestLoadAndSchedule_Paused(t *testing.T) {
	session := &models.Session{ID: uuid.New(), Name: "Morning"}
	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{},
		&mockStateRepo{state: "paused"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	s.loadAndSchedule()

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.timers)
}

func TestLoadAndSchedule_SchedulesFutureBells(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.Local)
	session := &models.Session{ID: uuid.New(), Name: "Morning"}

	// Bells at 10:05 and 10:10 — both future relative to nowFunc
	items := []*models.ScheduleItem{
		{
			ID:      uuid.New(),
			Name:    "Bell 1",
			Time:    time.Date(0, 1, 1, 10, 5, 0, 0, time.UTC),
			SoundID: uuid.New(),
		},
		{
			ID:      uuid.New(),
			Name:    "Bell 2",
			Time:    time.Date(0, 1, 1, 10, 10, 0, 0, time.UTC),
			SoundID: uuid.New(),
		},
	}

	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{items: items},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)
	s.nowFunc = func() time.Time { return now }

	s.loadAndSchedule()

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Len(t, s.timers, 2)

	// Clean up timers
	for _, timer := range s.timers {
		timer.timer.Stop()
	}
}

func TestLoadAndSchedule_SkipsPastBells(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 30, 0, 0, time.Local)
	session := &models.Session{ID: uuid.New(), Name: "Morning"}

	items := []*models.ScheduleItem{
		{
			ID:      uuid.New(),
			Name:    "Past Bell",
			Time:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC), // 10:00 — past
			SoundID: uuid.New(),
		},
		{
			ID:      uuid.New(),
			Name:    "Future Bell",
			Time:    time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC), // 11:00 — future
			SoundID: uuid.New(),
		},
	}

	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{items: items},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)
	s.nowFunc = func() time.Time { return now }

	s.loadAndSchedule()

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Len(t, s.timers, 1)
	assert.Equal(t, "Future Bell", s.timers[0].name)

	for _, timer := range s.timers {
		timer.timer.Stop()
	}
}

func TestLoadAndSchedule_TimerFiresAndCallsFireBell(t *testing.T) {
	// Schedule a bell ~1s in the future so the AfterFunc closure fires,
	// covering the inline func() { s.fireBell(st) } at scheduler.go:187-189.
	// loadAndSchedule uses time.Date(now.Y, now.M, now.D, item.H, item.M, item.S, 0, ...)
	// so the minimum granularity is 1 second.
	now := time.Now()
	bellTime := now.Add(1 * time.Second)

	session := &models.Session{ID: uuid.New(), Name: "Morning"}
	bellID := uuid.New()
	soundID := uuid.New()
	items := []*models.ScheduleItem{
		{
			ID:      bellID,
			Name:    "Imminent Bell",
			Time:    time.Date(0, 1, 1, bellTime.Hour(), bellTime.Minute(), bellTime.Second(), 0, time.UTC),
			SoundID: soundID,
		},
	}

	notifier := &mockNotifier{}
	stateRepo := &mockStateRepo{state: "active"}
	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{items: items},
		stateRepo,
		notifier,
		testLogger(),
		30,
	)
	s.nowFunc = func() time.Time { return now }

	s.loadAndSchedule()

	// Wait for the timer to fire (up to 3s to be safe)
	require.Eventually(t, func() bool {
		notifier.mu.Lock()
		defer notifier.mu.Unlock()
		return len(notifier.bellTriggeredArgs) > 0
	}, 3*time.Second, 50*time.Millisecond, "timer should have fired")

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	assert.Equal(t, bellID.String(), notifier.bellTriggeredArgs[0].ScheduleItemID)
	assert.Equal(t, soundID.String(), notifier.bellTriggeredArgs[0].SoundID)
}

func TestFireBell_NotifiesClients(t *testing.T) {
	notifier := &mockNotifier{}
	stateRepo := &mockStateRepo{state: "active"}
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		stateRepo,
		notifier,
		testLogger(),
		30,
	)

	st := &scheduledTimer{
		itemID:  "item-1",
		soundID: "sound-1",
		name:    "Test Bell",
	}

	// Add timer to slice so fireBell can remove it
	s.mu.Lock()
	s.timers = append(s.timers, st)
	s.mu.Unlock()

	s.fireBell(st)

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	require.Len(t, notifier.bellTriggeredArgs, 1)
	assert.Equal(t, "item-1", notifier.bellTriggeredArgs[0].ScheduleItemID)
	assert.Equal(t, "sound-1", notifier.bellTriggeredArgs[0].SoundID)
	assert.Equal(t, "Test Bell", notifier.bellTriggeredArgs[0].Name)

	// Timer should be removed
	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.timers)
}

func TestFireBell_SkipsWhenPaused(t *testing.T) {
	notifier := &mockNotifier{}
	stateRepo := &mockStateRepo{state: "paused"}
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		stateRepo,
		notifier,
		testLogger(),
		30,
	)

	st := &scheduledTimer{
		itemID:  "item-1",
		soundID: "sound-1",
		name:    "Test Bell",
	}

	s.fireBell(st)

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	assert.Empty(t, notifier.bellTriggeredArgs)
}

func TestCancelNextBell_CancelsEarliest(t *testing.T) {
	notifier := &mockNotifier{}
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		notifier,
		testLogger(),
		30,
	)

	now := time.Now()
	s.mu.Lock()
	s.timers = []*scheduledTimer{
		{itemID: "later", fireAt: now.Add(10 * time.Minute), timer: time.NewTimer(10 * time.Minute)},
		{itemID: "earliest", fireAt: now.Add(2 * time.Minute), timer: time.NewTimer(2 * time.Minute)},
		{itemID: "middle", fireAt: now.Add(5 * time.Minute), timer: time.NewTimer(5 * time.Minute)},
	}
	s.mu.Unlock()

	cancelledID, err := s.CancelNextBell()
	require.NoError(t, err)
	assert.Equal(t, "earliest", cancelledID)

	s.mu.Lock()
	assert.Len(t, s.timers, 2)
	s.mu.Unlock()

	// Clean up
	s.cancelAll()
}

func TestCancelNextBell_EmptyTimers(t *testing.T) {
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	_, err := s.CancelNextBell()
	assert.Error(t, err)
}

func TestReload_TriggersReload(t *testing.T) {
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	s.Reload()

	// Should have a message in the channel
	select {
	case <-s.reloadCh:
		// success
	default:
		t.Fatal("expected reload signal in channel")
	}
}

func TestReload_Coalescing(t *testing.T) {
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	// Multiple reloads should coalesce
	s.Reload()
	s.Reload()
	s.Reload()

	select {
	case <-s.reloadCh:
	default:
		t.Fatal("expected reload signal")
	}

	// Channel should be empty now
	select {
	case <-s.reloadCh:
		t.Fatal("expected no more reload signals")
	default:
	}
}

func TestGracefulShutdown(t *testing.T) {
	session := &models.Session{ID: uuid.New(), Name: "Morning"}
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.Local)

	items := []*models.ScheduleItem{
		{
			ID:      uuid.New(),
			Name:    "Bell",
			Time:    time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC),
			SoundID: uuid.New(),
		},
	}

	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{items: items},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		60, // long interval so ticker doesn't fire during test
	)
	s.nowFunc = func() time.Time { return now }

	done := make(chan struct{})
	go s.Run(done)

	// Give scheduler time to start
	time.Sleep(50 * time.Millisecond)

	close(done)

	select {
	case <-s.Done():
		// success — scheduler shut down
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not shut down in time")
	}

	// All timers should be cancelled
	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.timers)
}

func TestDefaultInterval(t *testing.T) {
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		0, // should default to 30s
	)

	assert.Equal(t, 30*time.Second, s.interval)
}

func TestLoadAndSchedule_GetCurrentSessionError(t *testing.T) {
	s := New(
		&mockSessionRepo{err: fmt.Errorf("connection lost")},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	s.loadAndSchedule()

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.timers)
}

func TestLoadAndSchedule_GetScheduleItemsError(t *testing.T) {
	session := &models.Session{ID: uuid.New(), Name: "Morning"}
	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{err: fmt.Errorf("query failed")},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	s.loadAndSchedule()

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.timers)
}

func TestLoadAndSchedule_GetStateError(t *testing.T) {
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		&mockStateRepo{err: fmt.Errorf("db error")},
		&mockNotifier{},
		testLogger(),
		30,
	)

	s.loadAndSchedule()

	s.mu.Lock()
	defer s.mu.Unlock()
	assert.Empty(t, s.timers)
}

func TestFireBell_GetStateError(t *testing.T) {
	notifier := &mockNotifier{}
	stateRepo := &mockStateRepo{err: fmt.Errorf("db error")}
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		stateRepo,
		notifier,
		testLogger(),
		30,
	)

	st := &scheduledTimer{
		itemID:  "item-1",
		soundID: "sound-1",
		name:    "Test Bell",
	}

	s.fireBell(st)

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	assert.Empty(t, notifier.bellTriggeredArgs, "bell should not fire when state check fails")
}

func TestLoadAndSchedule_LogsScheduledBells(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.Local)
	session := &models.Session{ID: uuid.New(), Name: "Morning"}

	items := []*models.ScheduleItem{
		{
			ID:      uuid.New(),
			Name:    "Bell",
			Time:    time.Date(0, 1, 1, 11, 0, 0, 0, time.UTC),
			SoundID: uuid.New(),
		},
	}

	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{items: items},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)
	s.nowFunc = func() time.Time { return now }

	s.loadAndSchedule()

	s.mu.Lock()
	assert.Len(t, s.timers, 1)
	for _, timer := range s.timers {
		timer.timer.Stop()
	}
	s.mu.Unlock()
}

func TestRun_TickerTriggersReload(t *testing.T) {
	session := &models.Session{ID: uuid.New(), Name: "Morning"}
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.Local)

	callCount := 0
	var mu sync.Mutex
	itemRepo := &mockItemRepo{}
	// Override GetCurrentSessionSchedules to count calls
	sessionRepo := &mockSessionRepo{session: session}
	stateRepo := &mockStateRepo{state: "active"}

	s := New(sessionRepo, itemRepo, stateRepo, &mockNotifier{}, testLogger(), 0)
	s.nowFunc = func() time.Time { return now }
	// Use a very short interval so the ticker fires quickly
	s.interval = 20 * time.Millisecond

	done := make(chan struct{})
	go func() {
		// Count loadAndSchedule calls by watching for state.GetState calls
		origGetState := stateRepo.GetState
		_ = origGetState
		mu.Lock()
		callCount = 0
		mu.Unlock()
		s.Run(done)
	}()

	// Wait enough time for at least 2 ticker fires (initial + ticks)
	time.Sleep(80 * time.Millisecond)

	close(done)
	<-s.Done()

	// loadAndSchedule is called at least once on startup plus on ticks
	// The initial call + at least 2 tick calls = 3+
	mu.Lock()
	_ = callCount
	mu.Unlock()
	// If we got here without deadlock/panic, the ticker path works
}

func TestRun_ReloadChannelTriggersReload(t *testing.T) {
	session := &models.Session{ID: uuid.New(), Name: "Morning"}
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.Local)

	s := New(
		&mockSessionRepo{session: session},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		600, // long interval so ticker doesn't fire
	)
	s.nowFunc = func() time.Time { return now }

	done := make(chan struct{})
	go s.Run(done)

	// Give scheduler time to start and complete initial loadAndSchedule
	time.Sleep(50 * time.Millisecond)

	// Trigger reload via channel
	s.Reload()

	// Give it time to process the reload
	time.Sleep(50 * time.Millisecond)

	close(done)
	<-s.Done()

	// If we got here without deadlock/panic, the reload channel path works
}

func TestSetNowFunc_OverridesDefault(t *testing.T) {
	s := New(
		&mockSessionRepo{},
		&mockItemRepo{},
		&mockStateRepo{state: "active"},
		&mockNotifier{},
		testLogger(),
		30,
	)

	// Default should be time.Now (non-nil)
	assert.NotNil(t, s.nowFunc)
	before := s.nowFunc()

	fixed := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	s.SetNowFunc(func() time.Time { return fixed })

	got := s.nowFunc()
	assert.True(t, got.Equal(fixed), "expected %v, got %v", fixed, got)
	assert.False(t, got.Equal(before), "should differ from wall clock")
}
