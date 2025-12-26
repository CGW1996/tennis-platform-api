package services

import (
	"errors"
	"math"
	"sort"
	"tennis-platform/backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// IntelligentSchedulingService 智能排課服務
type IntelligentSchedulingService struct {
	db *gorm.DB
}

// NewIntelligentSchedulingService 創建新的智能排課服務
func NewIntelligentSchedulingService(db *gorm.DB) *IntelligentSchedulingService {
	return &IntelligentSchedulingService{
		db: db,
	}
}

// SchedulingConstraints 排課約束條件
type SchedulingConstraints struct {
	CoachAvailability   []models.TimeSlot `json:"coachAvailability"`
	StudentAvailability []models.TimeSlot `json:"studentAvailability"`
	SkillLevelMatch     bool              `json:"skillLevelMatch"`
	LocationPreference  string            `json:"locationPreference"`
	LessonType          string            `json:"lessonType"`    // individual, group
	MaxDistance         float64           `json:"maxDistance"`   // 最大距離（公里）
	PreferredDays       []int             `json:"preferredDays"` // 偏好的星期幾
	MinPrice            *float64          `json:"minPrice"`
	MaxPrice            *float64          `json:"maxPrice"`
}

// RecommendedLesson 推薦課程
type RecommendedLesson struct {
	CoachID      string                    `json:"coachId"`
	Coach        *models.Coach             `json:"coach"`
	TimeSlot     SchedulingTimeSlot        `json:"timeSlot"`
	MatchScore   float64                   `json:"matchScore"`
	Price        float64                   `json:"price"`
	Location     string                    `json:"location"`
	Distance     float64                   `json:"distance"`
	MatchFactors SchedulingMatchingFactors `json:"matchFactors"`
	LessonTypeID *string                   `json:"lessonTypeId"`
	LessonType   *models.LessonType        `json:"lessonType"`
}

// SchedulingTimeSlot 排課時間段
type SchedulingTimeSlot struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	DayOfWeek int       `json:"dayOfWeek"`
}

// SchedulingMatchingFactors 排課匹配因子
type SchedulingMatchingFactors struct {
	SkillLevel        float64 `json:"skillLevel"`        // 技術等級匹配度 (0-1)
	TimeCompatibility float64 `json:"timeCompatibility"` // 時間相容性 (0-1)
	LocationScore     float64 `json:"locationScore"`     // 位置評分 (0-1)
	PriceScore        float64 `json:"priceScore"`        // 價格評分 (0-1)
	ExperienceScore   float64 `json:"experienceScore"`   // 經驗評分 (0-1)
	RatingScore       float64 `json:"ratingScore"`       // 評分評分 (0-1)
}

// StudentPreferences 學生偏好
type StudentPreferences struct {
	UserID              string    `json:"userId"`
	NTRPLevel           float64   `json:"ntrpLevel"`
	PreferredTimes      []string  `json:"preferredTimes"` // ["09:00-12:00", "14:00-18:00"]
	PreferredDays       []int     `json:"preferredDays"`  // [1,2,3,4,5] (Monday-Friday)
	MaxDistance         float64   `json:"maxDistance"`    // 公里
	MinPrice            *float64  `json:"minPrice"`
	MaxPrice            *float64  `json:"maxPrice"`
	PreferredLessonType string    `json:"preferredLessonType"` // individual, group
	Location            *Location `json:"location"`
}

// Location 位置信息
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

// RecommendCoaches 推薦教練和課程時間
func (iss *IntelligentSchedulingService) RecommendCoaches(studentPrefs *StudentPreferences, dateRange []string) ([]RecommendedLesson, error) {
	// 1. 獲取符合基本條件的教練
	coaches, err := iss.getEligibleCoaches(studentPrefs)
	if err != nil {
		return nil, err
	}

	if len(coaches) == 0 {
		return []RecommendedLesson{}, nil
	}

	var recommendations []RecommendedLesson

	// 2. 為每個教練計算推薦課程
	for _, coach := range coaches {
		coachRecommendations, err := iss.generateCoachRecommendations(&coach, studentPrefs, dateRange)
		if err != nil {
			continue // 跳過有錯誤的教練
		}
		recommendations = append(recommendations, coachRecommendations...)
	}

	// 3. 排序推薦結果
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].MatchScore > recommendations[j].MatchScore
	})

	// 4. 限制返回數量
	maxRecommendations := 20
	if len(recommendations) > maxRecommendations {
		recommendations = recommendations[:maxRecommendations]
	}

	return recommendations, nil
}

// getEligibleCoaches 獲取符合條件的教練
func (iss *IntelligentSchedulingService) getEligibleCoaches(studentPrefs *StudentPreferences) ([]models.Coach, error) {
	query := iss.db.Model(&models.Coach{}).
		Preload("User").
		Preload("User.Profile").
		Where("is_active = ? AND is_verified = ?", true, true)

	// 技術等級篩選 - 教練應該能教導學生的等級
	if studentPrefs.NTRPLevel > 0 {
		// 教練的專長應該包含學生的等級
		level := iss.getNTRPLevelCategory(studentPrefs.NTRPLevel)
		query = query.Where("specialties && ?", []string{level})
	}

	// 價格篩選
	if studentPrefs.MinPrice != nil || studentPrefs.MaxPrice != nil {
		subQuery := iss.db.Model(&models.LessonType{}).
			Select("coach_id").
			Where("is_active = ?", true)

		if studentPrefs.MinPrice != nil {
			subQuery = subQuery.Where("price >= ?", *studentPrefs.MinPrice)
		}
		if studentPrefs.MaxPrice != nil {
			subQuery = subQuery.Where("price <= ?", *studentPrefs.MaxPrice)
		}

		query = query.Where("id IN (?)", subQuery)
	}

	var coaches []models.Coach
	if err := query.Find(&coaches).Error; err != nil {
		return nil, errors.New("獲取教練列表失敗")
	}

	// 地理位置篩選
	if studentPrefs.Location != nil && studentPrefs.MaxDistance > 0 {
		coaches = iss.filterCoachesByDistance(coaches, studentPrefs.Location, studentPrefs.MaxDistance)
	}

	return coaches, nil
}

// generateCoachRecommendations 為特定教練生成推薦
func (iss *IntelligentSchedulingService) generateCoachRecommendations(coach *models.Coach, studentPrefs *StudentPreferences, dateRange []string) ([]RecommendedLesson, error) {
	var recommendations []RecommendedLesson

	// 獲取教練的時間表
	var schedules []models.LessonSchedule
	if err := iss.db.Where("coach_id = ? AND is_active = ?", coach.ID, true).Find(&schedules).Error; err != nil {
		return nil, err
	}

	// 為每個日期生成推薦
	for _, dateStr := range dateRange {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		dayOfWeek := int(date.Weekday())

		// 檢查是否是學生偏好的日期
		if len(studentPrefs.PreferredDays) > 0 && !iss.containsInt(studentPrefs.PreferredDays, dayOfWeek) {
			continue
		}

		// 獲取該日教練的可用時間
		availableSlots, err := iss.getCoachAvailabilityForDate(coach.ID, date)
		if err != nil {
			continue
		}

		// 篩選符合學生時間偏好的時間段
		matchingSlots := iss.filterSlotsByStudentPreferences(availableSlots, studentPrefs.PreferredTimes)

		// 獲取教練的課程類型
		var lessonTypes []models.LessonType
		if err := iss.db.Where("coach_id = ? AND is_active = ?", coach.ID, true).Find(&lessonTypes).Error; err != nil {
			continue
		}

		// 為每個匹配的時間段和課程類型生成推薦
		for _, slot := range matchingSlots {
			for _, lessonType := range lessonTypes {
				if studentPrefs.PreferredLessonType != "" && lessonType.Type != studentPrefs.PreferredLessonType {
					continue
				}

				recommendation := iss.createRecommendation(coach, &lessonType, slot, date, studentPrefs)
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	return recommendations, nil
}

// createRecommendation 創建推薦
func (iss *IntelligentSchedulingService) createRecommendation(coach *models.Coach, lessonType *models.LessonType, slot models.TimeSlot, date time.Time, studentPrefs *StudentPreferences) RecommendedLesson {
	// 解析時間
	startTime, _ := time.Parse("15:04", slot.StartTime)
	endTime, _ := time.Parse("15:04", slot.EndTime)

	// 創建完整的時間戳
	start := time.Date(date.Year(), date.Month(), date.Day(), startTime.Hour(), startTime.Minute(), 0, 0, date.Location())
	end := time.Date(date.Year(), date.Month(), date.Day(), endTime.Hour(), endTime.Minute(), 0, 0, date.Location())

	// 計算匹配因子
	factors := iss.calculateMatchingFactors(coach, lessonType, studentPrefs)

	// 計算總體匹配分數
	matchScore := iss.calculateOverallMatchScore(factors)

	// 計算距離
	distance := 0.0
	location := "未指定"
	if studentPrefs.Location != nil && coach.User != nil && coach.User.Profile != nil {
		// 這裡應該從用戶檔案中獲取位置信息
		// 暫時使用默認值
		location = "台北市"
	}

	return RecommendedLesson{
		CoachID: coach.ID,
		Coach:   coach,
		TimeSlot: SchedulingTimeSlot{
			Start:     start,
			End:       end,
			DayOfWeek: int(date.Weekday()),
		},
		MatchScore:   matchScore,
		Price:        lessonType.Price,
		Location:     location,
		Distance:     distance,
		MatchFactors: factors,
		LessonTypeID: &lessonType.ID,
		LessonType:   lessonType,
	}
}

// calculateMatchingFactors 計算匹配因子
func (iss *IntelligentSchedulingService) calculateMatchingFactors(coach *models.Coach, lessonType *models.LessonType, studentPrefs *StudentPreferences) SchedulingMatchingFactors {
	factors := SchedulingMatchingFactors{}

	// 1. 技術等級匹配度
	factors.SkillLevel = iss.calculateSkillLevelMatch(coach, studentPrefs.NTRPLevel)

	// 2. 時間相容性（這裡簡化處理，實際應該更複雜）
	factors.TimeCompatibility = 0.8 // 假設時間匹配度較高

	// 3. 位置評分
	factors.LocationScore = iss.calculateLocationScore(coach, studentPrefs.Location, studentPrefs.MaxDistance)

	// 4. 價格評分
	factors.PriceScore = iss.calculatePriceScore(lessonType.Price, studentPrefs.MinPrice, studentPrefs.MaxPrice)

	// 5. 經驗評分
	factors.ExperienceScore = iss.calculateExperienceScore(coach.Experience)

	// 6. 評分評分
	factors.RatingScore = iss.calculateRatingScore(coach.AverageRating)

	return factors
}

// calculateSkillLevelMatch 計算技術等級匹配度
func (iss *IntelligentSchedulingService) calculateSkillLevelMatch(coach *models.Coach, studentNTRP float64) float64 {
	if studentNTRP <= 0 {
		return 0.5 // 默認中等匹配度
	}

	studentLevel := iss.getNTRPLevelCategory(studentNTRP)

	// 檢查教練是否有對應的專長
	for _, specialty := range coach.Specialties {
		if specialty == studentLevel {
			return 1.0 // 完美匹配
		}
	}

	// 檢查相鄰等級
	if studentLevel == "intermediate" {
		for _, specialty := range coach.Specialties {
			if specialty == "beginner" || specialty == "advanced" {
				return 0.7 // 較好匹配
			}
		}
	}

	return 0.3 // 基本匹配
}

// calculateLocationScore 計算位置評分
func (iss *IntelligentSchedulingService) calculateLocationScore(coach *models.Coach, studentLocation *Location, maxDistance float64) float64 {
	if studentLocation == nil || maxDistance <= 0 {
		return 0.5 // 默認中等評分
	}

	// 這裡應該計算實際距離，暫時返回默認值
	// 實際實現需要從教練檔案中獲取位置信息並計算距離
	return 0.8
}

// calculatePriceScore 計算價格評分
func (iss *IntelligentSchedulingService) calculatePriceScore(price float64, minPrice, maxPrice *float64) float64 {
	if minPrice == nil && maxPrice == nil {
		return 0.5 // 無價格偏好
	}

	score := 1.0

	if minPrice != nil && price < *minPrice {
		score *= 0.5 // 價格太低可能品質有問題
	}

	if maxPrice != nil && price > *maxPrice {
		// 超出預算，評分降低
		overBudget := (price - *maxPrice) / *maxPrice
		score *= math.Max(0.1, 1.0-overBudget)
	}

	return score
}

// calculateExperienceScore 計算經驗評分
func (iss *IntelligentSchedulingService) calculateExperienceScore(experience int) float64 {
	if experience <= 0 {
		return 0.2
	}
	if experience >= 10 {
		return 1.0
	}
	return float64(experience) / 10.0
}

// calculateRatingScore 計算評分評分
func (iss *IntelligentSchedulingService) calculateRatingScore(rating float64) float64 {
	if rating <= 0 {
		return 0.5 // 無評分時給予中等分數
	}
	return rating / 5.0 // 假設最高評分是5
}

// calculateOverallMatchScore 計算總體匹配分數
func (iss *IntelligentSchedulingService) calculateOverallMatchScore(factors SchedulingMatchingFactors) float64 {
	// 權重設定
	weights := map[string]float64{
		"skillLevel":        0.25, // 技術等級最重要
		"timeCompatibility": 0.20, // 時間相容性
		"locationScore":     0.15, // 位置
		"priceScore":        0.15, // 價格
		"experienceScore":   0.15, // 經驗
		"ratingScore":       0.10, // 評分
	}

	score := factors.SkillLevel*weights["skillLevel"] +
		factors.TimeCompatibility*weights["timeCompatibility"] +
		factors.LocationScore*weights["locationScore"] +
		factors.PriceScore*weights["priceScore"] +
		factors.ExperienceScore*weights["experienceScore"] +
		factors.RatingScore*weights["ratingScore"]

	return math.Min(1.0, math.Max(0.0, score)) // 確保分數在0-1之間
}

// getNTRPLevelCategory 獲取NTRP等級分類
func (iss *IntelligentSchedulingService) getNTRPLevelCategory(ntrp float64) string {
	if ntrp <= 2.5 {
		return "beginner"
	} else if ntrp <= 4.0 {
		return "intermediate"
	} else {
		return "advanced"
	}
}

// filterCoachesByDistance 根據距離篩選教練
func (iss *IntelligentSchedulingService) filterCoachesByDistance(coaches []models.Coach, studentLocation *Location, maxDistance float64) []models.Coach {
	// 這裡應該實現實際的地理距離計算
	// 暫時返回所有教練
	return coaches
}

// getCoachAvailabilityForDate 獲取教練在特定日期的可用時間
func (iss *IntelligentSchedulingService) getCoachAvailabilityForDate(coachID string, date time.Time) ([]models.TimeSlot, error) {
	dayOfWeek := int(date.Weekday())

	// 獲取教練的時間表
	var schedules []models.LessonSchedule
	if err := iss.db.Where("coach_id = ? AND day_of_week = ? AND is_active = ?", coachID, dayOfWeek, true).Find(&schedules).Error; err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return []models.TimeSlot{}, nil
	}

	// 獲取當天已預訂的課程
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var bookedLessons []models.Lesson
	if err := iss.db.Where("coach_id = ? AND scheduled_at >= ? AND scheduled_at < ? AND status IN ?",
		coachID, startOfDay, endOfDay, []string{"scheduled", "in_progress"}).Find(&bookedLessons).Error; err != nil {
		return nil, err
	}

	// 生成可用時間段
	var availableSlots []models.TimeSlot
	for _, schedule := range schedules {
		slots := iss.generateTimeSlots(schedule.StartTime, schedule.EndTime, 60) // 60分鐘間隔
		for _, slot := range slots {
			isBooked := iss.isTimeSlotBooked(slot, bookedLessons, date)
			if !isBooked {
				availableSlots = append(availableSlots, slot)
			}
		}
	}

	return availableSlots, nil
}

// filterSlotsByStudentPreferences 根據學生時間偏好篩選時間段
func (iss *IntelligentSchedulingService) filterSlotsByStudentPreferences(slots []models.TimeSlot, preferredTimes []string) []models.TimeSlot {
	if len(preferredTimes) == 0 {
		return slots // 無偏好時返回所有時間段
	}

	var matchingSlots []models.TimeSlot
	for _, slot := range slots {
		for _, prefTime := range preferredTimes {
			if iss.isTimeInRange(slot.StartTime, prefTime) {
				matchingSlots = append(matchingSlots, slot)
				break
			}
		}
	}

	return matchingSlots
}

// isTimeInRange 檢查時間是否在範圍內
func (iss *IntelligentSchedulingService) isTimeInRange(timeStr, rangeStr string) bool {
	// 解析範圍 "09:00-12:00"
	if len(rangeStr) != 11 || rangeStr[5] != '-' {
		return false
	}

	rangeStart := rangeStr[:5]
	rangeEnd := rangeStr[6:]

	return timeStr >= rangeStart && timeStr <= rangeEnd
}

// generateTimeSlots 生成時間段
func (iss *IntelligentSchedulingService) generateTimeSlots(startTime, endTime string, intervalMinutes int) []models.TimeSlot {
	var slots []models.TimeSlot

	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)

	current := start
	for current.Before(end) {
		next := current.Add(time.Duration(intervalMinutes) * time.Minute)
		if next.After(end) {
			break
		}

		slots = append(slots, models.TimeSlot{
			StartTime: current.Format("15:04"),
			EndTime:   next.Format("15:04"),
			IsBooked:  false,
		})

		current = next
	}

	return slots
}

// isTimeSlotBooked 檢查時間段是否已預訂
func (iss *IntelligentSchedulingService) isTimeSlotBooked(slot models.TimeSlot, bookedLessons []models.Lesson, targetDate time.Time) bool {
	slotStart, _ := time.Parse("15:04", slot.StartTime)
	slotEnd, _ := time.Parse("15:04", slot.EndTime)

	// 將時間段轉換為目標日期的完整時間
	slotStartTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		slotStart.Hour(), slotStart.Minute(), 0, 0, targetDate.Location())
	slotEndTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		slotEnd.Hour(), slotEnd.Minute(), 0, 0, targetDate.Location())

	for _, lesson := range bookedLessons {
		lessonEnd := lesson.ScheduledAt.Add(time.Duration(lesson.Duration) * time.Minute)

		// 檢查時間重疊
		if lesson.ScheduledAt.Before(slotEndTime) && lessonEnd.After(slotStartTime) {
			return true
		}
	}

	return false
}

// containsInt 檢查整數數組是否包含特定值
func (iss *IntelligentSchedulingService) containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FindOptimalLessonTime 尋找最佳課程時間
func (iss *IntelligentSchedulingService) FindOptimalLessonTime(coachID, studentID string, preferences *StudentPreferences, dateRange []string) (*RecommendedLesson, error) {
	// 獲取教練信息
	var coach models.Coach
	if err := iss.db.Preload("User").Preload("User.Profile").
		Where("id = ? AND is_active = ? AND is_verified = ?", coachID, true, true).First(&coach).Error; err != nil {
		return nil, errors.New("教練不存在或未認證")
	}

	// 生成推薦
	recommendations, err := iss.generateCoachRecommendations(&coach, preferences, dateRange)
	if err != nil {
		return nil, err
	}

	if len(recommendations) == 0 {
		return nil, errors.New("未找到合適的課程時間")
	}

	// 返回最佳推薦
	return &recommendations[0], nil
}

// DetectSchedulingConflicts 檢測排課衝突
func (iss *IntelligentSchedulingService) DetectSchedulingConflicts(coachID string, scheduledAt time.Time, duration int, excludeLessonID *string) ([]models.Lesson, error) {
	endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)

	query := iss.db.Model(&models.Lesson{}).
		Preload("Student").
		Where("coach_id = ? AND status IN ? AND ((scheduled_at < ? AND scheduled_at + INTERVAL '1 minute' * duration > ?) OR (scheduled_at < ? AND scheduled_at + INTERVAL '1 minute' * duration > ?))",
			coachID, []string{"scheduled", "in_progress"}, endTime, scheduledAt, scheduledAt, endTime)

	if excludeLessonID != nil {
		query = query.Where("id != ?", *excludeLessonID)
	}

	var conflictingLessons []models.Lesson
	if err := query.Find(&conflictingLessons).Error; err != nil {
		return nil, errors.New("檢測排課衝突失敗")
	}

	return conflictingLessons, nil
}

// ResolveSchedulingConflict 解決排課衝突
func (iss *IntelligentSchedulingService) ResolveSchedulingConflict(conflictingLessonID string, newScheduledAt time.Time) error {
	// 檢查新時間是否有衝突
	var lesson models.Lesson
	if err := iss.db.Where("id = ?", conflictingLessonID).First(&lesson).Error; err != nil {
		return errors.New("課程不存在")
	}

	conflicts, err := iss.DetectSchedulingConflicts(lesson.CoachID, newScheduledAt, lesson.Duration, &conflictingLessonID)
	if err != nil {
		return err
	}

	if len(conflicts) > 0 {
		return errors.New("新時間仍有衝突")
	}

	// 更新課程時間
	if err := iss.db.Model(&lesson).Update("scheduled_at", newScheduledAt).Error; err != nil {
		return errors.New("更新課程時間失敗")
	}

	return nil
}
