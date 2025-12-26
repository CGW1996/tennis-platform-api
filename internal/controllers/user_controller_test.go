package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockUserUsecase 模擬用戶用例
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) GetUserByID(userID string) (*models.User, error) {
	args := m.Called(userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserUsecase) CreateUserProfile(userID string, req *dto.CreateProfileRequest) (*models.User, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserProfile(userID string, req *dto.UpdateProfileRequest) (*models.User, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserPreferences(userID string, req *dto.UserPreferencesRequest) (*models.User, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserLocation(userID string, req *dto.LocationUpdateRequest) (*models.User, error) {
	args := m.Called(userID, req)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserAvatar(userID string, avatarURL string) (*models.User, error) {
	args := m.Called(userID, avatarURL)
	return args.Get(0).(*models.User), args.Error(1)
}

// setupTestDB 設置測試數據庫
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自動遷移
	db.AutoMigrate(&models.User{}, &models.UserProfile{})

	return db
}

// setupTestController 設置測試控制器
func setupTestController() (*UserController, *MockUserUsecase) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxFileSize: 10 * 1024 * 1024, // 10MB
			AllowedExts: "jpg,jpeg,png,gif",
			UploadPath:  "./test_uploads",
		},
	}

	mockUsecase := new(MockUserUsecase)
	uploadService := services.NewUploadService(cfg)
	controller := NewUserController(mockUsecase, uploadService)

	return controller, mockUsecase
}

// TestGetProfile 測試獲取用戶檔案
func TestGetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := setupTestController()

	// 準備測試數據
	userID := "test-user-id"
	expectedUser := &models.User{
		ID:    userID,
		Email: "test@example.com",
		Profile: &models.UserProfile{
			UserID:    userID,
			FirstName: "Test",
			LastName:  "User",
			NTRPLevel: func() *float64 { v := 3.5; return &v }(),
		},
	}

	mockUsecase.On("GetUserByID", userID).Return(expectedUser, nil)

	// 創建請求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", userID)

	// 執行測試
	controller.GetProfile(c)

	// 驗證結果
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)
	assert.Equal(t, expectedUser.Email, response.Email)

	mockUsecase.AssertExpectations(t)
}

// TestCreateProfile 測試創建用戶檔案
func TestCreateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := setupTestController()

	userID := "test-user-id"
	createReq := dto.CreateProfileRequest{
		FirstName: "Test",
		LastName:  "User",
		NTRPLevel: func() *float64 { v := 3.5; return &v }(),
	}

	expectedUser := &models.User{
		ID:    userID,
		Email: "test@example.com",
		Profile: &models.UserProfile{
			UserID:    userID,
			FirstName: createReq.FirstName,
			LastName:  createReq.LastName,
			NTRPLevel: createReq.NTRPLevel,
		},
	}

	mockUsecase.On("CreateUserProfile", userID, &createReq).Return(expectedUser, nil)

	// 創建請求
	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/users/profile", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", userID)

	// 執行測試
	controller.CreateProfile(c)

	// 驗證結果
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, response.ID)

	mockUsecase.AssertExpectations(t)
}

// TestUpdateProfile 測試更新用戶檔案
func TestUpdateProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := setupTestController()

	userID := "test-user-id"
	updateReq := dto.UpdateProfileRequest{
		FirstName: func() *string { v := "Updated"; return &v }(),
		NTRPLevel: func() *float64 { v := 4.0; return &v }(),
	}

	expectedUser := &models.User{
		ID:    userID,
		Email: "test@example.com",
		Profile: &models.UserProfile{
			UserID:    userID,
			FirstName: *updateReq.FirstName,
			LastName:  "User",
			NTRPLevel: updateReq.NTRPLevel,
		},
	}

	mockUsecase.On("UpdateUserProfile", userID, &updateReq).Return(expectedUser, nil)

	// 創建請求
	reqBody, _ := json.Marshal(updateReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/users/profile", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", userID)

	// 執行測試
	controller.UpdateProfile(c)

	// 驗證結果
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.Profile.FirstName, response.Profile.FirstName)

	mockUsecase.AssertExpectations(t)
}

// TestUpdatePreferences 測試更新用戶偏好
func TestUpdatePreferences(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := setupTestController()

	userID := "test-user-id"
	prefsReq := dto.UserPreferencesRequest{
		PlayingStyle:     func() *string { v := "aggressive"; return &v }(),
		PlayingFrequency: func() *string { v := "regular"; return &v }(),
		PreferredTimes:   []string{"morning", "evening"},
	}

	expectedUser := &models.User{
		ID:    userID,
		Email: "test@example.com",
		Profile: &models.UserProfile{
			UserID:           userID,
			FirstName:        "Test",
			LastName:         "User",
			PlayingStyle:     prefsReq.PlayingStyle,
			PlayingFrequency: prefsReq.PlayingFrequency,
			PreferredTimes:   prefsReq.PreferredTimes,
		},
	}

	mockUsecase.On("UpdateUserPreferences", userID, &prefsReq).Return(expectedUser, nil)

	// 創建請求
	reqBody, _ := json.Marshal(prefsReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/users/preferences", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", userID)

	// 執行測試
	controller.UpdatePreferences(c)

	// 驗證結果
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, *expectedUser.Profile.PlayingStyle, *response.Profile.PlayingStyle)

	mockUsecase.AssertExpectations(t)
}

// TestUpdateLocation 測試更新用戶位置
func TestUpdateLocation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, mockUsecase := setupTestController()

	userID := "test-user-id"
	locationReq := dto.LocationUpdateRequest{
		Latitude:        func() *float64 { v := 25.0330; return &v }(),
		Longitude:       func() *float64 { v := 121.5654; return &v }(),
		LocationPrivacy: func() *bool { v := true; return &v }(),
	}

	expectedUser := &models.User{
		ID:    userID,
		Email: "test@example.com",
		Profile: &models.UserProfile{
			UserID:          userID,
			FirstName:       "Test",
			LastName:        "User",
			Latitude:        locationReq.Latitude,
			Longitude:       locationReq.Longitude,
			LocationPrivacy: *locationReq.LocationPrivacy,
		},
	}

	mockUsecase.On("UpdateUserLocation", userID, &locationReq).Return(expectedUser, nil)

	// 創建請求
	reqBody, _ := json.Marshal(locationReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/users/location", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", userID)

	// 執行測試
	controller.UpdateLocation(c)

	// 驗證結果
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, *expectedUser.Profile.Latitude, *response.Profile.Latitude)
	assert.Equal(t, expectedUser.Profile.LocationPrivacy, response.Profile.LocationPrivacy)

	mockUsecase.AssertExpectations(t)
}

// TestGetNTRPLevels 測試獲取 NTRP 等級列表
func TestGetNTRPLevels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, _ := setupTestController()

	// 創建請求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 執行測試
	controller.GetNTRPLevels(c)

	// 驗證結果
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	levels, exists := response["levels"]
	assert.True(t, exists)

	levelsArray, ok := levels.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 13, len(levelsArray)) // 1.0 到 7.0，共 13 個等級
}

// TestUploadAvatar 測試上傳頭像（簡化版本）
func TestUploadAvatar_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, _ := setupTestController()

	userID := "test-user-id"

	// 創建沒有文件的請求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/users/avatar", nil)
	c.Set("userID", userID)

	// 執行測試
	controller.UploadAvatar(c)

	// 驗證結果
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "未找到上傳文件")
}

// TestValidationErrors 測試驗證錯誤
func TestCreateProfile_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller, _ := setupTestController()

	userID := "test-user-id"

	// 創建無效的請求（缺少必填欄位）
	invalidReq := map[string]interface{}{
		"firstName": "", // 空字符串應該失敗
	}

	reqBody, _ := json.Marshal(invalidReq)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/users/profile", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", userID)

	// 執行測試
	controller.CreateProfile(c)

	// 驗證結果
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "請求參數錯誤")
}
