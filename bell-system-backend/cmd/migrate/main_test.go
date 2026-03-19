package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- getMigrationFiles tests ---

func TestGetMigrationFiles_UpFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"002_users_up.sql",
		"001_init_up.sql",
		"001_init_down.sql",
		"002_users_down.sql",
	}
	for _, f := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, f), []byte("SELECT 1"), 0644))
	}

	result, err := getMigrationFiles(dir, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"001_init_up.sql", "002_users_up.sql"}, result)
}

func TestGetMigrationFiles_DownFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"002_users_up.sql",
		"001_init_up.sql",
		"001_init_down.sql",
		"002_users_down.sql",
	}
	for _, f := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, f), []byte("SELECT 1"), 0644))
	}

	result, err := getMigrationFiles(dir, true)
	require.NoError(t, err)
	// Down files are reverse-sorted
	assert.Equal(t, []string{"002_users_down.sql", "001_init_down.sql"}, result)
}

func TestGetMigrationFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	result, err := getMigrationFiles(dir, false)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetMigrationFiles_NonExistentDir(t *testing.T) {
	result, err := getMigrationFiles("/nonexistent/path/that/does/not/exist", false)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestGetMigrationFiles_SkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_up.sql"), []byte("SELECT 1"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "subdir_up.sql"), 0755))

	result, err := getMigrationFiles(dir, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"001_init_up.sql"}, result)
}

func TestGetMigrationFiles_SkipsNonSQLFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_up.sql"), []byte("SELECT 1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("notes"), 0644))

	result, err := getMigrationFiles(dir, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"001_init_up.sql"}, result)
}

// --- runUpMigrations tests ---

func TestRunUpMigrations_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_up.sql"), []byte("CREATE TABLE test (id INT)"), 0644))

	// Check if already applied → count = 0
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO migrations").
		WithArgs("001_init").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = runUpMigrations(db, dir, []string{"001_init_up.sql"})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRunUpMigrations_AlreadyApplied(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_up.sql"), []byte("CREATE TABLE test (id INT)"), 0644))

	// Already applied → count = 1
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	err = runUpMigrations(db, dir, []string{"001_init_up.sql"})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRunUpMigrations_CheckStatusError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnError(errors.New("query failed"))

	err = runUpMigrations(db, "", []string{"001_init_up.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check migration status")
}

func TestRunUpMigrations_ReadFileError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// dir doesn't have the file
	err = runUpMigrations(db, t.TempDir(), []string{"001_init_up.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read migration file")
}

func TestRunUpMigrations_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_up.sql"), []byte("INVALID SQL"), 0644))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectBegin()
	mock.ExpectExec("INVALID SQL").WillReturnError(errors.New("syntax error"))
	mock.ExpectRollback()

	err = runUpMigrations(db, dir, []string{"001_init_up.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute migration")
}

func TestRunUpMigrations_RecordError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_up.sql"), []byte("CREATE TABLE test (id INT)"), 0644))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO migrations").
		WithArgs("001_init").
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err = runUpMigrations(db, dir, []string{"001_init_up.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record migration")
}

// --- runDownMigrations tests ---

func TestRunDownMigrations_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_down.sql"), []byte("DROP TABLE test"), 0644))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectBegin()
	mock.ExpectExec("DROP TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM migrations").
		WithArgs("001_init").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = runDownMigrations(db, dir, []string{"001_init_down.sql"})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRunDownMigrations_NotApplied(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	err = runDownMigrations(db, "", []string{"001_init_down.sql"})
	require.NoError(t, err)
}

func TestRunDownMigrations_CheckStatusError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnError(errors.New("query failed"))

	err = runDownMigrations(db, "", []string{"001_init_down.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check migration status")
}

func TestRunDownMigrations_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_down.sql"), []byte("INVALID SQL"), 0644))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectBegin()
	mock.ExpectExec("INVALID SQL").WillReturnError(errors.New("syntax error"))
	mock.ExpectRollback()

	err = runDownMigrations(db, dir, []string{"001_init_down.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute migration")
}

func TestRunDownMigrations_DeleteRecordError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_down.sql"), []byte("DROP TABLE test"), 0644))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectBegin()
	mock.ExpectExec("DROP TABLE test").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM migrations").
		WithArgs("001_init").
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	err = runDownMigrations(db, dir, []string{"001_init_down.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record migration rollback")
}

func TestRunUpMigrations_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_up.sql"), []byte("SELECT 1"), 0644))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

	err = runUpMigrations(db, dir, []string{"001_init_up.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
}

func TestRunDownMigrations_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "001_init_down.sql"), []byte("SELECT 1"), 0644))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

	err = runDownMigrations(db, dir, []string{"001_init_down.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
}

func TestRunDownMigrations_ReadFileError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("001_init").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	err = runDownMigrations(db, t.TempDir(), []string{"001_init_down.sql"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read migration file")
}
