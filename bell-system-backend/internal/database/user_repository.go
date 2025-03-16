package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/google/uuid"
)

// UserRepository handles database operations for users
type UserRepository struct {
	*Repository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, logger *logger.Logger) *UserRepository {
	return &UserRepository{
		Repository: NewRepository(db, logger),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO Users (Id, Username, PasswordHash, Role, CreatedAt)
        VALUES (@p1, @p2, @p3, @p4, @p5)
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
        SELECT Id, Username, PasswordHash, Role, CreatedAt
        FROM Users
        WHERE Id = @p1
    `
	var user models.User
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByUsername gets a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT Id, Username, PasswordHash, Role, CreatedAt
        FROM Users
        WHERE Username = @p1
    `
	var user models.User
	err := r.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// List gets all users
func (r *UserRepository) List(ctx context.Context) ([]*models.User, error) {
	query := `
        SELECT Id, Username, PasswordHash, Role, CreatedAt
        FROM Users
        ORDER BY Username
    `
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE Users
        SET Username = @p1, PasswordHash = @p2, Role = @p3
        WHERE Id = @p4
    `
	_, err := r.DB.ExecContext(
		ctx, query,
		user.Username, user.PasswordHash, user.Role, user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM Users WHERE Id = @p1"
	_, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
