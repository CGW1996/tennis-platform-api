package dto

import "tennis-platform/backend/internal/models"

// ===== 球拍相關 =====

// CreateRacketRequest 創建球拍請求
type CreateRacketRequest struct {
	Brand          string   `json:"brand" binding:"required,min=1,max=100"`
	Model          string   `json:"model" binding:"required,min=1,max=100"`
	Year           *int     `json:"year" binding:"omitempty,min=1900,max=2030"`
	HeadSize       int      `json:"headSize" binding:"required,min=80,max=140"`
	Weight         int      `json:"weight" binding:"required,min=200,max=400"`
	Balance        *int     `json:"balance" binding:"omitempty,min=280,max=380"`
	StringPattern  string   `json:"stringPattern" binding:"required"`
	BeamWidth      *float64 `json:"beamWidth" binding:"omitempty,min=15,max=35"`
	Length         int      `json:"length" binding:"omitempty,min=26,max=29"`
	Stiffness      *int     `json:"stiffness" binding:"omitempty,min=40,max=80"`
	SwingWeight    *int     `json:"swingWeight" binding:"omitempty,min=250,max=400"`
	PowerLevel     *int     `json:"powerLevel" binding:"omitempty,min=1,max=10"`
	ControlLevel   *int     `json:"controlLevel" binding:"omitempty,min=1,max=10"`
	ManeuverLevel  *int     `json:"maneuverLevel" binding:"omitempty,min=1,max=10"`
	StabilityLevel *int     `json:"stabilityLevel" binding:"omitempty,min=1,max=10"`
	Description    *string  `json:"description" binding:"omitempty,max=2000"`
	Images         []string `json:"images"`
	MSRP           *float64 `json:"msrp" binding:"omitempty,min=0"`
	Currency       string   `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
}

// UpdateRacketRequest 更新球拍請求
type UpdateRacketRequest struct {
	Brand          *string  `json:"brand" binding:"omitempty,min=1,max=100"`
	Model          *string  `json:"model" binding:"omitempty,min=1,max=100"`
	Year           *int     `json:"year" binding:"omitempty,min=1900,max=2030"`
	HeadSize       *int     `json:"headSize" binding:"omitempty,min=80,max=140"`
	Weight         *int     `json:"weight" binding:"omitempty,min=200,max=400"`
	Balance        *int     `json:"balance" binding:"omitempty,min=280,max=380"`
	StringPattern  *string  `json:"stringPattern"`
	BeamWidth      *float64 `json:"beamWidth" binding:"omitempty,min=15,max=35"`
	Length         *int     `json:"length" binding:"omitempty,min=26,max=29"`
	Stiffness      *int     `json:"stiffness" binding:"omitempty,min=40,max=80"`
	SwingWeight    *int     `json:"swingWeight" binding:"omitempty,min=250,max=400"`
	PowerLevel     *int     `json:"powerLevel" binding:"omitempty,min=1,max=10"`
	ControlLevel   *int     `json:"controlLevel" binding:"omitempty,min=1,max=10"`
	ManeuverLevel  *int     `json:"maneuverLevel" binding:"omitempty,min=1,max=10"`
	StabilityLevel *int     `json:"stabilityLevel" binding:"omitempty,min=1,max=10"`
	Description    *string  `json:"description" binding:"omitempty,max=2000"`
	Images         []string `json:"images"`
	MSRP           *float64 `json:"msrp" binding:"omitempty,min=0"`
	Currency       *string  `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	IsActive       *bool    `json:"isActive"`
}

// RacketSearchRequest 球拍搜尋請求
type RacketSearchRequest struct {
	Query          *string  `form:"query"`
	Brand          *string  `form:"brand"`
	MinHeadSize    *int     `form:"minHeadSize" binding:"omitempty,min=80"`
	MaxHeadSize    *int     `form:"maxHeadSize" binding:"omitempty,max=140"`
	MinWeight      *int     `form:"minWeight" binding:"omitempty,min=200"`
	MaxWeight      *int     `form:"maxWeight" binding:"omitempty,max=400"`
	MinPrice       *float64 `form:"minPrice" binding:"omitempty,min=0"`
	MaxPrice       *float64 `form:"maxPrice" binding:"omitempty,min=0"`
	PowerLevel     *int     `form:"powerLevel" binding:"omitempty,min=1,max=10"`
	ControlLevel   *int     `form:"controlLevel" binding:"omitempty,min=1,max=10"`
	ManeuverLevel  *int     `form:"maneuverLevel" binding:"omitempty,min=1,max=10"`
	StabilityLevel *int     `form:"stabilityLevel" binding:"omitempty,min=1,max=10"`
	MinRating      *float64 `form:"minRating" binding:"omitempty,min=0,max=5"`
	SortBy         *string  `form:"sortBy" binding:"omitempty,oneof=brand model price rating popularity"`
	SortOrder      *string  `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page           int      `form:"page" binding:"omitempty,min=1"`
	PageSize       int      `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// RacketSearchResponse 球拍搜尋回應
type RacketSearchResponse struct {
	Rackets    []models.Racket `json:"rackets"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"pageSize"`
	TotalPages int             `json:"totalPages"`
}

// ===== 球拍價格相關 =====

// CreateRacketPriceRequest 創建球拍價格請求
type CreateRacketPriceRequest struct {
	RacketID    string  `json:"racketId" binding:"required"`
	Retailer    string  `json:"retailer" binding:"required,min=1,max=100"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Currency    string  `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	URL         *string `json:"url" binding:"omitempty,url"`
	IsAvailable bool    `json:"isAvailable"`
}

// UpdateRacketPriceRequest 更新球拍價格請求
type UpdateRacketPriceRequest struct {
	Retailer    *string  `json:"retailer" binding:"omitempty,min=1,max=100"`
	Price       *float64 `json:"price" binding:"omitempty,min=0"`
	Currency    *string  `json:"currency" binding:"omitempty,oneof=TWD USD EUR"`
	URL         *string  `json:"url" binding:"omitempty,url"`
	IsAvailable *bool    `json:"isAvailable"`
}

// PriceComparisonResponse 價格比較響應
type PriceComparisonResponse struct {
	RacketID     string               `json:"racketId"`
	Brand        string               `json:"brand"`
	Model        string               `json:"model"`
	LowestPrice  *models.RacketPrice  `json:"lowestPrice,omitempty"`
	HighestPrice *models.RacketPrice  `json:"highestPrice,omitempty"`
	AveragePrice float64              `json:"averagePrice"`
	PriceRange   float64              `json:"priceRange"`
	TotalPrices  int                  `json:"totalPrices"`
	AllPrices    []models.RacketPrice `json:"allPrices"`
}

// ===== 球拍評價相關 =====

// CreateRacketReviewRequest 創建球拍評價請求
type CreateRacketReviewRequest struct {
	RacketID      string  `json:"racketId" binding:"required"`
	Rating        int     `json:"rating" binding:"required,min=1,max=5"`
	PowerRating   *int    `json:"powerRating" binding:"omitempty,min=1,max=5"`
	ControlRating *int    `json:"controlRating" binding:"omitempty,min=1,max=5"`
	ComfortRating *int    `json:"comfortRating" binding:"omitempty,min=1,max=5"`
	Comment       *string `json:"comment" binding:"omitempty,max=2000"`
	PlayingStyle  string  `json:"playingStyle" binding:"required,oneof=aggressive defensive all-court"`
	UsageDuration *int    `json:"usageDuration" binding:"omitempty,min=0,max=120"` // 使用月數
}

// UpdateRacketReviewRequest 更新球拍評價請求
type UpdateRacketReviewRequest struct {
	Rating        *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	PowerRating   *int    `json:"powerRating" binding:"omitempty,min=1,max=5"`
	ControlRating *int    `json:"controlRating" binding:"omitempty,min=1,max=5"`
	ComfortRating *int    `json:"comfortRating" binding:"omitempty,min=1,max=5"`
	Comment       *string `json:"comment" binding:"omitempty,max=2000"`
	PlayingStyle  *string `json:"playingStyle" binding:"omitempty,oneof=aggressive defensive all-court"`
	UsageDuration *int    `json:"usageDuration" binding:"omitempty,min=0,max=120"`
}

// RacketReviewListRequest 球拍評價列表請求
type RacketReviewListRequest struct {
	RacketID     *string `form:"racketId"`
	UserID       *string `form:"userId"`
	MinRating    *int    `form:"minRating" binding:"omitempty,min=1,max=5"`
	MaxRating    *int    `form:"maxRating" binding:"omitempty,min=1,max=5"`
	PlayingStyle *string `form:"playingStyle" binding:"omitempty,oneof=aggressive defensive all-court"`
	SortBy       *string `form:"sortBy" binding:"omitempty,oneof=rating date helpful"`
	SortOrder    *string `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page         int     `form:"page" binding:"omitempty,min=1"`
	PageSize     int     `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// RacketReviewListResponse 球拍評價列表回應
type RacketReviewListResponse struct {
	Reviews    []models.RacketReview `json:"reviews"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	TotalPages int                   `json:"totalPages"`
}

// RacketReviewStatistics 球拍評價統計
type RacketReviewStatistics struct {
	RacketID           string                      `json:"racketId"`
	TotalReviews       int                         `json:"totalReviews"`
	AverageRating      float64                     `json:"averageRating"`
	RatingDistribution map[string]int              `json:"ratingDistribution"`
	PowerRating        *float64                    `json:"powerRating,omitempty"`
	ControlRating      *float64                    `json:"controlRating,omitempty"`
	ComfortRating      *float64                    `json:"comfortRating,omitempty"`
	PlayingStyleStats  map[string]PlayingStyleStat `json:"playingStyleStats"`
	UsageDurationStats *UsageDurationStat          `json:"usageDurationStats,omitempty"`
}

// PlayingStyleStat 打法風格統計
type PlayingStyleStat struct {
	Count         int     `json:"count"`
	AverageRating float64 `json:"averageRating"`
}

// UsageDurationStat 使用時長統計
type UsageDurationStat struct {
	AverageDuration float64 `json:"averageDuration"`
	MinDuration     int     `json:"minDuration"`
	MaxDuration     int     `json:"maxDuration"`
}
