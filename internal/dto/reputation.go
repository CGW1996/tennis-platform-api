package dto

import "time"

// RecordMatchAttendanceRequest 記錄比賽出席請求
type RecordMatchAttendanceRequest struct {
	MatchID string `json:"matchId" binding:"required"`
	Status  string `json:"status" binding:"required,oneof=completed cancelled no_show"`
}

// RecordMatchPunctualityRequest 記錄比賽準時情況請求
type RecordMatchPunctualityRequest struct {
	MatchID     string    `json:"matchId" binding:"required"`
	ArrivalTime time.Time `json:"arrivalTime" binding:"required"`
}

// RecordSkillLevelAccuracyRequest 記錄技術等級準確度請求
type RecordSkillLevelAccuracyRequest struct {
	MatchID       string  `json:"matchId" binding:"required"`
	ReportedLevel float64 `json:"reportedLevel" binding:"required,min=1,max=7"`
	ObservedLevel float64 `json:"observedLevel" binding:"required,min=1,max=7"`
}

// SubmitBehaviorReviewRequest 提交行為評價請求
type SubmitBehaviorReviewRequest struct {
	MatchID string   `json:"matchId"`
	Rating  float64  `json:"rating" binding:"required,min=1,max=5"`
	Comment *string  `json:"comment"`
	Tags    []string `json:"tags"`
}
