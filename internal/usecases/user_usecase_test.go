package usecases

import (
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 手動創建表結構
	db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		phone TEXT,
		password_hash TEXT NOT NULL,
		email_verified BOOLEAN DEFAULT FALSE,
		phone_verified BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		last_login_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`)

	db.Exec(`CREATE TABLE user_profiles (
		user_id TEXT PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		avatar_url TEXT,
		ntrp_level REAL,
		playing_style TEXT,
		preferred_hand TEXT,
		latitude REAL,
		longitude REAL,
		location_privacy BOOLEAN DEFAULT FALSE,
		bio TEXT,
		birth_date DATE,
		gender TEXT,
		playing_frequency TEXT,
		preferred_times TEXT,
		max_travel_distance REAL,
		profile_privacy TEXT DEFAULT 'public',
		created_at DATETIME,
		updated_at DATETIME
	)`)

	return db
}

func createTestUser(db *gorm.DB) *models.User {
	user := models.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		IsActive:     true,
	}

	profile := models.UserProfile{
		UserID:    user.ID,
		FirstName: "Test",
		LastName:  "User",
	}

	db.Create(&user)
	db.Create(&profile)

	return &user
}

func TestUserUsecase_GetUserByID(t *testing.T) {
	db := setupUserTestDB()
	userUsecase := NewUserUsecase(db)

	// 創建測試用戶
	testUser := createTestUser(db)

	// 測試獲取用戶
	user, err := userUsecase.GetUserByID(testUser.ID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testUser.Email, user.Email)
	assert.NotNil(t, user.Profile)
	assert.Equal(t, "Test", user.Profile.FirstName)
	assert.Equal(t, "User", user.Profile.LastName)
}

func TestUserUsecase_GetUserByID_NotFound(t *testing.T) {
	db := setupUserTestDB()
	userUsecase := NewUserUsecase(db)

	// 測試獲取不存在的用戶
	user, err := userUsecase.GetUserByID("non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "用戶不存在")
}

func TestUserUsecase_UpdateUserProfile(t *testing.T) {
	db := setupUserTestDB()
	userUsecase := NewUserUsecase(db)

	// 創建測試用戶
	testUser := createTestUser(db)

	// 準備更新請求
	newFirstName := "Updated"
	newLastName := "Name"
	newBio := "This is my bio"
	newNTRPLevel := 4.5

	updateReq := &dto.UpdateProfileRequest{
		FirstName: &newFirstName,
		LastName:  &newLastName,
		Bio:       &newBio,
		NTRPLevel: &newNTRPLevel,
	}

	// 測試更新用戶檔案
	user, err := userUsecase.UpdateUserProfile(testUser.ID, updateReq)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, newFirstName, user.Profile.FirstName)
	assert.Equal(t, newLastName, user.Profile.LastName)
	assert.Equal(t, newBio, *user.Profile.Bio)
	assert.Equal(t, newNTRPLevel, *user.Profile.NTRPLevel)
}

func TestUserUsecase_UpdateUserProfile_PartialUpdate(t *testing.T) {
	db := setupUserTestDB()
	userUsecase := NewUserUsecase(db)

	// 創建測試用戶
	testUser := createTestUser(db)

	// 只更新部分字段
	newBio := "Updated bio only"
	updateReq := &dto.UpdateProfileRequest{
		Bio: &newBio,
	}

	// 測試部分更新
	user, err := userUsecase.UpdateUserProfile(testUser.ID, updateReq)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	// 原有字段應該保持不變
	assert.Equal(t, "Test", user.Profile.FirstName)
	assert.Equal(t, "User", user.Profile.LastName)
	// 更新的字段應該改變
	assert.Equal(t, newBio, *user.Profile.Bio)
}

func TestUserUsecase_UpdateUserProfile_UserNotFound(t *testing.T) {
	db := setupUserTestDB()
	userUsecase := NewUserUsecase(db)

	updateReq := &dto.UpdateProfileRequest{
		FirstName: stringPtr("New Name"),
	}

	// 測試更新不存在的用戶
	user, err := userUsecase.UpdateUserProfile("non-existent-id", updateReq)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "用戶不存在")
}

// 輔助函數
func stringPtr(s string) *string {
	return &s
}
