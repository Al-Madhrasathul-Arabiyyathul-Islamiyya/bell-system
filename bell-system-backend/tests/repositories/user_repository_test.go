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

func TestUserRepository_Create_Success(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashed",
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	mock.ExpectExec("INSERT INTO Users").
		WithArgs(user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DBError(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	user := &models.User{
		ID:           uuid.New(),
		Username:     "duplicate",
		PasswordHash: "hashed",
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	mock.ExpectExec("INSERT INTO Users").
		WithArgs(user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt).
		WillReturnError(errors.New("duplicate key"))

	err := repo.Create(ctx, user)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
}

func TestUserRepository_GetByID_Found(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	id := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"Id", "Username", "PasswordHash", "Role", "CreatedAt"}).
		AddRow(id, "admin", "hash123", "admin", now)

	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").
		WithArgs(id).
		WillReturnRows(rows)

	user, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, id, user.ID)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "hash123", user.PasswordHash)
	assert.Equal(t, models.RoleAdmin, user.Role)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetByID_DBError(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("connection lost"))

	user, err := repo.GetByID(ctx, uuid.New())
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user")
}

func TestUserRepository_GetByUsername_Found(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	id := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"Id", "Username", "PasswordHash", "Role", "CreatedAt"}).
		AddRow(id, "morning_user", "hash", "morning_user", now)

	mock.ExpectQuery("SELECT .+ FROM Users WHERE Username").
		WithArgs("morning_user").
		WillReturnRows(rows)

	user, err := repo.GetByUsername(ctx, "morning_user")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "morning_user", user.Username)
	assert.Equal(t, models.RoleMorningUser, user.Role)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM Users WHERE Username").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByUsername(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_List_MultipleUsers(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"Id", "Username", "PasswordHash", "Role", "CreatedAt"}).
		AddRow(uuid.New(), "admin", "hash1", "admin", now).
		AddRow(uuid.New(), "afternoon_user", "hash2", "afternoon_user", now).
		AddRow(uuid.New(), "morning_user", "hash3", "morning_user", now)

	mock.ExpectQuery("SELECT .+ FROM Users ORDER BY Username").
		WillReturnRows(rows)

	users, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 3)
	assert.Equal(t, "admin", users[0].Username)
	assert.Equal(t, "afternoon_user", users[1].Username)
	assert.Equal(t, "morning_user", users[2].Username)
}

func TestUserRepository_List_Empty(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "Username", "PasswordHash", "Role", "CreatedAt"})
	mock.ExpectQuery("SELECT .+ FROM Users ORDER BY Username").
		WillReturnRows(rows)

	users, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestUserRepository_List_DBError(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM Users ORDER BY Username").
		WillReturnError(errors.New("query failed"))

	users, err := repo.List(ctx)
	require.Error(t, err)
	assert.Nil(t, users)
}

func TestUserRepository_Update_Success(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	user := &models.User{
		ID:           uuid.New(),
		Username:     "updated_user",
		PasswordHash: "newhash",
		Role:         models.RoleAfternoonUser,
	}

	mock.ExpectExec("UPDATE Users").
		WithArgs(user.Username, user.PasswordHash, user.Role, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, user)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_DBError(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	user := &models.User{ID: uuid.New(), Username: "fail", PasswordHash: "h", Role: models.RoleAdmin}

	mock.ExpectExec("UPDATE Users").
		WithArgs(user.Username, user.PasswordHash, user.Role, user.ID).
		WillReturnError(errors.New("update failed"))

	err := repo.Update(ctx, user)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update user")
}

func TestUserRepository_Delete_Success(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectExec("DELETE FROM Users WHERE").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, id)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_DBError(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM Users WHERE").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("foreign key constraint"))

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user")
}

// --- Spec-driven: Three distinct roles ---

func TestUserRepository_AllThreeRoles(t *testing.T) {
	repo, mock := mocks.NewMockUserRepo(t)
	ctx := context.Background()

	roles := []models.Role{models.RoleAdmin, models.RoleMorningUser, models.RoleAfternoonUser}

	for _, role := range roles {
		user := &models.User{
			ID:           uuid.New(),
			Username:     string(role),
			PasswordHash: "hash",
			Role:         role,
			CreatedAt:    time.Now(),
		}

		mock.ExpectExec("INSERT INTO Users").
			WithArgs(user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Create(ctx, user)
		require.NoError(t, err, "should create user with role %s", role)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}
