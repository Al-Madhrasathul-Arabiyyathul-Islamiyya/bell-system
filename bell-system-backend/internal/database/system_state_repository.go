package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"arabiyya.edu.mv/bell-system-backend/pkg/logger"
)

// SystemStateRepository handles database operations for system state.
type SystemStateRepository struct {
	*Repository
}

// NewSystemStateRepository creates a new system state repository.
func NewSystemStateRepository(db *sql.DB, logger *logger.Logger) *SystemStateRepository {
	return &SystemStateRepository{
		Repository: NewRepository(db, logger),
	}
}

// GetState returns the current system state value and its last update time.
func (r *SystemStateRepository) GetState(ctx context.Context) (string, time.Time, error) {
	query := `SELECT Value, UpdatedAt FROM SystemState WHERE [Key] = 'system_state'`

	var value string
	var updatedAt time.Time
	err := r.DB.QueryRowContext(ctx, query).Scan(&value, &updatedAt)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get system state: %w", err)
	}

	return value, updatedAt, nil
}

// SetState updates the system state value.
func (r *SystemStateRepository) SetState(ctx context.Context, state string) error {
	query := `UPDATE SystemState SET Value = @p1, UpdatedAt = GETDATE() WHERE [Key] = 'system_state'`

	_, err := r.DB.ExecContext(ctx, query, state)
	if err != nil {
		return fmt.Errorf("failed to set system state: %w", err)
	}

	return nil
}
