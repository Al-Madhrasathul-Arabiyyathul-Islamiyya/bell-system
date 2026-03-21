package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/helpers"
	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"
	"go.uber.org/zap"

	sq "github.com/Masterminds/squirrel"
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
		now := time.Now()
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.CreatedAt = now
		item.UpdatedAt = now

		query, args, err := qb.Insert("ScheduleItems").
			Columns("Id", "SessionId", "Name", "Time", "SoundId", "CreatedAt", "UpdatedAt").
			Values(item.ID, item.SessionID, item.Name, item.Time, item.SoundID, item.CreatedAt, item.UpdatedAt).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build schedule item insert query: %w", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to insert schedule item: %w", err)
		}

		for _, day := range item.Days {
			dayQuery, dayArgs, err := qb.Insert("ScheduleDays").
				Columns("ScheduleItemId", "DayOfWeek").
				Values(item.ID, day).
				ToSql()
			if err != nil {
				return fmt.Errorf("failed to build schedule day insert query: %w", err)
			}
			_, err = tx.ExecContext(ctx, dayQuery, dayArgs...)
			if err != nil {
				return fmt.Errorf("failed to insert schedule day: %w", err)
			}
		}

		return nil
	})
}

// GetByID gets a schedule item by ID with related sound, session, and days
func (r *ScheduleItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ScheduleItem, error) {
	query, args, err := qb.Select(
		"CONVERT(NVARCHAR(36), Id) AS Id",
		"CONVERT(NVARCHAR(36), SessionId) AS SessionId",
		"Name", "Time",
		"CONVERT(NVARCHAR(36), SoundId) AS SoundId",
		"CreatedAt", "UpdatedAt",
	).From("ScheduleItems").
		Where(sq.Eq{"Id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build schedule item get query: %w", err)
	}

	var scheduleItem models.ScheduleItem
	var sessionID *uuid.UUID

	err = r.DB.QueryRowContext(ctx, query, args...).Scan(
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

// scheduleSortColumns maps JSON:API sort field names to SQL column names.
var scheduleSortColumns = map[string]string{
	"name": "Name",
	"time": "Time",
}

// List gets all schedule items with their associations, with optional filters and sorting.
func (r *ScheduleItemRepository) List(ctx context.Context, filterSessionID string, filterDay int, sorts []jsonapi.SortField) ([]*models.ScheduleItem, error) {
	sb := qb.Select(
		"CONVERT(NVARCHAR(36), Id) AS Id",
		"CONVERT(NVARCHAR(36), SessionId) AS SessionId",
		"Name", "Time",
		"CONVERT(NVARCHAR(36), SoundId) AS SoundId",
		"CreatedAt", "UpdatedAt",
	).From("ScheduleItems")

	if filterSessionID != "" {
		sb = sb.Where(sq.Eq{"SessionId": filterSessionID})
	}

	if filterDay > 0 {
		sb = sb.Where("Id IN (SELECT ScheduleItemId FROM ScheduleDays WHERE DayOfWeek = ?)", filterDay)
	}

	orderBy := jsonapi.SortToSQL(sorts, scheduleSortColumns, "ORDER BY Time")
	sb = sb.Suffix(orderBy)

	query, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build schedule item list query: %w", err)
	}

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

	query, args, err := qb.Select(
		"CONVERT(NVARCHAR(36), si.Id) AS Id",
		"CONVERT(NVARCHAR(36), si.SessionId) AS SessionId",
		"si.Name", "si.Time",
		"CONVERT(NVARCHAR(36), si.SoundId) AS SoundId",
		"si.CreatedAt", "si.UpdatedAt",
	).From("ScheduleItems si").
		Join("ScheduleDays sd ON si.Id = sd.ScheduleItemId").
		Where(sq.Eq{"si.SessionId": sessionID}).
		Where(sq.Eq{"sd.DayOfWeek": currentDayOfWeek}).
		Suffix("ORDER BY si.Time").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build current session schedules query: %w", err)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
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
		item.UpdatedAt = time.Now()

		query, args, err := qb.Update("ScheduleItems").
			Set("SessionId", item.SessionID).
			Set("Name", item.Name).
			Set("Time", item.Time).
			Set("SoundId", item.SoundID).
			Set("UpdatedAt", item.UpdatedAt).
			Where(sq.Eq{"Id": item.ID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build schedule item update query: %w", err)
		}

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update schedule item: %w", err)
		}

		delQuery, delArgs, err := qb.Delete("ScheduleDays").
			Where(sq.Eq{"ScheduleItemId": item.ID}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build schedule days delete query: %w", err)
		}
		_, err = tx.ExecContext(ctx, delQuery, delArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete existing schedule days: %w", err)
		}

		for _, day := range item.Days {
			dayQuery, dayArgs, err := qb.Insert("ScheduleDays").
				Columns("ScheduleItemId", "DayOfWeek").
				Values(item.ID, day).
				ToSql()
			if err != nil {
				return fmt.Errorf("failed to build schedule day insert query: %w", err)
			}
			_, err = tx.ExecContext(ctx, dayQuery, dayArgs...)
			if err != nil {
				return fmt.Errorf("failed to insert schedule day: %w", err)
			}
		}

		return nil
	})
}

// Delete deletes a schedule item and its associated days
func (r *ScheduleItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.WithTx(ctx, func(tx *sql.Tx) error {
		// Delete associated days first (FK constraint)
		daysQuery, daysArgs, err := qb.Delete("ScheduleDays").
			Where(sq.Eq{"ScheduleItemId": id}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build schedule days delete query: %w", err)
		}
		if _, err := tx.ExecContext(ctx, daysQuery, daysArgs...); err != nil {
			return fmt.Errorf("failed to delete schedule days: %w", err)
		}

		itemQuery, itemArgs, err := qb.Delete("ScheduleItems").
			Where(sq.Eq{"Id": id}).
			ToSql()
		if err != nil {
			return fmt.Errorf("failed to build schedule item delete query: %w", err)
		}
		if _, err := tx.ExecContext(ctx, itemQuery, itemArgs...); err != nil {
			return fmt.Errorf("failed to delete the schedule item: %w", err)
		}
		return nil
	})
}
