package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Club 網球俱樂部
type Club struct {
	ID             string             `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name           string             `json:"name" gorm:"not null"`
	Description    *string            `json:"description" gorm:"type:text"`
	Address        string             `json:"address" gorm:"not null"`
	Latitude       float64            `json:"latitude" gorm:"not null"`
	Longitude      float64            `json:"longitude" gorm:"not null"`
	ContactPhone   *string            `json:"contactPhone"`
	ContactEmail   *string            `json:"contactEmail"`
	Website        *string            `json:"website"`
	Images         pq.StringArray     `json:"images" gorm:"type:text[]" swaggertype:"array,string"`
	Facilities     pq.StringArray     `json:"facilities" gorm:"type:text[]" swaggertype:"array,string"`
	MembershipFees map[string]float64 `json:"membershipFees" gorm:"type:jsonb"` // {"monthly": 2000, "yearly": 20000}
	Currency       string             `json:"currency" gorm:"default:'TWD'"`
	MaxMembers     *int               `json:"maxMembers"`
	CurrentMembers int                `json:"currentMembers" gorm:"default:0"`
	AverageRating  float64            `json:"averageRating" gorm:"default:0"`
	TotalReviews   int                `json:"totalReviews" gorm:"default:0"`
	IsActive       bool               `json:"isActive" gorm:"default:true"`
	OwnerID        *string            `json:"ownerId" gorm:"type:uuid"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt     `json:"-" gorm:"index"`

	// 關聯
	Members []ClubMember `json:"members,omitempty" gorm:"foreignKey:ClubID;constraint:OnDelete:CASCADE"`
	Events  []ClubEvent  `json:"events,omitempty" gorm:"foreignKey:ClubID;constraint:OnDelete:CASCADE"`
	Reviews []ClubReview `json:"reviews,omitempty" gorm:"foreignKey:ClubID;constraint:OnDelete:CASCADE"`
}

// ClubMember 俱樂部會員
type ClubMember struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ClubID         string         `json:"clubId" gorm:"type:uuid;not null"`
	UserID         string         `json:"userId" gorm:"type:uuid;not null"`
	MembershipType string         `json:"membershipType" gorm:"not null"` // monthly, yearly, lifetime
	Status         string         `json:"status" gorm:"default:'active'"` // active, suspended, expired
	JoinedAt       time.Time      `json:"joinedAt" gorm:"not null"`
	ExpiresAt      *time.Time     `json:"expiresAt"`
	PaymentID      *string        `json:"paymentId"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Club *Club `json:"club,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// ClubEvent 俱樂部活動
type ClubEvent struct {
	ID                  string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ClubID              string         `json:"clubId" gorm:"type:uuid;not null"`
	Title               string         `json:"title" gorm:"not null"`
	Description         *string        `json:"description" gorm:"type:text"`
	EventType           string         `json:"eventType" gorm:"not null"` // tournament, social, training, meeting
	StartTime           time.Time      `json:"startTime" gorm:"not null"`
	EndTime             time.Time      `json:"endTime" gorm:"not null"`
	Location            *string        `json:"location"`
	MaxParticipants     *int           `json:"maxParticipants"`
	CurrentParticipants int            `json:"currentParticipants" gorm:"default:0"`
	RegistrationFee     *float64       `json:"registrationFee"`
	Currency            string         `json:"currency" gorm:"default:'TWD'"`
	Images              []string       `json:"images" gorm:"type:text[]"`
	Status              string         `json:"status" gorm:"default:'upcoming'"` // upcoming, ongoing, completed, cancelled
	IsPublic            bool           `json:"isPublic" gorm:"default:false"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Club         *Club                  `json:"club,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	Participants []ClubEventParticipant `json:"participants,omitempty" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
}

// ClubEventParticipant 俱樂部活動參與者
type ClubEventParticipant struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID      string         `json:"eventId" gorm:"type:uuid;not null"`
	UserID       string         `json:"userId" gorm:"type:uuid;not null"`
	Status       string         `json:"status" gorm:"default:'registered'"` // registered, attended, no_show, cancelled
	PaymentID    *string        `json:"paymentId"`
	RegisteredAt time.Time      `json:"registeredAt" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt    time.Time      `json:"createdAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Event *ClubEvent `json:"event,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User  *User      `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// ClubReview 俱樂部評價
type ClubReview struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ClubID    string         `json:"clubId" gorm:"type:uuid;not null"`
	UserID    string         `json:"userId" gorm:"type:uuid;not null"`
	Rating    int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment   *string        `json:"comment" gorm:"type:text"`
	Tags      []string       `json:"tags" gorm:"type:text[]"` // friendly, professional, well-maintained
	IsHelpful int            `json:"isHelpful" gorm:"default:0"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Club *Club `json:"club,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate 創建前的鉤子
func (c *Club) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cm *ClubMember) BeforeCreate(tx *gorm.DB) error {
	if cm.ID == "" {
		cm.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (ce *ClubEvent) BeforeCreate(tx *gorm.DB) error {
	if ce.ID == "" {
		ce.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cep *ClubEventParticipant) BeforeCreate(tx *gorm.DB) error {
	if cep.ID == "" {
		cep.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (cr *ClubReview) BeforeCreate(tx *gorm.DB) error {
	if cr.ID == "" {
		cr.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (Club) TableName() string {
	return "clubs"
}

// TableName 指定表名
func (ClubMember) TableName() string {
	return "club_members"
}

// TableName 指定表名
func (ClubEvent) TableName() string {
	return "club_events"
}

// TableName 指定表名
func (ClubEventParticipant) TableName() string {
	return "club_event_participants"
}

// TableName 指定表名
func (ClubReview) TableName() string {
	return "club_reviews"
}
