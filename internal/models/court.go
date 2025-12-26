package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OperatingHours 營業時間自定義類型
type OperatingHours map[string]string

// Value 實現 driver.Valuer 接口
func (oh OperatingHours) Value() (driver.Value, error) {
	if oh == nil {
		return nil, nil
	}
	return json.Marshal(oh)
}

// Scan 實現 sql.Scanner 接口
func (oh *OperatingHours) Scan(value interface{}) error {
	if value == nil {
		*oh = make(OperatingHours)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, oh)
}

// Court 網球場地
type Court struct {
	ID             string         `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name           string         `json:"name" gorm:"type:text;not null"`
	Description    *string        `json:"description" gorm:"type:text"`
	Address        string         `json:"address" gorm:"type:text;not null"`
	Latitude       float64        `json:"latitude" gorm:"type:numeric;not null"`
	Longitude      float64        `json:"longitude" gorm:"type:numeric;not null"`
	Facilities     pq.StringArray `json:"facilities" gorm:"type:text[]" swaggertype:"array,string"`
	CourtType      string         `json:"courtType" gorm:"type:text"`
	PricePerHour   float64        `json:"pricePerHour" gorm:"type:numeric;not null"`
	Currency       string         `json:"currency" gorm:"type:text;default:'TWD'"`
	Images         pq.StringArray `json:"images" gorm:"type:text[]" swaggertype:"array,string"`
	OperatingHours datatypes.JSON `json:"operatingHours" gorm:"type:jsonb" swaggertype:"object"`
	ContactPhone   *string        `json:"contactPhone" gorm:"type:text"`
	ContactEmail   *string        `json:"contactEmail" gorm:"type:text"`
	Website        *string        `json:"website" gorm:"type:text"`
	AverageRating  float64        `json:"averageRating" gorm:"type:numeric;default:0"`
	TotalReviews   int64          `json:"totalReviews" gorm:"type:bigint;default:0"`
	IsActive       bool           `json:"isActive" gorm:"default:true"`
	OwnerID        *string        `json:"ownerId" gorm:"type:uuid"`

	CreatedAt time.Time      `json:"createdAt" gorm:"type:timestamptz"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"type:timestamptz"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Reviews  []CourtReview `json:"reviews" gorm:"foreignKey:CourtID;constraint:OnDelete:CASCADE"`
	Bookings []Booking     `json:"bookings,omitempty" gorm:"foreignKey:CourtID;constraint:OnDelete:CASCADE"`
}

// CourtReview 場地評價
type CourtReview struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CourtID     string         `json:"courtId" gorm:"type:uuid;not null"`
	UserID      string         `json:"userId" gorm:"type:uuid;not null"`
	Rating      int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment     *string        `json:"comment" gorm:"type:text"`
	Images      pq.StringArray `json:"images" gorm:"type:text[]" swaggertype:"array,string"`
	IsHelpful   int            `json:"isHelpful" gorm:"default:0"`
	IsReported  bool           `json:"isReported" gorm:"default:false"`
	ReportCount int            `json:"reportCount" gorm:"default:0"`
	Status      string         `json:"status" gorm:"default:'active';check:status IN ('active', 'hidden', 'deleted')"`
	ModeratedAt *time.Time     `json:"moderatedAt"`
	ModeratedBy *string        `json:"moderatedBy" gorm:"type:uuid"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Court   *Court         `json:"court,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User    *User          `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Reports []ReviewReport `json:"reports,omitempty" gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE"`
}

// ReviewReport 評價舉報
type ReviewReport struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReviewID  string         `json:"reviewId" gorm:"type:uuid;not null"`
	UserID    string         `json:"userId" gorm:"type:uuid;not null"`
	Reason    string         `json:"reason" gorm:"not null;check:reason IN ('spam', 'inappropriate', 'fake', 'offensive', 'other')"`
	Comment   *string        `json:"comment" gorm:"type:text"`
	Status    string         `json:"status" gorm:"default:'pending';check:status IN ('pending', 'reviewed', 'dismissed')"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Review *CourtReview `json:"review,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User   *User        `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// Booking 場地預訂
type Booking struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CourtID    string         `json:"courtId" gorm:"type:uuid;not null"`
	UserID     string         `json:"userId" gorm:"type:uuid;not null"`
	StartTime  time.Time      `json:"startTime" gorm:"not null"`
	EndTime    time.Time      `json:"endTime" gorm:"not null"`
	TotalPrice float64        `json:"totalPrice" gorm:"not null"`
	Status     string         `json:"status" gorm:"default:'pending'"` // pending, confirmed, cancelled, completed
	PaymentID  *string        `json:"paymentId"`
	Notes      *string        `json:"notes" gorm:"type:text"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Court *Court `json:"court,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User  *User  `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate 創建前的鉤子
func (c *Court) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cr *CourtReview) BeforeCreate(tx *gorm.DB) error {
	if cr.ID == "" {
		cr.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (rr *ReviewReport) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == "" {
		rr.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Court) TableName() string {
	return "courts"
}

// TableName 指定表名
func (CourtReview) TableName() string {
	return "court_reviews"
}

// TableName 指定表名
func (Booking) TableName() string {
	return "bookings"
}

// TableName 指定表名
func (ReviewReport) TableName() string {
	return "review_reports"
}
