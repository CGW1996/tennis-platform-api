package main

import (
	"log"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/db"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	log.Println("Migrating database...")

	// Add target_criteria column to matches table
	err = database.DB.Exec("ALTER TABLE matches ADD COLUMN IF NOT EXISTS target_criteria JSONB").Error
	if err != nil {
		log.Fatal("Failed to add column:", err)
	}

	log.Println("Migration completed successfully: target_criteria column added to matches.")
}
