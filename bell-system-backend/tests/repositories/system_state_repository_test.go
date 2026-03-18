package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemStateRepository_GetState_Success(t *testing.T) {
	repo, mock := mocks.NewMockSystemStateRepo(t)
	ctx := context.Background()

	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"Value", "UpdatedAt"}).
		AddRow("active", now)

	mock.ExpectQuery("SELECT .+ FROM SystemState WHERE").
		WillReturnRows(rows)

	state, updatedAt, err := repo.GetState(ctx)
	require.NoError(t, err)
	assert.Equal(t, "active", state)
	assert.Equal(t, now, updatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSystemStateRepository_GetState_NoRows(t *testing.T) {
	repo, mock := mocks.NewMockSystemStateRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM SystemState WHERE").
		WillReturnError(sql.ErrNoRows)

	state, updatedAt, err := repo.GetState(ctx)
	require.Error(t, err)
	assert.Empty(t, state)
	assert.True(t, updatedAt.IsZero())
	assert.Contains(t, err.Error(), "failed to get system state")
}

func TestSystemStateRepository_GetState_DBError(t *testing.T) {
	repo, mock := mocks.NewMockSystemStateRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM SystemState WHERE").
		WillReturnError(errors.New("connection lost"))

	state, updatedAt, err := repo.GetState(ctx)
	require.Error(t, err)
	assert.Empty(t, state)
	assert.True(t, updatedAt.IsZero())
	assert.Contains(t, err.Error(), "failed to get system state")
}

func TestSystemStateRepository_GetState_ScanError(t *testing.T) {
	repo, mock := mocks.NewMockSystemStateRepo(t)
	ctx := context.Background()

	// Return wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"Value"}).
		AddRow("active")

	mock.ExpectQuery("SELECT .+ FROM SystemState WHERE").
		WillReturnRows(rows)

	_, _, err := repo.GetState(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get system state")
}

func TestSystemStateRepository_SetState_Success(t *testing.T) {
	repo, mock := mocks.NewMockSystemStateRepo(t)
	ctx := context.Background()

	mock.ExpectExec("UPDATE SystemState SET").
		WithArgs("paused").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetState(ctx, "paused")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSystemStateRepository_SetState_DBError(t *testing.T) {
	repo, mock := mocks.NewMockSystemStateRepo(t)
	ctx := context.Background()

	mock.ExpectExec("UPDATE SystemState SET").
		WithArgs("active").
		WillReturnError(errors.New("constraint violation"))

	err := repo.SetState(ctx, "active")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set system state")
}
