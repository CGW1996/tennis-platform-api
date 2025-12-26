package dto

import "time"

// UpdateProfileRequest 更新檔案請求
type UpdateProfileRequest struct {
	FirstName         *string    `json:"firstName" binding:"omitempty,min=1,max=100"`
	LastName          *string    `json:"lastName" binding:"omitempty,min=1,max=100"`
	NTRPLevel         *float64   `json:"ntrpLevel" binding:"omitempty,min=1.0,max=7.0"`
	PlayingStyle      *string    `json:"playingStyle" binding:"omitempty,oneof=aggressive defensive all-court"`
	PreferredHand     *string    `json:"preferredHand" binding:"omitempty,oneof=right left both"`
	Latitude          *float64   `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude         *float64   `json:"longitude" binding:"omitempty,min=-180,max=180"`
	Bio               *string    `json:"bio" binding:"omitempty,max=500"`
	BirthDate         *time.Time `json:"birthDate"`
	Gender            *string    `json:"gender" binding:"omitempty,oneof=male female other"`
	PlayingFrequency  *string    `json:"playingFrequency" binding:"omitempty,oneof=casual regular competitive"`
	PreferredTimes    []string   `json:"preferredTimes"`
	MaxTravelDistance *float64   `json:"maxTravelDistance" binding:"omitempty,min=0,max=100"`
	LocationPrivacy   *bool      `json:"locationPrivacy"`
	ProfilePrivacy    *string    `json:"profilePrivacy" binding:"omitempty,oneof=public friends private"`
}

// CreateProfileRequest 創建檔案請求
type CreateProfileRequest struct {
	FirstName         string     `json:"firstName" binding:"required,min=1,max=100"`
	LastName          string     `json:"lastName" binding:"required,min=1,max=100"`
	NTRPLevel         *float64   `json:"ntrpLevel" binding:"omitempty,min=1.0,max=7.0"`
	PlayingStyle      *string    `json:"playingStyle" binding:"omitempty,oneof=aggressive defensive all-court"`
	PreferredHand     *string    `json:"preferredHand" binding:"omitempty,oneof=right left both"`
	Latitude          *float64   `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude         *float64   `json:"longitude" binding:"omitempty,min=-180,max=180"`
	Bio               *string    `json:"bio" binding:"omitempty,max=500"`
	BirthDate         *time.Time `json:"birthDate"`
	Gender            *string    `json:"gender" binding:"omitempty,oneof=male female other"`
	PlayingFrequency  *string    `json:"playingFrequency" binding:"omitempty,oneof=casual regular competitive"`
	PreferredTimes    []string   `json:"preferredTimes"`
	MaxTravelDistance *float64   `json:"maxTravelDistance" binding:"omitempty,min=0,max=100"`
	LocationPrivacy   *bool      `json:"locationPrivacy"`
	ProfilePrivacy    *string    `json:"profilePrivacy" binding:"omitempty,oneof=public friends private"`
}

// UserPreferencesRequest 用戶偏好設定請求
type UserPreferencesRequest struct {
	PlayingStyle      *string  `json:"playingStyle" binding:"omitempty,oneof=aggressive defensive all-court"`
	PlayingFrequency  *string  `json:"playingFrequency" binding:"omitempty,oneof=casual regular competitive"`
	PreferredTimes    []string `json:"preferredTimes"`
	MaxTravelDistance *float64 `json:"maxTravelDistance" binding:"omitempty,min=0,max=100"`
	LocationPrivacy   *bool    `json:"locationPrivacy"`
	ProfilePrivacy    *string  `json:"profilePrivacy" binding:"omitempty,oneof=public friends private"`
}

// LocationUpdateRequest 位置更新請求
type LocationUpdateRequest struct {
	Latitude        *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude       *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`
	LocationPrivacy *bool    `json:"locationPrivacy"`
}
