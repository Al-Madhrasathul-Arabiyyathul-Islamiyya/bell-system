package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"arabiyya.edu.mv/bell-system-backend/config"

	_ "github.com/denisenkom/go-mssqldb"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(cfg config.DatabaseConfig) (*DB, error) {
	connStr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		cfg.Host, cfg.User, cfg.Password, cfg.Port, cfg.Name)

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
