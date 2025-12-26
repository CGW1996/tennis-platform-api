package dto

import (
	"tennis-platform/backend/internal/models"
	"time"
)

// ===== 教練檔案相關 =====

// CreateCoachProfileRequest 創建教練檔案請求
type CreateCoachProfileRequest struct {
	LicenseNumber  *string               `json:"licenseNumber" binding:"omitempty,min=1,max=50"`
	Certifications []string              `json:"certifications" binding:"omitempty,dive,min=1,max=100"`
	Experience     int                   `json:"experience" binding:"required,min=0,max=50"`
	Specialties    []string              `json:"specialties" binding:"required,dive,oneof=beginner intermediate advanced junior"`
	Biography      *string               `json:"biography" binding:"omitempty,max=1000"`
	HourlyRate     float64               `json:"hourlyRate" binding:"required,min=0"`
	Currency       string                `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	Languages      []string              `json:"languages" binding:"omitempty,dive,min=1,max=50"`
	AvailableHours models.AvailableHours `json:"availableHours" binding:"omitempty"`
}

// UpdateCoachProfileRequest 更新教練檔案請求
type UpdateCoachProfileRequest struct {
	LicenseNumber  *string                `json:"licenseNumber" binding:"omitempty,min=1,max=50"`
	Certifications []string               `json:"certifications" binding:"omitempty,dive,min=1,max=100"`
	Experience     *int                   `json:"experience" binding:"omitempty,min=0,max=50"`
	Specialties    []string               `json:"specialties" binding:"omitempty,dive,oneof=beginner intermediate advanced junior"`
	Biography      *string                `json:"biography" binding:"omitempty,max=1000"`
	HourlyRate     *float64               `json:"hourlyRate" binding:"omitempty,min=0"`
	Currency       *string                `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	Languages      []string               `json:"languages" binding:"omitempty,dive,min=1,max=50"`
	AvailableHours *models.AvailableHours `json:"availableHours" binding:"omitempty"`
	IsActive       *bool                  `json:"isActive"`
}

// CoachVerificationRequest 教練認證請求
type CoachVerificationRequest struct {
	CoachID           string  `json:"coachId" binding:"required"`
	IsVerified        bool    `json:"isVerified" binding:"required"`
	VerificationNotes *string `json:"verificationNotes" binding:"omitempty,max=500"`
}

// CoachSearchRequest 教練搜尋請求
type CoachSearchRequest struct {
	Specialties   []string `json:"specialties" form:"specialties"`
	MinExperience *int     `json:"minExperience" form:"min_experience,minExperience" binding:"omitempty,min=0"`
	MaxExperience *int     `json:"maxExperience" form:"max_experience,maxExperience" binding:"omitempty,min=0"`
	MinHourlyRate *float64 `json:"minHourlyRate" form:"price_min,minHourlyRate" binding:"omitempty,min=0"`
	MaxHourlyRate *float64 `json:"maxHourlyRate" form:"price_max,maxHourlyRate" binding:"omitempty,min=0"`
	Languages     []string `json:"languages" form:"languages"`
	MinRating     *float64 `json:"minRating" form:"min_rating,minRating" binding:"omitempty,min=0,max=5"`
	IsVerified    *bool    `json:"isVerified" form:"is_verified,isVerified"`
	IsActive      *bool    `json:"isActive" form:"is_active,isActive"`
	Page          int      `json:"page" form:"page" binding:"omitempty,min=1"`
	Limit         int      `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
	SortBy        string   `json:"sortBy" form:"sort_by,sortBy" binding:"omitempty,oneof=rating experience hourlyRate createdAt"`
	SortOrder     string   `json:"sortOrder" form:"sort_order,sortOrder" binding:"omitempty,oneof=asc desc"`
}

// ===== 課程類型相關 =====

// CreateLessonTypeRequest 創建課程類型請求
type CreateLessonTypeRequest struct {
	Name            string   `json:"name" binding:"required,min=1,max=100"`
	Description     *string  `json:"description" binding:"omitempty,max=500"`
	Type            string   `json:"type" binding:"required,oneof=individual group clinic"`
	Level           string   `json:"level" binding:"omitempty,oneof=beginner intermediate advanced"`
	Duration        int      `json:"duration" binding:"required,min=30,max=480"` // 30分鐘到8小時
	Price           float64  `json:"price" binding:"required,min=0"`
	Currency        string   `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	MaxParticipants *int     `json:"maxParticipants" binding:"omitempty,min=1,max=20"`
	MinParticipants *int     `json:"minParticipants" binding:"omitempty,min=1,max=20"`
	Equipment       []string `json:"equipment" binding:"omitempty,dive,min=1,max=50"`
	Prerequisites   *string  `json:"prerequisites" binding:"omitempty,max=500"`
}

// UpdateLessonTypeRequest 更新課程類型請求
type UpdateLessonTypeRequest struct {
	Name            *string  `json:"name" binding:"omitempty,min=1,max=100"`
	Description     *string  `json:"description" binding:"omitempty,max=500"`
	Type            *string  `json:"type" binding:"omitempty,oneof=individual group clinic"`
	Level           *string  `json:"level" binding:"omitempty,oneof=beginner intermediate advanced"`
	Duration        *int     `json:"duration" binding:"omitempty,min=30,max=480"`
	Price           *float64 `json:"price" binding:"omitempty,min=0"`
	Currency        *string  `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	MaxParticipants *int     `json:"maxParticipants" binding:"omitempty,min=1,max=20"`
	MinParticipants *int     `json:"minParticipants" binding:"omitempty,min=1,max=20"`
	Equipment       []string `json:"equipment" binding:"omitempty,dive,min=1,max=50"`
	Prerequisites   *string  `json:"prerequisites" binding:"omitempty,max=500"`
	IsActive        *bool    `json:"isActive"`
}

// ===== 課程相關 =====

// CreateLessonRequest 創建課程請求
type CreateLessonRequest struct {
	CoachID      string    `json:"coachId" binding:"required"`
	StudentID    string    `json:"studentId"` // 由控制器設置
	LessonTypeID *string   `json:"lessonTypeId" binding:"omitempty"`
	CourtID      *string   `json:"courtId" binding:"omitempty"`
	Type         string    `json:"type" binding:"required,oneof=individual group clinic"`
	Level        string    `json:"level" binding:"omitempty,oneof=beginner intermediate advanced"`
	Duration     int       `json:"duration" binding:"required,min=30,max=480"`
	Price        float64   `json:"price" binding:"required,min=0"`
	Currency     string    `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	ScheduledAt  time.Time `json:"scheduledAt" binding:"required"`
	Notes        *string   `json:"notes" binding:"omitempty,max=500"`
}

// UpdateLessonRequest 更新課程請求
type UpdateLessonRequest struct {
	CourtID     *string    `json:"courtId"`
	ScheduledAt *time.Time `json:"scheduledAt"`
	Notes       *string    `json:"notes" binding:"omitempty,max=500"`
	Status      *string    `json:"status" binding:"omitempty,oneof=scheduled in_progress completed cancelled"`
}

// CancelLessonRequest 取消課程請求
type CancelLessonRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=500"`
}

// GetLessonsRequest 獲取課程列表請求
type GetLessonsRequest struct {
	CoachID   *string `json:"coachId" form:"coachId"`
	StudentID *string `json:"studentId" form:"studentId"`
	Status    *string `json:"status" form:"status" binding:"omitempty,oneof=scheduled in_progress completed cancelled"`
	StartDate *string `json:"startDate" form:"startDate"`
	EndDate   *string `json:"endDate" form:"endDate"`
	Page      int     `json:"page" form:"page" binding:"omitempty,min=1"`
	Limit     int     `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
}

// ===== 排程相關 =====

// UpdateScheduleRequest 更新時間表請求
type UpdateScheduleRequest struct {
	Schedules []ScheduleItem `json:"schedules" binding:"required,dive"`
}

// ScheduleItem 時間表項目
type ScheduleItem struct {
	DayOfWeek int    `json:"dayOfWeek" binding:"required,min=0,max=6"`
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
	IsActive  bool   `json:"isActive"`
}

// ===== 智能排課相關 =====

// IntelligentSchedulingRequest 智能排課請求
type IntelligentSchedulingRequest struct {
	StudentID           string    `json:"studentId" binding:"required"`
	NTRPLevel           float64   `json:"ntrpLevel" binding:"omitempty,min=1,max=7"`
	PreferredTimes      []string  `json:"preferredTimes" binding:"omitempty,dive,len=11"`     // ["09:00-12:00", "14:00-18:00"]
	PreferredDays       []int     `json:"preferredDays" binding:"omitempty,dive,min=0,max=6"` // [1,2,3,4,5] (Monday-Friday)
	MaxDistance         float64   `json:"maxDistance" binding:"omitempty,min=0"`              // 公里
	MinPrice            *float64  `json:"minPrice" binding:"omitempty,min=0"`
	MaxPrice            *float64  `json:"maxPrice" binding:"omitempty,min=0"`
	PreferredLessonType string    `json:"preferredLessonType" binding:"omitempty,oneof=individual group clinic"`
	DateRange           []string  `json:"dateRange" binding:"required,dive,len=10"` // ["2024-12-01", "2024-12-02"]
	Location            *Location `json:"location"`
}

// Location 位置信息
type Location struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Address   string  `json:"address" binding:"omitempty,max=200"`
}

// OptimalTimeRequest 最佳時間查詢請求
type OptimalTimeRequest struct {
	CoachID             string    `json:"coachId" binding:"required"`
	StudentID           string    `json:"studentId" binding:"required"`
	NTRPLevel           float64   `json:"ntrpLevel" binding:"omitempty,min=1,max=7"`
	PreferredTimes      []string  `json:"preferredTimes" binding:"omitempty,dive,len=11"`
	PreferredDays       []int     `json:"preferredDays" binding:"omitempty,dive,min=0,max=6"`
	MaxDistance         float64   `json:"maxDistance" binding:"omitempty,min=0"`
	MinPrice            *float64  `json:"minPrice" binding:"omitempty,min=0"`
	MaxPrice            *float64  `json:"maxPrice" binding:"omitempty,min=0"`
	PreferredLessonType string    `json:"preferredLessonType" binding:"omitempty,oneof=individual group clinic"`
	DateRange           []string  `json:"dateRange" binding:"required,dive,len=10"`
	Location            *Location `json:"location"`
}

// ConflictDetectionRequest 衝突檢測請求
type ConflictDetectionRequest struct {
	CoachID         string    `json:"coachId" binding:"required"`
	ScheduledAt     time.Time `json:"scheduledAt" binding:"required"`
	Duration        int       `json:"duration" binding:"required,min=30,max=480"`
	ExcludeLessonID *string   `json:"excludeLessonId"`
}

// ConflictResolutionRequest 衝突解決請求
type ConflictResolutionRequest struct {
	ConflictingLessonID string    `json:"conflictingLessonId" binding:"required"`
	NewScheduledAt      time.Time `json:"newScheduledAt" binding:"required"`
}

// ===== 教練評價相關 =====

// CreateCoachReviewRequest 創建教練評價請求
type CreateCoachReviewRequest struct {
	CoachID  string   `json:"coachId" binding:"required"`
	LessonID *string  `json:"lessonId" binding:"omitempty"`
	Rating   int      `json:"rating" binding:"required,min=1,max=5"`
	Comment  *string  `json:"comment" binding:"omitempty,max=1000"`
	Tags     []string `json:"tags" binding:"omitempty,dive,min=1,max=50"`
}

// UpdateCoachReviewRequest 更新教練評價請求
type UpdateCoachReviewRequest struct {
	Rating  *int     `json:"rating" binding:"omitempty,min=1,max=5"`
	Comment *string  `json:"comment" binding:"omitempty,max=1000"`
	Tags    []string `json:"tags" binding:"omitempty,dive,min=1,max=50"`
}

// CoachReviewSearchRequest 教練評價搜尋請求
type CoachReviewSearchRequest struct {
	CoachID    string   `json:"coachId" form:"coachId" binding:"required"`
	Rating     *int     `json:"rating" form:"rating" binding:"omitempty,min=1,max=5"`
	HasComment *bool    `json:"hasComment" form:"hasComment"`
	Tags       []string `json:"tags" form:"tags"`
	Page       int      `json:"page" form:"page" binding:"omitempty,min=1"`
	Limit      int      `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
	SortBy     string   `json:"sortBy" form:"sortBy" binding:"omitempty,oneof=rating createdAt isHelpful"`
	SortOrder  string   `json:"sortOrder" form:"sortOrder" binding:"omitempty,oneof=asc desc"`
}

// MarkReviewHelpfulRequest 標記評價有用請求
type MarkReviewHelpfulRequest struct {
	ReviewID  string `json:"reviewId" binding:"required"`
	IsHelpful bool   `json:"isHelpful" binding:"required"`
}
