package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

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
	query := `
        INSERT INTO Sessions (Id, Name, StartTime, EndTime)
        VALUES (@p1, @p2, @p3, @p4)
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		session.ID, session.Name,
		session.StartTime.Format("15:04"),
		session.EndTime.Format("15:04"),
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetByID gets a session by ID
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	query := `
        SELECT CONVERT(NVARCHAR(36), Id) AS Id, Name, StartTime, EndTime
        FROM Sessions
        WHERE Id = @p1
    `
	var session models.Session
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
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

	// Convert UUIDs to strings for the query
	idStrings := make([]string, len(sessionIDs))
	for i, id := range sessionIDs {
		idStrings[i] = "'" + id.String() + "'"
	}

	query := fmt.Sprintf(`
		SELECT CONVERT(NVARCHAR(36), Id) AS Id, Name, StartTime, EndTime
		FROM Sessions
		WHERE Id IN (%s)
	`, strings.Join(idStrings, ", "))

	rows, err := r.DB.QueryContext(ctx, query)
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

// List gets all sessions
func (r *SessionRepository) List(ctx context.Context) ([]*models.Session, error) {
	query := `
        SELECT CONVERT(NVARCHAR(36), Id) AS Id, Name, StartTime, EndTime
        FROM Sessions
        ORDER BY StartTime
    `
	rows, err := r.DB.QueryContext(ctx, query)
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

	query := `
        SELECT CONVERT(NVARCHAR(36), Id) AS Id, Name, StartTime, EndTime
        FROM Sessions
        WHERE CAST(@p1 AS TIME) BETWEEN StartTime AND EndTime
    `
	var session models.Session
	err := r.DB.QueryRowContext(ctx, query, currentTime).Scan(
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
	query := `
        UPDATE Sessions
        SET Name = @p1, StartTime = @p2, EndTime = @p3
        WHERE Id = @p4
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		session.Name,
		session.StartTime.Format("15:04"),
		session.EndTime.Format("15:04"),
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// Delete deletes a session
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM Sessions WHERE Id = @p1"
	_, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
