package controllers

import (
	"net/http"
	"strconv"

	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/services"
	"tennis-platform/backend/internal/usecases"

	"github.com/gin-gonic/gin"
)

// MatchesController 對戰配對控制器
type MatchesController struct {
	matchingUsecase *usecases.MatchingUsecase
}

// NewMatchesController 創建對戰配對控制器實例
func NewMatchesController(matchingUsecase *usecases.MatchingUsecase) *MatchesController {
	return &MatchesController{
		matchingUsecase: matchingUsecase,
	}
}

// FindMatches 尋找對手（競賽性質）
// @Summary 尋找對手
// @Description 尋找適合對戰的對手，重視技術匹配和信譽分數
// @Tags matches
// @Accept json
// @Produce json
// @Param request body dto.FindCompetitiveMatchesRequest true "對手篩選條件"
// @Success 200 {object} map[string]interface{} "對手列表"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matches/find [post]
func (c *MatchesController) FindMatches(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req dto.FindCompetitiveMatchesRequest
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
		Gender:             req.Gender,
		AgeRange:           req.AgeRange,
		MinReputationScore: req.MinReputationScore,
		Location:           req.Location,
		MaxDistance:        req.MaxDistance,
		Availability:       req.Availability,
	}

	// 尋找對手
	results, err := c.matchingUsecase.FindCompetitiveMatches(ctx, userID.(string), criteria, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find matches",
		})
		return
	}

	// 格式化結果
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

		// 添加位置信息
		if user.Profile.Latitude != nil && user.Profile.Longitude != nil {
			match["location"] = map[string]interface{}{
				"latitude":  *user.Profile.Latitude,
				"longitude": *user.Profile.Longitude,
			}
		}

		// 添加性別信息
		if user.Profile.Gender != nil {
			match["gender"] = *user.Profile.Gender
		}

		// 獲取信譽分數
		reputation, err := c.matchingUsecase.GetUserReputationScore(ctx, user.ID)
		if err == nil {
			match["reputation"] = map[string]interface{}{
				"overallScore":     reputation.OverallScore,
				"attendanceRate":   reputation.AttendanceRate,
				"punctualityScore": reputation.PunctualityScore,
				"behaviorRating":   reputation.BehaviorRating,
				"totalMatches":     reputation.TotalMatches,
				"completedMatches": reputation.CompletedMatches,
			}
		}

		// 添加配對因子
		match["factors"] = result.Factors

		matches = append(matches, match)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"total":   len(matches),
		"type":    "competitive", // 標記為競賽性質
	})
}

// GetMatchHistory 獲取對戰歷史
// @Summary 獲取對戰歷史
// @Description 獲取過往的競賽對戰記錄
// @Tags matches
// @Produce json
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(10)
// @Success 200 {object} map[string]interface{} "對戰歷史"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matches/history [get]
func (c *MatchesController) GetMatchHistory(ctx *gin.Context) {
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
		limit = 50
	}

	offset := (page - 1) * limit

	// 獲取配對歷史（篩選競賽類型）
	matches, err := c.matchingUsecase.GetMatchingHistory(ctx, userID.(string), limit*2, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get match history",
		})
		return
	}

	// 篩選競賽類型的配對
	var competitiveMatches []interface{}
	for _, match := range matches {
		if match.Type == "tournament" || match.Type == "competitive" {
			competitiveMatches = append(competitiveMatches, match)
			if len(competitiveMatches) >= limit {
				break
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"matches": competitiveMatches,
		"page":    page,
		"limit":   limit,
		"total":   len(competitiveMatches),
	})
}

// CreateMatch 創建競賽配對
// @Summary 創建競賽配對
// @Description 創建新的競賽對戰配對
// @Tags matches
// @Accept json
// @Produce json
// @Param request body dto.CreateMatchRequest true "配對資訊"
// @Success 201 {object} map[string]interface{} "創建成功"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/matches/create [post]
func (c *MatchesController) CreateMatch(ctx *gin.Context) {
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

	// 驗證配對類型（只允許競賽類型）
	validTypes := map[string]bool{
		"tournament":  true,
		"competitive": true,
	}
	if !validTypes[req.MatchType] {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid match type for competitive match. Use 'tournament' or 'competitive'",
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
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create competitive match",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"match": match,
	})
}
