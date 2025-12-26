package main

import (
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

	// 執行種子數據邏輯
	seeder := db.NewSeeder(database.DB.DB)
	if err := seeder.SeedAll(); err != nil {
		log.Fatal("Failed to seed database:", err)
	}
}
