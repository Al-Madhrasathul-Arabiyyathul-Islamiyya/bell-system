package unit_test

import (
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Close(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectClose()

	db := &database.DB{DB: sqlDB}
	err = db.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
