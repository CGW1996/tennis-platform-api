package controllers

import (
	"net/http"
	"strconv"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/usecases"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReputationController 信譽評分控制器
type ReputationController struct {
	reputationUseCase *usecases.ReputationUseCase
}

// NewReputationController 創建新的信譽評分控制器
func NewReputationController(db *gorm.DB) *ReputationController {
	return &ReputationController{
		reputationUseCase: usecases.NewReputationUseCase(db),
	}
}

// GetUserReputationScore 獲取用戶信譽分數
// @Summary 獲取用戶信譽分數
// @Description 獲取指定用戶的信譽分數和統計信息
// @Tags reputation
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Success 200 {object} models.ReputationScore
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/users/{userId}/score [get]
func (rc *ReputationController) GetUserReputationScore(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	reputation, err := rc.reputationUseCase.GetUserReputationScore(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_REPUTATION_FAILED",
			"message": "獲取信譽分數失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reputation)
}

// GetUserReputationHistory 獲取用戶信譽歷史記錄
// @Summary 獲取用戶信譽歷史記錄
// @Description 獲取指定用戶的詳細信譽歷史記錄
// @Tags reputation
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Success 200 {object} models.ReputationHistory
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/users/{userId}/history [get]
func (rc *ReputationController) GetUserReputationHistory(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	history, err := rc.reputationUseCase.GetUserReputationHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_HISTORY_FAILED",
			"message": "獲取信譽歷史失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, history)
}

// RecordMatchAttendance 記錄比賽出席情況
// @Summary 記錄比賽出席情況
// @Description 記錄用戶的比賽出席情況，影響出席率評分
// @Tags reputation
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Param request body dto.RecordMatchAttendanceRequest true "出席記錄請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/users/{userId}/attendance [post]
func (rc *ReputationController) RecordMatchAttendance(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	var req dto.RecordMatchAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "請求參數無效",
			"details": err.Error(),
		})
		return
	}

	err := rc.reputationUseCase.RecordMatchAttendance(userID, req.MatchID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "RECORD_ATTENDANCE_FAILED",
			"message": "記錄出席情況失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "出席記錄已更新",
	})
}

// RecordMatchPunctuality 記錄比賽準時情況
// @Summary 記錄比賽準時情況
// @Description 記錄用戶的比賽準時情況，影響準時度評分
// @Tags reputation
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Param request body dto.RecordMatchPunctualityRequest true "準時記錄請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/users/{userId}/punctuality [post]
func (rc *ReputationController) RecordMatchPunctuality(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	var req dto.RecordMatchPunctualityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "請求參數無效",
			"details": err.Error(),
		})
		return
	}

	err := rc.reputationUseCase.RecordMatchPunctuality(userID, req.MatchID, req.ArrivalTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "RECORD_PUNCTUALITY_FAILED",
			"message": "記錄準時情況失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "準時記錄已更新",
	})
}

// RecordSkillLevelAccuracy 記錄技術等級準確度
// @Summary 記錄技術等級準確度
// @Description 記錄用戶的技術等級準確度，影響技術準確度評分
// @Tags reputation
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Param request body dto.RecordSkillLevelAccuracyRequest true "技術準確度記錄請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/users/{userId}/skill-accuracy [post]
func (rc *ReputationController) RecordSkillLevelAccuracy(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	var req dto.RecordSkillLevelAccuracyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "請求參數無效",
			"details": err.Error(),
		})
		return
	}

	err := rc.reputationUseCase.RecordSkillLevelAccuracy(userID, req.MatchID, req.ReportedLevel, req.ObservedLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "RECORD_SKILL_ACCURACY_FAILED",
			"message": "記錄技術準確度失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "技術準確度記錄已更新",
	})
}

// SubmitBehaviorReview 提交行為評價
// @Summary 提交行為評價
// @Description 對其他用戶提交行為評價，影響其行為評分
// @Tags reputation
// @Accept json
// @Produce json
// @Param userId path string true "被評價用戶ID"
// @Param request body dto.SubmitBehaviorReviewRequest true "行為評價請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/users/{userId}/behavior-review [post]
func (rc *ReputationController) SubmitBehaviorReview(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	// 從JWT中獲取評價者ID
	reviewerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	var req dto.SubmitBehaviorReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "請求參數無效",
			"details": err.Error(),
		})
		return
	}

	err := rc.reputationUseCase.SubmitBehaviorReview(
		reviewerID.(string),
		userID,
		req.MatchID,
		req.Rating,
		req.Comment,
		req.Tags,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "SUBMIT_REVIEW_FAILED",
			"message": "提交評價失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "行為評價已提交",
	})
}

// GetReputationLeaderboard 獲取信譽排行榜
// @Summary 獲取信譽排行榜
// @Description 獲取信譽分數排行榜
// @Tags reputation
// @Accept json
// @Produce json
// @Param limit query int false "返回數量限制" default(50)
// @Success 200 {array} models.ReputationScore
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/leaderboard [get]
func (rc *ReputationController) GetReputationLeaderboard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	leaderboard, err := rc.reputationUseCase.GetReputationLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_LEADERBOARD_FAILED",
			"message": "獲取排行榜失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

// GetReputationStats 獲取信譽統計信息
// @Summary 獲取信譽統計信息
// @Description 獲取平台信譽系統的統計信息
// @Tags reputation
// @Accept json
// @Produce json
// @Success 200 {object} usecases.ReputationStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/stats [get]
func (rc *ReputationController) GetReputationStats(c *gin.Context) {
	stats, err := rc.reputationUseCase.GetReputationStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_STATS_FAILED",
			"message": "獲取統計信息失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateUserNTRPLevel 更新用戶NTRP等級
// @Summary 更新用戶NTRP等級
// @Description 基於信譽系統數據自動調整用戶NTRP等級
// @Tags reputation
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/reputation/users/{userId}/update-ntrp [post]
func (rc *ReputationController) UpdateUserNTRPLevel(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	err := rc.reputationUseCase.UpdateUserNTRPLevel(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "UPDATE_NTRP_FAILED",
			"message": "更新NTRP等級失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "NTRP等級已更新",
	})
}
