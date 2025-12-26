package services

import (
	"fmt"
	"math"
	"tennis-platform/backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// MatchStatisticsService 配對統計服務
type MatchStatisticsService struct {
	db *gorm.DB
}

// NewMatchStatisticsService 創建新的配對統計服務
func NewMatchStatisticsService(db *gorm.DB) *MatchStatisticsService {
	return &MatchStatisticsService{
		db: db,
	}
}

// GetMatchStatistics 獲取用戶配對統計資訊
func (mss *MatchStatisticsService) GetMatchStatistics(userID string) (*models.MatchStatistics, error) {
	stats := &models.MatchStatistics{
		UserID: userID,
	}

	// 獲取基本統計數據
	if err := mss.calculateBasicStats(userID, stats); err != nil {
		return nil, fmt.Errorf("failed to calculate basic stats: %w", err)
	}

	// 獲取勝負記錄
	if err := mss.calculateWinLossStats(userID, stats); err != nil {
		return nil, fmt.Errorf("failed to calculate win/loss stats: %w", err)
	}

	// 獲取平均比賽時長
	if err := mss.calculateAverageMatchDuration(userID, stats); err != nil {
		return nil, fmt.Errorf("failed to calculate average match duration: %w", err)
	}

	// 獲取最常配對的用戶
	if err := mss.getMostPlayedWith(userID, stats); err != nil {
		return nil, fmt.Errorf("failed to get most played with: %w", err)
	}

	// 獲取最近比賽記錄
	if err := mss.getRecentMatches(userID, stats); err != nil {
		return nil, fmt.Errorf("failed to get recent matches: %w", err)
	}

	// 獲取月度統計
	if err := mss.getMonthlyStats(userID, stats); err != nil {
		return nil, fmt.Errorf("failed to get monthly stats: %w", err)
	}

	// 獲取技術等級進展
	if err := mss.getSkillLevelProgression(userID, stats); err != nil {
		return nil, fmt.Errorf("failed to get skill level progression: %w", err)
	}

	return stats, nil
}

// calculateBasicStats 計算基本統計數據
func (mss *MatchStatisticsService) calculateBasicStats(userID string, stats *models.MatchStatistics) error {
	// 總比賽數
	err := mss.db.Table("match_participants").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ?", userID).
		Count(&stats.TotalMatches).Error
	if err != nil {
		return err
	}

	// 已完成比賽數
	err = mss.db.Table("match_participants").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ? AND matches.status = ?", userID, "completed").
		Count(&stats.CompletedMatches).Error
	if err != nil {
		return err
	}

	// 取消比賽數
	err = mss.db.Table("match_participants").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ? AND matches.status = ?", userID, "cancelled").
		Count(&stats.CancelledMatches).Error
	if err != nil {
		return err
	}

	// 計算出席率
	if stats.TotalMatches > 0 {
		stats.AttendanceRate = float64(stats.CompletedMatches) / float64(stats.TotalMatches) * 100
	}

	return nil
}

// calculateWinLossStats 計算勝負統計
func (mss *MatchStatisticsService) calculateWinLossStats(userID string, stats *models.MatchStatistics) error {
	// 勝場數
	err := mss.db.Table("match_results").
		Where("winner_id = ?", userID).
		Count(&stats.WonMatches).Error
	if err != nil {
		return err
	}

	// 敗場數
	err = mss.db.Table("match_results").
		Where("loser_id = ?", userID).
		Count(&stats.LostMatches).Error
	if err != nil {
		return err
	}

	// 計算勝率
	totalGames := stats.WonMatches + stats.LostMatches
	if totalGames > 0 {
		stats.WinRate = float64(stats.WonMatches) / float64(totalGames) * 100
	}

	return nil
}

// calculateAverageMatchDuration 計算平均比賽時長
func (mss *MatchStatisticsService) calculateAverageMatchDuration(userID string, stats *models.MatchStatistics) error {
	var avgDuration *float64

	err := mss.db.Table("match_participants").
		Select("AVG(matches.duration)").
		Joins("JOIN matches ON match_participants.match_id = matches.id").
		Where("match_participants.user_id = ? AND matches.status = ? AND matches.duration IS NOT NULL", userID, "completed").
		Scan(&avgDuration).Error

	if err != nil {
		return err
	}

	if avgDuration != nil {
		stats.AverageMatchDuration = int(*avgDuration)
	}

	return nil
}

// getMostPlayedWith 獲取最常配對的用戶
func (mss *MatchStatisticsService) getMostPlayedWith(userID string, stats *models.MatchStatistics) error {
	type PartnerCount struct {
		UserID string `json:"userId"`
		Count  int    `json:"count"`
	}

	var partners []PartnerCount

	err := mss.db.Raw(`
		SELECT 
			mp2.user_id,
			COUNT(*) as count
		FROM match_participants mp1
		JOIN match_participants mp2 ON mp1.match_id = mp2.match_id
		JOIN matches m ON mp1.match_id = m.id
		WHERE mp1.user_id = ? 
		AND mp2.user_id != ? 
		AND m.status = 'completed'
		GROUP BY mp2.user_id
		ORDER BY count DESC
		LIMIT 5
	`, userID, userID).Scan(&partners).Error

	if err != nil {
		return err
	}

	for _, partner := range partners {
		stats.MostPlayedWith = append(stats.MostPlayedWith, partner.UserID)
	}

	return nil
}

// getRecentMatches 獲取最近比賽記錄
func (mss *MatchStatisticsService) getRecentMatches(userID string, stats *models.MatchStatistics) error {
	var matches []models.Match

	err := mss.db.
		Joins("JOIN match_participants ON matches.id = match_participants.match_id").
		Where("match_participants.user_id = ?", userID).
		Preload("Participants").
		Preload("Participants.Profile").
		Preload("Court").
		Preload("Results").
		Order("matches.created_at DESC").
		Limit(10).
		Find(&matches).Error

	if err != nil {
		return err
	}

	stats.RecentMatches = matches
	return nil
}

// getMonthlyStats 獲取月度統計
func (mss *MatchStatisticsService) getMonthlyStats(userID string, stats *models.MatchStatistics) error {
	// 獲取過去12個月的統計數據
	var monthlyStats []models.MonthlyMatchStats

	for i := 11; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		year := date.Year()
		month := int(date.Month())

		monthStat := models.MonthlyMatchStats{
			Year:  year,
			Month: month,
		}

		// 該月總比賽數
		err := mss.db.Table("match_participants").
			Joins("JOIN matches ON match_participants.match_id = matches.id").
			Where("match_participants.user_id = ? AND EXTRACT(YEAR FROM matches.created_at) = ? AND EXTRACT(MONTH FROM matches.created_at) = ?",
				userID, year, month).
			Count(&monthStat.TotalMatches).Error
		if err != nil {
			return err
		}

		// 該月完成比賽數
		err = mss.db.Table("match_participants").
			Joins("JOIN matches ON match_participants.match_id = matches.id").
			Where("match_participants.user_id = ? AND matches.status = ? AND EXTRACT(YEAR FROM matches.created_at) = ? AND EXTRACT(MONTH FROM matches.created_at) = ?",
				userID, "completed", year, month).
			Count(&monthStat.CompletedMatches).Error
		if err != nil {
			return err
		}

		// 該月勝場數
		var wonMatches int64
		err = mss.db.Table("match_results").
			Joins("JOIN matches ON match_results.match_id = matches.id").
			Where("match_results.winner_id = ? AND EXTRACT(YEAR FROM matches.created_at) = ? AND EXTRACT(MONTH FROM matches.created_at) = ?",
				userID, year, month).
			Count(&wonMatches).Error
		if err != nil {
			return err
		}

		// 該月敗場數
		var lostMatches int64
		err = mss.db.Table("match_results").
			Joins("JOIN matches ON match_results.match_id = matches.id").
			Where("match_results.loser_id = ? AND EXTRACT(YEAR FROM matches.created_at) = ? AND EXTRACT(MONTH FROM matches.created_at) = ?",
				userID, year, month).
			Count(&lostMatches).Error
		if err != nil {
			return err
		}

		// 計算該月勝率
		totalGames := wonMatches + lostMatches
		if totalGames > 0 {
			monthStat.WinRate = float64(wonMatches) / float64(totalGames) * 100
		}

		// 該月平均行為評分
		var avgRating *float64
		err = mss.db.Table("behavior_reviews").
			Where("user_id = ? AND EXTRACT(YEAR FROM created_at) = ? AND EXTRACT(MONTH FROM created_at) = ?",
				userID, year, month).
			Select("AVG(rating)").
			Scan(&avgRating).Error
		if err != nil {
			return err
		}

		if avgRating != nil {
			monthStat.AverageRating = *avgRating
		}

		monthlyStats = append(monthlyStats, monthStat)
	}

	stats.MonthlyStats = monthlyStats
	return nil
}

// getSkillLevelProgression 獲取技術等級進展
func (mss *MatchStatisticsService) getSkillLevelProgression(userID string, stats *models.MatchStatistics) error {
	var skillRecords []models.SkillLevelRecord

	err := mss.db.Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&skillRecords).Error

	if err != nil {
		return err
	}

	stats.SkillLevelProgression = skillRecords
	return nil
}

// RecordMatchResult 記錄比賽結果
func (mss *MatchStatisticsService) RecordMatchResult(matchID, winnerID, loserID string, score string) error {
	// 開始事務
	tx := mss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 創建比賽結果記錄
	matchResult := models.MatchResult{
		MatchID:     matchID,
		WinnerID:    &winnerID,
		LoserID:     &loserID,
		Score:       &score,
		IsConfirmed: false,
		ConfirmedBy: []string{},
	}

	if err := tx.Create(&matchResult).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create match result: %w", err)
	}

	// 更新比賽狀態
	if err := tx.Model(&models.Match{}).Where("id = ?", matchID).Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update match status: %w", err)
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ConfirmMatchResult 確認比賽結果
func (mss *MatchStatisticsService) ConfirmMatchResult(matchResultID, userID string) error {
	var matchResult models.MatchResult

	err := mss.db.Where("id = ?", matchResultID).First(&matchResult).Error
	if err != nil {
		return fmt.Errorf("match result not found: %w", err)
	}

	// 檢查用戶是否有權限確認
	if matchResult.WinnerID == nil || matchResult.LoserID == nil {
		return fmt.Errorf("invalid match result")
	}

	if userID != *matchResult.WinnerID && userID != *matchResult.LoserID {
		return fmt.Errorf("user not authorized to confirm this result")
	}

	// 檢查是否已經確認過
	for _, confirmedBy := range matchResult.ConfirmedBy {
		if confirmedBy == userID {
			return fmt.Errorf("user has already confirmed this result")
		}
	}

	// 添加確認
	matchResult.ConfirmedBy = append(matchResult.ConfirmedBy, userID)

	// 如果雙方都確認了，標記為已確認
	if len(matchResult.ConfirmedBy) >= 2 {
		matchResult.IsConfirmed = true
	}

	return mss.db.Save(&matchResult).Error
}

// AutoAdjustSkillLevel 自動調整技術等級
func (mss *MatchStatisticsService) AutoAdjustSkillLevel(userID string) error {
	// 獲取用戶當前等級
	var userProfile models.UserProfile
	err := mss.db.Where("user_id = ?", userID).First(&userProfile).Error
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	if userProfile.NTRPLevel == nil {
		return nil // 沒有設定等級，不進行調整
	}

	currentLevel := *userProfile.NTRPLevel

	// 獲取最近的技術準確度記錄
	var skillRecords []models.SkillAccuracyRecord
	err = mss.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&skillRecords).Error
	if err != nil {
		return fmt.Errorf("failed to get skill records: %w", err)
	}

	if len(skillRecords) < 5 {
		return nil // 記錄不足，不進行調整
	}

	// 計算建議等級
	var totalActualLevel float64
	for _, record := range skillRecords {
		totalActualLevel += record.ActualLevel
	}
	suggestedLevel := totalActualLevel / float64(len(skillRecords))

	// 獲取最近的比賽勝率
	winRate, err := mss.getRecentWinRate(userID, 20) // 最近20場比賽
	if err != nil {
		return fmt.Errorf("failed to get recent win rate: %w", err)
	}

	// 根據勝率調整建議等級
	if winRate > 70 {
		suggestedLevel += 0.2 // 勝率高，建議提升等級
	} else if winRate < 30 {
		suggestedLevel -= 0.2 // 勝率低，建議降低等級
	}

	// 計算等級差異
	levelDiff := suggestedLevel - currentLevel

	// 如果差異超過閾值，進行調整
	if math.Abs(levelDiff) > 0.3 {
		// 逐步調整，避免劇烈變化
		adjustment := levelDiff * 0.3
		newLevel := currentLevel + adjustment

		// 確保等級在合理範圍內
		if newLevel < 1.0 {
			newLevel = 1.0
		} else if newLevel > 7.0 {
			newLevel = 7.0
		}

		// 記錄等級變更
		skillLevelRecord := models.SkillLevelRecord{
			UserID:   userID,
			OldLevel: currentLevel,
			NewLevel: newLevel,
			Reason:   "auto_adjustment",
		}

		if err := mss.db.Create(&skillLevelRecord).Error; err != nil {
			return fmt.Errorf("failed to create skill level record: %w", err)
		}

		// 更新用戶等級
		if err := mss.db.Model(&userProfile).Update("ntrp_level", newLevel).Error; err != nil {
			return fmt.Errorf("failed to update user NTRP level: %w", err)
		}
	}

	return nil
}

// getRecentWinRate 獲取最近的勝率
func (mss *MatchStatisticsService) getRecentWinRate(userID string, matchCount int) (float64, error) {
	// 獲取最近的比賽結果
	var wonMatches int64
	err := mss.db.Table("match_results").
		Joins("JOIN matches ON match_results.match_id = matches.id").
		Where("match_results.winner_id = ? AND matches.status = 'completed'", userID).
		Order("matches.created_at DESC").
		Limit(matchCount).
		Count(&wonMatches).Error
	if err != nil {
		return 0, err
	}

	var lostMatches int64
	err = mss.db.Table("match_results").
		Joins("JOIN matches ON match_results.match_id = matches.id").
		Where("match_results.loser_id = ? AND matches.status = 'completed'", userID).
		Order("matches.created_at DESC").
		Limit(matchCount).
		Count(&lostMatches).Error
	if err != nil {
		return 0, err
	}

	totalMatches := wonMatches + lostMatches
	if totalMatches == 0 {
		return 0, nil
	}

	return float64(wonMatches) / float64(totalMatches) * 100, nil
}

// GetUserPrivacySettings 獲取用戶隱私設定
func (mss *MatchStatisticsService) GetUserPrivacySettings(userID string) (*models.UserPrivacySettings, error) {
	var settings models.UserPrivacySettings

	err := mss.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 創建默認隱私設定
			settings = models.UserPrivacySettings{
				UserID:                 userID,
				ShowReputationScore:    true,
				ShowMatchHistory:       true,
				ShowWinLossRecord:      true,
				ShowSkillProgression:   true,
				ShowBehaviorReviews:    false,
				ShowDetailedStats:      true,
				AllowStatisticsSharing: false,
			}

			if err := mss.db.Create(&settings).Error; err != nil {
				return nil, fmt.Errorf("failed to create privacy settings: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get privacy settings: %w", err)
		}
	}

	return &settings, nil
}

// UpdateUserPrivacySettings 更新用戶隱私設定
func (mss *MatchStatisticsService) UpdateUserPrivacySettings(userID string, settings *models.UserPrivacySettings) error {
	settings.UserID = userID
	settings.UpdatedAt = time.Now()

	return mss.db.Save(settings).Error
}

// FilterStatisticsByPrivacy 根據隱私設定過濾統計資訊
func (mss *MatchStatisticsService) FilterStatisticsByPrivacy(stats *models.MatchStatistics, privacy *models.UserPrivacySettings, isOwner bool) *models.MatchStatistics {
	if isOwner {
		return stats // 用戶自己可以看到所有資訊
	}

	filtered := &models.MatchStatistics{
		UserID: stats.UserID,
	}

	if privacy.ShowMatchHistory {
		filtered.TotalMatches = stats.TotalMatches
		filtered.CompletedMatches = stats.CompletedMatches
		filtered.AttendanceRate = stats.AttendanceRate
		filtered.RecentMatches = stats.RecentMatches
	}

	if privacy.ShowWinLossRecord {
		filtered.WonMatches = stats.WonMatches
		filtered.LostMatches = stats.LostMatches
		filtered.WinRate = stats.WinRate
	}

	if privacy.ShowSkillProgression {
		filtered.SkillLevelProgression = stats.SkillLevelProgression
	}

	if privacy.ShowDetailedStats {
		filtered.AverageMatchDuration = stats.AverageMatchDuration
		filtered.FavoriteCourtType = stats.FavoriteCourtType
		filtered.MostPlayedWith = stats.MostPlayedWith
		filtered.MonthlyStats = stats.MonthlyStats
	}

	return filtered
}
