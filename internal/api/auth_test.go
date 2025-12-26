package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/controllers"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"tennis-platform/backend/internal/usecases"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestServer() (*Server, *gorm.DB) {
	// 設置測試模式
	gin.SetMode(gin.TestMode)

	// 創建內存數據庫
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

	// 創建測試配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-secret",
			AccessTokenTTL:  15,
			RefreshTokenTTL: 7,
		},
		Env: "test",
	}

	// 初始化服務層
	jwtService := services.NewJWTService(cfg)

	// 初始化用例層
	authUsecase := usecases.NewAuthUsecase(db, cfg)
	userUsecase := usecases.NewUserUsecase(db)

	// 初始化控制器層
	authController := controllers.NewAuthController(authUsecase)
	uploadService := services.NewUploadService(cfg)
	userController := controllers.NewUserController(userUsecase, uploadService)

	// 創建服務器
	server := &Server{
		config:         cfg,
		router:         gin.New(),
		jwtService:     jwtService,
		authController: authController,
		userController: userController,
	}

	server.setupRoutes()

	return server, db
}

func TestRegisterAPI(t *testing.T) {
	server, _ := setupTestServer()

	registerReq := dto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	jsonData, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, registerReq.Email, response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
}

func TestLoginAPI(t *testing.T) {
	server, _ := setupTestServer()

	// 先註冊用戶
	registerReq := dto.RegisterRequest{
		Email:     "login@example.com",
		Password:  "password123",
		FirstName: "Login",
		LastName:  "User",
	}

	jsonData, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// 測試登入
	loginReq := dto.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}

	jsonData, _ = json.Marshal(loginReq)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, loginReq.Email, response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
}

func TestGetUserProfileAPI(t *testing.T) {
	server, _ := setupTestServer()

	// 先註冊用戶
	registerReq := dto.RegisterRequest{
		Email:     "profile@example.com",
		Password:  "password123",
		FirstName: "Profile",
		LastName:  "User",
	}

	jsonData, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse dto.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	assert.NoError(t, err)

	// 測試獲取用戶檔案
	req, _ = http.NewRequest("GET", "/api/v1/users/profile", nil)
	req.Header.Set("Authorization", "Bearer "+registerResponse.AccessToken)

	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user models.User
	err = json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, registerReq.Email, user.Email)
	assert.Equal(t, registerReq.FirstName, user.Profile.FirstName)
	assert.Equal(t, registerReq.LastName, user.Profile.LastName)
}

func TestUnauthorizedAccess(t *testing.T) {
	server, _ := setupTestServer()

	// 測試未授權訪問受保護的端點
	req, _ := http.NewRequest("GET", "/api/v1/users/profile", nil)

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInvalidToken(t *testing.T) {
	server, _ := setupTestServer()

	// 測試無效令牌
	req, _ := http.NewRequest("GET", "/api/v1/users/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
