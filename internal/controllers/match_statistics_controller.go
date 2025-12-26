package controllers

import (
	"net/http"
	"strconv"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/usecases"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MatchStatisticsController 配對統計控制器
type MatchStatisticsController struct {
	matchStatisticsUseCase *usecases.MatchStatisticsUseCase
}

// NewMatchStatisticsController 創建新的配對統計控制器
func NewMatchStatisticsController(db *gorm.DB) *MatchStatisticsController {
	return &MatchStatisticsController{
		matchStatisticsUseCase: usecases.NewMatchStatisticsUseCase(db),
	}
}

// GetUserMatchStatistics 獲取用戶配對統計資訊
// @Summary 獲取用戶配對統計資訊
// @Description 獲取指定用戶的詳細配對統計資訊
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Success 200 {object} models.MatchStatistics
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/users/{userId} [get]
func (msc *MatchStatisticsController) GetUserMatchStatistics(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	// 獲取請求者ID
	requestingUserID, exists := c.Get("userID")
	if !exists {
		requestingUserID = ""
	}

	stats, err := msc.matchStatisticsUseCase.GetUserMatchStatistics(userID, requestingUserID.(string))
	if err != nil {
		if err.Error() == "match statistics are private" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "PRIVATE_STATISTICS",
				"message": "用戶統計資訊為私人設定",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_STATISTICS_FAILED",
			"message": "獲取統計資訊失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserMatchHistory 獲取用戶配對歷史
// @Summary 獲取用戶配對歷史
// @Description 獲取指定用戶的配對歷史記錄
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Param limit query int false "返回數量限制" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} models.Match
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/users/{userId}/history [get]
func (msc *MatchStatisticsController) GetUserMatchHistory(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	// 解析查詢參數
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 獲取請求者ID
	requestingUserID, exists := c.Get("userID")
	if !exists {
		requestingUserID = ""
	}

	matches, err := msc.matchStatisticsUseCase.GetUserMatchHistory(userID, requestingUserID.(string), limit, offset)
	if err != nil {
		if err.Error() == "match history is private" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "PRIVATE_HISTORY",
				"message": "用戶配對歷史為私人設定",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_HISTORY_FAILED",
			"message": "獲取配對歷史失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, matches)
}

// RecordMatchResult 記錄比賽結果
// @Summary 記錄比賽結果
// @Description 記錄比賽的勝負結果和比分
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param matchId path string true "比賽ID"
// @Param request body dto.RecordMatchResultRequest true "比賽結果請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/matches/{matchId}/result [post]
func (msc *MatchStatisticsController) RecordMatchResult(c *gin.Context) {
	matchID := c.Param("matchId")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_MATCH_ID",
			"message": "比賽ID不能為空",
		})
		return
	}

	// 獲取記錄者ID
	recordedBy, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	var req dto.RecordMatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "請求參數無效",
			"details": err.Error(),
		})
		return
	}

	err := msc.matchStatisticsUseCase.RecordMatchResult(matchID, req.WinnerID, req.LoserID, req.Score, recordedBy.(string))
	if err != nil {
		if err.Error() == "only match participants can record results" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "NOT_PARTICIPANT",
				"message": "只有比賽參與者可以記錄結果",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "RECORD_RESULT_FAILED",
			"message": "記錄比賽結果失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "比賽結果已記錄",
	})
}

// ConfirmMatchResult 確認比賽結果
// @Summary 確認比賽結果
// @Description 確認比賽結果的準確性
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param resultId path string true "比賽結果ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/results/{resultId}/confirm [post]
func (msc *MatchStatisticsController) ConfirmMatchResult(c *gin.Context) {
	resultID := c.Param("resultId")
	if resultID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_RESULT_ID",
			"message": "結果ID不能為空",
		})
		return
	}

	// 獲取確認者ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	err := msc.matchStatisticsUseCase.ConfirmMatchResult(resultID, userID.(string))
	if err != nil {
		if err.Error() == "user not authorized to confirm this result" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "NOT_AUTHORIZED",
				"message": "無權限確認此結果",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "CONFIRM_RESULT_FAILED",
			"message": "確認比賽結果失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "比賽結果已確認",
	})
}

// GetSkillLevelProgression 獲取技術等級進展
// @Summary 獲取技術等級進展
// @Description 獲取用戶的技術等級變化歷史
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Success 200 {array} models.SkillLevelRecord
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/users/{userId}/skill-progression [get]
func (msc *MatchStatisticsController) GetSkillLevelProgression(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	// 獲取請求者ID
	requestingUserID, exists := c.Get("userID")
	if !exists {
		requestingUserID = ""
	}

	skillRecords, err := msc.matchStatisticsUseCase.GetSkillLevelProgression(userID, requestingUserID.(string))
	if err != nil {
		if err.Error() == "skill progression is private" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "PRIVATE_PROGRESSION",
				"message": "技術等級進展為私人設定",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_PROGRESSION_FAILED",
			"message": "獲取技術等級進展失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, skillRecords)
}

// ManuallyAdjustSkillLevel 手動調整技術等級
// @Summary 手動調整技術等級
// @Description 手動調整用戶的NTRP技術等級
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Param request body dto.ManuallyAdjustSkillLevelRequest true "調整等級請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/users/{userId}/adjust-skill-level [post]
func (msc *MatchStatisticsController) ManuallyAdjustSkillLevel(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	// 獲取調整者ID
	adjustedBy, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	var req dto.ManuallyAdjustSkillLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "請求參數無效",
			"details": err.Error(),
		})
		return
	}

	err := msc.matchStatisticsUseCase.ManuallyAdjustSkillLevel(userID, req.NewLevel, req.Reason, adjustedBy.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ADJUST_LEVEL_FAILED",
			"message": "調整技術等級失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "技術等級已調整",
	})
}

// GetUserPrivacySettings 獲取用戶隱私設定
// @Summary 獲取用戶隱私設定
// @Description 獲取用戶的統計資訊隱私設定
// @Tags match-statistics
// @Accept json
// @Produce json
// @Success 200 {object} models.UserPrivacySettings
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/privacy-settings [get]
func (msc *MatchStatisticsController) GetUserPrivacySettings(c *gin.Context) {
	// 獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	settings, err := msc.matchStatisticsUseCase.GetUserPrivacySettings(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_SETTINGS_FAILED",
			"message": "獲取隱私設定失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateUserPrivacySettings 更新用戶隱私設定
// @Summary 更新用戶隱私設定
// @Description 更新用戶的統計資訊隱私設定
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param request body models.UserPrivacySettings true "隱私設定"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/privacy-settings [put]
func (msc *MatchStatisticsController) UpdateUserPrivacySettings(c *gin.Context) {
	// 獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	var settings models.UserPrivacySettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "請求參數無效",
			"details": err.Error(),
		})
		return
	}

	err := msc.matchStatisticsUseCase.UpdateUserPrivacySettings(userID.(string), &settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "UPDATE_SETTINGS_FAILED",
			"message": "更新隱私設定失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "隱私設定已更新",
	})
}

// GetMatchResultsForConfirmation 獲取待確認的比賽結果
// @Summary 獲取待確認的比賽結果
// @Description 獲取用戶需要確認的比賽結果列表
// @Tags match-statistics
// @Accept json
// @Produce json
// @Success 200 {array} models.MatchResult
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/pending-confirmations [get]
func (msc *MatchStatisticsController) GetMatchResultsForConfirmation(c *gin.Context) {
	// 獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	matchResults, err := msc.matchStatisticsUseCase.GetMatchResultsForConfirmation(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_CONFIRMATIONS_FAILED",
			"message": "獲取待確認結果失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, matchResults)
}

// GetReputationScoreWithPrivacy 根據隱私設定獲取信譽分數
// @Summary 根據隱私設定獲取信譽分數
// @Description 根據用戶隱私設定獲取信譽分數
// @Tags match-statistics
// @Accept json
// @Produce json
// @Param userId path string true "用戶ID"
// @Success 200 {object} models.ReputationScore
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/users/{userId}/reputation [get]
func (msc *MatchStatisticsController) GetReputationScoreWithPrivacy(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_USER_ID",
			"message": "用戶ID不能為空",
		})
		return
	}

	// 獲取請求者ID
	requestingUserID, exists := c.Get("userID")
	if !exists {
		requestingUserID = ""
	}

	reputation, err := msc.matchStatisticsUseCase.GetReputationScoreWithPrivacy(userID, requestingUserID.(string))
	if err != nil {
		if err.Error() == "reputation score is private" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "PRIVATE_REPUTATION",
				"message": "信譽分數為私人設定",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_REPUTATION_FAILED",
			"message": "獲取信譽分數失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reputation)
}

// GetMatchStatisticsSummary 獲取配對統計摘要
// @Summary 獲取配對統計摘要
// @Description 獲取用戶配對統計的簡要摘要
// @Tags match-statistics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/match-statistics/summary [get]
func (msc *MatchStatisticsController) GetMatchStatisticsSummary(c *gin.Context) {
	// 獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "UNAUTHORIZED",
			"message": "未授權的請求",
		})
		return
	}

	summary, err := msc.matchStatisticsUseCase.GetMatchStatisticsSummary(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "GET_SUMMARY_FAILED",
			"message": "獲取統計摘要失敗",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}
