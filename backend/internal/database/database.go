package database

import (
	"bell-schedule-system/internal/config"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
)

type DB struct {
	*sql.DB
}

func New(cfg *config.Config) (*DB, error) {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	err = db.PingContext(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &DB{db}, nil
}

func (db *DB) TestQuery() error {
	var testResult string
	err := db.QueryRow("SELECT TOP 1 Name FROM Sessions").Scan(&testResult)
	if err != nil {
		return fmt.Errorf("test query failed: %v", err)
	}
	return nil
}
