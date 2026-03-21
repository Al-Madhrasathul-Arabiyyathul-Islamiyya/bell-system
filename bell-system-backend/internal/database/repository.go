package database

import (
	"context"
	"database/sql"
	"fmt"

	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	sq "github.com/Masterminds/squirrel"
)

// qb is the package-level statement builder configured for SQL Server (@p1, @p2 placeholders).
var qb = sq.StatementBuilder.PlaceholderFormat(sq.AtP)

// Repository provides common database operations
type Repository struct {
	DB     *sql.DB
	Logger *logger.Logger
}

// NewRepository creates a new repository with the given database and logger
func NewRepository(db *sql.DB, logger *logger.Logger) *Repository {
	return &Repository{DB: db, Logger: logger}
}

// WithTx executes a function within a transaction
func (r *Repository) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
