package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	sq "github.com/Masterminds/squirrel"
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
	query, args, err := sq.Insert("SystemAudioFiles").
		Columns("Id", "Name", "FilePath", "FileType", "Checksum").
		Values(audio.ID, audio.Name, audio.FilePath, audio.FileType, audio.Checksum).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build audio file insert query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create system audio file: %w", err)
	}
	return nil
}

// GetByID gets a system audio file by ID
func (r *SystemAudioFileRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SystemAudioFile, error) {
	query, args, err := sq.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Name", "FilePath", "FileType", "Checksum").
		From("SystemAudioFiles").
		Where(sq.Eq{"Id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build audio file get query: %w", err)
	}
	var audio models.SystemAudioFile
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(
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

	query, args, err := sq.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Name", "FilePath", "FileType", "Checksum", "CreatedAt", "UpdatedAt").
		From("SystemAudioFiles").
		Where(sq.Eq{"Id": soundIDs}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sounds by IDs query: %w", err)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sound files: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*models.SystemAudioFile)

	for rows.Next() {
		var sound models.SystemAudioFile
		var updatedAt sql.NullTime

		if err := rows.Scan(
			&sound.ID,
			&sound.Name,
			&sound.FilePath,
			&sound.FileType,
			&sound.Checksum,
			&sound.CreatedAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sound file: %w", err)
		}

		if updatedAt.Valid {
			sound.UpdatedAt = updatedAt.Time
		}

		result[sound.ID] = &sound
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sound files rows: %w", err)
	}

	return result, nil
}

// audioSortColumns maps JSON:API sort field names to SQL column names.
var audioSortColumns = map[string]string{
	"name":      "Name",
	"fileType":  "FileType",
	"createdAt": "CreatedAt",
}

// List gets audio files with pagination, optional file type filter, and sorting.
func (r *SystemAudioFileRepository) List(ctx context.Context, page, size int, filterFileType string, sorts []jsonapi.SortField) ([]*models.SystemAudioFile, int, error) {
	countQB := sq.Select("COUNT(*)").From("SystemAudioFiles")
	if filterFileType != "" {
		countQB = countQB.Where(sq.Eq{"FileType": filterFileType})
	}

	countQuery, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build audio file count query: %w", err)
	}

	var total int
	if err := r.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audio files: %w", err)
	}

	offset := (page - 1) * size
	listQB := sq.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Name", "FilePath", "FileType", "Checksum").
		From("SystemAudioFiles")
	if filterFileType != "" {
		listQB = listQB.Where(sq.Eq{"FileType": filterFileType})
	}

	orderBy := jsonapi.SortToSQL(sorts, audioSortColumns, "ORDER BY Name")
	listQB = listQB.Suffix(orderBy+" OFFSET ? ROWS FETCH NEXT ? ROWS ONLY", offset, size)

	query, args, err := listQB.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build audio file list query: %w", err)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audio files: %w", err)
	}
	defer rows.Close()

	var audios []*models.SystemAudioFile
	for rows.Next() {
		var audio models.SystemAudioFile
		if err := rows.Scan(&audio.ID, &audio.Name, &audio.FilePath, &audio.FileType, &audio.Checksum); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audio file: %w", err)
		}
		audios = append(audios, &audio)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audio file rows: %w", err)
	}

	return audios, total, nil
}

// Update updates a system audio file
func (r *SystemAudioFileRepository) Update(ctx context.Context, audio *models.SystemAudioFile) error {
	query, args, err := sq.Update("SystemAudioFiles").
		Set("Name", audio.Name).
		Set("FilePath", audio.FilePath).
		Set("FileType", audio.FileType).
		Set("Checksum", audio.Checksum).
		Where(sq.Eq{"Id": audio.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build audio file update query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// Delete deletes a system audio file
func (r *SystemAudioFileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := sq.Delete("SystemAudioFiles").
		Where(sq.Eq{"Id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build audio file delete query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete system audio file: %w", err)
	}
	return nil
}
