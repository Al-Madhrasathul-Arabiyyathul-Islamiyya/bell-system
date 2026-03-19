package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"arabiyya.edu.mv/bell-system-backend/config"
	"arabiyya.edu.mv/bell-system-backend/internal/database"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	down := flag.Bool("down", false, "Run down migrations")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := ensureMigrationsTable(db.DB); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationsDir := "migrations"
	files, err := getMigrationFiles(migrationsDir, *down)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	if len(files) == 0 {
		log.Println("No migrations to run")
		return nil
	}

	if *down {
		return runDownMigrations(db.DB, migrationsDir, files)
	}
	return runUpMigrations(db.DB, migrationsDir, files)
}

func ensureMigrationsTable(db *sql.DB) error {
	query := `
	IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='migrations' AND xtype='U')
	CREATE TABLE migrations (
		id INT PRIMARY KEY IDENTITY(1,1),
		name NVARCHAR(255) NOT NULL,
		applied_at DATETIME2 NOT NULL DEFAULT GETDATE()
	)
	`
	_, err := db.Exec(query)
	return err
}

func getMigrationFiles(dir string, down bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if strings.HasSuffix(fileName, ".sql") {
			if down && strings.HasSuffix(fileName, "_down.sql") {
				files = append(files, fileName)
			} else if !down && strings.HasSuffix(fileName, "_up.sql") {
				files = append(files, fileName)
			}
		}
	}

	sort.Strings(files)
	if down {
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	return files, nil
}

func runUpMigrations(db *sql.DB, dir string, files []string) error {
	for _, file := range files {
		baseName := strings.TrimSuffix(file, "_up.sql")
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = @p1", baseName).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", baseName, err)
		}

		if count > 0 {
			log.Printf("Skipping migration %s (already applied)", baseName)
			continue
		}

		log.Printf("Applying migration: %s", baseName)
		filePath := filepath.Join(dir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", baseName, err)
		}

		_, err = tx.Exec("INSERT INTO migrations (name) VALUES (@p1)", baseName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", baseName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for %s: %w", baseName, err)
		}

		log.Printf("Successfully applied migration: %s", baseName)
	}
	return nil
}

func runDownMigrations(db *sql.DB, dir string, files []string) error {
	for _, file := range files {
		baseName := strings.TrimSuffix(file, "_down.sql")
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = @p1", baseName).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", baseName, err)
		}

		if count == 0 {
			log.Printf("Skipping migration %s (not applied)", baseName)
			continue
		}

		log.Printf("Reverting migration: %s", baseName)
		filePath := filepath.Join(dir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", baseName, err)
		}

		_, err = tx.Exec("DELETE FROM migrations WHERE name = @p1", baseName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration rollback %s: %w", baseName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for %s: %w", baseName, err)
		}

		log.Printf("Successfully reverted migration: %s", baseName)
	}
	return nil
}
