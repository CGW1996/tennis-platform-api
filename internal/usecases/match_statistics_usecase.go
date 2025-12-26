package usecases

import (
	"fmt"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"

	"gorm.io/gorm"
)

// MatchStatisticsUseCase 配對統計用例
type MatchStatisticsUseCase struct {
	db                     *gorm.DB
	matchStatisticsService *services.MatchStatisticsService
	reputationService      *services.ReputationService
}

// NewMatchStatisticsUseCase 創建新的配對統計用例
func NewMatchStatisticsUseCase(db *gorm.DB) *MatchStatisticsUseCase {
	return &MatchStatisticsUseCase{
		db:                     db,
		matchStatisticsService: services.NewMatchStatisticsService(db),
		reputationService:      services.NewReputationService(db),
	}
}

// GetUserMatchStatistics 獲取用戶配對統計資訊
func (msuc *MatchStatisticsUseCase) GetUserMatchStatistics(userID, requestingUserID string) (*models.MatchStatistics, error) {
	// 獲取統計資訊
	stats, err := msuc.matchStatisticsService.GetMatchStatistics(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match statistics: %w", err)
	}

	// 獲取隱私設定
	privacy, err := msuc.matchStatisticsService.GetUserPrivacySettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	// 根據隱私設定過濾資訊
	isOwner := userID == requestingUserID
	filteredStats := msuc.matchStatisticsService.FilterStatisticsByPrivacy(stats, privacy, isOwner)

	return filteredStats, nil
}

// GetUserMatchHistory 獲取用戶配對歷史
func (msuc *MatchStatisticsUseCase) GetUserMatchHistory(userID, requestingUserID string, limit, offset int) ([]models.Match, error) {
	// 檢查隱私設定
	privacy, err := msuc.matchStatisticsService.GetUserPrivacySettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	isOwner := userID == requestingUserID
	if !isOwner && !privacy.ShowMatchHistory {
		return nil, fmt.Errorf("match history is private")
	}

	// 獲取配對歷史
	var matches []models.Match
	err = msuc.db.
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
		return nil, fmt.Errorf("failed to get match history: %w", err)
	}

	return matches, nil
}

// RecordMatchResult 記錄比賽結果
func (msuc *MatchStatisticsUseCase) RecordMatchResult(matchID, winnerID, loserID, score string, recordedBy string) error {
	// 驗證記錄者是比賽參與者
	var match models.Match
	err := msuc.db.Preload("Participants").Where("id = ?", matchID).First(&match).Error
	if err != nil {
		return fmt.Errorf("match not found: %w", err)
	}

	isParticipant := false
	for _, participant := range match.Participants {
		if participant.ID == recordedBy {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return fmt.Errorf("only match participants can record results")
	}

	// 驗證勝負者都是比賽參與者
	winnerIsParticipant := false
	loserIsParticipant := false
	for _, participant := range match.Participants {
		if participant.ID == winnerID {
			winnerIsParticipant = true
		}
		if participant.ID == loserID {
			loserIsParticipant = true
		}
	}

	if !winnerIsParticipant || !loserIsParticipant {
		return fmt.Errorf("winner and loser must be match participants")
	}

	// 記錄比賽結果
	err = msuc.matchStatisticsService.RecordMatchResult(matchID, winnerID, loserID, score)
	if err != nil {
		return fmt.Errorf("failed to record match result: %w", err)
	}

	// 更新信譽分數
	err = msuc.reputationService.UpdateAttendanceRate(winnerID, "completed")
	if err != nil {
		return fmt.Errorf("failed to update winner reputation: %w", err)
	}

	err = msuc.reputationService.UpdateAttendanceRate(loserID, "completed")
	if err != nil {
		return fmt.Errorf("failed to update loser reputation: %w", err)
	}

	// 自動調整技術等級
	err = msuc.matchStatisticsService.AutoAdjustSkillLevel(winnerID)
	if err != nil {
		return fmt.Errorf("failed to auto adjust winner skill level: %w", err)
	}

	err = msuc.matchStatisticsService.AutoAdjustSkillLevel(loserID)
	if err != nil {
		return fmt.Errorf("failed to auto adjust loser skill level: %w", err)
	}

	return nil
}

// ConfirmMatchResult 確認比賽結果
func (msuc *MatchStatisticsUseCase) ConfirmMatchResult(matchResultID, userID string) error {
	return msuc.matchStatisticsService.ConfirmMatchResult(matchResultID, userID)
}

// GetSkillLevelProgression 獲取技術等級進展
func (msuc *MatchStatisticsUseCase) GetSkillLevelProgression(userID, requestingUserID string) ([]models.SkillLevelRecord, error) {
	// 檢查隱私設定
	privacy, err := msuc.matchStatisticsService.GetUserPrivacySettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	isOwner := userID == requestingUserID
	if !isOwner && !privacy.ShowSkillProgression {
		return nil, fmt.Errorf("skill progression is private")
	}

	var skillRecords []models.SkillLevelRecord
	err = msuc.db.Where("user_id = ?", userID).
		Preload("Match").
		Order("created_at ASC").
		Find(&skillRecords).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get skill level progression: %w", err)
	}

	return skillRecords, nil
}

// ManuallyAdjustSkillLevel 手動調整技術等級
func (msuc *MatchStatisticsUseCase) ManuallyAdjustSkillLevel(userID string, newLevel float64, reason string, adjustedBy string) error {
	// 驗證等級範圍
	if newLevel < 1.0 || newLevel > 7.0 {
		return fmt.Errorf("NTRP level must be between 1.0 and 7.0")
	}

	// 獲取用戶當前等級
	var userProfile models.UserProfile
	err := msuc.db.Where("user_id = ?", userID).First(&userProfile).Error
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	oldLevel := 0.0
	if userProfile.NTRPLevel != nil {
		oldLevel = *userProfile.NTRPLevel
	}

	// 記錄等級變更
	skillLevelRecord := models.SkillLevelRecord{
		UserID:   userID,
		OldLevel: oldLevel,
		NewLevel: newLevel,
		Reason:   reason,
	}

	if err := msuc.db.Create(&skillLevelRecord).Error; err != nil {
		return fmt.Errorf("failed to create skill level record: %w", err)
	}

	// 更新用戶等級
	if err := msuc.db.Model(&userProfile).Update("ntrp_level", newLevel).Error; err != nil {
		return fmt.Errorf("failed to update user NTRP level: %w", err)
	}

	return nil
}

// GetUserPrivacySettings 獲取用戶隱私設定
func (msuc *MatchStatisticsUseCase) GetUserPrivacySettings(userID string) (*models.UserPrivacySettings, error) {
	return msuc.matchStatisticsService.GetUserPrivacySettings(userID)
}

// UpdateUserPrivacySettings 更新用戶隱私設定
func (msuc *MatchStatisticsUseCase) UpdateUserPrivacySettings(userID string, settings *models.UserPrivacySettings) error {
	return msuc.matchStatisticsService.UpdateUserPrivacySettings(userID, settings)
}

// GetMatchResultsForConfirmation 獲取待確認的比賽結果
func (msuc *MatchStatisticsUseCase) GetMatchResultsForConfirmation(userID string) ([]models.MatchResult, error) {
	var matchResults []models.MatchResult

	err := msuc.db.
		Where("(winner_id = ? OR loser_id = ?) AND is_confirmed = ?", userID, userID, false).
		Preload("Match").
		Preload("Winner").
		Preload("Winner.Profile").
		Preload("Loser").
		Preload("Loser.Profile").
		Order("created_at DESC").
		Find(&matchResults).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get match results for confirmation: %w", err)
	}

	// 過濾出用戶尚未確認的結果
	var unconfirmedResults []models.MatchResult
	for _, result := range matchResults {
		hasConfirmed := false
		for _, confirmedBy := range result.ConfirmedBy {
			if confirmedBy == userID {
				hasConfirmed = true
				break
			}
		}
		if !hasConfirmed {
			unconfirmedResults = append(unconfirmedResults, result)
		}
	}

	return unconfirmedResults, nil
}

// GetReputationScoreWithPrivacy 根據隱私設定獲取信譽分數
func (msuc *MatchStatisticsUseCase) GetReputationScoreWithPrivacy(userID, requestingUserID string) (*models.ReputationScore, error) {
	// 獲取隱私設定
	privacy, err := msuc.matchStatisticsService.GetUserPrivacySettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	isOwner := userID == requestingUserID
	if !isOwner && !privacy.ShowReputationScore {
		return nil, fmt.Errorf("reputation score is private")
	}

	// 獲取信譽分數
	reputation, err := msuc.reputationService.GetReputationScore(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reputation score: %w", err)
	}

	return reputation, nil
}

// GetBehaviorReviewsWithPrivacy 根據隱私設定獲取行為評價
func (msuc *MatchStatisticsUseCase) GetBehaviorReviewsWithPrivacy(userID, requestingUserID string, limit, offset int) ([]models.BehaviorReview, error) {
	// 獲取隱私設定
	privacy, err := msuc.matchStatisticsService.GetUserPrivacySettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	isOwner := userID == requestingUserID
	if !isOwner && !privacy.ShowBehaviorReviews {
		return nil, fmt.Errorf("behavior reviews are private")
	}

	// 獲取行為評價
	var behaviorReviews []models.BehaviorReview
	err = msuc.db.Where("user_id = ?", userID).
		Preload("Reviewer").
		Preload("Reviewer.Profile").
		Preload("Match").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&behaviorReviews).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get behavior reviews: %w", err)
	}

	return behaviorReviews, nil
}

// AutoAdjustAllUserSkillLevels 自動調整所有用戶的技術等級
func (msuc *MatchStatisticsUseCase) AutoAdjustAllUserSkillLevels() error {
	// 獲取所有有NTRP等級的用戶
	var userProfiles []models.UserProfile
	err := msuc.db.Where("ntrp_level IS NOT NULL").Find(&userProfiles).Error
	if err != nil {
		return fmt.Errorf("failed to get user profiles: %w", err)
	}

	// 為每個用戶調整技術等級
	for _, profile := range userProfiles {
		err := msuc.matchStatisticsService.AutoAdjustSkillLevel(profile.UserID)
		if err != nil {
			// 記錄錯誤但繼續處理其他用戶
			fmt.Printf("Failed to auto adjust skill level for user %s: %v\n", profile.UserID, err)
		}
	}

	return nil
}

// GetMatchStatisticsSummary 獲取配對統計摘要
func (msuc *MatchStatisticsUseCase) GetMatchStatisticsSummary(userID string) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// 基本統計
	var totalMatches int64
	err := msuc.db.Table("match_participants").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ?", userID).
		Count(&totalMatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total matches: %w", err)
	}
	summary["totalMatches"] = totalMatches

	// 本月比賽數
	var monthlyMatches int64
	err = msuc.db.Table("match_participants").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ? AND EXTRACT(YEAR FROM matches.created_at) = ? AND EXTRACT(MONTH FROM matches.created_at) = ?",
			userID, 2024, 11).
		Count(&monthlyMatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count monthly matches: %w", err)
	}
	summary["monthlyMatches"] = monthlyMatches

	// 勝率
	var wonMatches int64
	err = msuc.db.Table("match_results").
		Where("winner_id = ?", userID).
		Count(&wonMatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count won matches: %w", err)
	}

	var lostMatches int64
	err = msuc.db.Table("match_results").
		Where("loser_id = ?", userID).
		Count(&lostMatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count lost matches: %w", err)
	}

	totalGames := wonMatches + lostMatches
	winRate := 0.0
	if totalGames > 0 {
		winRate = float64(wonMatches) / float64(totalGames) * 100
	}
	summary["winRate"] = winRate

	// 信譽分數
	reputation, err := msuc.reputationService.GetReputationScore(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reputation score: %w", err)
	}
	summary["reputationScore"] = reputation.OverallScore

	// 當前NTRP等級
	var userProfile models.UserProfile
	err = msuc.db.Where("user_id = ?", userID).First(&userProfile).Error
	if err == nil && userProfile.NTRPLevel != nil {
		summary["ntrpLevel"] = *userProfile.NTRPLevel
	}

	return summary, nil
}
