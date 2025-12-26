package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Match 球友配對/比賽
type Match struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type        string         `json:"type" gorm:"not null"`            // casual, practice, tournament
	Status      string         `json:"status" gorm:"default:'pending'"` // pending, confirmed, in_progress, completed, cancelled
	CourtID     *string        `json:"courtId" gorm:"type:uuid"`
	ScheduledAt *time.Time     `json:"scheduledAt"`
	StartedAt   *time.Time     `json:"startedAt"`
	CompletedAt *time.Time     `json:"completedAt"`
	Duration    *int           `json:"duration"` // 分鐘
	Notes       *string        `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Court        *Court        `json:"court,omitempty" gorm:"constraint:OnDelete:SET NULL"`
	Participants []User        `json:"participants,omitempty" gorm:"many2many:match_participants;constraint:OnDelete:CASCADE"`
	Results      []MatchResult `json:"results,omitempty" gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE"`
	ChatRoom     *ChatRoom     `json:"chatRoom,omitempty" gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE"`
}

// MatchParticipant 比賽參與者 (中間表)
type MatchParticipant struct {
	MatchID   string    `json:"matchId" gorm:"type:uuid;primaryKey"`
	UserID    string    `json:"userId" gorm:"type:uuid;primaryKey"`
	Role      string    `json:"role" gorm:"default:'player'"`    // player, organizer
	Status    string    `json:"status" gorm:"default:'pending'"` // pending, accepted, declined
	JoinedAt  time.Time `json:"joinedAt" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt time.Time `json:"createdAt"`
}

// MatchResult 比賽結果
type MatchResult struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MatchID     string         `json:"matchId" gorm:"type:uuid;not null"`
	WinnerID    *string        `json:"winnerId" gorm:"type:uuid"`
	LoserID     *string        `json:"loserId" gorm:"type:uuid"`
	Score       *string        `json:"score"` // "6-4, 6-2"
	IsConfirmed bool           `json:"isConfirmed" gorm:"default:false"`
	ConfirmedBy []string       `json:"confirmedBy" gorm:"type:text[]"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Match  *Match `json:"match,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Winner *User  `json:"winner,omitempty" gorm:"foreignKey:WinnerID;constraint:OnDelete:SET NULL"`
	Loser  *User  `json:"loser,omitempty" gorm:"foreignKey:LoserID;constraint:OnDelete:SET NULL"`
}

// ChatRoom 聊天室
type ChatRoom struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MatchID   *string        `json:"matchId" gorm:"type:uuid;uniqueIndex"`
	Type      string         `json:"type" gorm:"default:'match'"` // match, group, direct
	Name      *string        `json:"name"`
	IsActive  bool           `json:"isActive" gorm:"default:true"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Match        *Match            `json:"match,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Messages     []ChatMessage     `json:"messages,omitempty" gorm:"foreignKey:ChatRoomID;constraint:OnDelete:CASCADE"`
	Participants []ChatParticipant `json:"participants,omitempty" gorm:"foreignKey:ChatRoomID;constraint:OnDelete:CASCADE"`
}

// ChatMessage 聊天訊息
type ChatMessage struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ChatRoomID  string         `json:"chatRoomId" gorm:"type:uuid;not null"`
	SenderID    string         `json:"senderId" gorm:"type:uuid;not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	MessageType string         `json:"messageType" gorm:"default:'text'"` // text, image, file
	IsRead      bool           `json:"isRead" gorm:"default:false"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	ChatRoom *ChatRoom `json:"chatRoom,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Sender   *User     `json:"sender,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// ChatParticipant 聊天室參與者
type ChatParticipant struct {
	ID         string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ChatRoomID string     `json:"chatRoomId" gorm:"type:uuid;not null"`
	UserID     string     `json:"userId" gorm:"type:uuid;not null"`
	JoinedAt   time.Time  `json:"joinedAt" gorm:"default:CURRENT_TIMESTAMP"`
	LastReadAt *time.Time `json:"lastReadAt"`
	IsActive   bool       `json:"isActive" gorm:"default:true"`
	CreatedAt  time.Time  `json:"createdAt"`

	// 關聯
	ChatRoom *ChatRoom `json:"chatRoom,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User     *User     `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// ReputationScore 信譽分數
type ReputationScore struct {
	ID               string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           string    `json:"userId" gorm:"type:uuid;uniqueIndex;not null"`
	AttendanceRate   float64   `json:"attendanceRate" gorm:"default:100"`   // 出席率 (0-100)
	PunctualityScore float64   `json:"punctualityScore" gorm:"default:100"` // 準時度 (0-100)
	SkillAccuracy    float64   `json:"skillAccuracy" gorm:"default:100"`    // 技術等級準確度 (0-100)
	BehaviorRating   float64   `json:"behaviorRating" gorm:"default:5"`     // 行為評分 (1-5)
	TotalMatches     int       `json:"totalMatches" gorm:"default:0"`
	CompletedMatches int       `json:"completedMatches" gorm:"default:0"`
	CancelledMatches int       `json:"cancelledMatches" gorm:"default:0"`
	OverallScore     float64   `json:"overallScore" gorm:"default:100"` // 綜合分數 (0-100)
	UpdatedAt        time.Time `json:"updatedAt"`

	// 關聯
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate 創建前的鉤子
func (m *Match) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (mr *MatchResult) BeforeCreate(tx *gorm.DB) error {
	if mr.ID == "" {
		mr.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cr *ChatRoom) BeforeCreate(tx *gorm.DB) error {
	if cr.ID == "" {
		cr.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cm *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		cm.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cp *ChatParticipant) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == "" {
		cp.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (rs *ReputationScore) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == "" {
		rs.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Match) TableName() string {
	return "matches"
}

// TableName 指定表名
func (MatchParticipant) TableName() string {
	return "match_participants"
}

// TableName 指定表名
func (MatchResult) TableName() string {
	return "match_results"
}

// TableName 指定表名
func (ChatRoom) TableName() string {
	return "chat_rooms"
}

// TableName 指定表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// TableName 指定表名
func (ChatParticipant) TableName() string {
	return "chat_participants"
}

// TableName 指定表名
func (ReputationScore) TableName() string {
	return "reputation_scores"
}

// CardInteraction 抽卡互動記錄
type CardInteraction struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string         `json:"userId" gorm:"type:uuid;not null"`       // 執行動作的用戶
	TargetUserID string         `json:"targetUserId" gorm:"type:uuid;not null"` // 被動作的用戶
	Action       string         `json:"action" gorm:"not null"`                 // like, dislike, skip
	IsMatch      bool           `json:"isMatch" gorm:"default:false"`           // 是否配對成功
	MatchID      *string        `json:"matchId" gorm:"type:uuid"`               // 配對ID（如果配對成功）
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	User       *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	TargetUser *User  `json:"targetUser,omitempty" gorm:"foreignKey:TargetUserID;constraint:OnDelete:CASCADE"`
	Match      *Match `json:"match,omitempty" gorm:"constraint:OnDelete:SET NULL"`
}

// MatchNotification 配對通知
type MatchNotification struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string         `json:"userId" gorm:"type:uuid;not null"`
	Type      string         `json:"type" gorm:"not null"` // match_success, match_request, match_cancelled
	Title     string         `json:"title" gorm:"not null"`
	Message   string         `json:"message" gorm:"type:text"`
	Data      string         `json:"data" gorm:"type:jsonb"` // 額外數據（JSON格式）
	IsRead    bool           `json:"isRead" gorm:"default:false"`
	ReadAt    *time.Time     `json:"readAt"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate 創建前的鉤子
func (ci *CardInteraction) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == "" {
		ci.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (mn *MatchNotification) BeforeCreate(tx *gorm.DB) error {
	if mn.ID == "" {
		mn.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (CardInteraction) TableName() string {
	return "card_interactions"
}

// TableName 指定表名
func (MatchNotification) TableName() string {
	return "match_notifications"
}

// PunctualityRecord 準時記錄
type PunctualityRecord struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string    `json:"userId" gorm:"type:uuid;not null"`
	MatchID      *string   `json:"matchId" gorm:"type:uuid"`
	IsOnTime     bool      `json:"isOnTime"`
	DelayMinutes int       `json:"delayMinutes" gorm:"default:0"`
	CreatedAt    time.Time `json:"createdAt"`

	// 關聯
	User  *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Match *Match `json:"match,omitempty" gorm:"constraint:OnDelete:SET NULL"`
}

// SkillAccuracyRecord 技術等級準確度記錄
type SkillAccuracyRecord struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        string    `json:"userId" gorm:"type:uuid;not null"`
	MatchID       *string   `json:"matchId" gorm:"type:uuid"`
	ReportedLevel float64   `json:"reportedLevel"` // 用戶自報等級
	ActualLevel   float64   `json:"actualLevel"`   // 實際表現等級
	Accuracy      float64   `json:"accuracy"`      // 準確度分數 (0-100)
	CreatedAt     time.Time `json:"createdAt"`

	// 關聯
	User  *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Match *Match `json:"match,omitempty" gorm:"constraint:OnDelete:SET NULL"`
}

// BehaviorReview 行為評價
type BehaviorReview struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string    `json:"userId" gorm:"type:uuid;not null"`     // 被評價者
	ReviewerID string    `json:"reviewerId" gorm:"type:uuid;not null"` // 評價者
	MatchID    *string   `json:"matchId" gorm:"type:uuid"`
	Rating     float64   `json:"rating" gorm:"check:rating >= 1 AND rating <= 5"` // 1-5分
	Comment    *string   `json:"comment" gorm:"type:text"`
	Tags       []string  `json:"tags" gorm:"type:text[]"` // 評價標籤：friendly, punctual, skilled, etc.
	CreatedAt  time.Time `json:"createdAt"`

	// 關聯
	User     *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Reviewer *User  `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID;constraint:OnDelete:CASCADE"`
	Match    *Match `json:"match,omitempty" gorm:"constraint:OnDelete:SET NULL"`
}

// ReputationHistory 信譽歷史記錄（用於API響應）
type ReputationHistory struct {
	UserID             string                `json:"userId"`
	PunctualityRecords []PunctualityRecord   `json:"punctualityRecords"`
	SkillRecords       []SkillAccuracyRecord `json:"skillRecords"`
	BehaviorReviews    []BehaviorReview      `json:"behaviorReviews"`
}

// MatchStatistics 配對統計資訊
type MatchStatistics struct {
	UserID                string              `json:"userId"`
	TotalMatches          int64               `json:"totalMatches"`
	CompletedMatches      int64               `json:"completedMatches"`
	CancelledMatches      int64               `json:"cancelledMatches"`
	WonMatches            int64               `json:"wonMatches"`
	LostMatches           int64               `json:"lostMatches"`
	WinRate               float64             `json:"winRate"`
	AttendanceRate        float64             `json:"attendanceRate"`
	AverageMatchDuration  int                 `json:"averageMatchDuration"` // 分鐘
	FavoriteCourtType     string              `json:"favoriteCourtType"`
	MostPlayedWith        []string            `json:"mostPlayedWith"` // 最常配對的用戶ID
	RecentMatches         []Match             `json:"recentMatches"`
	MonthlyStats          []MonthlyMatchStats `json:"monthlyStats"`
	SkillLevelProgression []SkillLevelRecord  `json:"skillLevelProgression"`
}

// MonthlyMatchStats 月度配對統計
type MonthlyMatchStats struct {
	Year             int     `json:"year"`
	Month            int     `json:"month"`
	TotalMatches     int64   `json:"totalMatches"`
	CompletedMatches int64   `json:"completedMatches"`
	WinRate          float64 `json:"winRate"`
	AverageRating    float64 `json:"averageRating"`
}

// SkillLevelRecord 技術等級記錄
type SkillLevelRecord struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"userId" gorm:"type:uuid;not null"`
	OldLevel  float64   `json:"oldLevel"`
	NewLevel  float64   `json:"newLevel"`
	Reason    string    `json:"reason"` // manual, auto_adjustment, match_result
	MatchID   *string   `json:"matchId" gorm:"type:uuid"`
	CreatedAt time.Time `json:"createdAt"`

	// 關聯
	User  *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Match *Match `json:"match,omitempty" gorm:"constraint:OnDelete:SET NULL"`
}

// UserPrivacySettings 用戶隱私設定
type UserPrivacySettings struct {
	UserID                 string    `json:"userId" gorm:"type:uuid;primaryKey"`
	ShowReputationScore    bool      `json:"showReputationScore" gorm:"default:true"`
	ShowMatchHistory       bool      `json:"showMatchHistory" gorm:"default:true"`
	ShowWinLossRecord      bool      `json:"showWinLossRecord" gorm:"default:true"`
	ShowSkillProgression   bool      `json:"showSkillProgression" gorm:"default:true"`
	ShowBehaviorReviews    bool      `json:"showBehaviorReviews" gorm:"default:false"`
	ShowDetailedStats      bool      `json:"showDetailedStats" gorm:"default:true"`
	AllowStatisticsSharing bool      `json:"allowStatisticsSharing" gorm:"default:false"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`

	// 關聯
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate 創建前的鉤子
func (pr *PunctualityRecord) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == "" {
		pr.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (sar *SkillAccuracyRecord) BeforeCreate(tx *gorm.DB) error {
	if sar.ID == "" {
		sar.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (br *BehaviorReview) BeforeCreate(tx *gorm.DB) error {
	if br.ID == "" {
		br.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (slr *SkillLevelRecord) BeforeCreate(tx *gorm.DB) error {
	if slr.ID == "" {
		slr.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (PunctualityRecord) TableName() string {
	return "punctuality_records"
}

// TableName 指定表名
func (SkillAccuracyRecord) TableName() string {
	return "skill_accuracy_records"
}

// TableName 指定表名
func (BehaviorReview) TableName() string {
	return "behavior_reviews"
}

// TableName 指定表名
func (SkillLevelRecord) TableName() string {
	return "skill_level_records"
}

// TableName 指定表名
func (UserPrivacySettings) TableName() string {
	return "user_privacy_settings"
}
