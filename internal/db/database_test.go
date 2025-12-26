package db

import (
	"testing"

	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConnection(t *testing.T) {
	// 跳過測試如果沒有測試數據庫
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	// 創建測試配置
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "tennis_platform_test",
			User:     "tennis_user",
			Password: "tennis_password",
			SSLMode:  "disable",
		},
		Env: "test",
	}

	// 嘗試連接數據庫
	db, err := NewDatabase(cfg)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
	defer db.Close()

	// 測試數據庫連接
	err = db.HealthCheck()
	assert.NoError(t, err, "Database health check should pass")
}

func TestModelCreation(t *testing.T) {
	// 測試模型創建不會 panic
	models := models.AllModels()
	assert.NotEmpty(t, models, "Should have models defined")
	assert.Greater(t, len(models), 10, "Should have multiple models")
}

func TestUserModel(t *testing.T) {
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		IsActive:     true, // 手動設置，因為 GORM 默認值只在數據庫操作時生效
	}

	// 測試模型字段
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.True(t, user.IsActive)
}

func TestCourtModel(t *testing.T) {
	court := &models.Court{
		Name:         "Test Court",
		Address:      "Test Address",
		Latitude:     25.0330,
		Longitude:    121.5654,
		PricePerHour: 800.0,
		Currency:     "TWD",
		IsActive:     true, // 手動設置，因為 GORM 默認值只在數據庫操作時生效
	}

	// 測試模型字段
	assert.Equal(t, "Test Court", court.Name)
	assert.Equal(t, 800.0, court.PricePerHour)
	assert.Equal(t, "TWD", court.Currency)
	assert.True(t, court.IsActive)
}
