package services

import (
	"fmt"
	"math"
	"tennis-platform/backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// ReputationService 信譽評分服務
type ReputationService struct {
	db *gorm.DB
}

// NewReputationService 創建新的信譽評分服務
func NewReputationService(db *gorm.DB) *ReputationService {
	return &ReputationService{
		db: db,
	}
}

// ReputationWeights 信譽評分權重配置
type ReputationWeights struct {
	AttendanceRate   float64 // 出席率權重
	PunctualityScore float64 // 準時度權重
	SkillAccuracy    float64 // 技術等級準確度權重
	BehaviorRating   float64 // 行為評分權重
}

// DefaultWeights 默認權重配置
var DefaultWeights = ReputationWeights{
	AttendanceRate:   0.3,
	PunctualityScore: 0.2,
	SkillAccuracy:    0.2,
	BehaviorRating:   0.3,
}

// GetOrCreateReputationScore 獲取或創建用戶信譽分數
func (rs *ReputationService) GetOrCreateReputationScore(userID string) (*models.ReputationScore, error) {
	var reputation models.ReputationScore

	err := rs.db.Where("user_id = ?", userID).First(&reputation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 創建新的信譽記錄
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
				UpdatedAt:        time.Now(),
			}

			if err := rs.db.Create(&reputation).Error; err != nil {
				return nil, fmt.Errorf("failed to create reputation score: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get reputation score: %w", err)
		}
	}

	return &reputation, nil
}

// UpdateAttendanceRate 更新出席率
func (rs *ReputationService) UpdateAttendanceRate(userID string, matchStatus string) error {
	reputation, err := rs.GetOrCreateReputationScore(userID)
	if err != nil {
		return err
	}

	// 更新比賽統計
	reputation.TotalMatches++

	switch matchStatus {
	case "completed":
		reputation.CompletedMatches++
	case "cancelled":
		reputation.CancelledMatches++
	}

	// 計算出席率 (完成的比賽 / 總比賽數)
	if reputation.TotalMatches > 0 {
		reputation.AttendanceRate = float64(reputation.CompletedMatches) / float64(reputation.TotalMatches) * 100
	}

	// 重新計算綜合分數
	rs.calculateOverallScore(reputation)

	return rs.db.Save(reputation).Error
}

// UpdatePunctualityScore 更新準時度評分
func (rs *ReputationService) UpdatePunctualityScore(userID string, isOnTime bool, delayMinutes int) error {
	reputation, err := rs.GetOrCreateReputationScore(userID)
	if err != nil {
		return err
	}

	// 獲取用戶最近的準時記錄
	var punctualityRecords []models.PunctualityRecord
	err = rs.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10). // 取最近10次記錄
		Find(&punctualityRecords).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to get punctuality records: %w", err)
	}

	// 創建新的準時記錄
	newRecord := models.PunctualityRecord{
		UserID:       userID,
		IsOnTime:     isOnTime,
		DelayMinutes: delayMinutes,
		CreatedAt:    time.Now(),
	}

	if err := rs.db.Create(&newRecord).Error; err != nil {
		return fmt.Errorf("failed to create punctuality record: %w", err)
	}

	// 重新獲取最新的記錄
	punctualityRecords = append([]models.PunctualityRecord{newRecord}, punctualityRecords...)
	if len(punctualityRecords) > 10 {
		punctualityRecords = punctualityRecords[:10]
	}

	// 計算準時度分數
	reputation.PunctualityScore = rs.calculatePunctualityScore(punctualityRecords)

	// 重新計算綜合分數
	rs.calculateOverallScore(reputation)

	return rs.db.Save(reputation).Error
}

// UpdateSkillAccuracy 更新技術等級準確度
func (rs *ReputationService) UpdateSkillAccuracy(userID string, reportedLevel, actualLevel float64) error {
	reputation, err := rs.GetOrCreateReputationScore(userID)
	if err != nil {
		return err
	}

	// 獲取用戶的技術等級記錄
	var skillRecords []models.SkillAccuracyRecord
	err = rs.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10). // 取最近10次記錄
		Find(&skillRecords).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to get skill accuracy records: %w", err)
	}

	// 創建新的技術準確度記錄
	newRecord := models.SkillAccuracyRecord{
		UserID:        userID,
		ReportedLevel: reportedLevel,
		ActualLevel:   actualLevel,
		Accuracy:      rs.calculateLevelAccuracy(reportedLevel, actualLevel),
		CreatedAt:     time.Now(),
	}

	if err := rs.db.Create(&newRecord).Error; err != nil {
		return fmt.Errorf("failed to create skill accuracy record: %w", err)
	}

	// 重新獲取最新的記錄
	skillRecords = append([]models.SkillAccuracyRecord{newRecord}, skillRecords...)
	if len(skillRecords) > 10 {
		skillRecords = skillRecords[:10]
	}

	// 計算技術準確度分數
	reputation.SkillAccuracy = rs.calculateSkillAccuracyScore(skillRecords)

	// 重新計算綜合分數
	rs.calculateOverallScore(reputation)

	return rs.db.Save(reputation).Error
}

// UpdateBehaviorRating 更新行為評分
func (rs *ReputationService) UpdateBehaviorRating(userID string, rating float64, reviewerID string) error {
	reputation, err := rs.GetOrCreateReputationScore(userID)
	if err != nil {
		return err
	}

	// 創建行為評價記錄
	behaviorReview := models.BehaviorReview{
		UserID:     userID,
		ReviewerID: reviewerID,
		Rating:     rating,
		CreatedAt:  time.Now(),
	}

	if err := rs.db.Create(&behaviorReview).Error; err != nil {
		return fmt.Errorf("failed to create behavior review: %w", err)
	}

	// 獲取用戶最近的行為評價
	var behaviorReviews []models.BehaviorReview
	err = rs.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(20). // 取最近20次評價
		Find(&behaviorReviews).Error

	if err != nil {
		return fmt.Errorf("failed to get behavior reviews: %w", err)
	}

	// 計算平均行為評分
	if len(behaviorReviews) > 0 {
		var totalRating float64
		for _, review := range behaviorReviews {
			totalRating += review.Rating
		}
		reputation.BehaviorRating = totalRating / float64(len(behaviorReviews))
	}

	// 重新計算綜合分數
	rs.calculateOverallScore(reputation)

	return rs.db.Save(reputation).Error
}

// calculateOverallScore 計算綜合信譽分數
func (rs *ReputationService) calculateOverallScore(reputation *models.ReputationScore) {
	weights := DefaultWeights

	// 將行為評分轉換為百分制 (1-5 -> 0-100)
	behaviorScore := (reputation.BehaviorRating - 1) / 4 * 100

	// 計算加權平均分
	reputation.OverallScore = reputation.AttendanceRate*weights.AttendanceRate +
		reputation.PunctualityScore*weights.PunctualityScore +
		reputation.SkillAccuracy*weights.SkillAccuracy +
		behaviorScore*weights.BehaviorRating

	// 確保分數在 0-100 範圍內
	if reputation.OverallScore < 0 {
		reputation.OverallScore = 0
	} else if reputation.OverallScore > 100 {
		reputation.OverallScore = 100
	}

	reputation.UpdatedAt = time.Now()
}

// calculatePunctualityScore 計算準時度分數
func (rs *ReputationService) calculatePunctualityScore(records []models.PunctualityRecord) float64 {
	if len(records) == 0 {
		return 100.0
	}

	var totalScore float64
	for _, record := range records {
		if record.IsOnTime {
			totalScore += 100
		} else {
			// 根據遲到時間計算扣分
			penalty := float64(record.DelayMinutes) * 2 // 每分鐘扣2分
			score := 100 - penalty
			if score < 0 {
				score = 0
			}
			totalScore += score
		}
	}

	return totalScore / float64(len(records))
}

// calculateSkillAccuracyScore 計算技術準確度分數
func (rs *ReputationService) calculateSkillAccuracyScore(records []models.SkillAccuracyRecord) float64 {
	if len(records) == 0 {
		return 100.0
	}

	var totalAccuracy float64
	for _, record := range records {
		totalAccuracy += record.Accuracy
	}

	return totalAccuracy / float64(len(records))
}

// calculateLevelAccuracy 計算等級準確度
func (rs *ReputationService) calculateLevelAccuracy(reported, actual float64) float64 {
	diff := math.Abs(reported - actual)

	// 差距越小，準確度越高
	// 0差距 = 100分，1.0差距 = 50分，2.0差距 = 0分
	accuracy := 100 - (diff * 50)
	if accuracy < 0 {
		accuracy = 0
	}

	return accuracy
}

// GetReputationScore 獲取用戶信譽分數
func (rs *ReputationService) GetReputationScore(userID string) (*models.ReputationScore, error) {
	var reputation models.ReputationScore
	err := rs.db.Where("user_id = ?", userID).First(&reputation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果沒有記錄，創建默認記錄
			return rs.GetOrCreateReputationScore(userID)
		}
		return nil, fmt.Errorf("failed to get reputation score: %w", err)
	}

	return &reputation, nil
}

// GetReputationHistory 獲取信譽歷史記錄
func (rs *ReputationService) GetReputationHistory(userID string) (*models.ReputationHistory, error) {
	// 獲取準時記錄
	var punctualityRecords []models.PunctualityRecord
	rs.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&punctualityRecords)

	// 獲取技術準確度記錄
	var skillRecords []models.SkillAccuracyRecord
	rs.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&skillRecords)

	// 獲取行為評價記錄
	var behaviorReviews []models.BehaviorReview
	rs.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(20).
		Find(&behaviorReviews)

	return &models.ReputationHistory{
		UserID:             userID,
		PunctualityRecords: punctualityRecords,
		SkillRecords:       skillRecords,
		BehaviorReviews:    behaviorReviews,
	}, nil
}

// RecalculateAllScores 重新計算所有用戶的信譽分數
func (rs *ReputationService) RecalculateAllScores() error {
	var reputations []models.ReputationScore
	if err := rs.db.Find(&reputations).Error; err != nil {
		return fmt.Errorf("failed to get all reputation scores: %w", err)
	}

	for _, reputation := range reputations {
		rs.calculateOverallScore(&reputation)
		if err := rs.db.Save(&reputation).Error; err != nil {
			return fmt.Errorf("failed to update reputation score for user %s: %w", reputation.UserID, err)
		}
	}

	return nil
}
