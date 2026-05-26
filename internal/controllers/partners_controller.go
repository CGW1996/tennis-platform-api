package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"tennis-platform/backend/internal/usecases"

	"github.com/gin-gonic/gin"
)

// PartnersController 球友配對控制器
type PartnersController struct {
	matchingUsecase *usecases.MatchingUsecase
}

// NewPartnersController 創建球友配對控制器實例
func NewPartnersController(matchingUsecase *usecases.MatchingUsecase) *PartnersController {
	return &PartnersController{
		matchingUsecase: matchingUsecase,
	}
}

// FindPartners 尋找球友（練習性質）
// @Summary 尋找球友
// @Description 尋找適合練習的球友，重視位置接近度和時間彈性
// @Tags partners
// @Accept json
// @Produce json
// @Param request body dto.FindPartnersRequest true "球友篩選條件"
// @Success 200 {object} map[string]interface{} "球友列表"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/partners/find [post]
func (c *PartnersController) FindPartners(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req dto.FindPartnersRequest
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
		Location:           req.Location,
		PlayTypes:          req.PlayType,
		Availability:       req.Availability,
		PlayingFrequency:   req.PlayingFrequency,
		Gender:             req.Gender,
		AgeRange:           req.AgeRange,
		NTRPRange:          req.NTRPRange,
		MaxDistance:        req.MaxDistance,
		MinReputationScore: req.MinReputationScore,
	}

	// 尋找球友請求
	matches, err := c.matchingUsecase.FindPartnerRequests(ctx, userID.(string), criteria, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find partners",
		})
		return
	}

	// 格式化結果
	var partners []map[string]interface{}
	for _, match := range matches {
		// Find organizer
		var organizer *models.User
		for _, p := range match.Participants {
			if p.ID == userID.(string) {
				continue
			}
			// Assuming the first other participant is the organizer or the relevant user to show
			// logic in usecase filtered for organizer role, so match.Participants should contain organizer
			// We need to find the one that is NOT the current user (if current user happened to be in it?)
			// Actually, the usecase already filters out matches where current user is a participant.
			// So we just take the organizer.
			if p.ID != userID.(string) {
				// In a real scenario, we check p.Role == "organizer", but here we just take the first valid user profile
				organizer = &p
				break
			}
		}

		if organizer == nil || organizer.Profile == nil {
			continue
		}

		user := organizer
		partner := map[string]interface{}{
			"id":                  user.ID,
			"name":                getFullName(user),
			"age":                 calculateAge(user.Profile.BirthDate),
			"ntrpLevel":           user.Profile.NTRPLevel,
			"playingStyle":        user.Profile.PlayingStyle,
			"playingFrequency":    user.Profile.PlayingFrequency,
			"bio":                 user.Profile.Bio,
			"specialRequirements": match.SpecialRequirements,
			"avatarUrl":           user.Profile.AvatarURL,
			"matchScore":          90, // Static score for requests for now, or calculate
			"lastActive":          user.LastLoginAt,
			"requestId":           match.ID,                       // Pass the match ID
			"availabilitySlots":   user.Profile.AvailabilitySlots, // Pass slots directly
		}

		// Parse TargetCriteria for "Looking For"
		if match.TargetCriteria != nil {
			var target struct {
				NtrpMin   *float64 `json:"ntrpMin"`
				NtrpMax   *float64 `json:"ntrpMax"`
				PlayTypes []string `json:"playTypes"`
			}
			if err := json.Unmarshal([]byte(*match.TargetCriteria), &target); err == nil {
				partner["lookingFor"] = map[string]interface{}{
					"ntrpMin":   target.NtrpMin,
					"ntrpMax":   target.NtrpMax,
					"playTypes": target.PlayTypes,
				}
			}
		}

		// 添加位置信息 (Use first slot location if available as primary, or profile lat/long)
		if len(user.Profile.AvailabilitySlots) > 0 {
			// We can pass the whole slots array as above, frontend handles display
		}

		// Keep legacy gender field
		if user.Profile.Gender != nil {
			partner["gender"] = *user.Profile.Gender
		}

		partners = append(partners, partner)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"partners": partners,
		"total":    len(partners),
		"type":     "practice",
	})
}

// GetPartnerHistory 獲取球友歷史
// @Summary 獲取球友歷史
// @Description 獲取過往的練習球友記錄
// @Tags partners
// @Produce json
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(10)
// @Success 200 {object} map[string]interface{} "球友歷史"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/partners/history [get]
func (c *PartnersController) GetPartnerHistory(ctx *gin.Context) {
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

	// 獲取配對歷史（篩選練習類型）
	matches, err := c.matchingUsecase.GetMatchingHistory(ctx, userID.(string), limit*2, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get partner history",
		})
		return
	}

	// 篩選練習類型的配對
	var practiceMatches []interface{}
	for _, match := range matches {
		if match.Type == "practice" || match.Type == "casual" {
			practiceMatches = append(practiceMatches, match)
			if len(practiceMatches) >= limit {
				break
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"matches": practiceMatches,
		"page":    page,
		"limit":   limit,
		"total":   len(practiceMatches),
	})
}

// CreatePartnerMatch 創建球友練習配對
// @Summary 創建球友練習配對
// @Description 創建新的球友練習配對
// @Tags partners
// @Accept json
// @Produce json
// @Param request body dto.CreateMatchRequest true "配對資訊"
// @Success 201 {object} map[string]interface{} "創建成功"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Failure 500 {object} map[string]interface{} "伺服器錯誤"
// @Router /api/v1/partners/create [post]
func (c *PartnersController) CreatePartnerMatch(ctx *gin.Context) {
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

	// 強制設定為練習類型，除非明確指定為 casual
	if req.MatchType == "" {
		req.MatchType = "practice"
	}

	// 轉換 AvailabilitySlots
	var availabilitySlots []models.AvailabilitySlot
	if len(req.AvailabilitySlots) > 0 {
		for _, slot := range req.AvailabilitySlots {
			availabilitySlots = append(availabilitySlots, models.AvailabilitySlot{
				Day:       slot.Day,
				StartTime: slot.StartTime,
				EndTime:   slot.EndTime,
				Location:  slot.Location,
			})
		}
	}

	// 創建配對
	match, err := c.matchingUsecase.CreateMatch(
		ctx,
		userID.(string),
		req.ParticipantIDs,
		req.MatchType,
		req.CourtID,
		req.ScheduledAt,
		availabilitySlots,
		req.SpecialRequirements,
		req.NtrpMin,
		req.NtrpMax,
		req.PlayTypes,
	)
	if err != nil {
		// Add logging to see the actual error
		println("Error creating partner match:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create partner match: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"match": match,
	})
}
