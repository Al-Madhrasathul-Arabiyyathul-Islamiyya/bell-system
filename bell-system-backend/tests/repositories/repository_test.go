package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	repo, _ := mocks.NewMockDB(t)
	assert.NotNil(t, repo)
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	repo, mock := mocks.NewMockDB(t)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := repo.WithTx(context.Background(), func(tx *sql.Tx) error {
		return nil
	})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_RollbackOnError(t *testing.T) {
	repo, mock := mocks.NewMockDB(t)

	mock.ExpectBegin()
	mock.ExpectRollback()

	fnErr := errors.New("something went wrong")
	err := repo.WithTx(context.Background(), func(tx *sql.Tx) error {
		return fnErr
	})

	require.Error(t, err)
	assert.Equal(t, fnErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_BeginError(t *testing.T) {
	repo, mock := mocks.NewMockDB(t)

	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

	err := repo.WithTx(context.Background(), func(tx *sql.Tx) error {
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_RollbackOnPanic(t *testing.T) {
	repo, mock := mocks.NewMockDB(t)

	mock.ExpectBegin()
	mock.ExpectRollback()

	assert.Panics(t, func() {
		repo.WithTx(context.Background(), func(tx *sql.Tx) error {
			panic("unexpected error")
		})
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_CommitError(t *testing.T) {
	repo, mock := mocks.NewMockDB(t)

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := repo.WithTx(context.Background(), func(tx *sql.Tx) error {
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}
