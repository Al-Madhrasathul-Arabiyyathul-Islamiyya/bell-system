package database

import (
	"context"
	"database/sql"
	"fmt"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// ScheduleDayRepository handles database operations for schedule days
type ScheduleDayRepository struct {
	*Repository
}

// NewScheduleDayRepository creates a new schedule day repository
func NewScheduleDayRepository(db *sql.DB, logger *logger.Logger) *ScheduleDayRepository {
	return &ScheduleDayRepository{
		Repository: NewRepository(db, logger),
	}
}

// Create creates a new schedule day
func (r *ScheduleDayRepository) Create(ctx context.Context, scheduleDay *models.ScheduleDay) error {
	query, args, err := qb.Insert("ScheduleDays").
		Columns("ScheduleItemID", "DayOfWeek").
		Values(scheduleDay.ScheduleItemID, scheduleDay.DayOfWeek).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build schedule day insert query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create schedule day: %w", err)
	}
	return nil
}

// GetDaysForScheduleItems gets all days for a list of schedule item IDs
func (r *ScheduleDayRepository) GetDaysForScheduleItems(ctx context.Context, itemIDs []uuid.UUID) (map[uuid.UUID][]int, error) {
	if len(itemIDs) == 0 {
		return make(map[uuid.UUID][]int), nil
	}

	query, args, err := qb.Select("CONVERT(NVARCHAR(36), ScheduleItemId) AS ScheduleItemId", "DayOfWeek").
		From("ScheduleDays").
		Where(sq.Eq{"ScheduleItemId": itemIDs}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build days for schedule items query: %w", err)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get days for schedule items: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]int)

	for rows.Next() {
		var itemID uuid.UUID
		var dayOfWeek int

		if err := rows.Scan(&itemID, &dayOfWeek); err != nil {
			return nil, fmt.Errorf("failed to scan schedule day: %w", err)
		}

		result[itemID] = append(result[itemID], dayOfWeek)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule days rows: %w", err)
	}

	return result, nil
}

// Update updates a schedule day
func (r *ScheduleDayRepository) Update(ctx context.Context, scheduleDay *models.ScheduleDay) error {
	query, args, err := qb.Update("ScheduleDays").
		Set("DayOfWeek", scheduleDay.DayOfWeek).
		Where(sq.Eq{"ScheduleItemId": scheduleDay.ScheduleItemID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build schedule day update query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update schedule day: %w", err)
	}
	return nil
}

// Delete deletes a schedule day
func (r *ScheduleDayRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := qb.Delete("ScheduleDays").
		Where(sq.Eq{"ScheduleItemId": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build schedule day delete query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete schedule day: %w", err)
	}
	return nil
}
