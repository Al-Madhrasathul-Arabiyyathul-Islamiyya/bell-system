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
	mock.ExpectExec("INSERT INTO ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
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
	mock.ExpectExec("INSERT INTO ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
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
	mock.ExpectExec("DELETE FROM ScheduleItems WHERE").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, id)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleItemRepository_Delete_DBError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleItemRepo(t)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM ScheduleItems WHERE").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("constraint"))

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete the schedule item")
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
	mock.ExpectExec("UPDATE ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
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
	mock.ExpectExec("INSERT INTO ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
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
	mock.ExpectExec("INSERT INTO ScheduleItems").WillReturnResult(sqlmock.NewResult(0, 1))
	for range item.Days {
		mock.ExpectExec("INSERT INTO ScheduleDays").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	err := repo.Create(ctx, item)
	require.NoError(t, err)
	require.NotNil(t, item.SessionID)
	assert.Equal(t, sessionID, *item.SessionID)
}
