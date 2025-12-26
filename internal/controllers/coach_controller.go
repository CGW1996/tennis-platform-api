package controllers

import (
	"net/http"
	"strconv"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// CoachUsecaseInterface 教練用例接口
type CoachUsecaseInterface interface {
	CreateCoachProfile(userID string, req *dto.CreateCoachProfileRequest) (*models.Coach, error)
	GetCoachByID(coachID string) (*models.Coach, error)
	GetCoachByUserID(userID string) (*models.Coach, error)
	UpdateCoachProfile(coachID string, req *dto.UpdateCoachProfileRequest) (*models.Coach, error)
	SearchCoaches(req *dto.CoachSearchRequest) ([]models.Coach, int64, error)
	VerifyCoach(req *dto.CoachVerificationRequest) (*models.Coach, error)
	GetCoachSpecialties() []map[string]interface{}
	GetCoachCertifications() []map[string]interface{}

	// 課程管理相關方法
	CreateLessonType(coachID string, req *dto.CreateLessonTypeRequest) (*models.LessonType, error)
	GetLessonTypes(coachID string) ([]models.LessonType, error)
	UpdateLessonType(lessonTypeID string, req *dto.UpdateLessonTypeRequest) (*models.LessonType, error)
	DeleteLessonType(lessonTypeID string) error

	CreateLesson(req *dto.CreateLessonRequest) (*models.Lesson, error)
	GetLesson(lessonID string) (*models.Lesson, error)
	GetLessons(req *dto.GetLessonsRequest) ([]models.Lesson, int64, error)
	UpdateLesson(lessonID string, req *dto.UpdateLessonRequest) (*models.Lesson, error)
	CancelLesson(lessonID string, req *dto.CancelLessonRequest) (*models.Lesson, error)

	GetCoachAvailability(coachID string, date string) ([]models.TimeSlot, error)
	UpdateCoachSchedule(coachID string, req *dto.UpdateScheduleRequest) error
	GetCoachSchedule(coachID string) ([]models.LessonSchedule, error)

	// 智能排課相關方法
	GetIntelligentRecommendations(req *dto.IntelligentSchedulingRequest) ([]interface{}, error)
	FindOptimalLessonTime(req *dto.OptimalTimeRequest) (interface{}, error)
	DetectSchedulingConflicts(req *dto.ConflictDetectionRequest) ([]models.Lesson, error)
	ResolveSchedulingConflict(req *dto.ConflictResolutionRequest) error
	GetCoachRecommendationFactors(coachID string, studentPrefs *dto.IntelligentSchedulingRequest) (map[string]interface{}, error)

	// 教練評價相關方法
	CreateCoachReview(userID string, req *dto.CreateCoachReviewRequest) (*models.CoachReview, error)
	GetCoachReview(reviewID string) (*models.CoachReview, error)
	GetCoachReviews(req *dto.CoachReviewSearchRequest) ([]models.CoachReview, int64, error)
	UpdateCoachReview(reviewID string, userID string, req *dto.UpdateCoachReviewRequest) (*models.CoachReview, error)
	DeleteCoachReview(reviewID string, userID string) error
	MarkReviewHelpful(userID string, req *dto.MarkReviewHelpfulRequest) (*models.CoachReview, error)
	GetCoachReviewStatistics(coachID string) (map[string]interface{}, error)
	GetAvailableReviewTags() []map[string]interface{}
	CheckCanReviewCoach(userID string, coachID string, lessonID *string) (bool, string, error)
}

// CoachController 教練控制器
type CoachController struct {
	coachUsecase CoachUsecaseInterface
}

// NewCoachController 創建新的教練控制器
func NewCoachController(coachUsecase CoachUsecaseInterface) *CoachController {
	return &CoachController{
		coachUsecase: coachUsecase,
	}
}

// CreateCoachProfile 創建教練檔案
// @Summary 創建教練檔案
// @Description 為用戶創建教練檔案
// @Tags coaches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCoachProfileRequest true "教練檔案創建請求"
// @Success 201 {object} models.Coach
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/coaches [post]
func (cc *CoachController) CreateCoachProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var profileCreate dto.CreateCoachProfileRequest
	if err := c.ShouldBindJSON(&profileCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	coach, err := cc.coachUsecase.CreateCoachProfile(userID.(string), &profileCreate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, coach)
}

// GetCoach 獲取教練詳情
// @Summary 獲取教練詳情
// @Description 根據教練ID獲取教練詳細信息
// @Tags coaches
// @Accept json
// @Produce json
// @Param id path string true "教練ID"
// @Success 200 {object} models.Coach
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/coaches/{id} [get]
func (cc *CoachController) GetCoach(c *gin.Context) {
	coachID := c.Param("id")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	coach, err := cc.coachUsecase.GetCoachByID(coachID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, coach)
}

// GetMyCoachProfile 獲取我的教練檔案
// @Summary 獲取我的教練檔案
// @Description 獲取當前用戶的教練檔案
// @Tags coaches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Coach
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/coaches/my-profile [get]
func (cc *CoachController) GetMyCoachProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	coach, err := cc.coachUsecase.GetCoachByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, coach)
}

// UpdateCoachProfile 更新教練檔案
// @Summary 更新教練檔案
// @Description 更新教練檔案信息
// @Tags coaches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "教練ID"
// @Param request body dto.UpdateCoachProfileRequest true "教練檔案更新請求"
// @Success 200 {object} models.Coach
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/coaches/{id} [put]
func (cc *CoachController) UpdateCoachProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	coachID := c.Param("id")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	// 檢查權限：只有教練本人可以更新自己的檔案
	coach, err := cc.coachUsecase.GetCoachByID(coachID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	if coach.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "無權限更新此教練檔案",
		})
		return
	}

	var profileUpdate dto.UpdateCoachProfileRequest
	if err := c.ShouldBindJSON(&profileUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	updatedCoach, err := cc.coachUsecase.UpdateCoachProfile(coachID, &profileUpdate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedCoach)
}

// SearchCoaches 搜尋教練
// @Summary 搜尋教練
// @Description 根據條件搜尋教練
// @Tags coaches
// @Accept json
// @Produce json
// @Param specialties query []string false "專長" collectionFormat(multi)
// @Param min_experience query int false "最少經驗年數"
// @Param max_experience query int false "最多經驗年數"
// @Param price_min query number false "最低時薪"
// @Param price_max query number false "最高時薪"
// @Param languages query []string false "語言" collectionFormat(multi)
// @Param min_rating query number false "最低評分"
// @Param is_verified query bool false "是否已認證"
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(20)
// @Param sort_by query string false "排序欄位" Enums(rating, experience, hourlyRate, createdAt)
// @Param sort_order query string false "排序順序" Enums(asc, desc)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches [get]
func (cc *CoachController) SearchCoaches(c *gin.Context) {
	var searchReq dto.CoachSearchRequest

	// 綁定查詢參數
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "查詢參數錯誤",
			"details": err.Error(),
		})
		return
	}

	coaches, total, err := cc.coachUsecase.SearchCoaches(&searchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 計算分頁信息
	page := searchReq.Page
	if page < 1 {
		page = 1
	}
	limit := searchReq.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	totalPages := (int(total) + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"coaches": coaches,
		"pagination": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
	})
}

// VerifyCoach 認證教練
// @Summary 認證教練
// @Description 管理員認證教練資格
// @Tags coaches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CoachVerificationRequest true "教練認證請求"
// @Success 200 {object} models.Coach
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/coaches/verify [post]
func (cc *CoachController) VerifyCoach(c *gin.Context) {
	// TODO: 添加管理員權限檢查
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	// 這裡應該檢查用戶是否為管理員
	// 暫時跳過權限檢查，在後續任務中實現
	_ = userID

	var verificationReq dto.CoachVerificationRequest
	if err := c.ShouldBindJSON(&verificationReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	coach, err := cc.coachUsecase.VerifyCoach(&verificationReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, coach)
}

// GetCoachSpecialties 獲取教練專長選項
// @Summary 獲取教練專長選項
// @Description 獲取所有可用的教練專長選項
// @Tags coaches
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches/specialties [get]
func (cc *CoachController) GetCoachSpecialties(c *gin.Context) {
	specialties := cc.coachUsecase.GetCoachSpecialties()
	c.JSON(http.StatusOK, gin.H{
		"specialties": specialties,
	})
}

// GetCoachCertifications 獲取教練認證選項
// @Summary 獲取教練認證選項
// @Description 獲取所有可用的教練認證選項
// @Tags coaches
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches/certifications [get]
func (cc *CoachController) GetCoachCertifications(c *gin.Context) {
	certifications := cc.coachUsecase.GetCoachCertifications()
	c.JSON(http.StatusOK, gin.H{
		"certifications": certifications,
	})
}

// GetCoachStatistics 獲取教練統計信息
// @Summary 獲取教練統計信息
// @Description 獲取教練的課程統計和評價統計
// @Tags coaches
// @Accept json
// @Produce json
// @Param id path string true "教練ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/coaches/{id}/statistics [get]
func (cc *CoachController) GetCoachStatistics(c *gin.Context) {
	coachID := c.Param("id")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	// 檢查教練是否存在
	coach, err := cc.coachUsecase.GetCoachByID(coachID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 返回基本統計信息（從模型中獲取）
	statistics := gin.H{
		"totalLessons":   coach.TotalLessons,
		"totalReviews":   coach.TotalReviews,
		"averageRating":  coach.AverageRating,
		"experience":     coach.Experience,
		"isVerified":     coach.IsVerified,
		"specialties":    coach.Specialties,
		"certifications": coach.Certifications,
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": statistics,
	})
}

// GetAvailableLanguages 獲取可用語言選項
// @Summary 獲取可用語言選項
// @Description 獲取教練可以選擇的語言選項
// @Tags coaches
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches/languages [get]
func (cc *CoachController) GetAvailableLanguages(c *gin.Context) {
	languages := []map[string]interface{}{
		{"value": "zh-TW", "label": "繁體中文"},
		{"value": "zh-CN", "label": "簡體中文"},
		{"value": "en", "label": "English"},
		{"value": "ja", "label": "日本語"},
		{"value": "ko", "label": "한국어"},
		{"value": "es", "label": "Español"},
		{"value": "fr", "label": "Français"},
		{"value": "de", "label": "Deutsch"},
		{"value": "it", "label": "Italiano"},
		{"value": "pt", "label": "Português"},
	}

	c.JSON(http.StatusOK, gin.H{
		"languages": languages,
	})
}

// GetAvailableCurrencies 獲取可用貨幣選項
// @Summary 獲取可用貨幣選項
// @Description 獲取教練可以選擇的貨幣選項
// @Tags coaches
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches/currencies [get]
func (cc *CoachController) GetAvailableCurrencies(c *gin.Context) {
	currencies := []map[string]interface{}{
		{"value": "TWD", "label": "新台幣 (TWD)", "symbol": "NT$"},
		{"value": "USD", "label": "美元 (USD)", "symbol": "$"},
		{"value": "EUR", "label": "歐元 (EUR)", "symbol": "€"},
	}

	c.JSON(http.StatusOK, gin.H{
		"currencies": currencies,
	})
}

// ===== 課程管理相關方法 =====

// CreateLessonType 創建課程類型
// @Summary 創建課程類型
// @Description 教練創建新的課程類型和價格設定
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateLessonTypeRequest true "課程類型創建請求"
// @Success 201 {object} models.LessonType
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/coaches/lesson-types [post]
func (cc *CoachController) CreateLessonType(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	// 獲取教練信息
	coach, err := cc.coachUsecase.GetCoachByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練檔案不存在",
		})
		return
	}

	var req dto.CreateLessonTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	lessonType, err := cc.coachUsecase.CreateLessonType(coach.ID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, lessonType)
}

// GetLessonTypes 獲取課程類型列表
// @Summary 獲取課程類型列表
// @Description 獲取教練的所有課程類型
// @Tags lessons
// @Accept json
// @Produce json
// @Param id path string true "教練ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches/{id}/lesson-types [get]
func (cc *CoachController) GetLessonTypes(c *gin.Context) {
	coachID := c.Param("id")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	lessonTypes, err := cc.coachUsecase.GetLessonTypes(coachID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lessonTypes": lessonTypes,
	})
}

// UpdateLessonType 更新課程類型
// @Summary 更新課程類型
// @Description 更新課程類型信息和價格
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "課程類型ID"
// @Param request body dto.UpdateLessonTypeRequest true "課程類型更新請求"
// @Success 200 {object} models.LessonType
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/lesson-types/{id} [put]
func (cc *CoachController) UpdateLessonType(c *gin.Context) {
	lessonTypeID := c.Param("id")
	if lessonTypeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "課程類型ID不能為空",
		})
		return
	}

	var req dto.UpdateLessonTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	lessonType, err := cc.coachUsecase.UpdateLessonType(lessonTypeID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, lessonType)
}

// DeleteLessonType 刪除課程類型
// @Summary 刪除課程類型
// @Description 刪除課程類型
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "課程類型ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/lesson-types/{id} [delete]
func (cc *CoachController) DeleteLessonType(c *gin.Context) {
	lessonTypeID := c.Param("id")
	if lessonTypeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "課程類型ID不能為空",
		})
		return
	}

	if err := cc.coachUsecase.DeleteLessonType(lessonTypeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateLesson 創建課程
// @Summary 創建課程
// @Description 學生預訂課程
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateLessonRequest true "課程創建請求"
// @Success 201 {object} models.Lesson
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/lessons [post]
func (cc *CoachController) CreateLesson(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var req dto.CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 設置學生ID
	req.StudentID = userID.(string)

	lesson, err := cc.coachUsecase.CreateLesson(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, lesson)
}

// GetLesson 獲取課程詳情
// @Summary 獲取課程詳情
// @Description 獲取課程詳細信息
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "課程ID"
// @Success 200 {object} models.Lesson
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/lessons/{id} [get]
func (cc *CoachController) GetLesson(c *gin.Context) {
	lessonID := c.Param("id")
	if lessonID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "課程ID不能為空",
		})
		return
	}

	lesson, err := cc.coachUsecase.GetLesson(lessonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, lesson)
}

// GetLessons 獲取課程列表
// @Summary 獲取課程列表
// @Description 獲取課程列表（支援篩選）
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param coachId query string false "教練ID"
// @Param studentId query string false "學生ID"
// @Param status query string false "課程狀態"
// @Param startDate query string false "開始日期"
// @Param endDate query string false "結束日期"
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/lessons [get]
func (cc *CoachController) GetLessons(c *gin.Context) {
	var req dto.GetLessonsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "查詢參數錯誤",
			"details": err.Error(),
		})
		return
	}

	lessons, total, err := cc.coachUsecase.GetLessons(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 計算分頁信息
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	totalPages := (int(total) + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"lessons": lessons,
		"pagination": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
	})
}

// UpdateLesson 更新課程
// @Summary 更新課程
// @Description 更新課程信息
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "課程ID"
// @Param request body dto.UpdateLessonRequest true "課程更新請求"
// @Success 200 {object} models.Lesson
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/lessons/{id} [put]
func (cc *CoachController) UpdateLesson(c *gin.Context) {
	lessonID := c.Param("id")
	if lessonID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "課程ID不能為空",
		})
		return
	}

	var req dto.UpdateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	lesson, err := cc.coachUsecase.UpdateLesson(lessonID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, lesson)
}

// CancelLesson 取消課程
// @Summary 取消課程
// @Description 取消課程
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "課程ID"
// @Param request body dto.CancelLessonRequest true "取消課程請求"
// @Success 200 {object} models.Lesson
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/lessons/{id}/cancel [post]
func (cc *CoachController) CancelLesson(c *gin.Context) {
	lessonID := c.Param("id")
	if lessonID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "課程ID不能為空",
		})
		return
	}

	var req dto.CancelLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	lesson, err := cc.coachUsecase.CancelLesson(lessonID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, lesson)
}

// GetCoachAvailability 獲取教練可用時間
// @Summary 獲取教練可用時間
// @Description 獲取教練在指定日期的可用時間段
// @Tags lessons
// @Accept json
// @Produce json
// @Param id path string true "教練ID"
// @Param date query string true "日期 (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches/{id}/availability [get]
func (cc *CoachController) GetCoachAvailability(c *gin.Context) {
	coachID := c.Param("id")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "日期參數不能為空",
		})
		return
	}

	availability, err := cc.coachUsecase.GetCoachAvailability(coachID, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":         date,
		"availability": availability,
	})
}

// UpdateCoachSchedule 更新教練時間表
// @Summary 更新教練時間表
// @Description 更新教練的可用時間表
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateScheduleRequest true "時間表更新請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/coaches/schedule [put]
func (cc *CoachController) UpdateCoachSchedule(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	// 獲取教練信息
	coach, err := cc.coachUsecase.GetCoachByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練檔案不存在",
		})
		return
	}

	var req dto.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	if err := cc.coachUsecase.UpdateCoachSchedule(coach.ID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "時間表更新成功",
	})
}

// GetCoachSchedule 獲取教練時間表
// @Summary 獲取教練時間表
// @Description 獲取教練的時間表設定
// @Tags lessons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "教練ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coaches/{id}/schedule [get]
func (cc *CoachController) GetCoachSchedule(c *gin.Context) {
	coachID := c.Param("id")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	schedule, err := cc.coachUsecase.GetCoachSchedule(coachID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schedule": schedule,
	})
}

// parseIntParam 解析整數參數
func parseIntParam(c *gin.Context, key string) *int {
	if value := c.Query(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return &intValue
		}
	}
	return nil
}

// parseFloatParam 解析浮點數參數
func parseFloatParam(c *gin.Context, key string) *float64 {
	if value := c.Query(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return &floatValue
		}
	}
	return nil
}

// parseBoolParam 解析布林參數
func parseBoolParam(c *gin.Context, key string) *bool {
	if value := c.Query(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return &boolValue
		}
	}
	return nil
}

// ===== 智能排課相關方法 =====

// GetIntelligentRecommendations 獲取智能教練推薦
// @Summary 獲取智能教練推薦
// @Description 根據學生偏好和技術水平推薦合適的教練和課程時間
// @Tags intelligent-scheduling
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.IntelligentSchedulingRequest true "智能排課請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/intelligent-scheduling/recommendations [post]
func (cc *CoachController) GetIntelligentRecommendations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var req dto.IntelligentSchedulingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 設置學生ID為當前用戶
	req.StudentID = userID.(string)

	recommendations, err := cc.coachUsecase.GetIntelligentRecommendations(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"total":           len(recommendations),
	})
}

// FindOptimalLessonTime 尋找最佳課程時間
// @Summary 尋找最佳課程時間
// @Description 為特定教練和學生尋找最佳的課程時間
// @Tags intelligent-scheduling
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.OptimalTimeRequest true "最佳時間查詢請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/intelligent-scheduling/optimal-time [post]
func (cc *CoachController) FindOptimalLessonTime(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var req dto.OptimalTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 設置學生ID為當前用戶
	req.StudentID = userID.(string)

	recommendation, err := cc.coachUsecase.FindOptimalLessonTime(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendation": recommendation,
	})
}

// DetectSchedulingConflicts 檢測排課衝突
// @Summary 檢測排課衝突
// @Description 檢測指定時間是否與現有課程衝突
// @Tags intelligent-scheduling
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ConflictDetectionRequest true "衝突檢測請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/intelligent-scheduling/detect-conflicts [post]
func (cc *CoachController) DetectSchedulingConflicts(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var req dto.ConflictDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 檢查權限：只有教練本人可以檢測自己的衝突
	coach, err := cc.coachUsecase.GetCoachByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "只有教練可以檢測排課衝突",
		})
		return
	}

	if coach.ID != req.CoachID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "無權限檢測其他教練的排課衝突",
		})
		return
	}

	conflicts, err := cc.coachUsecase.DetectSchedulingConflicts(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conflicts":     conflicts,
		"hasConflicts":  len(conflicts) > 0,
		"conflictCount": len(conflicts),
	})
}

// ResolveSchedulingConflict 解決排課衝突
// @Summary 解決排課衝突
// @Description 通過調整課程時間來解決排課衝突
// @Tags intelligent-scheduling
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ConflictResolutionRequest true "衝突解決請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/intelligent-scheduling/resolve-conflict [post]
func (cc *CoachController) ResolveSchedulingConflict(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var req dto.ConflictResolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 檢查權限：只有相關的教練或學生可以解決衝突
	lesson, err := cc.coachUsecase.GetLesson(req.ConflictingLessonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "課程不存在",
		})
		return
	}

	// 檢查是否為教練本人或學生本人
	isCoach := false
	if coach, err := cc.coachUsecase.GetCoachByUserID(userID.(string)); err == nil {
		isCoach = coach.ID == lesson.CoachID
	}
	isStudent := lesson.StudentID == userID.(string)

	if !isCoach && !isStudent {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "無權限解決此課程的衝突",
		})
		return
	}

	if err := cc.coachUsecase.ResolveSchedulingConflict(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "排課衝突已解決",
	})
}

// GetCoachRecommendationFactors 獲取教練推薦因子
// @Summary 獲取教練推薦因子
// @Description 獲取教練推薦的詳細因子信息（用於調試和優化）
// @Tags intelligent-scheduling
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param coachId path string true "教練ID"
// @Param request body dto.IntelligentSchedulingRequest true "學生偏好"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/intelligent-scheduling/coaches/{coachId}/factors [post]
func (cc *CoachController) GetCoachRecommendationFactors(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	coachID := c.Param("coachId")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	var req dto.IntelligentSchedulingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 設置學生ID為當前用戶
	req.StudentID = userID.(string)

	factors, err := cc.coachUsecase.GetCoachRecommendationFactors(coachID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"factors": factors,
	})
}

// GetIntelligentSchedulingOptions 獲取智能排課選項
// @Summary 獲取智能排課選項
// @Description 獲取智能排課的可用選項和配置
// @Tags intelligent-scheduling
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/intelligent-scheduling/options [get]
func (cc *CoachController) GetIntelligentSchedulingOptions(c *gin.Context) {
	options := gin.H{
		"lessonTypes": []map[string]interface{}{
			{"value": "individual", "label": "個人課程", "description": "一對一教學"},
			{"value": "group", "label": "團體課程", "description": "小組教學"},
			{"value": "clinic", "label": "訓練營", "description": "大型團體訓練"},
		},
		"skillLevels": []map[string]interface{}{
			{"value": "beginner", "label": "初學者", "nterpRange": "1.0-2.5"},
			{"value": "intermediate", "label": "中級", "nterpRange": "2.5-4.0"},
			{"value": "advanced", "label": "高級", "nterpRange": "4.0-7.0"},
		},
		"timeSlots": []map[string]interface{}{
			{"value": "09:00-12:00", "label": "上午 (09:00-12:00)"},
			{"value": "12:00-14:00", "label": "中午 (12:00-14:00)"},
			{"value": "14:00-18:00", "label": "下午 (14:00-18:00)"},
			{"value": "18:00-21:00", "label": "晚上 (18:00-21:00)"},
		},
		"daysOfWeek": []map[string]interface{}{
			{"value": 0, "label": "星期日"},
			{"value": 1, "label": "星期一"},
			{"value": 2, "label": "星期二"},
			{"value": 3, "label": "星期三"},
			{"value": 4, "label": "星期四"},
			{"value": 5, "label": "星期五"},
			{"value": 6, "label": "星期六"},
		},
		"maxDistance": gin.H{
			"min":     0,
			"max":     50,
			"default": 10,
			"unit":    "公里",
		},
		"priceRange": gin.H{
			"min":      0,
			"max":      5000,
			"default":  1500,
			"currency": "TWD",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"options": options,
	})
}

// ===== 教練評價系統相關方法 =====

// CreateCoachReview 創建教練評價
// @Summary 創建教練評價
// @Description 學生對教練進行評價
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCoachReviewRequest true "評價創建請求"
// @Success 201 {object} models.CoachReview
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/coach-reviews [post]
func (cc *CoachController) CreateCoachReview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var req dto.CreateCoachReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	review, err := cc.coachUsecase.CreateCoachReview(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GetCoachReview 獲取教練評價詳情
// @Summary 獲取教練評價詳情
// @Description 獲取特定教練評價的詳細信息
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Param id path string true "評價ID"
// @Success 200 {object} models.CoachReview
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/coach-reviews/{id} [get]
func (cc *CoachController) GetCoachReview(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	review, err := cc.coachUsecase.GetCoachReview(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetCoachReviews 獲取教練評價列表
// @Summary 獲取教練評價列表
// @Description 獲取特定教練的所有評價
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Param coachId query string true "教練ID"
// @Param rating query int false "評分篩選"
// @Param hasComment query bool false "是否有評論"
// @Param tags query []string false "標籤篩選" collectionFormat(multi)
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(20)
// @Param sortBy query string false "排序欄位" Enums(rating, createdAt, isHelpful)
// @Param sortOrder query string false "排序順序" Enums(asc, desc)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coach-reviews [get]
func (cc *CoachController) GetCoachReviews(c *gin.Context) {
	var req dto.CoachReviewSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "查詢參數錯誤",
			"details": err.Error(),
		})
		return
	}

	reviews, total, err := cc.coachUsecase.GetCoachReviews(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 計算分頁信息
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	totalPages := (int(total) + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"pagination": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
	})
}

// UpdateCoachReview 更新教練評價
// @Summary 更新教練評價
// @Description 更新自己的教練評價
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "評價ID"
// @Param request body dto.UpdateCoachReviewRequest true "評價更新請求"
// @Success 200 {object} models.CoachReview
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/coach-reviews/{id} [put]
func (cc *CoachController) UpdateCoachReview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	var req dto.UpdateCoachReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	review, err := cc.coachUsecase.UpdateCoachReview(reviewID, userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// DeleteCoachReview 刪除教練評價
// @Summary 刪除教練評價
// @Description 刪除自己的教練評價
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "評價ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/v1/coach-reviews/{id} [delete]
func (cc *CoachController) DeleteCoachReview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	if err := cc.coachUsecase.DeleteCoachReview(reviewID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkReviewHelpful 標記評價有用
// @Summary 標記評價有用
// @Description 標記或取消標記評價為有用
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.MarkReviewHelpfulRequest true "標記有用請求"
// @Success 200 {object} models.CoachReview
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/coach-reviews/mark-helpful [post]
func (cc *CoachController) MarkReviewHelpful(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	var req dto.MarkReviewHelpfulRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	review, err := cc.coachUsecase.MarkReviewHelpful(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetCoachReviewStatistics 獲取教練評價統計
// @Summary 獲取教練評價統計
// @Description 獲取教練的評價統計信息
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Param id path string true "教練ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/coaches/{id}/review-statistics [get]
func (cc *CoachController) GetCoachReviewStatistics(c *gin.Context) {
	coachID := c.Param("id")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	statistics, err := cc.coachUsecase.GetCoachReviewStatistics(coachID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": statistics,
	})
}

// GetAvailableReviewTags 獲取可用的評價標籤
// @Summary 獲取可用的評價標籤
// @Description 獲取所有可用的評價標籤選項
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/coach-reviews/available-tags [get]
func (cc *CoachController) GetAvailableReviewTags(c *gin.Context) {
	tags := cc.coachUsecase.GetAvailableReviewTags()
	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}

// CheckCanReviewCoach 檢查是否可以評價教練
// @Summary 檢查是否可以評價教練
// @Description 檢查當前用戶是否可以評價指定教練
// @Tags coach-reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param coachId query string true "教練ID"
// @Param lessonId query string false "課程ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/coach-reviews/can-review [get]
func (cc *CoachController) CheckCanReviewCoach(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未找到用戶信息",
		})
		return
	}

	coachID := c.Query("coachId")
	if coachID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "教練ID不能為空",
		})
		return
	}

	var lessonID *string
	if lessonIDStr := c.Query("lessonId"); lessonIDStr != "" {
		lessonID = &lessonIDStr
	}

	canReview, message, err := cc.coachUsecase.CheckCanReviewCoach(userID.(string), coachID, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"canReview": canReview,
		"message":   message,
	})
}
