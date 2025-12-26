package usecases

import (
	"fmt"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"time"

	"gorm.io/gorm"
)

// ReputationUseCase 信譽評分用例
type ReputationUseCase struct {
	db                *gorm.DB
	reputationService *services.ReputationService
}

// NewReputationUseCase 創建新的信譽評分用例
func NewReputationUseCase(db *gorm.DB) *ReputationUseCase {
	return &ReputationUseCase{
		db:                db,
		reputationService: services.NewReputationService(db),
	}
}

// GetUserReputationScore 獲取用戶信譽分數
func (ruc *ReputationUseCase) GetUserReputationScore(userID string) (*models.ReputationScore, error) {
	return ruc.reputationService.GetReputationScore(userID)
}

// GetUserReputationHistory 獲取用戶信譽歷史記錄
func (ruc *ReputationUseCase) GetUserReputationHistory(userID string) (*models.ReputationHistory, error) {
	return ruc.reputationService.GetReputationHistory(userID)
}

// RecordMatchAttendance 記錄比賽出席情況
func (ruc *ReputationUseCase) RecordMatchAttendance(userID, matchID, status string) error {
	// 驗證比賽存在且用戶是參與者
	var match models.Match
	err := ruc.db.Preload("Participants").Where("id = ?", matchID).First(&match).Error
	if err != nil {
		return fmt.Errorf("match not found: %w", err)
	}

	// 檢查用戶是否是比賽參與者
	isParticipant := false
	for _, participant := range match.Participants {
		if participant.ID == userID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		return fmt.Errorf("user is not a participant of this match")
	}

	// 更新出席率
	return ruc.reputationService.UpdateAttendanceRate(userID, status)
}

// RecordMatchPunctuality 記錄比賽準時情況
func (ruc *ReputationUseCase) RecordMatchPunctuality(userID, matchID string, arrivalTime time.Time) error {
	// 獲取比賽信息
	var match models.Match
	err := ruc.db.Where("id = ?", matchID).First(&match).Error
	if err != nil {
		return fmt.Errorf("match not found: %w", err)
	}

	if match.ScheduledAt == nil {
		return fmt.Errorf("match has no scheduled time")
	}

	// 計算是否準時和遲到時間
	scheduledTime := *match.ScheduledAt
	isOnTime := arrivalTime.Before(scheduledTime) || arrivalTime.Equal(scheduledTime)
	delayMinutes := 0

	if !isOnTime {
		delayMinutes = int(arrivalTime.Sub(scheduledTime).Minutes())
	}

	// 更新準時度評分
	return ruc.reputationService.UpdatePunctualityScore(userID, isOnTime, delayMinutes)
}

// RecordSkillLevelAccuracy 記錄技術等級準確度
func (ruc *ReputationUseCase) RecordSkillLevelAccuracy(userID, matchID string, reportedLevel, observedLevel float64) error {
	// 驗證等級範圍
	if reportedLevel < 1.0 || reportedLevel > 7.0 || observedLevel < 1.0 || observedLevel > 7.0 {
		return fmt.Errorf("NTRP level must be between 1.0 and 7.0")
	}

	// 驗證比賽存在
	var match models.Match
	err := ruc.db.Where("id = ?", matchID).First(&match).Error
	if err != nil {
		return fmt.Errorf("match not found: %w", err)
	}

	// 更新技術等級準確度
	return ruc.reputationService.UpdateSkillAccuracy(userID, reportedLevel, observedLevel)
}

// SubmitBehaviorReview 提交行為評價
func (ruc *ReputationUseCase) SubmitBehaviorReview(reviewerID, userID, matchID string, rating float64, comment *string, tags []string) error {
	// 驗證評分範圍
	if rating < 1.0 || rating > 5.0 {
		return fmt.Errorf("rating must be between 1.0 and 5.0")
	}

	// 驗證評價者和被評價者不是同一人
	if reviewerID == userID {
		return fmt.Errorf("cannot review yourself")
	}

	// 驗證比賽存在且雙方都是參與者
	if matchID != "" {
		var match models.Match
		err := ruc.db.Preload("Participants").Where("id = ?", matchID).First(&match).Error
		if err != nil {
			return fmt.Errorf("match not found: %w", err)
		}

		reviewerIsParticipant := false
		userIsParticipant := false

		for _, participant := range match.Participants {
			if participant.ID == reviewerID {
				reviewerIsParticipant = true
			}
			if participant.ID == userID {
				userIsParticipant = true
			}
		}

		if !reviewerIsParticipant || !userIsParticipant {
			return fmt.Errorf("both users must be participants of the match")
		}
	}

	// 檢查是否已經評價過
	var existingReview models.BehaviorReview
	err := ruc.db.Where("reviewer_id = ? AND user_id = ? AND match_id = ?", reviewerID, userID, matchID).First(&existingReview).Error
	if err == nil {
		return fmt.Errorf("you have already reviewed this user for this match")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing review: %w", err)
	}

	// 更新行為評分
	return ruc.reputationService.UpdateBehaviorRating(userID, rating, reviewerID)
}

// GetReputationLeaderboard 獲取信譽排行榜
func (ruc *ReputationUseCase) GetReputationLeaderboard(limit int) ([]models.ReputationScore, error) {
	var reputations []models.ReputationScore

	err := ruc.db.Preload("User").
		Preload("User.Profile").
		Where("total_matches >= ?", 5). // 至少參與5場比賽才能上榜
		Order("overall_score DESC").
		Limit(limit).
		Find(&reputations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get reputation leaderboard: %w", err)
	}

	return reputations, nil
}

// GetReputationStats 獲取信譽統計信息
func (ruc *ReputationUseCase) GetReputationStats() (*ReputationStats, error) {
	var stats ReputationStats

	// 總用戶數
	err := ruc.db.Model(&models.ReputationScore{}).Count(&stats.TotalUsers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// 平均信譽分數
	err = ruc.db.Model(&models.ReputationScore{}).
		Select("AVG(overall_score)").
		Scan(&stats.AverageScore).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average score: %w", err)
	}

	// 高信譽用戶數 (分數 >= 80)
	err = ruc.db.Model(&models.ReputationScore{}).
		Where("overall_score >= ?", 80).
		Count(&stats.HighReputationUsers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count high reputation users: %w", err)
	}

	// 活躍用戶數 (最近30天有比賽記錄)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = ruc.db.Model(&models.ReputationScore{}).
		Where("updated_at >= ?", thirtyDaysAgo).
		Count(&stats.ActiveUsers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	return &stats, nil
}

// UpdateUserNTRPLevel 更新用戶NTRP等級（基於信譽系統建議）
func (ruc *ReputationUseCase) UpdateUserNTRPLevel(userID string) error {
	// 獲取用戶的技術準確度記錄
	var skillRecords []models.SkillAccuracyRecord
	err := ruc.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&skillRecords).Error

	if err != nil {
		return fmt.Errorf("failed to get skill records: %w", err)
	}

	if len(skillRecords) < 3 {
		// 記錄不足，不進行調整
		return nil
	}

	// 計算建議的NTRP等級
	var totalActualLevel float64
	for _, record := range skillRecords {
		totalActualLevel += record.ActualLevel
	}
	suggestedLevel := totalActualLevel / float64(len(skillRecords))

	// 獲取用戶當前等級
	var userProfile models.UserProfile
	err = ruc.db.Where("user_id = ?", userID).First(&userProfile).Error
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	// 如果建議等級與當前等級差距較大，更新等級
	if userProfile.NTRPLevel != nil {
		currentLevel := *userProfile.NTRPLevel
		levelDiff := suggestedLevel - currentLevel

		// 如果差距超過0.5，進行調整
		if levelDiff > 0.5 || levelDiff < -0.5 {
			newLevel := currentLevel + (levelDiff * 0.3) // 逐步調整，避免劇烈變化

			// 確保等級在合理範圍內
			if newLevel < 1.0 {
				newLevel = 1.0
			} else if newLevel > 7.0 {
				newLevel = 7.0
			}

			// 更新用戶等級
			err = ruc.db.Model(&userProfile).Update("ntrp_level", newLevel).Error
			if err != nil {
				return fmt.Errorf("failed to update user NTRP level: %w", err)
			}
		}
	}

	return nil
}

// ReputationStats 信譽統計信息
type ReputationStats struct {
	TotalUsers          int64   `json:"totalUsers"`
	AverageScore        float64 `json:"averageScore"`
	HighReputationUsers int64   `json:"highReputationUsers"`
	ActiveUsers         int64   `json:"activeUsers"`
}
