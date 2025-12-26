package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// StringArray is a custom type that works with both GORM and Swagger
// swagger:model StringArray
type StringArray []string

// Value implements the driver.Valuer interface for GORM
func (sa StringArray) Value() (driver.Value, error) {
	return pq.StringArray(sa).Value()
}

// Scan implements the sql.Scanner interface for GORM
func (sa *StringArray) Scan(value interface{}) error {
	var pqArray pq.StringArray
	if err := pqArray.Scan(value); err != nil {
		return err
	}
	*sa = StringArray(pqArray)
	return nil
}

// Coach 教練
type Coach struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         string         `json:"userId" gorm:"type:uuid;uniqueIndex;not null"`
	LicenseNumber  *string        `json:"licenseNumber"`
	Certifications StringArray    `json:"certifications" gorm:"type:text[]"`
	Experience     int            `json:"experience"`                     // 年數
	Specialties    StringArray    `json:"specialties" gorm:"type:text[]"` // beginner, intermediate, advanced, junior
	Biography      *string        `json:"biography" gorm:"type:text"`
	HourlyRate     float64        `json:"hourlyRate" gorm:"not null"`
	Currency       string         `json:"currency" gorm:"default:'TWD'"`
	Languages      StringArray    `json:"languages" gorm:"type:text[]"`
	AverageRating  float64        `json:"averageRating" gorm:"default:0"`
	TotalReviews   int            `json:"totalReviews" gorm:"default:0"`
	TotalLessons   int            `json:"totalLessons" gorm:"default:0"`
	IsVerified     bool           `json:"isVerified" gorm:"default:false"`
	IsActive       bool           `json:"isActive" gorm:"default:true"`
	AvailableHours AvailableHours `json:"availableHours" gorm:"type:jsonb"` // {"monday": ["09:00-12:00", "14:00-18:00"]}
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	User    *User         `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Reviews []CoachReview `json:"reviews,omitempty" gorm:"foreignKey:CoachID;constraint:OnDelete:CASCADE"`
	Lessons []Lesson      `json:"lessons,omitempty" gorm:"foreignKey:CoachID;constraint:OnDelete:CASCADE"`
}

// CoachReview 教練評價
type CoachReview struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CoachID   string         `json:"coachId" gorm:"type:uuid;not null"`
	UserID    string         `json:"userId" gorm:"type:uuid;not null"`
	LessonID  *string        `json:"lessonId" gorm:"type:uuid"`
	Rating    int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment   *string        `json:"comment" gorm:"type:text"`
	Tags      []string       `json:"tags" gorm:"type:text[]"` // patient, professional, knowledgeable
	IsHelpful int            `json:"isHelpful" gorm:"default:0"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Coach  *Coach  `json:"coach,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User   *User   `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Lesson *Lesson `json:"lesson,omitempty" gorm:"constraint:OnDelete:SET NULL"`
}

// Lesson 課程
type Lesson struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CoachID      string         `json:"coachId" gorm:"type:uuid;not null"`
	StudentID    string         `json:"studentId" gorm:"type:uuid;not null"`
	LessonTypeID *string        `json:"lessonTypeId" gorm:"type:uuid"`
	CourtID      *string        `json:"courtId" gorm:"type:uuid"`
	Type         string         `json:"type" gorm:"not null"`     // individual, group, clinic
	Level        string         `json:"level"`                    // beginner, intermediate, advanced
	Duration     int            `json:"duration" gorm:"not null"` // 分鐘
	Price        float64        `json:"price" gorm:"not null"`
	Currency     string         `json:"currency" gorm:"default:'TWD'"`
	ScheduledAt  time.Time      `json:"scheduledAt" gorm:"not null"`
	Status       string         `json:"status" gorm:"default:'scheduled'"` // scheduled, in_progress, completed, cancelled
	Notes        *string        `json:"notes" gorm:"type:text"`
	PaymentID    *string        `json:"paymentId"`
	CancelReason *string        `json:"cancelReason" gorm:"type:text"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Coach      *Coach       `json:"coach,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Student    *User        `json:"student,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	LessonType *LessonType  `json:"lessonType,omitempty" gorm:"constraint:OnDelete:SET NULL"`
	Court      *Court       `json:"court,omitempty" gorm:"constraint:OnDelete:SET NULL"`
	Review     *CoachReview `json:"review,omitempty" gorm:"foreignKey:LessonID;constraint:OnDelete:SET NULL"`
}

// LessonType 課程類型
type LessonType struct {
	ID              string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CoachID         string         `json:"coachId" gorm:"type:uuid;not null"`
	Name            string         `json:"name" gorm:"not null"`
	Description     *string        `json:"description" gorm:"type:text"`
	Type            string         `json:"type" gorm:"not null"`     // individual, group, clinic
	Level           string         `json:"level"`                    // beginner, intermediate, advanced
	Duration        int            `json:"duration" gorm:"not null"` // 分鐘
	Price           float64        `json:"price" gorm:"not null"`
	Currency        string         `json:"currency" gorm:"default:'TWD'"`
	MaxParticipants *int           `json:"maxParticipants"`                // 最大參與人數（團體課程用）
	MinParticipants *int           `json:"minParticipants"`                // 最小參與人數（團體課程用）
	Equipment       StringArray    `json:"equipment" gorm:"type:text[]"`   // 需要的設備
	Prerequisites   *string        `json:"prerequisites" gorm:"type:text"` // 先決條件
	IsActive        bool           `json:"isActive" gorm:"default:true"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Coach   *Coach   `json:"coach,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Lessons []Lesson `json:"lessons,omitempty" gorm:"foreignKey:LessonTypeID;constraint:OnDelete:SET NULL"`
}

// LessonSchedule 課程時間表
type LessonSchedule struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CoachID   string    `json:"coachId" gorm:"type:uuid;not null"`
	DayOfWeek int       `json:"dayOfWeek" gorm:"not null;check:day_of_week >= 0 AND day_of_week <= 6"` // 0=Sunday, 6=Saturday
	StartTime string    `json:"startTime" gorm:"not null"`                                             // "09:00"
	EndTime   string    `json:"endTime" gorm:"not null"`                                               // "17:00"
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 關聯
	Coach *Coach `json:"coach,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// TimeSlot 時間段
type TimeSlot struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	IsBooked  bool   `json:"isBooked"`
}

// BeforeCreate 創建前的鉤子
func (c *Coach) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cr *CoachReview) BeforeCreate(tx *gorm.DB) error {
	if cr.ID == "" {
		cr.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (l *Lesson) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (lt *LessonType) BeforeCreate(tx *gorm.DB) error {
	if lt.ID == "" {
		lt.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (ls *LessonSchedule) BeforeCreate(tx *gorm.DB) error {
	if ls.ID == "" {
		ls.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Coach) TableName() string {
	return "coaches"
}

// TableName 指定表名
func (CoachReview) TableName() string {
	return "coach_reviews"
}

// TableName 指定表名
func (Lesson) TableName() string {
	return "lessons"
}

// TableName 指定表名
func (LessonType) TableName() string {
	return "lesson_types"
}

// TableName 指定表名
func (LessonSchedule) TableName() string {
	return "lesson_schedules"
}

// AvailableHours 可用時間自定義類型
type AvailableHours map[string][]string

// Value 實現 driver.Valuer 接口
func (ah AvailableHours) Value() (driver.Value, error) {
	if ah == nil {
		return nil, nil
	}
	return json.Marshal(ah)
}

// Scan 實現 sql.Scanner 接口
func (ah *AvailableHours) Scan(value interface{}) error {
	if value == nil {
		*ah = make(AvailableHours)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, ah)
}
