package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Racket 網球拍
type Racket struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Brand          string         `json:"brand" gorm:"not null"`
	Model          string         `json:"model" gorm:"not null"`
	Year           *int           `json:"year"`
	HeadSize       int            `json:"headSize"`                 // 平方英寸
	Weight         int            `json:"weight"`                   // 克
	Balance        *int           `json:"balance"`                  // 毫米
	StringPattern  string         `json:"stringPattern"`            // "16x19"
	BeamWidth      *float64       `json:"beamWidth"`                // 毫米
	Length         int            `json:"length" gorm:"default:27"` // 英寸
	Stiffness      *int           `json:"stiffness"`                // RA值
	SwingWeight    *int           `json:"swingWeight"`
	PowerLevel     *int           `json:"powerLevel" gorm:"check:power_level >= 1 AND power_level <= 10"`
	ControlLevel   *int           `json:"controlLevel" gorm:"check:control_level >= 1 AND control_level <= 10"`
	ManeuverLevel  *int           `json:"maneuverLevel" gorm:"check:maneuver_level >= 1 AND maneuver_level <= 10"`
	StabilityLevel *int           `json:"stabilityLevel" gorm:"check:stability_level >= 1 AND stability_level <= 10"`
	Description    *string        `json:"description" gorm:"type:text"`
	Images         pq.StringArray `json:"images" gorm:"type:text[]" swaggertype:"array,string"`
	MSRP           *float64       `json:"msrp"` // 建議售價
	Currency       string         `json:"currency" gorm:"default:'TWD'"`
	AverageRating  float64        `json:"averageRating" gorm:"default:0"`
	TotalReviews   int            `json:"totalReviews" gorm:"default:0"`
	IsActive       bool           `json:"isActive" gorm:"default:true"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Reviews []RacketReview `json:"reviews,omitempty" gorm:"foreignKey:RacketID;constraint:OnDelete:CASCADE"`
	Prices  []RacketPrice  `json:"prices,omitempty" gorm:"foreignKey:RacketID;constraint:OnDelete:CASCADE"`
}

// RacketReview 球拍評價
type RacketReview struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RacketID      string         `json:"racketId" gorm:"type:uuid;not null"`
	UserID        string         `json:"userId" gorm:"type:uuid;not null"`
	Rating        int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	PowerRating   *int           `json:"powerRating" gorm:"check:power_rating >= 1 AND power_rating <= 5"`
	ControlRating *int           `json:"controlRating" gorm:"check:control_rating >= 1 AND control_rating <= 5"`
	ComfortRating *int           `json:"comfortRating" gorm:"check:comfort_rating >= 1 AND comfort_rating <= 5"`
	Comment       *string        `json:"comment" gorm:"type:text"`
	PlayingStyle  string         `json:"playingStyle"`  // aggressive, defensive, all-court
	UsageDuration *int           `json:"usageDuration"` // 使用月數
	IsHelpful     int            `json:"isHelpful" gorm:"default:0"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Racket *Racket `json:"racket,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User   *User   `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// RacketPrice 球拍價格
type RacketPrice struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RacketID    string         `json:"racketId" gorm:"type:uuid;not null"`
	Retailer    string         `json:"retailer" gorm:"not null"`
	Price       float64        `json:"price" gorm:"not null"`
	Currency    string         `json:"currency" gorm:"default:'TWD'"`
	URL         *string        `json:"url"`
	IsAvailable bool           `json:"isAvailable" gorm:"default:true"`
	LastChecked time.Time      `json:"lastChecked" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Racket *Racket `json:"racket,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// RacketRecommendation 球拍推薦記錄
type RacketRecommendation struct {
	ID        string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string     `json:"userId" gorm:"type:uuid;not null"`
	RacketID  string     `json:"racketId" gorm:"type:uuid;not null"`
	Score     float64    `json:"score" gorm:"not null"`      // 推薦分數 0-100
	Reasons   []string   `json:"reasons" gorm:"type:text[]"` // 推薦原因
	IsClicked bool       `json:"isClicked" gorm:"default:false"`
	ClickedAt *time.Time `json:"clickedAt"`
	CreatedAt time.Time  `json:"createdAt"`

	// 關聯
	User   *User   `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Racket *Racket `json:"racket,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate 創建前的鉤子
func (r *Racket) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (rr *RacketReview) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == "" {
		rr.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (rp *RacketPrice) BeforeCreate(tx *gorm.DB) error {
	if rp.ID == "" {
		rp.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (rr *RacketRecommendation) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == "" {
		rr.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Racket) TableName() string {
	return "rackets"
}

// TableName 指定表名
func (RacketReview) TableName() string {
	return "racket_reviews"
}

// TableName 指定表名
func (RacketPrice) TableName() string {
	return "racket_prices"
}

// TableName 指定表名
func (RacketRecommendation) TableName() string {
	return "racket_recommendations"
}
