package tests

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

func TestSystemAudioFileRepository_Create_Success(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	audio := &models.SystemAudioFile{
		ID:       uuid.New(),
		Name:     "bell.mp3",
		FilePath: "/audio/bell.mp3",
		FileType: models.FileTypeBell,
		Checksum: "sha256hash",
	}

	mock.ExpectExec("INSERT INTO SystemAudioFiles").
		WithArgs(audio.ID, audio.Name, audio.FilePath, audio.FileType, audio.Checksum).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, audio)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSystemAudioFileRepository_Create_DBError(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	audio := &models.SystemAudioFile{ID: uuid.New(), Name: "fail.mp3", FilePath: "/f", FileType: models.FileTypeBell, Checksum: "x"}

	mock.ExpectExec("INSERT INTO SystemAudioFiles").
		WithArgs(audio.ID, audio.Name, audio.FilePath, audio.FileType, audio.Checksum).
		WillReturnError(errors.New("insert failed"))

	err := repo.Create(ctx, audio)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create system audio file")
}

func TestSystemAudioFileRepository_GetByID_NotFound(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	audio, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, audio)
}

func TestSystemAudioFileRepository_GetSoundsByIDs_Empty(t *testing.T) {
	repo, _ := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	result, err := repo.GetSoundsByIDs(ctx, []uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSystemAudioFileRepository_GetSoundsByIDs_Multiple(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	id1 := uuid.New()
	id2 := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt"}).
		AddRow(id1, "bell.mp3", "/audio/bell.mp3", "bell", "hash1", now, now).
		AddRow(id2, "anthem.mp3", "/audio/anthem.mp3", "anthem", "hash2", now, now)

	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles WHERE Id IN").
		WillReturnRows(rows)

	result, err := repo.GetSoundsByIDs(ctx, []uuid.UUID{id1, id2})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "bell.mp3", result[id1].Name)
	assert.Equal(t, models.FileType("bell"), result[id1].FileType)
	assert.Equal(t, "anthem.mp3", result[id2].Name)
}

func TestSystemAudioFileRepository_List_Success(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum"}).
		AddRow(uuid.New(), "bell.mp3", "/audio/bell.mp3", "bell", "h1").
		AddRow(uuid.New(), "anthem.mp3", "/audio/anthem.mp3", "anthem", "h2").
		AddRow(uuid.New(), "song.mp3", "/audio/song.mp3", "school_song", "h3")

	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles").
		WillReturnRows(rows)

	audios, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, audios, 3)
}

func TestSystemAudioFileRepository_List_Empty(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"Id", "Name", "FilePath", "FileType", "Checksum"})
	mock.ExpectQuery("SELECT .+ FROM SystemAudioFiles").
		WillReturnRows(rows)

	audios, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, audios)
}

func TestSystemAudioFileRepository_Update_Success(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	audio := &models.SystemAudioFile{
		ID:       uuid.New(),
		Name:     "updated.mp3",
		FilePath: "/audio/updated.mp3",
		FileType: models.FileTypeOther,
		Checksum: "newhash",
	}

	mock.ExpectExec("UPDATE SystemAudioFiles").
		WithArgs(audio.Name, audio.FilePath, audio.FileType, audio.Checksum, audio.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, audio)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSystemAudioFileRepository_Delete_Success(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	id := uuid.New()
	mock.ExpectExec("DELETE FROM SystemAudioFiles WHERE").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, id)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSystemAudioFileRepository_Delete_DBError(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM SystemAudioFiles WHERE").
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("constraint violation"))

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete system audio file")
}

// --- Spec-driven: Four audio file types ---

func TestSystemAudioFileRepository_AllFileTypes(t *testing.T) {
	repo, mock := mocks.NewMockSystemAudioFileRepo(t)
	ctx := context.Background()

	fileTypes := []models.FileType{
		models.FileTypeBell,
		models.FileTypeAnthem,
		models.FileTypeSchoolSong,
		models.FileTypeOther,
	}

	for _, ft := range fileTypes {
		audio := &models.SystemAudioFile{
			ID:       uuid.New(),
			Name:     string(ft) + ".mp3",
			FilePath: "/audio/" + string(ft) + ".mp3",
			FileType: ft,
			Checksum: "hash_" + string(ft),
		}

		mock.ExpectExec("INSERT INTO SystemAudioFiles").
			WithArgs(audio.ID, audio.Name, audio.FilePath, audio.FileType, audio.Checksum).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Create(ctx, audio)
		require.NoError(t, err, "should create audio file with type %s", ft)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
