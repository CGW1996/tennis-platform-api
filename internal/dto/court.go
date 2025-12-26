package dto

import (
	"tennis-platform/backend/internal/models"
	"time"
)

// ===== 場地相關 =====

// CreateCourtRequest 創建場地請求
type CreateCourtRequest struct {
	Name           string            `json:"name" binding:"required,min=1,max=200"`
	Description    *string           `json:"description" binding:"omitempty,max=1000"`
	Address        string            `json:"address" binding:"required,min=1,max=500"`
	Latitude       float64           `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude      float64           `json:"longitude" binding:"required,min=-180,max=180"`
	Facilities     []string          `json:"facilities"`
	CourtType      string            `json:"courtType" binding:"required,oneof=hard clay grass indoor outdoor"`
	PricePerHour   float64           `json:"pricePerHour" binding:"required,min=0"`
	Currency       string            `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	Images         []string          `json:"images"`
	OperatingHours map[string]string `json:"operatingHours"`
	ContactPhone   *string           `json:"contactPhone" binding:"omitempty,max=20"`
	ContactEmail   *string           `json:"contactEmail" binding:"omitempty,email"`
	Website        *string           `json:"website" binding:"omitempty,url"`
	OwnerID        *string           `json:"ownerId"`
}

// UpdateCourtRequest 更新場地請求
type UpdateCourtRequest struct {
	Name           *string           `json:"name" binding:"omitempty,min=1,max=200"`
	Description    *string           `json:"description" binding:"omitempty,max=1000"`
	Address        *string           `json:"address" binding:"omitempty,min=1,max=500"`
	Latitude       *float64          `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude      *float64          `json:"longitude" binding:"omitempty,min=-180,max=180"`
	Facilities     []string          `json:"facilities"`
	CourtType      *string           `json:"courtType" binding:"omitempty,oneof=hard clay grass indoor outdoor"`
	PricePerHour   *float64          `json:"pricePerHour" binding:"omitempty,min=0"`
	Currency       *string           `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	Images         []string          `json:"images"`
	OperatingHours map[string]string `json:"operatingHours"`
	ContactPhone   *string           `json:"contactPhone" binding:"omitempty,max=20"`
	ContactEmail   *string           `json:"contactEmail" binding:"omitempty,email"`
	Website        *string           `json:"website" binding:"omitempty,url"`
	IsActive       *bool             `json:"isActive"`
}

// CourtSearchRequest 場地搜尋請求
type CourtSearchRequest struct {
	Query      *string  `form:"query"` // 文字搜尋
	Latitude   *float64 `form:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude  *float64 `form:"longitude" binding:"omitempty,min=-180,max=180"`
	Radius     *float64 `form:"radius" binding:"omitempty,min=0,max=100"` // 公里
	MinPrice   *float64 `form:"minPrice" binding:"omitempty,min=0"`
	MaxPrice   *float64 `form:"maxPrice" binding:"omitempty,min=0"`
	CourtType  *string  `form:"courtType" binding:"omitempty,oneof=hard clay grass indoor outdoor"`
	Facilities []string `form:"facilities"`
	MinRating  *float64 `form:"minRating" binding:"omitempty,min=0,max=5"`
	SortBy     *string  `form:"sortBy" binding:"omitempty,oneof=distance price rating name"`
	SortOrder  *string  `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page       int      `form:"page" binding:"omitempty,min=1"`
	PageSize   int      `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// CourtSearchResponse 場地搜尋回應
type CourtSearchResponse struct {
	Courts     []CourtWithDistance `json:"courts"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"pageSize"`
	TotalPages int                 `json:"totalPages"`
}

// CourtWithDistance 帶距離的場地
type CourtWithDistance struct {
	*models.Court
	Distance *float64 `json:"distance,omitempty"` // 公里
}

// ===== 預訂相關 =====

// CreateBookingRequest 創建預訂請求
type CreateBookingRequest struct {
	CourtID   string    `json:"courtId" binding:"required,uuid"`
	StartTime time.Time `json:"startTime" binding:"required"`
	EndTime   time.Time `json:"endTime" binding:"required"`
	Notes     *string   `json:"notes" binding:"omitempty,max=500"`
}

// UpdateBookingRequest 更新預訂請求
type UpdateBookingRequest struct {
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	Notes     *string    `json:"notes" binding:"omitempty,max=500"`
	Status    *string    `json:"status" binding:"omitempty,oneof=pending confirmed cancelled completed"`
}

// BookingListRequest 預訂列表請求
type BookingListRequest struct {
	CourtID   *string    `form:"courtId" binding:"omitempty,uuid"`
	UserID    *string    `form:"userId" binding:"omitempty,uuid"`
	Status    *string    `form:"status" binding:"omitempty,oneof=pending confirmed cancelled completed"`
	StartDate *time.Time `form:"startDate"`
	EndDate   *time.Time `form:"endDate"`
	Page      int        `form:"page" binding:"omitempty,min=1"`
	PageSize  int        `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// BookingListResponse 預訂列表回應
type BookingListResponse struct {
	Bookings   []models.Booking `json:"bookings"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

// AvailabilityRequest 可用時間查詢請求
type AvailabilityRequest struct {
	CourtID  string    `form:"courtId" binding:"required,uuid"`
	Date     time.Time `form:"date" binding:"required"`
	Duration int       `form:"duration" binding:"omitempty,min=30,max=480"` // 分鐘，默認60分鐘
}

// TimeSlot 時間段
type TimeSlot struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Available bool      `json:"available"`
	Price     float64   `json:"price"`
}

// AvailabilityResponse 可用時間回應
type AvailabilityResponse struct {
	Date      time.Time  `json:"date"`
	CourtID   string     `json:"courtId"`
	TimeSlots []TimeSlot `json:"timeSlots"`
}

// ===== 評價相關 =====

// CreateReviewRequest 創建評價請求
type CreateReviewRequest struct {
	CourtID string   `json:"courtId" binding:"required,uuid"`
	Rating  int      `json:"rating" binding:"required,min=1,max=5"`
	Comment *string  `json:"comment" binding:"omitempty,max=1000"`
	Images  []string `json:"images"`
}

// UpdateReviewRequest 更新評價請求
type UpdateReviewRequest struct {
	Rating  *int     `json:"rating" binding:"omitempty,min=1,max=5"`
	Comment *string  `json:"comment" binding:"omitempty,max=1000"`
	Images  []string `json:"images"`
}

// ReportReviewRequest 舉報評價請求
type ReportReviewRequest struct {
	Reason  string  `json:"reason" binding:"required,oneof=spam inappropriate fake offensive other"`
	Comment *string `json:"comment" binding:"omitempty,max=500"`
}

// ReviewListRequest 評價列表請求
type ReviewListRequest struct {
	CourtID   *string `form:"courtId" binding:"omitempty,uuid"`
	UserID    *string `form:"userId" binding:"omitempty,uuid"`
	Rating    *int    `form:"rating" binding:"omitempty,min=1,max=5"`
	SortBy    *string `form:"sortBy" binding:"omitempty,oneof=rating created_at helpful"`
	SortOrder *string `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PageSize  int     `form:"pageSize" binding:"omitempty,min=1,max=50"`
}

// ReviewListResponse 評價列表回應
type ReviewListResponse struct {
	Reviews    []models.CourtReview `json:"reviews"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	TotalPages int                  `json:"totalPages"`
}

// ReviewStatistics 評價統計
type ReviewStatistics struct {
	TotalReviews    int                  `json:"totalReviews"`
	AverageRating   float64              `json:"averageRating"`
	RatingBreakdown map[string]int       `json:"ratingBreakdown"`
	RecentReviews   []models.CourtReview `json:"recentReviews"`
}
