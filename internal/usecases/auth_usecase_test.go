package usecases

import (
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/dto"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
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

	db.Exec(`CREATE TABLE refresh_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		is_revoked BOOLEAN DEFAULT FALSE,
		created_at DATETIME
	)`)

	db.Exec(`CREATE TABLE oauth_accounts (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		provider TEXT NOT NULL,
		provider_id TEXT NOT NULL,
		email TEXT,
		access_token TEXT,
		refresh_token TEXT,
		expires_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	return db
}

func setupTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-secret",
			AccessTokenTTL:  15,
			RefreshTokenTTL: 7,
		},
		Env: "test",
	}
}

func TestAuthUsecase_Register(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	authUsecase := NewAuthUsecase(db, cfg)

	req := &dto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	response, err := authUsecase.Register(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Email, response.User.Email)
	assert.Equal(t, req.FirstName, response.User.Profile.FirstName)
	assert.Equal(t, req.LastName, response.User.Profile.LastName)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)

	// 驗證密碼已加密
	err = bcrypt.CompareHashAndPassword([]byte(response.User.PasswordHash), []byte(req.Password))
	assert.NoError(t, err)
}

func TestAuthUsecase_Login(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	authUsecase := NewAuthUsecase(db, cfg)

	// 先註冊用戶
	registerReq := &dto.RegisterRequest{
		Email:     "login@example.com",
		Password:  "password123",
		FirstName: "Login",
		LastName:  "User",
	}
	_, err := authUsecase.Register(registerReq)
	assert.NoError(t, err)

	// 測試登入
	loginReq := &dto.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}

	response, err := authUsecase.Login(loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, loginReq.Email, response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
}

func TestAuthUsecase_Login_InvalidCredentials(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	authUsecase := NewAuthUsecase(db, cfg)

	loginReq := &dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}

	response, err := authUsecase.Login(loginReq)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "用戶不存在或密碼錯誤")
}

func TestAuthUsecase_RefreshToken(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	authUsecase := NewAuthUsecase(db, cfg)

	// 先註冊用戶
	registerReq := &dto.RegisterRequest{
		Email:     "refresh@example.com",
		Password:  "password123",
		FirstName: "Refresh",
		LastName:  "User",
	}
	registerResponse, err := authUsecase.Register(registerReq)
	assert.NoError(t, err)

	// 測試刷新令牌
	refreshReq := &dto.RefreshTokenRequest{
		RefreshToken: registerResponse.RefreshToken,
	}

	response, err := authUsecase.RefreshToken(refreshReq)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, registerReq.Email, response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	// 新的刷新令牌應該與舊的不同
	assert.NotEqual(t, registerResponse.RefreshToken, response.RefreshToken)
}
