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

func TestSessionRepository_Create_Success(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	session := &models.Session{
		ID:        uuid.New(),
		Name:      "Morning Session",
		StartTime: time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC),
		EndTime:   time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec("INSERT INTO Sessions").
		WithArgs(session.ID, session.Name, session.StartTime, session.EndTime).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, session)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Create_DBError(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	session := &models.Session{ID: uuid.New(), Name: "Duplicate"}

	mock.ExpectExec("INSERT INTO Sessions").
		WithArgs(session.ID, session.Name, session.StartTime, session.EndTime).
		WillReturnError(errors.New("duplicate name"))

	err := repo.Create(ctx, session)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create session")
}

func TestSessionRepository_GetByID_Found(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	id := uuid.New()
	start := time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC)
	end := time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"Id", "Name", "StartTime", "EndTime"}).
		AddRow(id, "Morning Session", start, end)

	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id").
		WithArgs(id).
		WillReturnRows(rows)

	session, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, id, session.ID)
	assert.Equal(t, "Morning Session", session.Name)
}

func TestSessionRepository_GetByID_NotFound(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	session, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, session)
}

func TestSessionRepository_GetSessionsByIDs_Empty(t *testing.T) {
	repo, _ := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	result, err := repo.GetSessionsByIDs(ctx, []uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSessionRepository_GetSessionsByIDs_Multiple(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()

	rows := sqlmock.NewRows([]string{"Id", "Name", "StartTime", "EndTime"}).
		AddRow(id1, "Morning Session", time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC), time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)).
		AddRow(id2, "Afternoon Session", time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC))

	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE Id IN").
		WillReturnRows(rows)

	result, err := repo.GetSessionsByIDs(ctx, []uuid.UUID{id1, id2})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Morning Session", result[id1].Name)
	assert.Equal(t, "Afternoon Session", result[id2].Name)
}

func TestSessionRepository_List_OrderedByStartTime(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "Name", "StartTime", "EndTime"}).
		AddRow(uuid.New(), "Morning Session", time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC), time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)).
		AddRow(uuid.New(), "Afternoon Session", time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC))

	mock.ExpectQuery("SELECT .+ FROM Sessions ORDER BY StartTime").
		WillReturnRows(rows)

	sessions, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, sessions, 2)
	assert.Equal(t, "Morning Session", sessions[0].Name)
	assert.Equal(t, "Afternoon Session", sessions[1].Name)
}

func TestSessionRepository_List_Empty(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "Name", "StartTime", "EndTime"})
	mock.ExpectQuery("SELECT .+ FROM Sessions ORDER BY StartTime").
		WillReturnRows(rows)

	sessions, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionRepository_GetCurrentSession_NoActiveSession(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM Sessions WHERE").
		WillReturnError(sql.ErrNoRows)

	session, err := repo.GetCurrentSession(ctx)
	require.NoError(t, err)
	assert.Nil(t, session)
}

func TestSessionRepository_Update_Success(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	session := &models.Session{
		ID:        uuid.New(),
		Name:      "Updated Session",
		StartTime: time.Date(0, 1, 1, 6, 45, 0, 0, time.UTC),
		EndTime:   time.Date(0, 1, 1, 12, 10, 0, 0, time.UTC),
	}

	mock.ExpectExec("UPDATE Sessions").
		WithArgs(session.Name, session.StartTime, session.EndTime, session.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, session)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Delete_Success(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	id := uuid.New()
	mock.ExpectExec("DELETE FROM Sessions WHERE").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, id)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Delete_DBError(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Sessions WHERE").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("constraint violation"))

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete session")
}

// --- Spec-driven: Two sessions (Morning 06:45-12:10, Afternoon 12:15-18:10) ---

func TestSessionRepository_SpecSessions(t *testing.T) {
	repo, mock := mocks.NewMockSessionRepo(t)
	ctx := context.Background()

	morning := &models.Session{
		ID:        uuid.New(),
		Name:      "Morning Session",
		StartTime: time.Date(0, 1, 1, 6, 45, 0, 0, time.UTC),
		EndTime:   time.Date(0, 1, 1, 12, 10, 0, 0, time.UTC),
	}
	afternoon := &models.Session{
		ID:        uuid.New(),
		Name:      "Afternoon Session",
		StartTime: time.Date(0, 1, 1, 12, 15, 0, 0, time.UTC),
		EndTime:   time.Date(0, 1, 1, 18, 10, 0, 0, time.UTC),
	}

	mock.ExpectExec("INSERT INTO Sessions").
		WithArgs(morning.ID, morning.Name, morning.StartTime, morning.EndTime).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO Sessions").
		WithArgs(afternoon.ID, afternoon.Name, afternoon.StartTime, afternoon.EndTime).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Create(ctx, morning))
	require.NoError(t, repo.Create(ctx, afternoon))
	assert.NoError(t, mock.ExpectationsWereMet())
}
