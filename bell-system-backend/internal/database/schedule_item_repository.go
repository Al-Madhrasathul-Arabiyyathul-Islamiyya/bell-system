package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/helpers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

// ScheduleItemRepository handles database operations for sessions
type ScheduleItemRepository struct {
	*Repository
	ScheduleDayRepo *ScheduleDayRepository
	SessionRepo     *SessionRepository
	SystemAudioRepo *SystemAudioFileRepository
}

// NewScheduleItemRepository creates a new schedule item repository
func NewScheduleItemRepository(
	db *sql.DB,
	logger *logger.Logger,
	scheduleDayRepo *ScheduleDayRepository,
	sessionRepo *SessionRepository,
	systemAudioRepo *SystemAudioFileRepository,
) *ScheduleItemRepository {
	return &ScheduleItemRepository{
		Repository:      NewRepository(db, logger),
		ScheduleDayRepo: scheduleDayRepo,
		SessionRepo:     sessionRepo,
		SystemAudioRepo: systemAudioRepo,
	}
}

// Create creates a new schedule item with its days
func (r *ScheduleItemRepository) Create(ctx context.Context, item *models.ScheduleItem) error {
	r.Logger.Info("Creating schedule item",
		zap.String("name", item.Name),
		zap.String("operation", "ScheduleItemRepository.Create"))

	return r.WithTx(ctx, func(tx *sql.Tx) error {
		// Insert the schedule item
		query := `
            INSERT INTO ScheduleItems (Id, SessionId, Name, Time, SoundId, CreatedAt, UpdatedAt)
            VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)
        `

		now := time.Now()
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.CreatedAt = now
		item.UpdatedAt = now

		_, err := tx.ExecContext(
			ctx, query,
			item.ID,
			item.SessionID,
			item.Name,
			item.Time,
			item.SoundID,
			item.CreatedAt,
			item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert schedule item: %w", err)
		}

		if len(item.Days) > 0 {
			for _, day := range item.Days {
				dayQuery := `
                    INSERT INTO ScheduleDays (ScheduleItemId, DayOfWeek)
                    VALUES (@p1, @p2)
                `
				_, err := tx.ExecContext(ctx, dayQuery, item.ID, day)
				if err != nil {
					return fmt.Errorf("failed to insert schedule day: %w", err)
				}
			}
		}

		return nil
	})
}

// GetByID gets a schedule item by ID with related sound, session, and days
func (r *ScheduleItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
	query := `
        SELECT
            CONVERT(NVARCHAR(36), Id) AS Id,
            CONVERT(NVARCHAR(36), SessionId) AS SessionId,
            Name,
            Time,
            CONVERT(NVARCHAR(36), SoundId) AS SoundId,
            CreatedAt,
            UpdatedAt
        FROM ScheduleItems
        WHERE Id = @id
    `

	var scheduleItem models.ScheduleItem
	var sessionID *uuid.UUID

	err := r.DB.QueryRowContext(ctx, query, sql.Named("id", id)).Scan(
		&scheduleItem.ID,
		&sessionID,
		&scheduleItem.Name,
		&scheduleItem.Time,
		&scheduleItem.SoundID,
		&scheduleItem.CreatedAt,
		&scheduleItem.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get schedule item: %w", err)
	}

	scheduleItem.SessionID = sessionID

	scheduleItemIDs := []uuid.UUID{scheduleItem.ID}
	daysMap, err := r.ScheduleDayRepo.GetDaysForScheduleItems(ctx, scheduleItemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule days: %w", err)
	}
	scheduleItem.Days = daysMap[scheduleItem.ID]

	scheduleItem.DayInfo = helpers.MapDaysToInfo(scheduleItem.Days)

	if scheduleItem.SoundID != uuid.Nil {
		soundIDs := []uuid.UUID{scheduleItem.SoundID}
		soundsMap, err := r.SystemAudioRepo.GetSoundsByIDs(ctx, soundIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get sound file: %w", err)
		}

		if sound, ok := soundsMap[scheduleItem.SoundID]; ok {
			scheduleItem.Sound = sound
		}
	}

	if sessionID != nil {
		sessionIDs := []uuid.UUID{*sessionID}
		sessionsMap, err := r.SessionRepo.GetSessionsByIDs(ctx, sessionIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get session: %w", err)
		}

		if session, ok := sessionsMap[*sessionID]; ok {
			scheduleItem.Session = session
		}
	}

	return &scheduleItem, nil
}

// List gets all schedule items with their associations, with optional filters and sorting.
func (r *ScheduleItemRepository) List(ctx context.Context, filterSessionID string, filterDay int, sortSQL string) ([]*models.ScheduleItem, error) {
	if sortSQL == "" {
		sortSQL = "ORDER BY Time"
	}

	where := ""
	var args []any
	paramIdx := 1

	if filterSessionID != "" {
		where += fmt.Sprintf(" WHERE SessionId = @p%d", paramIdx)
		args = append(args, sql.Named(fmt.Sprintf("p%d", paramIdx), filterSessionID))
		paramIdx++
	}

	if filterDay > 0 {
		if where == "" {
			where += " WHERE"
		} else {
			where += " AND"
		}
		where += fmt.Sprintf(" Id IN (SELECT ScheduleItemId FROM ScheduleDays WHERE DayOfWeek = @p%d)", paramIdx)
		args = append(args, sql.Named(fmt.Sprintf("p%d", paramIdx), filterDay))
	}

	query := fmt.Sprintf(`
		SELECT
			CONVERT(NVARCHAR(36), Id) AS Id,
			CONVERT(NVARCHAR(36), SessionId) AS SessionId,
			Name,
			Time,
			CONVERT(NVARCHAR(36), SoundId) AS SoundId,
			CreatedAt,
			UpdatedAt
		FROM ScheduleItems
		%s
		%s
	`, where, sortSQL)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedule items: %w", err)
	}
	defer rows.Close()

	var scheduleItems []*models.ScheduleItem
	scheduleItemIDs := make([]uuid.UUID, 0)
	soundIDs := make([]uuid.UUID, 0)
	sessionIDs := make([]uuid.UUID, 0)

	soundIDMap := make(map[uuid.UUID]bool)
	sessionIDMap := make(map[uuid.UUID]bool)

	for rows.Next() {
		var scheduleItem models.ScheduleItem
		var sessionID *uuid.UUID

		if err := rows.Scan(
			&scheduleItem.ID,
			&sessionID,
			&scheduleItem.Name,
			&scheduleItem.Time,
			&scheduleItem.SoundID,
			&scheduleItem.CreatedAt,
			&scheduleItem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan schedule item: %w", err)
		}

		scheduleItem.SessionID = sessionID
		scheduleItemIDs = append(scheduleItemIDs, scheduleItem.ID)

		if scheduleItem.SoundID != uuid.Nil && !soundIDMap[scheduleItem.SoundID] {
			soundIDs = append(soundIDs, scheduleItem.SoundID)
			soundIDMap[scheduleItem.SoundID] = true
		}

		if sessionID != nil && !sessionIDMap[*sessionID] {
			sessionIDs = append(sessionIDs, *sessionID)
			sessionIDMap[*sessionID] = true
		}

		scheduleItems = append(scheduleItems, &scheduleItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule item rows: %w", err)
	}

	if len(scheduleItems) == 0 {
		return scheduleItems, nil
	}

	daysMap, err := r.ScheduleDayRepo.GetDaysForScheduleItems(ctx, scheduleItemIDs)
	if err != nil {
		return nil, err
	}

	soundsMap, err := r.SystemAudioRepo.GetSoundsByIDs(ctx, soundIDs)
	if err != nil {
		return nil, err
	}

	sessionsMap, err := r.SessionRepo.GetSessionsByIDs(ctx, sessionIDs)
	if err != nil {
		return nil, err
	}

	for _, item := range scheduleItems {
		item.Days = daysMap[item.ID]

		item.DayInfo = helpers.MapDaysToInfo(item.Days)

		if item.SoundID != uuid.Nil {
			if sound, ok := soundsMap[item.SoundID]; ok {
				item.Sound = sound
			}
		}

		if item.SessionID != nil {
			if session, ok := sessionsMap[*item.SessionID]; ok {
				item.Session = session
			}
		}
	}

	return scheduleItems, nil
}

// GetCurrentSessionSchedules gets the schedule items for the current session
func (r *ScheduleItemRepository) GetCurrentSessionSchedules(ctx context.Context, sessionID uuid.UUID) ([]*models.ScheduleItem, error) {
	currentTime := time.Now()
	currentDayOfWeek := int(currentTime.Weekday()) + 1 // Adding 1 for Sunday = 1

	query := `
        SELECT
            CONVERT(NVARCHAR(36), si.Id) AS Id,
            CONVERT(NVARCHAR(36), si.SessionId) AS SessionId,
            si.Name,
            si.Time,
            CONVERT(NVARCHAR(36), si.SoundId) AS SoundId,
            si.CreatedAt,
            si.UpdatedAt
        FROM ScheduleItems si
        INNER JOIN ScheduleDays sd ON si.Id = sd.ScheduleItemId
        WHERE si.SessionId = @sessionId
        AND sd.DayOfWeek = @dayOfWeek
        ORDER BY si.Time
    `

	rows, err := r.DB.QueryContext(ctx, query,
		sql.Named("sessionId", sessionID),
		sql.Named("dayOfWeek", currentDayOfWeek),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get current session schedules: %w", err)
	}
	defer rows.Close()

	var scheduleItems []*models.ScheduleItem
	scheduleItemIDs := make([]uuid.UUID, 0)
	soundIDs := make([]uuid.UUID, 0)
	soundIDMap := make(map[uuid.UUID]bool)

	for rows.Next() {
		var item models.ScheduleItem
		if err := rows.Scan(
			&item.ID,
			&item.SessionID,
			&item.Name,
			&item.Time,
			&item.SoundID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan schedule item: %w", err)
		}

		scheduleItemIDs = append(scheduleItemIDs, item.ID)

		if item.SoundID != uuid.Nil && !soundIDMap[item.SoundID] {
			soundIDs = append(soundIDs, item.SoundID)
			soundIDMap[item.SoundID] = true
		}

		scheduleItems = append(scheduleItems, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule item rows: %w", err)
	}

	if len(scheduleItems) == 0 {
		return scheduleItems, nil
	}

	daysMap, err := r.ScheduleDayRepo.GetDaysForScheduleItems(ctx, scheduleItemIDs)
	if err != nil {
		return nil, err
	}

	soundsMap, err := r.SystemAudioRepo.GetSoundsByIDs(ctx, soundIDs)
	if err != nil {
		return nil, err
	}

	sessionIDs := []uuid.UUID{sessionID}
	sessionsMap, err := r.SessionRepo.GetSessionsByIDs(ctx, sessionIDs)
	if err != nil {
		return nil, err
	}

	session, hasSession := sessionsMap[sessionID]

	for _, item := range scheduleItems {
		item.Days = daysMap[item.ID]

		item.DayInfo = helpers.MapDaysToInfo(item.Days)

		if item.SoundID != uuid.Nil {
			if sound, ok := soundsMap[item.SoundID]; ok {
				item.Sound = sound
			}
		}

		if hasSession {
			item.Session = session
		}
	}

	return scheduleItems, nil
}

// Update updates a schedule item and its day associations
func (r *ScheduleItemRepository) Update(ctx context.Context, item *models.ScheduleItem) error {
	r.Logger.Info("Updating schedule item",
		zap.String("id", item.ID.String()),
		zap.String("operation", "ScheduleItemRepository.Update"))

	return r.WithTx(ctx, func(tx *sql.Tx) error {
		query := `
            UPDATE ScheduleItems
            SET SessionId = @p1, Name = @p2, Time = @p3, SoundId = @p4, UpdatedAt = @p5
            WHERE Id = @p6
        `

		item.UpdatedAt = time.Now()

		_, err := tx.ExecContext(
			ctx, query,
			item.SessionID,
			item.Name,
			item.Time,
			item.SoundID,
			item.UpdatedAt,
			item.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update schedule item: %w", err)
		}

		deleteQuery := `DELETE FROM ScheduleDays WHERE ScheduleItemId = @p1`
		_, err = tx.ExecContext(ctx, deleteQuery, item.ID)
		if err != nil {
			return fmt.Errorf("failed to delete existing schedule days: %w", err)
		}

		if len(item.Days) > 0 {
			for _, day := range item.Days {
				dayQuery := `
                    INSERT INTO ScheduleDays (ScheduleItemId, DayOfWeek)
                    VALUES (@p1, @p2)
                `
				_, err := tx.ExecContext(ctx, dayQuery, item.ID, day)
				if err != nil {
					return fmt.Errorf("failed to insert schedule day: %w", err)
				}
			}
		}

		return nil
	})
}

// Delete deletes a schedule item and its associated days
func (r *ScheduleItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.WithTx(ctx, func(tx *sql.Tx) error {
		// Delete associated days first (FK constraint)
		if _, err := tx.ExecContext(ctx, "DELETE FROM ScheduleDays WHERE ScheduleItemId = @p1", id); err != nil {
			return fmt.Errorf("failed to delete schedule days: %w", err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM ScheduleItems WHERE Id = @p1", id); err != nil {
			return fmt.Errorf("failed to delete the schedule item: %w", err)
		}
		return nil
	})
}
