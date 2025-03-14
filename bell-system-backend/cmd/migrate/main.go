package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"arabiyya.edu.mv/bell-system-backend/config"
	"arabiyya.edu.mv/bell-system-backend/internal/database"

	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	down := flag.Bool("down", false, "Run down migrations")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := ensureMigrationsTable(db.DB); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	migrationsDir := "migrations"
	files, err := getMigrationFiles(migrationsDir, *down)
	if err != nil {
		log.Fatalf("Failed to get migration files: %v", err)
	}

	if len(files) == 0 {
		log.Println("No migrations to run")
		return
	}

	if *down {
		runDownMigrations(db.DB, migrationsDir, files)
	} else {
		runUpMigrations(db.DB, migrationsDir, files)
	}
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

func runUpMigrations(db *sql.DB, dir string, files []string) {
	for _, file := range files {
		baseName := strings.TrimSuffix(file, "_up.sql")
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = @p1", baseName).Scan(&count)
		if err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		}

		if count > 0 {
			log.Printf("Skipping migration %s (already applied)", baseName)
			continue
		}

		log.Printf("Applying migration: %s", baseName)
		filePath := filepath.Join(dir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file: %v", err)
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration: %v", err)
		}

		_, err = tx.Exec("INSERT INTO migrations (name) VALUES (@p1)", baseName)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration: %v", err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit transaction: %v", err)
		}

		log.Printf("Successfully applied migration: %s", baseName)
	}
}

func runDownMigrations(db *sql.DB, dir string, files []string) {
	for _, file := range files {
		baseName := strings.TrimSuffix(file, "_down.sql")
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = @p1", baseName).Scan(&count)
		if err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		}

		if count == 0 {
			log.Printf("Skipping migration %s (not applied)", baseName)
			continue
		}

		log.Printf("Reverting migration: %s", baseName)
		filePath := filepath.Join(dir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file: %v", err)
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration: %v", err)
		}

		_, err = tx.Exec("DELETE FROM migrations WHERE name = @p1", baseName)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration rollback: %v", err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit transaction: %v", err)
		}

		log.Printf("Successfully reverted migration: %s", baseName)
	}
}
