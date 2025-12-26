package services

import (
	"math"
	"sort"
	"time"

	"tennis-platform/backend/internal/models"
)

// MatchingService 配對服務
type MatchingService struct{}

// NewMatchingService 創建配對服務實例
func NewMatchingService() *MatchingService {
	return &MatchingService{}
}

// MatchingCriteria 配對條件
type MatchingCriteria struct {
	UserID             string     `json:"userId"`
	NTRPRange          *NTRPRange `json:"ntrpRange,omitempty"`
	MaxDistance        *float64   `json:"maxDistance"`      // 公里
	PlayingFrequency   *string    `json:"playingFrequency"` // casual, regular, competitive
	AgeRange           *AgeRange  `json:"ageRange,omitempty"`
	Gender             *string    `json:"gender,omitempty"` // male, female, any
	MinReputationScore *float64   `json:"minReputationScore,omitempty"`

	// New fields
	Location     *LocationCriteria      `json:"location,omitempty"`
	PlayTypes    []string               `json:"playTypes,omitempty"`
	Availability []AvailabilityCriteria `json:"availability,omitempty"`
}

// AgeRange 年齡範圍
type AgeRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// NTRPRange NTRP等級範圍
type NTRPRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// LocationCriteria 地點條件
type LocationCriteria struct {
	City     string `json:"city"`
	District string `json:"district,omitempty"`
}

// AvailabilityCriteria 時間條件
type AvailabilityCriteria struct {
	Type string `json:"type"` // weekday, weekend
	Time string `json:"time"` // morning, afternoon, evening
}

// MatchingResult 配對結果
type MatchingResult struct {
	UserID  string          `json:"userId"`
	Score   float64         `json:"score"`
	Factors MatchingFactors `json:"factors"`
	User    *models.User    `json:"user,omitempty"`
}

// MatchingFactors 配對因子
type MatchingFactors struct {
	SkillLevel        float64 `json:"skillLevel"`        // NTRP等級匹配度 (0-1)
	Distance          float64 `json:"distance"`          // 距離匹配度 (0-1)
	TimeCompatibility float64 `json:"timeCompatibility"` // 時間相容性 (0-1)
	PlayingStyle      float64 `json:"playingStyle"`      // 打球風格匹配度 (0-1)
	Age               float64 `json:"age"`               // 年齡匹配度 (0-1)
	Reputation        float64 `json:"reputation"`        // 信譽匹配度 (0-1)
}

// MatchingWeights 配對權重
type MatchingWeights struct {
	SkillLevel        float64 `json:"skillLevel"`        // 預設: 0.35
	Distance          float64 `json:"distance"`          // 預設: 0.25
	TimeCompatibility float64 `json:"timeCompatibility"` // 預設: 0.20
	PlayingStyle      float64 `json:"playingStyle"`      // 預設: 0.10
	Age               float64 `json:"age"`               // 預設: 0.05
	Reputation        float64 `json:"reputation"`        // 預設: 0.05
}

// DefaultMatchingWeights 預設配對權重
func DefaultMatchingWeights() MatchingWeights {
	return MatchingWeights{
		SkillLevel:        0.35,
		Distance:          0.25,
		TimeCompatibility: 0.20,
		PlayingStyle:      0.10,
		Age:               0.05,
		Reputation:        0.05,
	}
}

// CalculateMatchingScore 計算配對分數
func (s *MatchingService) CalculateMatchingScore(
	requester *models.User,
	candidate *models.User,
	criteria MatchingCriteria,
	weights MatchingWeights,
) MatchingResult {
	factors := MatchingFactors{}

	// 1. NTRP等級匹配度
	factors.SkillLevel = s.calculateSkillLevelScore(
		requester.Profile.NTRPLevel,
		candidate.Profile.NTRPLevel,
	)

	// 2. 地理距離匹配度
	factors.Distance = s.calculateDistanceScore(
		requester.Profile.Latitude,
		requester.Profile.Longitude,
		candidate.Profile.Latitude,
		candidate.Profile.Longitude,
		criteria.MaxDistance,
	)

	// 3. 時間偏好匹配度
	factors.TimeCompatibility = s.calculateTimeCompatibilityScore(
		requester.Profile.PreferredTimes,
		candidate.Profile.PreferredTimes,
	)

	// 4. 打球風格匹配度
	factors.PlayingStyle = s.calculatePlayingStyleScore(
		requester.Profile.PlayingStyle,
		candidate.Profile.PlayingStyle,
		requester.Profile.PlayingFrequency,
		candidate.Profile.PlayingFrequency,
	)

	// 5. 年齡匹配度
	factors.Age = s.calculateAgeScore(
		requester.Profile.BirthDate,
		candidate.Profile.BirthDate,
		criteria.AgeRange,
	)

	// 6. 信譽匹配度
	factors.Reputation = s.calculateReputationScore(candidate, criteria.MinReputationScore)

	// 計算總分
	totalScore := factors.SkillLevel*weights.SkillLevel +
		factors.Distance*weights.Distance +
		factors.TimeCompatibility*weights.TimeCompatibility +
		factors.PlayingStyle*weights.PlayingStyle +
		factors.Age*weights.Age +
		factors.Reputation*weights.Reputation

	return MatchingResult{
		UserID:  candidate.ID,
		Score:   totalScore,
		Factors: factors,
		User:    candidate,
	}
}

// calculateSkillLevelScore 計算技術等級匹配分數
func (s *MatchingService) calculateSkillLevelScore(requesterLevel, candidateLevel *float64) float64 {
	if requesterLevel == nil || candidateLevel == nil {
		return 0.5 // 如果沒有等級資訊，給予中等分數
	}

	diff := math.Abs(*requesterLevel - *candidateLevel)

	// NTRP等級差異評分
	// 0.0-0.5差異: 1.0分
	// 0.5-1.0差異: 0.8分
	// 1.0-1.5差異: 0.6分
	// 1.5-2.0差異: 0.4分
	// 2.0+差異: 0.2分
	switch {
	case diff <= 0.5:
		return 1.0
	case diff <= 1.0:
		return 0.8
	case diff <= 1.5:
		return 0.6
	case diff <= 2.0:
		return 0.4
	default:
		return 0.2
	}
}

// calculateDistanceScore 計算距離匹配分數
func (s *MatchingService) calculateDistanceScore(
	reqLat, reqLng, candLat, candLng *float64,
	maxDistance *float64,
) float64 {
	if reqLat == nil || reqLng == nil || candLat == nil || candLng == nil {
		return 0.5 // 如果沒有位置資訊，給予中等分數
	}

	distance := s.calculateDistance(*reqLat, *reqLng, *candLat, *candLng)

	// 如果沒有設定最大距離，預設為20公里
	maxDist := 20.0
	if maxDistance != nil && *maxDistance > 0 {
		maxDist = *maxDistance
	}

	if distance > maxDist {
		return 0.0 // 超出最大距離
	}

	// 距離評分：距離越近分數越高
	// 0-5km: 1.0分
	// 5-10km: 0.8分
	// 10-15km: 0.6分
	// 15-20km: 0.4分
	// 20km+: 0.2分
	switch {
	case distance <= 5:
		return 1.0
	case distance <= 10:
		return 0.8
	case distance <= 15:
		return 0.6
	case distance <= 20:
		return 0.4
	default:
		return math.Max(0.2, 1.0-(distance/maxDist)*0.8)
	}
}

// calculateDistance 計算兩點間距離（公里）
func (s *MatchingService) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // 地球半徑（公里）

	// 轉換為弧度
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Haversine公式
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlng/2)*math.Sin(dlng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
}

// calculateTimeCompatibilityScore 計算時間相容性分數
func (s *MatchingService) calculateTimeCompatibilityScore(
	requesterTimes, candidateTimes []string,
) float64 {
	if len(requesterTimes) == 0 || len(candidateTimes) == 0 {
		return 0.5 // 如果沒有時間偏好，給予中等分數
	}

	// 計算共同時間偏好
	commonTimes := 0
	for _, reqTime := range requesterTimes {
		for _, candTime := range candidateTimes {
			if reqTime == candTime {
				commonTimes++
				break
			}
		}
	}

	if commonTimes == 0 {
		return 0.1 // 沒有共同時間
	}

	// 計算相容性分數
	maxTimes := math.Max(float64(len(requesterTimes)), float64(len(candidateTimes)))
	compatibility := float64(commonTimes) / maxTimes

	return math.Min(1.0, compatibility*1.5) // 最高1.0分
}

// calculatePlayingStyleScore 計算打球風格匹配分數
func (s *MatchingService) calculatePlayingStyleScore(
	reqStyle, candStyle *string,
	reqFreq, candFreq *string,
) float64 {
	score := 0.0

	// 打球風格匹配 (60%權重)
	if reqStyle != nil && candStyle != nil {
		if *reqStyle == *candStyle {
			score += 0.6
		} else if (*reqStyle == "all-court") || (*candStyle == "all-court") {
			score += 0.4 // all-court與任何風格都有一定相容性
		} else {
			score += 0.2
		}
	} else {
		score += 0.3 // 沒有風格資訊給予中等分數
	}

	// 打球頻率匹配 (40%權重)
	if reqFreq != nil && candFreq != nil {
		if *reqFreq == *candFreq {
			score += 0.4
		} else {
			// 頻率相容性評分
			freqScore := s.calculateFrequencyCompatibility(*reqFreq, *candFreq)
			score += freqScore * 0.4
		}
	} else {
		score += 0.2 // 沒有頻率資訊給予較低分數
	}

	return math.Min(1.0, score)
}

// calculateFrequencyCompatibility 計算打球頻率相容性
func (s *MatchingService) calculateFrequencyCompatibility(freq1, freq2 string) float64 {
	// 頻率相容性矩陣
	compatibility := map[string]map[string]float64{
		"casual": {
			"casual":      1.0,
			"regular":     0.7,
			"competitive": 0.3,
		},
		"regular": {
			"casual":      0.7,
			"regular":     1.0,
			"competitive": 0.8,
		},
		"competitive": {
			"casual":      0.3,
			"regular":     0.8,
			"competitive": 1.0,
		},
	}

	if compat, exists := compatibility[freq1]; exists {
		if score, exists := compat[freq2]; exists {
			return score
		}
	}

	return 0.5 // 預設中等相容性
}

// calculateAgeScore 計算年齡匹配分數
func (s *MatchingService) calculateAgeScore(
	reqBirthDate, candBirthDate *time.Time,
	ageRange *AgeRange,
) float64 {
	if reqBirthDate == nil || candBirthDate == nil {
		return 0.5 // 沒有年齡資訊給予中等分數
	}

	reqAge := s.calculateAge(*reqBirthDate)
	candAge := s.calculateAge(*candBirthDate)

	// 如果設定了年齡範圍限制
	if ageRange != nil {
		if candAge < ageRange.Min || candAge > ageRange.Max {
			return 0.0 // 不符合年齡範圍
		}
	}

	// 年齡差異評分
	ageDiff := math.Abs(float64(reqAge - candAge))

	switch {
	case ageDiff <= 3:
		return 1.0
	case ageDiff <= 5:
		return 0.8
	case ageDiff <= 10:
		return 0.6
	case ageDiff <= 15:
		return 0.4
	default:
		return 0.2
	}
}

// calculateAge 計算年齡
func (s *MatchingService) calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// 檢查是否還沒過生日
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	return age
}

// calculateReputationScore 計算信譽匹配分數
func (s *MatchingService) calculateReputationScore(
	candidate *models.User,
	minReputationScore *float64,
) float64 {
	// 這裡需要從數據庫獲取信譽分數，暫時返回預設值
	// TODO: 實作從ReputationScore表獲取實際分數
	defaultReputation := 85.0 // 預設信譽分數

	if minReputationScore != nil {
		if defaultReputation < *minReputationScore {
			return 0.0 // 不符合最低信譽要求
		}
	}

	// 信譽分數正規化到0-1
	return math.Min(1.0, defaultReputation/100.0)
}

// RankCandidates 對候選人進行排序
func (s *MatchingService) RankCandidates(results []MatchingResult) []MatchingResult {
	// 按分數降序排列
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// FilterCandidates 篩選候選人
func (s *MatchingService) FilterCandidates(
	candidates []models.User,
	criteria MatchingCriteria,
) []models.User {
	var filtered []models.User

	for _, candidate := range candidates {
		// 排除自己
		if candidate.ID == criteria.UserID {
			continue
		}

		// 檢查基本條件
		if !s.meetsCriteria(candidate, criteria) {
			continue
		}

		filtered = append(filtered, candidate)
	}

	return filtered
}

// meetsCriteria 檢查是否符合基本條件
func (s *MatchingService) meetsCriteria(candidate models.User, criteria MatchingCriteria) bool {
	// 檢查性別偏好
	if criteria.Gender != nil && *criteria.Gender != "any" {
		if candidate.Profile == nil || candidate.Profile.Gender == nil {
			return false
		}
		if *candidate.Profile.Gender != *criteria.Gender {
			return false
		}
	}

	// 檢查年齡範圍
	if criteria.AgeRange != nil && candidate.Profile != nil && candidate.Profile.BirthDate != nil {
		age := s.calculateAge(*candidate.Profile.BirthDate)
		if age < criteria.AgeRange.Min || age > criteria.AgeRange.Max {
			return false
		}
	}

	// 檢查 NTRP 範圍 (New)
	if criteria.NTRPRange != nil && candidate.Profile != nil && candidate.Profile.NTRPLevel != nil {
		if *candidate.Profile.NTRPLevel < criteria.NTRPRange.Min || *candidate.Profile.NTRPLevel > criteria.NTRPRange.Max {
			return false
		}
	}

	// 檢查 PlayTypes (New)
	if len(criteria.PlayTypes) > 0 && candidate.Profile != nil {
		// 如果候選人沒有設定 PlayTypes，假設不符合
		if len(candidate.Profile.PlayTypes) == 0 {
			// Fallback to PlayingStyle if PlayTypes is empty, but this logic depends on how strict we want to be.
			// For now, let's just check if PlayTypes has intersection.
			// Or check playing style as proxy? Let's check intersection if both exist.
		}

		hasIntersection := false
		for _, requiredType := range criteria.PlayTypes {
			for _, userType := range candidate.Profile.PlayTypes {
				if requiredType == userType {
					hasIntersection = true
					break
				}
			}
			if hasIntersection {
				break
			}
		}
		if !hasIntersection && len(candidate.Profile.PlayTypes) > 0 {
			return false
		}
	}

	// 檢查 Location (City/District) (New) - Fuzzy match
	// We don't have structured city/district in UserProfile yet, assuming it's part of Address or implicitly handled by Distance.
	// However, if frontend filters by city/district, we might want to check if we can support that.
	// Since UserProfile has Latitude/Longitude, and Address string.
	// We can't strictly enforce City/District without structured data in UserProfile.
	// For now, let's rely on Distance if available, or if Address contains the string.
	/*
		if criteria.Location != nil && candidate.Profile != nil {
			// Use Google Maps API or similar to check city? Too expensive.
			// Simple string check if address is populated?
			// if candidate.Profile.Address ...
		}
	*/

	// 檢查 Availability (New)
	/*
		if len(criteria.Availability) > 0 && candidate.Profile != nil {
			// Need to map candidate.Profile.PreferredTimes (flat list) to structured availability?
			// Frontend PlayTimes: ["morning", "afternoon", "evening"] (old)
			// Backend User Profile PreferredTimes: ["morning", "afternoon", "evening"] (flat)

			// Map Availability struct to flat list for check
			var requiredTimes []string
			for _, a := range criteria.Availability {
				// e.g., "weekday_morning", "weekend_afternoon" ??
				// Or just "morning", "afternoon"?
				// Struct says: Type & Time.
				// UserProfile.PreferredTimes is PQ StringArray... likely simplistic strings like "morning", "afternoon".
				// We need to know what UserProfile.PreferredTimes values look like.
				// Assuming they match the simple strings for now or we need to update UserProfile to support structured times.
			}
		}
	*/

	// 檢查最大距離
	if criteria.MaxDistance != nil && candidate.Profile != nil {
		// 這裡需要請求者的位置資訊來計算距離
		// 暫時跳過距離檢查，在實際使用時需要傳入請求者資訊
	}

	return true
}

// GenerateRandomMatches 生成隨機配對（抽卡功能）
func (s *MatchingService) GenerateRandomMatches(
	candidates []models.User,
	criteria MatchingCriteria,
	weights MatchingWeights,
	count int,
) []MatchingResult {
	if len(candidates) == 0 {
		return []MatchingResult{}
	}

	// 篩選候選人
	filtered := s.FilterCandidates(candidates, criteria)
	if len(filtered) == 0 {
		return []MatchingResult{}
	}

	// 如果候選人數量少於請求數量，返回所有候選人
	if len(filtered) <= count {
		var results []MatchingResult
		for _, candidate := range filtered {
			// 需要請求者資訊來計算分數，這裡暫時使用空的User
			requester := &models.User{ID: criteria.UserID}
			result := s.CalculateMatchingScore(requester, &candidate, criteria, weights)
			results = append(results, result)
		}
		return s.RankCandidates(results)
	}

	// 加權隨機選擇
	var results []MatchingResult
	selected := make(map[string]bool)

	for len(results) < count && len(results) < len(filtered) {
		// 簡單隨機選擇（可以改進為加權隨機）
		for _, candidate := range filtered {
			if selected[candidate.ID] {
				continue
			}

			if len(results) >= count {
				break
			}

			// 這裡可以加入隨機權重邏輯
			// 暫時使用簡單選擇
			selected[candidate.ID] = true

			// 需要請求者資訊來計算分數
			requester := &models.User{ID: criteria.UserID}
			result := s.CalculateMatchingScore(requester, &candidate, criteria, weights)
			results = append(results, result)
		}
	}

	return s.RankCandidates(results)
}

// CardAction 抽卡動作類型
type CardAction string

const (
	CardActionLike    CardAction = "like"
	CardActionDislike CardAction = "dislike"
	CardActionSkip    CardAction = "skip"
)

// CardMatchResult 抽卡配對結果
type CardMatchResult struct {
	IsMatch    bool   `json:"isMatch"`
	MatchID    string `json:"matchId,omitempty"`
	ChatRoomID string `json:"chatRoomId,omitempty"`
	Message    string `json:"message"`
}

// ProcessCardAction 處理抽卡動作
func (s *MatchingService) ProcessCardAction(
	userID string,
	targetUserID string,
	action CardAction,
) *CardMatchResult {
	result := &CardMatchResult{
		IsMatch: false,
	}

	switch action {
	case CardActionLike:
		result.Message = "已表達興趣，等待對方回應"
	case CardActionDislike:
		result.Message = "已跳過此用戶"
	case CardActionSkip:
		result.Message = "已跳過此用戶"
	default:
		result.Message = "無效的動作"
	}

	return result
}
