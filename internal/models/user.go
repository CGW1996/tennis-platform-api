package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// User 用戶基本資訊
type User struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email         string         `json:"email" gorm:"uniqueIndex;not null"`
	Phone         *string        `json:"phone" gorm:"uniqueIndex"`
	PasswordHash  string         `json:"-" gorm:"not null"`
	EmailVerified bool           `json:"emailVerified" gorm:"default:false"`
	PhoneVerified bool           `json:"phoneVerified" gorm:"default:false"`
	IsActive      bool           `json:"isActive" gorm:"default:true"`
	LastLoginAt   *time.Time     `json:"lastLoginAt"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 關聯
	Profile       *UserProfile   `json:"profile,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	OAuthAccounts []OAuthAccount `json:"oauthAccounts,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	CourtReviews  []CourtReview  `json:"courtReviews,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	CoachReviews  []CoachReview  `json:"coachReviews,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	RacketReviews []RacketReview `json:"racketReviews,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Matches       []Match        `json:"matches,omitempty" gorm:"many2many:match_participants;constraint:OnDelete:CASCADE"`
	Lessons       []Lesson       `json:"lessons,omitempty" gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	ClubMembers   []ClubMember   `json:"clubMembers,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// UserProfile 用戶詳細檔案
type UserProfile struct {
	UserID            string         `json:"userId" gorm:"type:uuid;primaryKey"`
	FirstName         string         `json:"firstName" gorm:"not null"`
	LastName          string         `json:"lastName" gorm:"not null"`
	AvatarURL         *string        `json:"avatarUrl"`
	NTRPLevel         *float64       `json:"ntrpLevel" gorm:"type:decimal(2,1);check:ntrp_level >= 1.0 AND ntrp_level <= 7.0"`
	PlayingStyle      *string        `json:"playingStyle"` // aggressive, defensive, all-court
	PreferredHand     *string        `json:"preferredHand" gorm:"check:preferred_hand IN ('right', 'left', 'both')"`
	Latitude          *float64       `json:"latitude"`
	Longitude         *float64       `json:"longitude"`
	LocationPrivacy   bool           `json:"locationPrivacy" gorm:"default:false"` // true = 隱藏精確位置
	Bio               *string        `json:"bio" gorm:"type:text"`
	BirthDate         *time.Time     `json:"birthDate" gorm:"type:date"`
	Gender            *string        `json:"gender" gorm:"check:gender IN ('male', 'female', 'other')"`
	PlayingFrequency  *string        `json:"playingFrequency"`             // casual, regular, competitive
	PlayTypes         pq.StringArray `json:"playTypes" gorm:"type:text[]"` // rally, singles, doubles
	PreferredTimes    pq.StringArray `json:"preferredTimes" gorm:"type:text[]" swaggertype:"array,string"`
	MaxTravelDistance *float64       `json:"maxTravelDistance"` // 公里
	ProfilePrivacy    string         `json:"profilePrivacy" gorm:"default:'public';check:profile_privacy IN ('public', 'friends', 'private')"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`

	// 關聯
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// OAuthAccount OAuth 帳號關聯
type OAuthAccount struct {
	ID           string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string     `json:"userId" gorm:"type:uuid;not null"`
	Provider     string     `json:"provider" gorm:"not null"` // google, facebook, apple
	ProviderID   string     `json:"providerId" gorm:"not null"`
	Email        string     `json:"email"`
	AccessToken  *string    `json:"-"`
	RefreshToken *string    `json:"-"`
	ExpiresAt    *time.Time `json:"expiresAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	// 關聯
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// RefreshToken JWT Refresh Token
type RefreshToken struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"userId" gorm:"type:uuid;not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"not null"`
	IsRevoked bool      `json:"isRevoked" gorm:"default:false"`
	CreatedAt time.Time `json:"createdAt"`

	// 關聯
	User *User `json:"user,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// BeforeCreate 創建前的鉤子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (oa *OAuthAccount) BeforeCreate(tx *gorm.DB) error {
	if oa.ID == "" {
		oa.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate 創建前的鉤子
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == "" {
		rt.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// TableName 指定表名
func (UserProfile) TableName() string {
	return "user_profiles"
}

// TableName 指定表名
func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}

// TableName 指定表名
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
