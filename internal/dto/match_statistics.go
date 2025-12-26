package dto

// RecordMatchResultRequest 記錄比賽結果請求
type RecordMatchResultRequest struct {
	WinnerID string `json:"winnerId" binding:"required"`
	LoserID  string `json:"loserId" binding:"required"`
	Score    string `json:"score" binding:"required"`
}

// ManuallyAdjustSkillLevelRequest 手動調整技術等級請求
type ManuallyAdjustSkillLevelRequest struct {
	NewLevel float64 `json:"newLevel" binding:"required,min=1,max=7"`
	Reason   string  `json:"reason" binding:"required"`
}
