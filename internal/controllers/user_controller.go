package controllers

import (
	"net/http"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// UserUsecaseInterface 用戶用例接口
type UserUsecaseInterface interface {
	GetUserByID(userID string) (*models.User, error)
	CreateUserProfile(userID string, req *dto.CreateProfileRequest) (*models.User, error)
	UpdateUserProfile(userID string, req *dto.UpdateProfileRequest) (*models.User, error)
	UpdateUserPreferences(userID string, req *dto.UserPreferencesRequest) (*models.User, error)
	UpdateUserLocation(userID string, req *dto.LocationUpdateRequest) (*models.User, error)
	UpdateUserAvatar(userID string, avatarURL string) (*models.User, error)
}

// UserController 用戶控制器
type UserController struct {
	userUsecase   UserUsecaseInterface
	uploadService *services.UploadService
}

// NewUserController 創建新的用戶控制器
func NewUserController(userUsecase UserUsecaseInterface, uploadService *services.UploadService) *UserController {
	return &UserController{
		userUsecase:   userUsecase,
		uploadService: uploadService,
	}
}

// GetProfile 獲取用戶檔案
// @Summary 獲取用戶檔案
// @Description 獲取當前用戶的詳細檔案信息
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/profile [get]
func (uc *UserController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	user, err := uc.userUsecase.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile 更新用戶檔案
// @Summary 更新用戶檔案
// @Description 更新當前用戶的檔案信息
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "用戶檔案更新請求"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/profile [put]
func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var profileUpdate dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&profileUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	user, err := uc.userUsecase.UpdateUserProfile(userID.(string), &profileUpdate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateProfile 創建用戶檔案
// @Summary 創建用戶檔案
// @Description 為新用戶創建詳細檔案
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateProfileRequest true "用戶檔案創建請求"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/profile [post]
func (uc *UserController) CreateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var profileCreate dto.CreateProfileRequest
	if err := c.ShouldBindJSON(&profileCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	user, err := uc.userUsecase.CreateUserProfile(userID.(string), &profileCreate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdatePreferences 更新用戶偏好設定
// @Summary 更新用戶偏好設定
// @Description 更新用戶的打球偏好和隱私設定
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UserPreferencesRequest true "用戶偏好設定請求"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/preferences [put]
func (uc *UserController) UpdatePreferences(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var preferencesUpdate dto.UserPreferencesRequest
	if err := c.ShouldBindJSON(&preferencesUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	user, err := uc.userUsecase.UpdateUserPreferences(userID.(string), &preferencesUpdate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateLocation 更新用戶位置
// @Summary 更新用戶位置
// @Description 更新用戶的地理位置信息
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.LocationUpdateRequest true "位置更新請求"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/location [put]
func (uc *UserController) UpdateLocation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var locationUpdate dto.LocationUpdateRequest
	if err := c.ShouldBindJSON(&locationUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	user, err := uc.userUsecase.UpdateUserLocation(userID.(string), &locationUpdate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UploadAvatar 上傳頭像
// @Summary 上傳用戶頭像
// @Description 上傳並更新用戶頭像
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "頭像文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/users/avatar [post]
func (uc *UserController) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	// 獲取上傳的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "未找到上傳文件",
		})
		return
	}

	// 上傳文件
	uploadResult, err := uc.uploadService.UploadAvatar(file, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 更新用戶檔案中的頭像URL
	user, err := uc.userUsecase.UpdateUserAvatar(userID.(string), uploadResult.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "頭像上傳成功",
		"user":    user,
		"upload":  uploadResult,
	})
}

// GetNTRPLevels 獲取 NTRP 等級列表
// @Summary 獲取 NTRP 等級列表
// @Description 獲取所有可用的 NTRP 等級和描述
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/ntrp-levels [get]
func (uc *UserController) GetNTRPLevels(c *gin.Context) {
	levels := []map[string]interface{}{
		{"level": 1.0, "description": "新手：剛開始學習網球"},
		{"level": 1.5, "description": "新手：有限的網球經驗"},
		{"level": 2.0, "description": "初學者：需要在場上練習"},
		{"level": 2.5, "description": "初學者：學習基本技巧"},
		{"level": 3.0, "description": "初級：能夠持續擊球"},
		{"level": 3.5, "description": "初級：有一定的一致性和方向控制"},
		{"level": 4.0, "description": "中級：能夠使用各種擊球技巧"},
		{"level": 4.5, "description": "中級：開始掌握戰術"},
		{"level": 5.0, "description": "高級：良好的擊球技巧和戰術意識"},
		{"level": 5.5, "description": "高級：能夠在比賽中運用各種戰術"},
		{"level": 6.0, "description": "專業級：具備專業水準的技巧"},
		{"level": 6.5, "description": "專業級：接近職業選手水準"},
		{"level": 7.0, "description": "職業級：世界級職業選手"},
	}

	c.JSON(http.StatusOK, gin.H{
		"levels": levels,
	})
}
