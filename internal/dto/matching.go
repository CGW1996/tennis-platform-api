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

// CreateMatchRequest 創建配對請求
type CreateMatchRequest struct {
	ParticipantIDs []string `json:"participantIds" binding:"required"`
	MatchType      string   `json:"matchType" binding:"required"`
	CourtID        *string  `json:"courtId,omitempty"`
	ScheduledAt    *string  `json:"scheduledAt,omitempty"`
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
