package db

import (
	"log"

	"tennis-platform/backend/internal/config"
)

// DatabaseManager 數據庫管理器
type DatabaseManager struct {
	DB    *Database
	Redis *RedisClient
}

// Initialize 初始化數據庫連接和遷移
func Initialize(cfg *config.Config) (*DatabaseManager, error) {
	log.Println("Initializing database connections...")

	// 初始化 PostgreSQL 連接
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	// 初始化 Redis 連接
	redis, err := NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	// 執行數據庫遷移 (暫時跳過)
	// migrationManager := NewMigrationManager(db.DB)
	// if err := migrationManager.RunMigrations(); err != nil {
	// 	return nil, err
	// }
	log.Println("Skipping migration manager due to GORM compatibility issues")

	// 執行種子數據（僅在開發環境）
	if cfg.Env == "development" {
		seeder := NewSeeder(db.DB)
		if err := seeder.SeedAll(); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		}
	}

	log.Println("Database initialization completed successfully")

	return &DatabaseManager{
		DB:    db,
		Redis: redis,
	}, nil
}

// Close 關閉所有數據庫連接
func (dm *DatabaseManager) Close() error {
	log.Println("Closing database connections...")

	if err := dm.DB.Close(); err != nil {
		log.Printf("Error closing PostgreSQL connection: %v", err)
	}

	if err := dm.Redis.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("Database connections closed")
	return nil
}

// HealthCheck 檢查所有數據庫連接的健康狀態
func (dm *DatabaseManager) HealthCheck() map[string]error {
	results := make(map[string]error)

	// 檢查 PostgreSQL
	if err := dm.DB.HealthCheck(); err != nil {
		results["postgresql"] = err
	} else {
		results["postgresql"] = nil
	}

	// 檢查 Redis
	if err := dm.Redis.HealthCheck(); err != nil {
		results["redis"] = err
	} else {
		results["redis"] = nil
	}

	return results
}
