package dto

import "tennis-platform/backend/internal/services"

// FindMatchesRequest 尋找配對請求
type FindMatchesRequest struct {
	NTRPRange          *services.NTRPRange             `json:"ntrp_range,omitempty"`
	MaxDistance        *float64                        `json:"max_distance,omitempty"`
	PlayingFrequency   *string                         `json:"playing_frequency,omitempty"`
	AgeRange           *services.AgeRange              `json:"age_range,omitempty"`
	Gender             *string                         `json:"gender,omitempty"`
	MinReputationScore *float64                        `json:"min_reputation_score,omitempty"`
	Limit              int                             `json:"limit,omitempty"`
	Location           *services.LocationCriteria      `json:"location,omitempty"`
	PlayType           []string                        `json:"play_type,omitempty"`
	Availability       []services.AvailabilityCriteria `json:"availability,omitempty"`
}

// AvailabilitySlot 可用時間段
type AvailabilitySlot struct {
	Day       string `json:"day"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Location  string `json:"location"`
}

// CreateMatchRequest 創建配對請求
type CreateMatchRequest struct {
	ParticipantIDs      []string           `json:"participantIds" binding:"required"`
	MatchType           string             `json:"matchType" binding:"required"`
	CourtID             *string            `json:"courtId,omitempty"`
	ScheduledAt         *string            `json:"scheduledAt,omitempty"`
	AvailabilitySlots   []AvailabilitySlot `json:"availabilitySlots"`
	PreferredCourt      *string            `json:"preferredCourt"`
	SpecialRequirements *string            `json:"specialRequirements"`
	NtrpMin             *float64           `json:"ntrpMin"`
	NtrpMax             *float64           `json:"ntrpMax"`
	PlayTypes           []string           `json:"playTypes"` // rally, singles, doubles,omitempty"`
}

// UpdateReputationRequest 更新信譽請求
type UpdateReputationRequest struct {
	MatchCompleted bool    `json:"matchCompleted"`
	WasOnTime      bool    `json:"wasOnTime"`
	BehaviorRating float64 `json:"behaviorRating"`
}

// CardActionRequest 抽卡動作請求
type CardActionRequest struct {
	TargetUserID string `json:"targetUserId" binding:"required"`
	Action       string `json:"action" binding:"required"` // like, dislike, skip
}

// FindPartnersRequest 找球友請求（練習性質）
type FindPartnersRequest struct {
	Location           *services.LocationCriteria      `json:"location,omitempty"`
	PlayType           []string                        `json:"play_type,omitempty"` // rally, doubles, singles
	Availability       []services.AvailabilityCriteria `json:"availability,omitempty"`
	PlayingFrequency   *string                         `json:"playing_frequency,omitempty"`
	Gender             *string                         `json:"gender,omitempty"`
	AgeRange           *services.AgeRange              `json:"age_range,omitempty"`
	NTRPRange          *services.NTRPRange             `json:"ntrp_range,omitempty"` // 較寬鬆的範圍
	MaxDistance        *float64                        `json:"max_distance,omitempty"`
	MinReputationScore *float64                        `json:"min_reputation_score,omitempty"`
	Limit              int                             `json:"limit,omitempty"`
}

// FindCompetitiveMatchesRequest 找對手請求（競賽性質）
type FindCompetitiveMatchesRequest struct {
	NTRPRange          *services.NTRPRange             `json:"ntrp_range,omitempty"` // 嚴格的等級匹配
	Gender             *string                         `json:"gender,omitempty"`
	AgeRange           *services.AgeRange              `json:"age_range,omitempty"`
	MatchType          string                          `json:"match_type,omitempty"`           // singles, doubles
	MinReputationScore *float64                        `json:"min_reputation_score,omitempty"` // 較高的信譽要求
	Location           *services.LocationCriteria      `json:"location,omitempty"`
	MaxDistance        *float64                        `json:"max_distance,omitempty"`
	PreferredCourtType *string                         `json:"preferred_court_type,omitempty"`
	Availability       []services.AvailabilityCriteria `json:"availability,omitempty"`
	Limit              int                             `json:"limit,omitempty"`
}
