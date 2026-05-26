package main

import (
	"flag"
	"log"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/db"
)

func main() {
	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化數據庫
	database, err := db.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// 手動修復 schema (因為 GORM AutoMigrate 在某些情況下會失敗)
	log.Println("Applying schema fixes...")
	if err := database.DB.DB.Exec("ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS play_types text[]").Error; err != nil {
		log.Printf("Warning: Failed to add play_types column: %v", err)
	}

	// 定義命令行參數
	reset := flag.Bool("reset", false, "Clear all data before seeding")
	flag.Parse()

	// 執行種子數據邏輯
	seeder := db.NewSeeder(database.DB.DB)

	if *reset {
		log.Println("Reset flag provided, clearing all data...")
		if err := seeder.ClearAll(); err != nil {
			log.Fatal("Failed to clear database:", err)
		}
	}

	if err := seeder.SeedAll(); err != nil {
		log.Fatal("Failed to seed database:", err)
	}
}
