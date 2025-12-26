package usecases

import (
	"errors"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// UserUsecase 用戶用例
type UserUsecase struct {
	db *gorm.DB
}

// NewUserUsecase 創建新的用戶用例
func NewUserUsecase(db *gorm.DB) *UserUsecase {
	return &UserUsecase{
		db: db,
	}
}

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

// GetUserByID 根據ID獲取用戶
func (uu *UserUsecase) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}
	return &user, nil
}

// CreateUserProfile 創建用戶檔案
func (uu *UserUsecase) CreateUserProfile(userID string, req *dto.CreateProfileRequest) (*models.User, error) {
	// 檢查用戶是否存在
	var user models.User
	if err := uu.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 檢查是否已有檔案
	var existingProfile models.UserProfile
	if err := uu.db.Where("user_id = ?", userID).First(&existingProfile).Error; err == nil {
		return nil, errors.New("用戶檔案已存在")
	}

	// 創建新檔案
	profile := models.UserProfile{
		UserID:            userID,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		NTRPLevel:         req.NTRPLevel,
		PlayingStyle:      req.PlayingStyle,
		PreferredHand:     req.PreferredHand,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		Bio:               req.Bio,
		BirthDate:         req.BirthDate,
		Gender:            req.Gender,
		PlayingFrequency:  req.PlayingFrequency,
		PreferredTimes:    pq.StringArray(req.PreferredTimes),
		MaxTravelDistance: req.MaxTravelDistance,
	}

	if err := uu.db.Create(&profile).Error; err != nil {
		return nil, errors.New("創建用戶檔案失敗")
	}

	// 重新載入用戶數據
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("載入用戶數據失敗")
	}

	return &user, nil
}

// UpdateUserProfile 更新用戶檔案
func (uu *UserUsecase) UpdateUserProfile(userID string, req *dto.UpdateProfileRequest) (*models.User, error) {
	// 查找用戶
	var user models.User
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 如果用戶沒有檔案，先創建一個
	if user.Profile == nil {
		profile := models.UserProfile{
			UserID: userID,
		}
		if err := uu.db.Create(&profile).Error; err != nil {
			return nil, errors.New("創建用戶檔案失敗")
		}
	}

	// 驗證 NTRP 等級
	if req.NTRPLevel != nil {
		if err := uu.validateNTRPLevel(*req.NTRPLevel); err != nil {
			return nil, err
		}
	}

	// 更新檔案
	updates := make(map[string]interface{})

	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.NTRPLevel != nil {
		updates["ntrp_level"] = *req.NTRPLevel
	}
	if req.PlayingStyle != nil {
		updates["playing_style"] = *req.PlayingStyle
	}
	if req.PreferredHand != nil {
		updates["preferred_hand"] = *req.PreferredHand
	}
	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.BirthDate != nil {
		updates["birth_date"] = *req.BirthDate
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.PlayingFrequency != nil {
		updates["playing_frequency"] = *req.PlayingFrequency
	}
	if req.PreferredTimes != nil {
		updates["preferred_times"] = pq.StringArray(req.PreferredTimes)
	}
	if req.MaxTravelDistance != nil {
		updates["max_travel_distance"] = *req.MaxTravelDistance
	}

	if len(updates) > 0 {
		if err := uu.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			return nil, errors.New("更新檔案失敗")
		}
	}

	// 重新載入用戶數據
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("載入用戶數據失敗")
	}

	return &user, nil
}

// UpdateUserPreferences 更新用戶偏好設定
func (uu *UserUsecase) UpdateUserPreferences(userID string, req *dto.UserPreferencesRequest) (*models.User, error) {
	// 查找用戶
	var user models.User
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 如果用戶沒有檔案，先創建一個
	if user.Profile == nil {
		profile := models.UserProfile{
			UserID: userID,
		}
		if err := uu.db.Create(&profile).Error; err != nil {
			return nil, errors.New("創建用戶檔案失敗")
		}
	}

	// 更新偏好設定
	updates := make(map[string]interface{})

	if req.PlayingStyle != nil {
		updates["playing_style"] = *req.PlayingStyle
	}
	if req.PlayingFrequency != nil {
		updates["playing_frequency"] = *req.PlayingFrequency
	}
	if req.PreferredTimes != nil {
		updates["preferred_times"] = pq.StringArray(req.PreferredTimes)
	}
	if req.MaxTravelDistance != nil {
		updates["max_travel_distance"] = *req.MaxTravelDistance
	}

	if len(updates) > 0 {
		if err := uu.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			return nil, errors.New("更新偏好設定失敗")
		}
	}

	// 重新載入用戶數據
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("載入用戶數據失敗")
	}

	return &user, nil
}

// UpdateUserLocation 更新用戶位置
func (uu *UserUsecase) UpdateUserLocation(userID string, req *dto.LocationUpdateRequest) (*models.User, error) {
	// 查找用戶
	var user models.User
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 如果用戶沒有檔案，先創建一個
	if user.Profile == nil {
		profile := models.UserProfile{
			UserID: userID,
		}
		if err := uu.db.Create(&profile).Error; err != nil {
			return nil, errors.New("創建用戶檔案失敗")
		}
	}

	// 更新位置信息
	updates := make(map[string]interface{})

	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}

	if len(updates) > 0 {
		if err := uu.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			return nil, errors.New("更新位置信息失敗")
		}
	}

	// 重新載入用戶數據
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("載入用戶數據失敗")
	}

	return &user, nil
}

// UpdateUserAvatar 更新用戶頭像
func (uu *UserUsecase) UpdateUserAvatar(userID string, avatarURL string) (*models.User, error) {
	// 查找用戶
	var user models.User
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 如果用戶沒有檔案，先創建一個
	if user.Profile == nil {
		profile := models.UserProfile{
			UserID: userID,
		}
		if err := uu.db.Create(&profile).Error; err != nil {
			return nil, errors.New("創建用戶檔案失敗")
		}
	}

	// 更新頭像URL
	if err := uu.db.Model(&models.UserProfile{}).Where("user_id = ?", userID).Update("avatar_url", avatarURL).Error; err != nil {
		return nil, errors.New("更新頭像失敗")
	}

	// 重新載入用戶數據
	if err := uu.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("載入用戶數據失敗")
	}

	return &user, nil
}

// validateNTRPLevel 驗證 NTRP 等級
func (uu *UserUsecase) validateNTRPLevel(level float64) error {
	if level < 1.0 || level > 7.0 {
		return errors.New("NTRP 等級必須在 1.0 到 7.0 之間")
	}

	// 檢查是否為有效的 NTRP 等級（0.5 的倍數）
	validLevels := []float64{1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0, 5.5, 6.0, 6.5, 7.0}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return nil
		}
	}

	return errors.New("NTRP 等級必須為 0.5 的倍數（例如：1.0, 1.5, 2.0...）")
}

// GetNTRPLevelDescription 獲取 NTRP 等級描述
func (uu *UserUsecase) GetNTRPLevelDescription(level float64) string {
	descriptions := map[float64]string{
		1.0: "新手：剛開始學習網球",
		1.5: "新手：有限的網球經驗",
		2.0: "初學者：需要在場上練習",
		2.5: "初學者：學習基本技巧",
		3.0: "初級：能夠持續擊球",
		3.5: "初級：有一定的一致性和方向控制",
		4.0: "中級：能夠使用各種擊球技巧",
		4.5: "中級：開始掌握戰術",
		5.0: "高級：良好的擊球技巧和戰術意識",
		5.5: "高級：能夠在比賽中運用各種戰術",
		6.0: "專業級：具備專業水準的技巧",
		6.5: "專業級：接近職業選手水準",
		7.0: "職業級：世界級職業選手",
	}

	if desc, exists := descriptions[level]; exists {
		return desc
	}
	return "未知等級"
}
