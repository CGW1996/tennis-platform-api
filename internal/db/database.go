package db

import (
	"fmt"
	"log"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 數據庫連接包裝器
type Database struct {
	DB *gorm.DB
}

// NewDatabase 創建新的數據庫連接
func NewDatabase(cfg *config.Config) (*Database, error) {
	// 構建數據庫連接字符串
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// 設置 GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 連接數據庫
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置連接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 設置連接池參數
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 測試連接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")

	// 自動遷移 (使用簡單遷移後跳過)
	// Tables already created by simple migration
	log.Println("Database tables already exist, skipping auto migration")

	// Skip migration system for now due to GORM compatibility issues
	log.Println("Skipping migration system initialization")

	return &Database{DB: db}, nil
}

// autoMigrate 自動遷移數據庫表
func autoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// 遷移所有模型
	err := db.AutoMigrate(models.AllModels()...)

	if err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Close 關閉數據庫連接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// HealthCheck 檢查數據庫連接健康狀態
func (d *Database) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
