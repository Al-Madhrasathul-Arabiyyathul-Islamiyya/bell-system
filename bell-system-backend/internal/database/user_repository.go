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
	query, args, err := buildQuery(qb.Insert("Users").
		Columns("Id", "Username", "PasswordHash", "Role", "CreatedAt").
		Values(user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt))
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query, args, err := buildQuery(qb.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Username", "PasswordHash", "Role", "CreatedAt").
		From("Users").
		Where(sq.Eq{"Id": id}))
	if err != nil {
		return nil, err
	}
	var user models.User
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(
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
	query, args, err := buildQuery(qb.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Username", "PasswordHash", "Role", "CreatedAt").
		From("Users").
		Where(sq.Eq{"Username": username}))
	if err != nil {
		return nil, err
	}
	var user models.User
	err = r.DB.QueryRowContext(ctx, query, args...).Scan(
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

// userSortColumns maps JSON:API sort field names to SQL column names.
var userSortColumns = map[string]string{
	"username":  "Username",
	"role":      "Role",
	"createdAt": "CreatedAt",
}

// List gets users with pagination, optional role filter, and sorting.
func (r *UserRepository) List(ctx context.Context, page, size int, filterRole string, sorts []jsonapi.SortField) ([]*models.User, int, error) {
	// Build base WHERE condition
	countQB := qb.Select("COUNT(*)").From("Users")
	if filterRole != "" {
		countQB = countQB.Where(sq.Eq{"Role": filterRole})
	}

	countQuery, countArgs, err := buildQuery(countQB)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.DB.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Fetch page
	offset := (page - 1) * size
	listQB := qb.Select("CONVERT(NVARCHAR(36), Id) AS Id", "Username", "PasswordHash", "Role", "CreatedAt").
		From("Users")
	if filterRole != "" {
		listQB = listQB.Where(sq.Eq{"Role": filterRole})
	}

	orderBy := jsonapi.SortToSQL(sorts, userSortColumns, "ORDER BY Username")
	listQB = listQB.Suffix(orderBy+" OFFSET ? ROWS FETCH NEXT ? ROWS ONLY", offset, size)

	query, args, err := buildQuery(listQB)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, total, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query, args, err := buildQuery(qb.Update("Users").
		Set("Username", user.Username).
		Set("PasswordHash", user.PasswordHash).
		Set("Role", user.Role).
		Where(sq.Eq{"Id": user.ID}))
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := buildQuery(qb.Delete("Users").
		Where(sq.Eq{"Id": id}))
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
