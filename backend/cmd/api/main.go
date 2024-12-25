package main

import (
	"fmt"
	"log"

	"bell-schedule-system/internal/config"
	"bell-schedule-system/internal/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := db.TestQuery(); err != nil {
		log.Fatalf("Database test query failed: %v", err)
	}

	fmt.Println("Successfully connected to database and ran test query")
}
