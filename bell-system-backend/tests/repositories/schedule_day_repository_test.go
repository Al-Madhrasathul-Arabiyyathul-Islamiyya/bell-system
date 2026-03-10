package repositories_test

import (
	"context"
	"errors"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleDayRepository_Create_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	sd := &models.ScheduleDay{ScheduleItemID: uuid.New(), DayOfWeek: 2}

	mock.ExpectExec("INSERT INTO ScheduleDays").
		WithArgs(sd.ScheduleItemID, sd.DayOfWeek).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, sd)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleDayRepository_Create_DBError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	sd := &models.ScheduleDay{ScheduleItemID: uuid.New(), DayOfWeek: 1}

	mock.ExpectExec("INSERT INTO ScheduleDays").
		WithArgs(sd.ScheduleItemID, sd.DayOfWeek).
		WillReturnError(errors.New("foreign key violation"))

	err := repo.Create(ctx, sd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create schedule day")
}

func TestScheduleDayRepository_GetDaysForScheduleItems_Empty(t *testing.T) {
	repo, _ := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	result, err := repo.GetDaysForScheduleItems(ctx, []uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestScheduleDayRepository_GetDaysForScheduleItems_Multiple(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	itemID1 := uuid.New()
	itemID2 := uuid.New()

	rows := sqlmock.NewRows([]string{"ScheduleItemId", "DayOfWeek"}).
		AddRow(itemID1, 2).
		AddRow(itemID1, 3).
		AddRow(itemID1, 4).
		AddRow(itemID2, 5).
		AddRow(itemID2, 6)

	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnRows(rows)

	result, err := repo.GetDaysForScheduleItems(ctx, []uuid.UUID{itemID1, itemID2})
	require.NoError(t, err)
	assert.Len(t, result[itemID1], 3)
	assert.Len(t, result[itemID2], 2)
	assert.Contains(t, result[itemID1], 2)
	assert.Contains(t, result[itemID1], 3)
	assert.Contains(t, result[itemID1], 4)
	assert.Contains(t, result[itemID2], 5)
	assert.Contains(t, result[itemID2], 6)
}

func TestScheduleDayRepository_GetDaysForScheduleItems_DBError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM ScheduleDays WHERE ScheduleItemId IN").
		WillReturnError(errors.New("query failed"))

	result, err := repo.GetDaysForScheduleItems(ctx, []uuid.UUID{uuid.New()})
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestScheduleDayRepository_Update_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	sd := &models.ScheduleDay{ScheduleItemID: uuid.New(), DayOfWeek: 5}

	mock.ExpectExec("UPDATE ScheduleDays").
		WithArgs(sd.DayOfWeek, sd.ScheduleItemID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, sd)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleDayRepository_Delete_Success(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	id := uuid.New()
	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.Delete(ctx, id)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleDayRepository_Delete_DBError(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM ScheduleDays WHERE").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("delete failed"))

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete schedule day")
}

// --- Spec-driven: Days 1-7 represent Sunday through Saturday ---

func TestScheduleDayRepository_AllSevenDays(t *testing.T) {
	repo, mock := mocks.NewMockScheduleDayRepo(t)
	ctx := context.Background()

	itemID := uuid.New()
	for day := 1; day <= 7; day++ {
		sd := &models.ScheduleDay{ScheduleItemID: itemID, DayOfWeek: day}
		mock.ExpectExec("INSERT INTO ScheduleDays").
			WithArgs(sd.ScheduleItemID, sd.DayOfWeek).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Create(ctx, sd)
		require.NoError(t, err, "should insert day %d", day)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
