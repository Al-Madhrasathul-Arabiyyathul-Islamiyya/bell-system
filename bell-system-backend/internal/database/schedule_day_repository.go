package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

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
	query := `
        INSERT INTO ScheduleDays (ScheduleItemID, DayOfWeek)
        VALUES (@p1, @p2)
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		scheduleDay.ScheduleItemID, scheduleDay.DayOfWeek,
	)
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

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(itemIDs))
	for i, id := range itemIDs {
		idStrings[i] = "'" + id.String() + "'"
	}

	// Build query with the IDs directly in the SQL
	query := fmt.Sprintf(`
        SELECT CONVERT(NVARCHAR(36), ScheduleItemId) AS ScheduleItemId, DayOfWeek
        FROM ScheduleDays
        WHERE ScheduleItemId IN (%s)
    `, strings.Join(idStrings, ", "))

	rows, err := r.DB.QueryContext(ctx, query)
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
	query := `
        UPDATE ScheduleDays
        SET DayOfWeek = @p1
        WHERE ScheduleItemId = @p2
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		scheduleDay.DayOfWeek, scheduleDay.ScheduleItemID,
	)
	if err != nil {
		return fmt.Errorf("failed to update schedule day: %w", err)
	}
	return nil
}

// Delete deletes a schedule day
func (r *ScheduleDayRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM ScheduleDays WHERE ScheduleItemId = @p1"
	_, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule day: %w", err)
	}
	return nil
}
