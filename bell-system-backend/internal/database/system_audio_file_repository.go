package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/google/uuid"
)

// SystemAudioFileRepository handles database operations for SystemAudioFiles
type SystemAudioFileRepository struct {
	*Repository
}

// NewSystemAudioFileRepository creates a new SystemAudioFile repository
func NewSystemAudioFileRepository(db *sql.DB, logger *logger.Logger) *SystemAudioFileRepository {
	return &SystemAudioFileRepository{
		Repository: NewRepository(db, logger),
	}
}

// Create creates a new system audio file
func (r *SystemAudioFileRepository) Create(ctx context.Context, audio *models.SystemAudioFile) error {
	query := `
        INSERT INTO SystemAudioFiles (Id, Name, FilePath, FileType, Checksum)
        VALUES (@p1, @p2, @p3, @p4, @p5)
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		audio.ID, audio.Name, audio.FilePath, audio.FileType, audio.Checksum,
	)
	if err != nil {
		return fmt.Errorf("failed to create system audio file: %w", err)
	}
	return nil
}

// GetByID gets a system audio file by ID
func (r *SystemAudioFileRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
	query := `
        SELECT Id, Name, FilePath, FileType, Checksum
        FROM SystemAudioFiles
        WHERE Id = @p1
    `
	var audio models.SystemAudioFile
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&audio.ID, &audio.Name, &audio.FilePath, &audio.FileType, &audio.Checksum,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system audio file: %w", err)
	}
	return &audio, nil
}

// GetSoundsByIDs gets multiple sound files by their IDs
func (r *SystemAudioFileRepository) GetSoundsByIDs(ctx context.Context, soundIDs []uuid.UUID) (map[uuid.UUID]*models.SystemAudioFile, error) {
	if len(soundIDs) == 0 {
		return make(map[uuid.UUID]*models.SystemAudioFile), nil
	}

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(soundIDs))
	for i, id := range soundIDs {
		idStrings[i] = "'" + id.String() + "'"
	}

	query := fmt.Sprintf(`
        SELECT Id, Name, FilePath, FileType, Checksum, CreatedAt, UpdatedAt
        FROM SystemAudioFiles
        WHERE Id IN (%s)
    `, strings.Join(idStrings, ", "))

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sound files: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*models.SystemAudioFile)

	for rows.Next() {
		var sound models.SystemAudioFile

		if err := rows.Scan(
			&sound.ID,
			&sound.Name,
			&sound.FilePath,
			&sound.FileType,
			&sound.Checksum,
			&sound.CreatedAt,
			&sound.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sound file: %w", err)
		}

		result[sound.ID] = &sound
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sound files rows: %w", err)
	}

	return result, nil
}

// List gets all system audio file
func (r *SystemAudioFileRepository) List(ctx context.Context) ([]*models.SystemAudioFile, error) {
	query := `
        SELECT Id, Name, FilePath, FileType, Checksum
        FROM SystemAudioFiles
    `
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var audios []*models.SystemAudioFile
	for rows.Next() {
		var audio models.SystemAudioFile
		if err := rows.Scan(&audio.ID, &audio.Name, &audio.FilePath, &audio.FileType, &audio.Checksum); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		audios = append(audios, &audio)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return audios, nil
}

// Update updates a system audio file
func (r *SystemAudioFileRepository) Update(ctx context.Context, audio *models.SystemAudioFile) error {
	query := `
        UPDATE SystemAudioFiles
        SET Name = @p1, FilePath = @p2, FileType = @p3, Checksum = @p4
        WHERE Id = @p5
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		audio.Name, audio.FilePath, audio.FileType, audio.Checksum, audio.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// Delete deletes a system audio file
func (r *SystemAudioFileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM SystemAudioFiles WHERE Id = @p1"
	_, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete system audio file: %w", err)
	}
	return nil
}
