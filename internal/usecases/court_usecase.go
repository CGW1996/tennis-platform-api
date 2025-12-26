package usecases

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"

	"gorm.io/gorm"
)

// CourtUsecase 場地用例
type CourtUsecase struct {
	db *gorm.DB
}

// NewCourtUsecase 創建新的場地用例
func NewCourtUsecase(db *gorm.DB) *CourtUsecase {
	return &CourtUsecase{
		db: db,
	}
}

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
	Courts     []dto.CourtWithDistance `json:"courts"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	TotalPages int                     `json:"totalPages"`
}

// dto.CourtWithDistance 帶距離的場地
type CourtWithDistance struct {
	*models.Court
	Distance *float64 `json:"distance,omitempty"` // 公里
}

// CreateCourt 創建場地
func (cu *CourtUsecase) CreateCourt(req *dto.CreateCourtRequest) (*models.Court, error) {
	// 驗證營業時間格式
	if err := cu.validateOperatingHours(req.OperatingHours); err != nil {
		return nil, err
	}

	// 驗證設施
	if err := cu.validateFacilities(req.Facilities); err != nil {
		return nil, err
	}

	// 轉換營業時間為 JSON
	operatingHoursJSON, err := json.Marshal(req.OperatingHours)
	if err != nil {
		return nil, errors.New("營業時間格式錯誤")
	}

	// 創建場地
	court := models.Court{
		Name:           req.Name,
		Description:    req.Description,
		Address:        req.Address,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Facilities:     req.Facilities,
		CourtType:      req.CourtType,
		PricePerHour:   req.PricePerHour,
		Currency:       req.Currency,
		Images:         req.Images,
		OperatingHours: operatingHoursJSON,
		ContactPhone:   req.ContactPhone,
		ContactEmail:   req.ContactEmail,
		Website:        req.Website,
		OwnerID:        req.OwnerID,
		IsActive:       true,
	}

	// 設置默認貨幣
	if court.Currency == "" {
		court.Currency = "TWD"
	}

	if err := cu.db.Create(&court).Error; err != nil {
		return nil, errors.New("創建場地失敗")
	}

	return &court, nil
}

// GetCourtByID 根據ID獲取場地
func (cu *CourtUsecase) GetCourtByID(courtID string) (*models.Court, error) {
	var court models.Court
	if err := cu.db.Preload("Reviews", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = 'active'").Order("created_at DESC").Limit(5)
	}).Preload("Reviews.User").Where("id = ? AND deleted_at IS NULL", courtID).First(&court).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("場地不存在")
		}
		return nil, errors.New("獲取場地失敗")
	}
	return &court, nil
}

// UpdateCourt 更新場地
func (cu *CourtUsecase) UpdateCourt(courtID string, req *dto.UpdateCourtRequest) (*models.Court, error) {
	// 檢查場地是否存在
	var court models.Court
	if err := cu.db.Where("id = ? AND deleted_at IS NULL", courtID).First(&court).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("場地不存在")
		}
		return nil, errors.New("獲取場地失敗")
	}

	// 驗證營業時間格式
	if req.OperatingHours != nil {
		if err := cu.validateOperatingHours(req.OperatingHours); err != nil {
			return nil, err
		}
	}

	// 驗證設施
	if req.Facilities != nil {
		if err := cu.validateFacilities(req.Facilities); err != nil {
			return nil, err
		}
	}

	// 準備更新數據
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}
	if req.Facilities != nil {
		updates["facilities"] = req.Facilities
	}
	if req.CourtType != nil {
		updates["court_type"] = *req.CourtType
	}
	if req.PricePerHour != nil {
		updates["price_per_hour"] = *req.PricePerHour
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.Images != nil {
		updates["images"] = req.Images
	}
	if req.OperatingHours != nil {
		updates["operating_hours"] = req.OperatingHours
	}
	if req.ContactPhone != nil {
		updates["contact_phone"] = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		updates["contact_email"] = *req.ContactEmail
	}
	if req.Website != nil {
		updates["website"] = *req.Website
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		if err := cu.db.Model(&court).Updates(updates).Error; err != nil {
			return nil, errors.New("更新場地失敗")
		}
	}

	// 重新載入場地數據
	if err := cu.db.Preload("Reviews").Where("id = ?", courtID).First(&court).Error; err != nil {
		return nil, errors.New("載入場地數據失敗")
	}

	return &court, nil
}

// DeleteCourt 刪除場地（軟刪除）
func (cu *CourtUsecase) DeleteCourt(courtID string) error {
	var court models.Court
	if err := cu.db.Where("id = ? AND deleted_at IS NULL", courtID).First(&court).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("場地不存在")
		}
		return errors.New("獲取場地失敗")
	}

	if err := cu.db.Delete(&court).Error; err != nil {
		return errors.New("刪除場地失敗")
	}

	return nil
}

// SearchCourts 搜尋場地
func (cu *CourtUsecase) SearchCourts(req *dto.CourtSearchRequest) (*dto.CourtSearchResponse, error) {
	// 設置默認值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == nil {
		sortBy := "name"
		req.SortBy = &sortBy
	}
	if req.SortOrder == nil {
		sortOrder := "asc"
		req.SortOrder = &sortOrder
	}

	// 回退到數據庫搜尋
	return cu.search(req)
}

// search 使用數據庫搜尋（回退方案）
func (cu *CourtUsecase) search(req *dto.CourtSearchRequest) (*dto.CourtSearchResponse, error) {
	// 構建查詢
	query := cu.db.Model(&models.Court{}).Where("deleted_at IS NULL AND is_active = true")

	// 價格篩選
	if req.MinPrice != nil {
		query = query.Where("price_per_hour >= ?", *req.MinPrice)
	}
	if req.MaxPrice != nil {
		query = query.Where("price_per_hour <= ?", *req.MaxPrice)
	}

	// 場地類型篩選
	if req.CourtType != nil {
		query = query.Where("court_type = ?", *req.CourtType)
	}

	// 設施篩選
	if len(req.Facilities) > 0 {
		query = query.Where("facilities @> ?", req.Facilities)
	}

	// 評分篩選
	if req.MinRating != nil {
		query = query.Where("average_rating >= ?", *req.MinRating)
	}

	// 地理位置篩選
	var courts []models.Court
	var total int64

	if req.Latitude != nil && req.Longitude != nil && req.Radius != nil {
		// 使用地理位置搜尋
		courts, total = cu.searchCourtsByLocation(query, *req.Latitude, *req.Longitude, *req.Radius, req.Page, req.PageSize, *req.SortBy, *req.SortOrder)
	} else {
		// 普通搜尋
		// 計算總數
		query.Count(&total)

		// 排序
		orderClause := fmt.Sprintf("%s %s", *req.SortBy, *req.SortOrder)
		query = query.Order(orderClause)

		// 分頁
		offset := (req.Page - 1) * req.PageSize
		query = query.Offset(offset).Limit(req.PageSize)

		if err := query.Find(&courts).Error; err != nil {
			return nil, errors.New("搜尋場地失敗")
		}
	}

	// 轉換為帶距離的場地
	courtsWithDistance := make([]dto.CourtWithDistance, len(courts))
	for i, court := range courts {
		courtWithDistance := dto.CourtWithDistance{
			Court: &court,
		}

		// 計算距離
		if req.Latitude != nil && req.Longitude != nil {
			distance := cu.calculateDistance(*req.Latitude, *req.Longitude, court.Latitude, court.Longitude)
			courtWithDistance.Distance = &distance
		}

		courtsWithDistance[i] = courtWithDistance
	}

	// 計算總頁數
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.CourtSearchResponse{
		Courts:     courtsWithDistance,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// searchCourtsByLocation 根據地理位置搜尋場地
func (cu *CourtUsecase) searchCourtsByLocation(baseQuery *gorm.DB, lat, lng, radius float64, page, pageSize int, sortBy, sortOrder string) ([]models.Court, int64) {
	// 使用 PostGIS 進行地理搜尋
	distanceQuery := fmt.Sprintf(
		"ST_DWithin(ST_Point(longitude, latitude)::geography, ST_Point(%f, %f)::geography, %f)",
		lng, lat, radius*1000, // 轉換為米
	)

	query := baseQuery.Where(distanceQuery)

	// 計算總數
	var total int64
	query.Count(&total)

	// 排序
	var orderClause string
	if sortBy == "distance" {
		orderClause = fmt.Sprintf(
			"ST_Distance(ST_Point(longitude, latitude)::geography, ST_Point(%f, %f)::geography) %s",
			lng, lat, sortOrder,
		)
	} else {
		orderClause = fmt.Sprintf("%s %s", sortBy, sortOrder)
	}

	query = query.Order(orderClause)

	// 分頁
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	var courts []models.Court
	query.Find(&courts)

	return courts, total
}

// calculateDistance 計算兩點間距離（公里）
func (cu *CourtUsecase) calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // 地球半徑（公里）

	// 轉換為弧度
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// 計算差值
	deltaLat := lat2Rad - lat1Rad
	deltaLng := lng2Rad - lng1Rad

	// 使用 Haversine 公式
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
}

// validateOperatingHours 驗證營業時間格式
func (cu *CourtUsecase) validateOperatingHours(hours map[string]string) error {
	validDays := map[string]bool{
		"monday":    true,
		"tuesday":   true,
		"wednesday": true,
		"thursday":  true,
		"friday":    true,
		"saturday":  true,
		"sunday":    true,
	}

	for day, timeRange := range hours {
		if !validDays[day] {
			return fmt.Errorf("無效的星期: %s", day)
		}

		// 驗證時間格式 (例如: "09:00-18:00" 或 "closed")
		if timeRange == "closed" {
			continue
		}

		// 簡單的時間格式驗證
		if len(timeRange) < 11 || timeRange[5] != '-' {
			return fmt.Errorf("無效的時間格式: %s，應為 HH:MM-HH:MM 或 closed", timeRange)
		}
	}

	return nil
}

// validateFacilities 驗證設施
func (cu *CourtUsecase) validateFacilities(facilities []string) error {
	validFacilities := map[string]bool{
		"parking":        true,
		"restroom":       true,
		"shower":         true,
		"locker":         true,
		"pro_shop":       true,
		"restaurant":     true,
		"lighting":       true,
		"air_condition":  true,
		"equipment_rent": true,
		"coaching":       true,
		"wifi":           true,
		"wheelchair":     true,
	}

	for _, facility := range facilities {
		if !validFacilities[facility] {
			return fmt.Errorf("無效的設施: %s", facility)
		}
	}

	return nil
}

// GetAvailableFacilities 獲取可用設施列表
func (cu *CourtUsecase) GetAvailableFacilities() []map[string]interface{} {
	facilities := []map[string]interface{}{
		{"key": "parking", "name": "停車場", "icon": "🅿️"},
		{"key": "restroom", "name": "洗手間", "icon": "🚻"},
		{"key": "shower", "name": "淋浴間", "icon": "🚿"},
		{"key": "locker", "name": "置物櫃", "icon": "🗄️"},
		{"key": "pro_shop", "name": "專業用品店", "icon": "🏪"},
		{"key": "restaurant", "name": "餐廳", "icon": "🍽️"},
		{"key": "lighting", "name": "夜間照明", "icon": "💡"},
		{"key": "air_condition", "name": "空調", "icon": "❄️"},
		{"key": "equipment_rent", "name": "器材租借", "icon": "🎾"},
		{"key": "coaching", "name": "教練服務", "icon": "👨‍🏫"},
		{"key": "wifi", "name": "無線網路", "icon": "📶"},
		{"key": "wheelchair", "name": "無障礙設施", "icon": "♿"},
	}

	return facilities
}

// GetCourtTypes 獲取場地類型列表
func (cu *CourtUsecase) GetCourtTypes() []map[string]interface{} {
	types := []map[string]interface{}{
		{"key": "hard", "name": "硬地球場", "description": "最常見的球場類型，適合各種打法"},
		{"key": "clay", "name": "紅土球場", "description": "球速較慢，適合底線型球員"},
		{"key": "grass", "name": "草地球場", "description": "球速快，彈跳低，適合發球上網"},
		{"key": "indoor", "name": "室內球場", "description": "不受天氣影響，全年可用"},
		{"key": "outdoor", "name": "室外球場", "description": "自然環境，通常價格較便宜"},
	}

	return types
}
