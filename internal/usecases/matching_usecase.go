package usecases

import (
	"context"
	"fmt"
	"time"

	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"

	"gorm.io/gorm"
)

// MatchingUsecase 配對業務邏輯
type MatchingUsecase struct {
	db              *gorm.DB
	matchingService *services.MatchingService
}

// NewMatchingUsecase 創建配對業務邏輯實例
func NewMatchingUsecase(db *gorm.DB) *MatchingUsecase {
	return &MatchingUsecase{
		db:              db,
		matchingService: services.NewMatchingService(),
	}
}

// FindMatches 尋找配對
func (uc *MatchingUsecase) FindMatches(
	ctx context.Context,
	userID string,
	criteria services.MatchingCriteria,
	limit int,
) ([]services.MatchingResult, error) {
	// 獲取請求者資訊
	requester, err := uc.getUserWithProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get requester profile: %w", err)
	}

	// 獲取候選人列表
	candidates, err := uc.getCandidates(ctx, criteria, limit*3) // 獲取更多候選人以提高匹配品質
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	// 篩選候選人
	filteredCandidates := uc.matchingService.FilterCandidates(candidates, criteria)

	// 計算配對分數
	var results []services.MatchingResult
	weights := services.DefaultMatchingWeights()

	for _, candidate := range filteredCandidates {
		result := uc.matchingService.CalculateMatchingScore(
			requester,
			&candidate,
			criteria,
			weights,
		)
		results = append(results, result)
	}

	// 排序並限制結果數量
	rankedResults := uc.matchingService.RankCandidates(results)
	if len(rankedResults) > limit {
		rankedResults = rankedResults[:limit]
	}

	return rankedResults, nil
}

// FindRandomMatches 尋找隨機配對（抽卡功能）
func (uc *MatchingUsecase) FindRandomMatches(
	ctx context.Context,
	userID string,
	criteria services.MatchingCriteria,
	count int,
) ([]services.MatchingResult, error) {
	// 獲取候選人列表
	candidates, err := uc.getCandidates(ctx, criteria, count*5) // 獲取更多候選人以增加隨機性
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	// 生成隨機配對
	weights := services.DefaultMatchingWeights()
	results := uc.matchingService.GenerateRandomMatches(candidates, criteria, weights, count)

	return results, nil
}

// GetUserReputationScore 獲取用戶信譽分數
func (uc *MatchingUsecase) GetUserReputationScore(ctx context.Context, userID string) (*models.ReputationScore, error) {
	var reputation models.ReputationScore

	err := uc.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&reputation).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果沒有信譽記錄，創建預設記錄
			reputation = models.ReputationScore{
				UserID:           userID,
				AttendanceRate:   100.0,
				PunctualityScore: 100.0,
				SkillAccuracy:    100.0,
				BehaviorRating:   5.0,
				TotalMatches:     0,
				CompletedMatches: 0,
				CancelledMatches: 0,
				OverallScore:     100.0,
			}

			if err := uc.db.WithContext(ctx).Create(&reputation).Error; err != nil {
				return nil, fmt.Errorf("failed to create reputation score: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get reputation score: %w", err)
		}
	}

	return &reputation, nil
}

// UpdateUserReputationScore 更新用戶信譽分數
func (uc *MatchingUsecase) UpdateUserReputationScore(
	ctx context.Context,
	userID string,
	matchCompleted bool,
	wasOnTime bool,
	behaviorRating float64,
) error {
	reputation, err := uc.GetUserReputationScore(ctx, userID)
	if err != nil {
		return err
	}

	// 更新統計數據
	reputation.TotalMatches++
	if matchCompleted {
		reputation.CompletedMatches++
	} else {
		reputation.CancelledMatches++
	}

	// 重新計算出席率
	if reputation.TotalMatches > 0 {
		reputation.AttendanceRate = float64(reputation.CompletedMatches) / float64(reputation.TotalMatches) * 100
	}

	// 更新準時度（移動平均）
	if matchCompleted {
		punctualityScore := 0.0
		if wasOnTime {
			punctualityScore = 100.0
		}

		// 使用加權平均更新準時度
		weight := 0.1 // 新數據權重
		reputation.PunctualityScore = reputation.PunctualityScore*(1-weight) + punctualityScore*weight
	}

	// 更新行為評分（移動平均）
	if behaviorRating > 0 {
		behaviorScore := behaviorRating / 5.0 * 100 // 轉換為0-100分
		weight := 0.1
		reputation.BehaviorRating = reputation.BehaviorRating*(1-weight) + behaviorScore*weight
	}

	// 重新計算綜合分數
	reputation.OverallScore = reputation.AttendanceRate*0.3 +
		reputation.PunctualityScore*0.2 +
		reputation.SkillAccuracy*0.2 +
		reputation.BehaviorRating*0.3

	// 保存更新
	err = uc.db.WithContext(ctx).Save(reputation).Error
	if err != nil {
		return fmt.Errorf("failed to update reputation score: %w", err)
	}

	return nil
}

// GetMatchingHistory 獲取配對歷史
func (uc *MatchingUsecase) GetMatchingHistory(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]models.Match, error) {
	var matches []models.Match

	err := uc.db.WithContext(ctx).
		Joins("JOIN match_participants ON matches.id = match_participants.match_id").
		Where("match_participants.user_id = ?", userID).
		Preload("Participants").
		Preload("Participants.Profile").
		Preload("Court").
		Preload("Results").
		Order("matches.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&matches).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get matching history: %w", err)
	}

	return matches, nil
}

// CreateMatch 創建配對
func (uc *MatchingUsecase) CreateMatch(
	ctx context.Context,
	organizerID string,
	participantIDs []string,
	matchType string,
	courtID *string,
	scheduledAt *string,
) (*models.Match, error) {
	// 開始事務
	tx := uc.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 創建比賽記錄
	match := models.Match{
		Type:    matchType,
		Status:  "pending",
		CourtID: courtID,
	}

	if scheduledAt != nil {
		// 解析時間字符串
		// TODO: 實作時間解析邏輯
	}

	if err := tx.Create(&match).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create match: %w", err)
	}

	// 添加組織者
	organizerParticipant := models.MatchParticipant{
		MatchID: match.ID,
		UserID:  organizerID,
		Role:    "organizer",
		Status:  "accepted",
	}

	if err := tx.Create(&organizerParticipant).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add organizer: %w", err)
	}

	// 添加其他參與者
	for _, participantID := range participantIDs {
		if participantID == organizerID {
			continue // 跳過組織者
		}

		participant := models.MatchParticipant{
			MatchID: match.ID,
			UserID:  participantID,
			Role:    "player",
			Status:  "pending",
		}

		if err := tx.Create(&participant).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to add participant %s: %w", participantID, err)
		}
	}

	// 創建聊天室
	chatRoom := models.ChatRoom{
		MatchID:  &match.ID,
		Type:     "match",
		IsActive: true,
	}

	if err := tx.Create(&chatRoom).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 重新載入完整的比賽資訊
	if err := uc.db.WithContext(ctx).
		Preload("Participants").
		Preload("Participants.Profile").
		Preload("Court").
		Preload("ChatRoom").
		First(&match, match.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload match: %w", err)
	}

	return &match, nil
}

// getUserWithProfile 獲取用戶及其檔案資訊
func (uc *MatchingUsecase) getUserWithProfile(userID string) (*models.User, error) {
	var user models.User

	err := uc.db.
		Preload("Profile").
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// getCandidates 獲取候選人列表
func (uc *MatchingUsecase) getCandidates(
	ctx context.Context,
	criteria services.MatchingCriteria,
	limit int,
) ([]models.User, error) {
	query := uc.db.WithContext(ctx).
		Preload("Profile").
		Joins("JOIN user_profiles ON users.id = user_profiles.user_id").
		Where("users.is_active = ?", true).
		Where("users.id != ?", criteria.UserID) // 排除自己

	// 基本篩選條件
	if criteria.Gender != nil && *criteria.Gender != "any" {
		query = query.Where("user_profiles.gender = ?", *criteria.Gender)
	}

	// NTRP等級篩選
	if criteria.NTRPRange != nil {
		query = query.Where("user_profiles.ntrp_level BETWEEN ? AND ?", criteria.NTRPRange.Min, criteria.NTRPRange.Max)
	}

	// 打球頻率篩選
	if criteria.PlayingFrequency != nil {
		query = query.Where("user_profiles.playing_frequency = ?", *criteria.PlayingFrequency)
	}

	// 地理位置篩選
	if criteria.MaxDistance != nil && *criteria.MaxDistance > 0 {
		// 使用PostGIS進行地理查詢
		// 這裡需要請求者的位置資訊
		// TODO: 實作地理位置篩選
	}

	var users []models.User
	err := query.Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetMatchingStatistics 獲取配對統計資訊
func (uc *MatchingUsecase) GetMatchingStatistics(ctx context.Context, userID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 總配對數
	var totalMatches int64
	err := uc.db.WithContext(ctx).
		Table("match_participants").
		Where("user_id = ?", userID).
		Count(&totalMatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total matches: %w", err)
	}
	stats["totalMatches"] = totalMatches

	// 已完成配對數
	var completedMatches int64
	err = uc.db.WithContext(ctx).
		Table("match_participants").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ? AND matches.status = ?", userID, "completed").
		Count(&completedMatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count completed matches: %w", err)
	}
	stats["completedMatches"] = completedMatches

	// 取消配對數
	var cancelledMatches int64
	err = uc.db.WithContext(ctx).
		Table("match_participants").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ? AND matches.status = ?", userID, "cancelled").
		Count(&cancelledMatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count cancelled matches: %w", err)
	}
	stats["cancelledMatches"] = cancelledMatches

	// 信譽分數
	reputation, err := uc.GetUserReputationScore(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reputation score: %w", err)
	}
	stats["reputationScore"] = reputation.OverallScore

	// 計算成功率
	if totalMatches > 0 {
		successRate := float64(completedMatches) / float64(totalMatches) * 100
		stats["successRate"] = successRate
	} else {
		stats["successRate"] = 0.0
	}

	return stats, nil
}

// ProcessCardAction 處理抽卡動作
func (uc *MatchingUsecase) ProcessCardAction(
	ctx context.Context,
	userID string,
	targetUserID string,
	action string,
) (*services.CardMatchResult, error) {
	// 開始事務
	tx := uc.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 記錄互動
	interaction := models.CardInteraction{
		UserID:       userID,
		TargetUserID: targetUserID,
		Action:       action,
		IsMatch:      false,
	}

	if err := tx.Create(&interaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create card interaction: %w", err)
	}

	result := &services.CardMatchResult{
		IsMatch: false,
	}

	// 如果是喜歡動作，檢查是否互相喜歡
	if action == "like" {
		// 檢查對方是否也喜歡過自己
		var mutualInteraction models.CardInteraction
		err := tx.Where("user_id = ? AND target_user_id = ? AND action = ?",
			targetUserID, userID, "like").
			First(&mutualInteraction).Error

		if err == nil {
			// 互相喜歡，創建配對
			match, err := uc.createMatchFromCardInteraction(tx, userID, targetUserID)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create match: %w", err)
			}

			// 更新互動記錄
			interaction.IsMatch = true
			interaction.MatchID = &match.ID
			if err := tx.Save(&interaction).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update interaction: %w", err)
			}

			// 更新對方的互動記錄
			mutualInteraction.IsMatch = true
			mutualInteraction.MatchID = &match.ID
			if err := tx.Save(&mutualInteraction).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update mutual interaction: %w", err)
			}

			// 創建通知
			if err := uc.createMatchNotifications(tx, userID, targetUserID, match.ID); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create notifications: %w", err)
			}

			result.IsMatch = true
			result.MatchID = match.ID
			result.ChatRoomID = match.ChatRoom.ID
			result.Message = "配對成功！你們可以開始聊天了"
		} else {
			result.Message = "已表達興趣，等待對方回應"
		}
	} else {
		result.Message = "已跳過此用戶"
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// createMatchFromCardInteraction 從抽卡互動創建配對
func (uc *MatchingUsecase) createMatchFromCardInteraction(
	tx *gorm.DB,
	userID1, userID2 string,
) (*models.Match, error) {
	// 創建比賽記錄
	match := models.Match{
		Type:   "casual",
		Status: "pending",
	}

	if err := tx.Create(&match).Error; err != nil {
		return nil, err
	}

	// 添加參與者
	participants := []models.MatchParticipant{
		{
			MatchID: match.ID,
			UserID:  userID1,
			Role:    "player",
			Status:  "accepted",
		},
		{
			MatchID: match.ID,
			UserID:  userID2,
			Role:    "player",
			Status:  "accepted",
		},
	}

	for _, participant := range participants {
		if err := tx.Create(&participant).Error; err != nil {
			return nil, err
		}
	}

	// 創建聊天室
	chatRoom := models.ChatRoom{
		MatchID:  &match.ID,
		Type:     "match",
		IsActive: true,
	}

	if err := tx.Create(&chatRoom).Error; err != nil {
		return nil, err
	}

	// 重新載入完整的比賽資訊
	if err := tx.Preload("Participants").
		Preload("Participants.Profile").
		Preload("ChatRoom").
		First(&match, match.ID).Error; err != nil {
		return nil, err
	}

	return &match, nil
}

// createMatchNotifications 創建配對通知
func (uc *MatchingUsecase) createMatchNotifications(
	tx *gorm.DB,
	userID1, userID2, matchID string,
) error {
	notifications := []models.MatchNotification{
		{
			UserID:  userID1,
			Type:    "match_success",
			Title:   "配對成功！",
			Message: "你們互相喜歡，現在可以開始聊天了",
			Data:    fmt.Sprintf(`{"matchId":"%s","targetUserId":"%s"}`, matchID, userID2),
		},
		{
			UserID:  userID2,
			Type:    "match_success",
			Title:   "配對成功！",
			Message: "你們互相喜歡，現在可以開始聊天了",
			Data:    fmt.Sprintf(`{"matchId":"%s","targetUserId":"%s"}`, matchID, userID1),
		},
	}

	for _, notification := range notifications {
		if err := tx.Create(&notification).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetCardInteractionHistory 獲取抽卡互動歷史
func (uc *MatchingUsecase) GetCardInteractionHistory(
	ctx context.Context,
	userID string,
	action string,
	limit, offset int,
) ([]models.CardInteraction, int64, error) {
	query := uc.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("TargetUser").
		Preload("TargetUser.Profile").
		Preload("Match")

	if action != "" {
		query = query.Where("action = ?", action)
	}

	// 獲取總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count interactions: %w", err)
	}

	// 獲取數據
	var interactions []models.CardInteraction
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&interactions).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get interactions: %w", err)
	}

	return interactions, total, nil
}

// GetMatchNotifications 獲取配對通知
func (uc *MatchingUsecase) GetMatchNotifications(
	ctx context.Context,
	userID string,
	unreadOnly bool,
	limit, offset int,
) ([]models.MatchNotification, int64, error) {
	query := uc.db.WithContext(ctx).Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// 獲取總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// 獲取數據
	var notifications []models.MatchNotification
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, total, nil
}

// MarkNotificationAsRead 標記通知為已讀
func (uc *MatchingUsecase) MarkNotificationAsRead(
	ctx context.Context,
	userID string,
	notificationID string,
) error {
	now := time.Now()
	result := uc.db.WithContext(ctx).
		Model(&models.MatchNotification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}
