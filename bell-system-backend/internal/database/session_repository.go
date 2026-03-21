package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// SessionRepository handles database operations for sessions
type SessionRepository struct {
	*Repository
	NowFunc func() time.Time
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB, logger *logger.Logger) *SessionRepository {
	return &SessionRepository{
		Repository: NewRepository(db, logger),
	}
}

// Create creates a new session
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	query, args, err := sq.Insert("Sessions").
		Columns("Id", "Name", "StartTime", "EndTime").
		Values(session.ID, session.Name, session.StartTime.Format("15:04"), session.EndTime.Format("15:04")).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build session insert query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetByID gets a session by ID
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	query, args, err := sq.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Name", "StartTime", "EndTime").
		From("Sessions").
		Where(sq.Eq{"Id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build session get query: %w", err)
	}
	var session models.Session
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(
		&session.ID, &session.Name, &session.StartTime, &session.EndTime,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// GetSessionsByIDs gets multiple sessions by their IDs
func (r *SessionRepository) GetSessionsByIDs(ctx context.Context, sessionIDs []uuid.UUID) (map[uuid.UUID]*models.Session, error) {
	if len(sessionIDs) == 0 {
		return make(map[uuid.UUID]*models.Session), nil
	}

	query, args, err := sq.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Name", "StartTime", "EndTime").
		From("Sessions").
		Where(sq.Eq{"Id": sessionIDs}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sessions by IDs query: %w", err)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*models.Session)

	for rows.Next() {
		var session models.Session

		if err := rows.Scan(
			&session.ID,
			&session.Name,
			&session.StartTime,
			&session.EndTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		result[session.ID] = &session
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return result, nil
}

// sessionSortColumns maps JSON:API sort field names to SQL column names.
var sessionSortColumns = map[string]string{
	"name":      "Name",
	"startTime": "StartTime",
	"endTime":   "EndTime",
}

// List gets all sessions with optional sorting.
func (r *SessionRepository) List(ctx context.Context, sorts []jsonapi.SortField) ([]*models.Session, error) {
	qb := sq.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Name", "StartTime", "EndTime").
		From("Sessions")

	orderBy := jsonapi.SortToSQL(sorts, sessionSortColumns, "ORDER BY StartTime")
	qb = qb.Suffix(orderBy)

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build session list query: %w", err)
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session
		if err := rows.Scan(&session.ID, &session.Name, &session.StartTime, &session.EndTime); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return sessions, nil
}

// GetCurrentSession gets the session that includes the current time
func (r *SessionRepository) GetCurrentSession(ctx context.Context) (*models.Session, error) {
	nowFn := r.NowFunc
	if nowFn == nil {
		nowFn = time.Now
	}
	now := nowFn()
	currentTime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	query, args, err := sq.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Name", "StartTime", "EndTime").
		From("Sessions").
		Where("CAST(? AS TIME) BETWEEN StartTime AND EndTime", currentTime).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build current session query: %w", err)
	}
	var session models.Session
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(
		&session.ID, &session.Name, &session.StartTime, &session.EndTime,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current session: %w", err)
	}
	return &session, nil
}

// Update updates a session
func (r *SessionRepository) Update(ctx context.Context, session *models.Session) error {
	query, args, err := sq.Update("Sessions").
		Set("Name", session.Name).
		Set("StartTime", session.StartTime.Format("15:04")).
		Set("EndTime", session.EndTime.Format("15:04")).
		Where(sq.Eq{"Id": session.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build session update query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// Delete deletes a session
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := sq.Delete("Sessions").
		Where(sq.Eq{"Id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build session delete query: %w", err)
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
