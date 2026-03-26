package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ScheduleItemRepository CRUD tests ---

func TestScheduleItemRepository_Create_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	sessionID := uuid.New()
	item := &models.ScheduleItem{
		SessionID: &sessionID,
		Name:      "First Period Bell",
		Time:      time.Date(0, 1, 1, 7, 45, 0, 0, time.UTC),
		SoundID:   uuid.New(),
		Days:      []int{2, 3, 4, 5, 6},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ScheduleItems").
		WithArgs(
			sqlmock.AnyArg(),
			sessionID,
			"First Period Bell",
			"07:45",
			item.SoundID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	for range item.Days {
		mock.ExpectExec("INSERT INTO ScheduleDays").
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	err := repo.Create(ctx, item)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, item.ID, "ID should be auto-generated if nil")
	assert.False(t, item.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, item.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_Create_WithPresetID(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	presetID := uuid.New()
	item := &models.ScheduleItem{
		ID:      presetID,
		Name:    "Custom ID Bell",
		Time:    time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		SoundID: uuid.New(),
		Days:    []int{2},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ScheduleItems").
		WithArgs(
			presetID,
			nil,
			"Custom ID Bell",
			"08:00",
			item.SoundID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Create(ctx, item)
	require.NoError(t, err)
	assert.Equal(t, presetID, item.ID, "preset ID should be preserved")
}

func TestScheduleItemRepository_Create_NoDays(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{
		Name:    "No Days Bell",
		Time:    time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		SoundID: uuid.New(),
		Days:    []int{},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ScheduleItems").
		WithArgs(
			sqlmock.AnyArg(),
			nil,
			"No Days Bell",
			"09:00",
			item.SoundID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Create(ctx, item)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_Create_TransactionRollbackOnItemInsertFailure(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{Name: "Fail Bell", SoundID: uuid.New(), Days: []int{2}}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ScheduleItems").
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err := repo.Create(ctx, item)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert schedule item")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_Create_TransactionRollbackOnDayInsertFailure(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{Name: "Day Fail Bell", SoundID: uuid.New(), Days: []int{2, 3}}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnError(errors.New("day insert failed"))
	mock.ExpectRollback()

	err := repo.Create(ctx, item)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert schedule day")
}

func TestScheduleItemRepository_GetByID_WithSoundAndSession(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	sessionID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	// Main item query
	itemRow := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, sessionID, "Period Bell", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE").
		WillReturnRows(itemRow)

	// Days sub-query
	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2).
		AddRow(itemID, 4)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	// Sounds sub-query
	soundRows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(soundID, "bell.wav", "/audio/bell.wav", "bell", "hash1", now, now)
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(soundRows)

	// Sessions sub-query
	sessionRows := sqlmock.NewRows([]string{"Id", "Name", "StartTime", "EndTime"}).
		AddRow(sessionID, "Morning", now, now)
	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id IN").
		WillReturnRows(sessionRows)

	item, err := repo.GetByID(ctx, itemID)
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, []int{2, 4}, item.Days)
	require.NotNil(t, item.Sound)
	assert.Equal(t, "bell.wav", item.Sound.Name)
	require.NotNil(t, item.Session)
	assert.Equal(t, "Morning", item.Session.Name)
	assert.NotNil(t, item.DayInfo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_GetByID_NoSoundNoSession(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	now := time.Now()

	// Main item with nil SessionID and zero SoundID
	itemRow := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, nil, "General Bell", now, uuid.Nil, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE").
		WillReturnRows(itemRow)

	// Days sub-query
	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 1)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	// No sound query expected (SoundID is Nil)
	// No session query expected (SessionID is nil)

	item, err := repo.GetByID(ctx, itemID)
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Nil(t, item.Sound)
	assert.Nil(t, item.Session)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_GetByID_DaysQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	now := time.Now()

	itemRow := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, nil, "Bell", now, uuid.New(), now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE").
		WillReturnRows(itemRow)

	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnError(errors.New("days query failed"))

	item, err := repo.GetByID(ctx, itemID)
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to get schedule days")
}

func TestScheduleItemRepository_GetByID_SoundsQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	itemRow := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, nil, "Bell", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE").
		WillReturnRows(itemRow)

	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnError(errors.New("sounds query failed"))

	item, err := repo.GetByID(ctx, itemID)
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to get sound file")
}

func TestScheduleItemRepository_GetByID_SessionsQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	sessionID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	itemRow := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, sessionID, "Bell", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE").
		WillReturnRows(itemRow)

	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	soundRows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(soundID, "bell.wav", "/audio/bell.wav", "bell", "hash1", now, now)
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(soundRows)

	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id IN").
		WillReturnError(errors.New("sessions query failed"))

	item, err := repo.GetByID(ctx, itemID)
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to get session")
}

func TestScheduleItemRepository_GetByID_NotFound(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE").
		WillReturnError(sql.ErrNoRows)

	item, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, item)
}

func TestScheduleItemRepository_GetByID_DBError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE").
		WillReturnError(errors.New("connection lost"))

	item, err := repo.GetByID(ctx, uuid.New())
	require.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to get schedule item")
}

func TestScheduleItemRepository_Delete_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("DELETE FROM ScheduleItems WHERE").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, id)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_Delete_DBError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ScheduleItems WHERE").
		WillReturnError(errors.New("constraint"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete the schedule item")
}

func TestScheduleItemRepository_Delete_DaysDeleteError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").
		WillReturnError(errors.New("fk error"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete schedule days")
}

func TestScheduleItemRepository_Update_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	sessionID := uuid.New()
	item := &models.ScheduleItem{
		ID:        uuid.New(),
		SessionID: &sessionID,
		Name:      "Updated Bell",
		Time:      time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
		SoundID:   uuid.New(),
		Days:      []int{2, 4, 6},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE ScheduleItems").
		WithArgs(
			sessionID,
			"Updated Bell",
			"10:00",
			item.SoundID,
			sqlmock.AnyArg(),
			item.ID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").WillReturnResult(sqlmock.NewResult(0, 5))
	for range item.Days {
		mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	err := repo.Update(ctx, item)
	require.NoError(t, err)
	assert.False(t, item.UpdatedAt.IsZero(), "UpdatedAt should be refreshed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_Update_RollbackOnUpdateFail(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{ID: uuid.New(), Name: "Fail Update", SoundID: uuid.New(), Days: []int{2}}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE ScheduleItems").WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	err := repo.Update(ctx, item)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update schedule item")
}

func TestScheduleItemRepository_Update_RollbackOnDeleteDaysFail(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{ID: uuid.New(), Name: "Day Delete Fail", SoundID: uuid.New(), Days: []int{2}}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").WillReturnError(errors.New("delete days failed"))
	mock.ExpectRollback()

	err := repo.Update(ctx, item)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete existing schedule days")
}

func TestScheduleItemRepository_Update_NoDays(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{ID: uuid.New(), Name: "No Days Update", SoundID: uuid.New(), Days: []int{}}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.Update(ctx, item)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Spec-driven: Schedule items can be session-specific or general ---

func TestScheduleItemRepository_Create_WithoutSession(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{
		SessionID: nil,
		Name:      "School Opening",
		Time:      time.Date(0, 1, 1, 6, 45, 0, 0, time.UTC),
		SoundID:   uuid.New(),
		Days:      []int{2, 3, 4, 5, 6},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ScheduleItems").
		WithArgs(
			sqlmock.AnyArg(),
			nil,
			"School Opening",
			"06:45",
			item.SoundID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	for range item.Days {
		mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	err := repo.Create(ctx, item)
	require.NoError(t, err)
}

func TestScheduleItemRepository_Create_WithSession(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	sessionID := uuid.New()
	item := &models.ScheduleItem{
		SessionID: &sessionID,
		Name:      "Morning First Period",
		Time:      time.Date(0, 1, 1, 7, 45, 0, 0, time.UTC),
		SoundID:   uuid.New(),
		Days:      []int{2, 3, 4, 5, 6},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ScheduleItems").
		WithArgs(
			sqlmock.AnyArg(),
			sessionID,
			"Morning First Period",
			"07:45",
			item.SoundID,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	for range item.Days {
		mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	err := repo.Create(ctx, item)
	require.NoError(t, err)
	require.NotNil(t, item.SessionID)
	assert.Equal(t, sessionID, *item.SessionID)
}

// --- ScheduleItemRepository.List tests ---

func TestScheduleItemRepository_List_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID1 := uuid.New()
	itemID2 := uuid.New()
	sessionID := uuid.New()
	soundID1 := uuid.New()
	soundID2 := uuid.New()
	now := time.Now()

	// Main query returns 2 items
	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID1, sessionID, "Bell 1", now, soundID1, now, now).
		AddRow(itemID2, nil, "Bell 2", now, soundID2, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnRows(itemRows)

	// Days sub-query
	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID1, 2).
		AddRow(itemID1, 3).
		AddRow(itemID2, 4)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	// Sounds sub-query
	soundRows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(soundID1, "bell.wav", "/audio/bell.wav", "bell", "hash1", now, now).
		AddRow(soundID2, "anthem.mp3", "/audio/anthem.mp3", "anthem", "hash2", now, now)
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(soundRows)

	// Sessions sub-query
	sessionRows := sqlmock.NewRows([]string{"Id", "Name", "StartTime", "EndTime"}).
		AddRow(sessionID, "Morning", now, now)
	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id IN").
		WillReturnRows(sessionRows)

	items, err := repo.List(ctx, "", 0, nil)
	require.NoError(t, err)
	require.Len(t, items, 2)

	assert.Equal(t, itemID1, items[0].ID)
	assert.Equal(t, []int{2, 3}, items[0].Days)
	assert.NotNil(t, items[0].Sound)
	assert.NotNil(t, items[0].Session)
	assert.Equal(t, "Morning", items[0].Session.Name)

	assert.Equal(t, itemID2, items[1].ID)
	assert.Equal(t, []int{4}, items[1].Days)
	assert.NotNil(t, items[1].Sound)
	assert.Nil(t, items[1].Session) // no session ID

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_List_FilterByDay(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, nil, "Bell 1", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems WHERE .+ IN .+ DayOfWeek .+ ORDER BY Time").
		WithArgs(3).
		WillReturnRows(itemRows)

	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 3)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	soundRows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(soundID, "bell.wav", "/audio/bell.wav", "bell", "hash1", now, now)
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(soundRows)

	items, err := repo.List(ctx, "", 3, nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, itemID, items[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_List_Empty(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"})
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnRows(rows)

	items, err := repo.List(ctx, "", 0, nil)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestScheduleItemRepository_List_QueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnError(errors.New("connection lost"))

	items, err := repo.List(ctx, "", 0, nil)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to list schedule items")
}

func TestScheduleItemRepository_List_ScanError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	// Return rows with wrong column count to trigger scan error
	rows := sqlmock.NewRows([]string{"Id", "SessionId", "Name"}).
		AddRow(uuid.New(), nil, "Bad Row")
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnRows(rows)

	items, err := repo.List(ctx, "", 0, nil)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to scan schedule item")
}

func TestScheduleItemRepository_List_DaysQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	now := time.Now()

	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, nil, "Bell", now, uuid.New(), now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnRows(itemRows)

	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnError(errors.New("days query failed"))

	items, err := repo.List(ctx, "", 0, nil)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestScheduleItemRepository_List_SoundsQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, nil, "Bell", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnRows(itemRows)

	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnError(errors.New("sounds query failed"))

	items, err := repo.List(ctx, "", 0, nil)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestScheduleItemRepository_List_SessionsQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	sessionID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, sessionID, "Bell", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnRows(itemRows)

	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	soundRows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(soundID, "bell.wav", "/audio/bell.wav", "bell", "hash1", now, now)
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(soundRows)

	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id IN").
		WillReturnError(errors.New("sessions query failed"))

	items, err := repo.List(ctx, "", 0, nil)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestScheduleItemRepository_List_RowsError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(uuid.New(), nil, "Bell", time.Now(), uuid.New(), time.Now(), time.Now()).
		RowError(0, errors.New("row iteration error"))
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems ORDER BY Time").
		WillReturnRows(rows)

	items, err := repo.List(ctx, "", 0, nil)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "error iterating schedule item rows")
}

// --- ScheduleItemRepository.GetCurrentSessionSchedules tests ---

func TestScheduleItemRepository_GetCurrentSessionSchedules_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	sessionID := uuid.New()
	itemID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	// Main query
	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, sessionID, "Period 1", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnRows(itemRows)

	// Days
	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2).
		AddRow(itemID, 3)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	// Sounds
	soundRows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(soundID, "bell.wav", "/audio/bell.wav", "bell", "hash1", now, now)
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(soundRows)

	// Sessions
	sessionRows := sqlmock.NewRows([]string{"Id", "Name", "StartTime", "EndTime"}).
		AddRow(sessionID, "Morning", now, now)
	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id IN").
		WillReturnRows(sessionRows)

	items, err := repo.GetCurrentSessionSchedules(ctx, sessionID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Period 1", items[0].Name)
	assert.Equal(t, []int{2, 3}, items[0].Days)
	assert.NotNil(t, items[0].Sound)
	assert.NotNil(t, items[0].Session)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_GetCurrentSessionSchedules_Empty(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"})
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnRows(rows)

	items, err := repo.GetCurrentSessionSchedules(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestScheduleItemRepository_GetCurrentSessionSchedules_QueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnError(errors.New("query failed"))

	items, err := repo.GetCurrentSessionSchedules(ctx, uuid.New())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to get current session schedules")
}

func TestScheduleItemRepository_GetCurrentSessionSchedules_ScanError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	// Return rows with wrong column count to trigger scan error
	rows := sqlmock.NewRows([]string{"Id", "SessionId", "Name"}).
		AddRow(uuid.New(), nil, "Bad Row")
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnRows(rows)

	items, err := repo.GetCurrentSessionSchedules(ctx, uuid.New())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to scan schedule item")
}

func TestScheduleItemRepository_GetCurrentSessionSchedules_RowsError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(uuid.New(), nil, "Bell", time.Now(), uuid.New(), time.Now(), time.Now()).
		RowError(0, errors.New("row iteration error"))
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnRows(rows)

	items, err := repo.GetCurrentSessionSchedules(ctx, uuid.New())
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "error iterating schedule item rows")
}

func TestScheduleItemRepository_GetCurrentSessionSchedules_DaysQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	sessionID := uuid.New()
	itemID := uuid.New()
	now := time.Now()

	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, sessionID, "Bell", now, uuid.New(), now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnRows(itemRows)

	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnError(errors.New("days query failed"))

	items, err := repo.GetCurrentSessionSchedules(ctx, sessionID)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestScheduleItemRepository_GetCurrentSessionSchedules_SoundsQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	sessionID := uuid.New()
	itemID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, sessionID, "Bell", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnRows(itemRows)

	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnError(errors.New("sounds query failed"))

	items, err := repo.GetCurrentSessionSchedules(ctx, sessionID)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestScheduleItemRepository_GetCurrentSessionSchedules_SessionsQueryError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	sessionID := uuid.New()
	itemID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	itemRows := sqlmock.NewRows([]string{"Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt"}).
		AddRow(itemID, sessionID, "Bell", now, soundID, now, now)
	mock.ExpectQuery("SELECT .+ FROM ScheduleItems si").
		WillReturnRows(itemRows)

	dayRows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID, 2)
	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(dayRows)

	soundRows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(soundID, "bell.wav", "/audio/bell.wav", "bell", "hash1", now, now)
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(soundRows)

	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id IN").
		WillReturnError(errors.New("sessions query failed"))

	items, err := repo.GetCurrentSessionSchedules(ctx, sessionID)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestScheduleItemRepository_Update_InsertDayError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	item := &models.ScheduleItem{ID: uuid.New(), Name: "Day Insert Fail", SoundID: uuid.New(), Days: []int{2, 3}}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnError(errors.New("day insert failed"))
	mock.ExpectRollback()

	err := repo.Update(ctx, item)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert schedule day")
}
