package controllers

import (
	"net/http"
	"strconv"
	"time"

	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"tennis-platform/backend/internal/usecases"

	"github.com/gin-gonic/gin"
)

// MatchingController 配對控制器
type MatchingController struct {
	matchingUsecase *usecases.MatchingUsecase
}

// NewMatchingController 創建配對控制器實例
func NewMatchingController(matchingUsecase *usecases.MatchingUsecase) *MatchingController {
	return &MatchingController{
		matchingUsecase: matchingUsecase,
	}
}

// FindMatches 尋找配對
// @Summary 尋找配對
// @Description 根據條件尋找合適的球友配對
// @Tags matching
// @Accept json
// @Produce json
// @Param request body dto.FindMatchesRequest true "配對條件"
// @Success 200 {object} map[string]interface{} "配對結果"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/find [post]
func (c *MatchingController) FindMatches(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req dto.FindMatchesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 設定預設值
	if req.Limit == 0 {
		req.Limit = 10
	}

	// 構建配對條件
	criteria := services.MatchingCriteria{
		UserID:             userID.(string),
		NTRPRange:          req.NTRPRange,
		MaxDistance:        req.MaxDistance,
		PlayingFrequency:   req.PlayingFrequency,
		AgeRange:           req.AgeRange,
		Gender:             req.Gender,
		MinReputationScore: req.MinReputationScore,
		// New mappings
		Location:     req.Location,
		PlayTypes:    req.PlayType,
		Availability: req.Availability,
	}

	// 尋找配對
	results, err := c.matchingUsecase.FindMatches(ctx, userID.(string), criteria, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find matches",
		})
		return
	}

	var matches []map[string]interface{}
	for _, result := range results {
		if result.User == nil {
			continue
		}

		user := result.User
		match := map[string]interface{}{
			"id":           user.ID,
			"name":         getFullName(user),
			"age":          calculateAge(user.Profile.BirthDate),
			"ntrpLevel":    user.Profile.NTRPLevel,
			"playingStyle": user.Profile.PlayingStyle,
			"bio":          user.Profile.Bio,
			"avatarUrl":    user.Profile.AvatarURL,
			"matchScore":   result.Score * 100,
			"lastActive":   user.LastLoginAt,
		}

		// 添加位置信息（如果有的話）
		if user.Profile.Latitude != nil && user.Profile.Longitude != nil {
			match["location"] = map[string]interface{}{
				"latitude":  *user.Profile.Latitude,
				"longitude": *user.Profile.Longitude,
				"address":   "", // 可以添加地址解析
			}
		}

		// Add additional fields for list view
		if user.Profile.Gender != nil {
			match["gender"] = *user.Profile.Gender
		}
		if user.Profile.PlayingFrequency != nil {
			match["playingFrequency"] = *user.Profile.PlayingFrequency
		}
		if len(user.Profile.PreferredTimes) > 0 {
			match["preferredTimes"] = user.Profile.PreferredTimes
		}

		matches = append(matches, match)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"total":   len(matches),
	})
}

// FindRandomMatches 尋找隨機配對（抽卡功能）
// @Summary 隨機配對
// @Description 抽卡式隨機配對功能
// @Tags matching
// @Accept json
// @Produce json
// @Param count query int false "配對數量" default(5)
// @Success 200 {object} map[string]interface{} "隨機配對結果"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/random [get]
func (c *MatchingController) FindRandomMatches(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 獲取數量參數
	countStr := ctx.DefaultQuery("count", "5")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		count = 5
	}
	if count > 20 {
		count = 20 // 限制最大數量
	}

	// 使用基本配對條件
	criteria := services.MatchingCriteria{
		UserID: userID.(string),
	}

	// 尋找隨機配對
	results, err := c.matchingUsecase.FindRandomMatches(ctx, userID.(string), criteria, count)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find random matches",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"matches": results,
		"total":   len(results),
	})
}

// GetReputationScore 獲取信譽分數
// @Summary 獲取信譽分數
// @Description 獲取用戶的信譽分數詳情
// @Tags matching
// @Produce json
// @Success 200 {object} map[string]interface{} "信譽分數"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/reputation [get]
func (c *MatchingController) GetReputationScore(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	reputation, err := c.matchingUsecase.GetUserReputationScore(ctx, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get reputation score",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"reputation": reputation,
	})
}

// GetMatchingHistory 獲取配對歷史
// @Summary 獲取配對歷史
// @Description 獲取用戶的配對歷史記錄
// @Tags matching
// @Produce json
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(10)
// @Success 200 {object} map[string]interface{} "配對歷史"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/history [get]
func (c *MatchingController) GetMatchingHistory(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 獲取分頁參數
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50 // 限制最大數量
	}

	offset := (page - 1) * limit

	// 獲取配對歷史
	matches, err := c.matchingUsecase.GetMatchingHistory(ctx, userID.(string), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get matching history",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"page":    page,
		"limit":   limit,
		"total":   len(matches),
	})
}

// CreateMatch 創建配對
// @Summary 創建配對
// @Description 創建新的球友配對
// @Tags matching
// @Accept json
// @Produce json
// @Param request body dto.CreateMatchRequest true "配對資訊"
// @Success 201 {object} map[string]interface{} "創建成功"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/create [post]
func (c *MatchingController) CreateMatch(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req dto.CreateMatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 驗證配對類型
	validTypes := map[string]bool{
		"casual":     true,
		"practice":   true,
		"tournament": true,
	}
	if !validTypes[req.MatchType] {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid match type",
		})
		return
	}

	// 創建配對
	match, err := c.matchingUsecase.CreateMatch(
		ctx,
		userID.(string),
		req.ParticipantIDs,
		req.MatchType,
		req.CourtID,
		req.ScheduledAt,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create match",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"match": match,
	})
}

// GetMatchingStatistics 獲取配對統計
// @Summary 獲取配對統計
// @Description 獲取用戶的配對統計資訊
// @Tags matching
// @Produce json
// @Success 200 {object} map[string]interface{} "統計資訊"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/statistics [get]
func (c *MatchingController) GetMatchingStatistics(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	stats, err := c.matchingUsecase.GetMatchingStatistics(ctx, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get statistics",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// UpdateReputation 更新信譽分數
// @Summary 更新信譽分數
// @Description 更新用戶的信譽分數（通常在比賽結束後調用）
// @Tags matching
// @Accept json
// @Produce json
// @Param userID path string true "用戶ID"
// @Param request body dto.UpdateReputationRequest true "信譽更新資訊"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/reputation/{userID} [put]
func (c *MatchingController) UpdateReputation(ctx *gin.Context) {
	// 這個API通常只有系統或管理員可以調用
	// 這裡簡化處理，實際應用中需要權限檢查

	targetUserID := ctx.Param("userID")
	if targetUserID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	var req dto.UpdateReputationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 驗證行為評分範圍
	if req.BehaviorRating < 1.0 || req.BehaviorRating > 5.0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Behavior rating must be between 1.0 and 5.0",
		})
		return
	}

	// 更新信譽分數
	err := c.matchingUsecase.UpdateUserReputationScore(
		ctx,
		targetUserID,
		req.MatchCompleted,
		req.WasOnTime,
		req.BehaviorRating,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update reputation",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Reputation updated successfully",
	})
}

// ProcessCardAction 處理抽卡動作
// @Summary 處理抽卡動作
// @Description 處理用戶對抽卡配對的動作（喜歡、不喜歡、跳過）
// @Tags matching
// @Accept json
// @Produce json
// @Param request body dto.CardActionRequest true "抽卡動作"
// @Success 200 {object} map[string]interface{} "處理結果"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/card-action [post]
func (c *MatchingController) ProcessCardAction(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req dto.CardActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 驗證動作類型
	validActions := map[string]bool{
		"like":    true,
		"dislike": true,
		"skip":    true,
	}
	if !validActions[req.Action] {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid action type",
		})
		return
	}

	// 不能對自己執行動作
	if req.TargetUserID == userID.(string) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot perform action on yourself",
		})
		return
	}

	// 處理抽卡動作
	result, err := c.matchingUsecase.ProcessCardAction(
		ctx,
		userID.(string),
		req.TargetUserID,
		req.Action,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process card action",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// GetCardInteractionHistory 獲取抽卡互動歷史
// @Summary 獲取抽卡互動歷史
// @Description 獲取用戶的抽卡互動歷史記錄
// @Tags matching
// @Produce json
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(20)
// @Param action query string false "動作類型篩選" Enums(like, dislike, skip)
// @Success 200 {object} map[string]interface{} "互動歷史"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/card-history [get]
func (c *MatchingController) GetCardInteractionHistory(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 獲取分頁參數
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")
	action := ctx.Query("action")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // 限制最大數量
	}

	offset := (page - 1) * limit

	// 獲取互動歷史
	interactions, total, err := c.matchingUsecase.GetCardInteractionHistory(
		ctx,
		userID.(string),
		action,
		limit,
		offset,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get card interaction history",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"interactions": interactions,
		"page":         page,
		"limit":        limit,
		"total":        total,
	})
}

// GetMatchNotifications 獲取配對通知
// @Summary 獲取配對通知
// @Description 獲取用戶的配對相關通知
// @Tags matching
// @Produce json
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(20)
// @Param unread_only query bool false "只顯示未讀" default(false)
// @Success 200 {object} map[string]interface{} "通知列表"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/notifications [get]
func (c *MatchingController) GetMatchNotifications(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 獲取分頁參數
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")
	unreadOnlyStr := ctx.DefaultQuery("unread_only", "false")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	unreadOnly := unreadOnlyStr == "true"
	offset := (page - 1) * limit

	// 獲取通知
	notifications, total, err := c.matchingUsecase.GetMatchNotifications(
		ctx,
		userID.(string),
		unreadOnly,
		limit,
		offset,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get notifications",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"page":          page,
		"limit":         limit,
		"total":         total,
	})
}

// MarkNotificationAsRead 標記通知為已讀
// @Summary 標記通知為已讀
// @Description 標記指定通知為已讀狀態
// @Tags matching
// @Accept json
// @Produce json
// @Param notificationID path string true "通知ID"
// @Success 200 {object} map[string]interface{} "標記成功"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 404 {object} map[string]interface{} "通知不存在"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matching/notifications/{notificationID}/read [put]
func (c *MatchingController) MarkNotificationAsRead(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	notificationID := ctx.Param("notificationID")
	if notificationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Notification ID is required",
		})
		return
	}

	// 標記通知為已讀
	err := c.matchingUsecase.MarkNotificationAsRead(
		ctx,
		userID.(string),
		notificationID,
	)
	if err != nil {
		if err.Error() == "notification not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Notification not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark notification as read",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

func getFullName(user *models.User) string {
	if user.Profile == nil {
		return ""
	}
	return user.Profile.FirstName + " " + user.Profile.LastName
}

// calculateAge 計算年齡
func calculateAge(birthDate *time.Time) int {
	if birthDate == nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}
