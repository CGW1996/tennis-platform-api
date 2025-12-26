package usecases

import (
	"errors"
	"fmt"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"

	"gorm.io/gorm"
)

// RacketUsecase 球拍用例
type RacketUsecase struct {
	db *gorm.DB
}

// NewRacketUsecase 創建新的球拍用例
func NewRacketUsecase(db *gorm.DB) *RacketUsecase {
	return &RacketUsecase{
		db: db,
	}
}

// CreateRacket 創建球拍
func (u *RacketUsecase) CreateRacket(req *dto.CreateRacketRequest) (*models.Racket, error) {
	// 檢查是否已存在相同品牌和型號的球拍
	var existingRacket models.Racket
	err := u.db.Where("brand = ? AND model = ? AND deleted_at IS NULL", req.Brand, req.Model).First(&existingRacket).Error
	if err == nil {
		return nil, errors.New("racket with same brand and model already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing racket: %w", err)
	}

	// 設置默認值
	length := req.Length
	if length == 0 {
		length = 27 // 默認27英寸
	}

	currency := req.Currency
	if currency == "" {
		currency = "TWD"
	}

	racket := &models.Racket{
		Brand:          req.Brand,
		Model:          req.Model,
		Year:           req.Year,
		HeadSize:       req.HeadSize,
		Weight:         req.Weight,
		Balance:        req.Balance,
		StringPattern:  req.StringPattern,
		BeamWidth:      req.BeamWidth,
		Length:         length,
		Stiffness:      req.Stiffness,
		SwingWeight:    req.SwingWeight,
		PowerLevel:     req.PowerLevel,
		ControlLevel:   req.ControlLevel,
		ManeuverLevel:  req.ManeuverLevel,
		StabilityLevel: req.StabilityLevel,
		Description:    req.Description,
		Images:         req.Images,
		MSRP:           req.MSRP,
		Currency:       currency,
		IsActive:       true,
	}

	if err := u.db.Create(racket).Error; err != nil {
		return nil, fmt.Errorf("failed to create racket: %w", err)
	}

	return racket, nil
}

// GetRacketByID 根據ID獲取球拍
func (u *RacketUsecase) GetRacketByID(racketID string) (*models.Racket, error) {
	var racket models.Racket
	err := u.db.Preload("Reviews").Preload("Prices").Where("id = ? AND deleted_at IS NULL", racketID).First(&racket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("racket not found")
		}
		return nil, fmt.Errorf("failed to get racket: %w", err)
	}

	return &racket, nil
}

// UpdateRacket 更新球拍
func (u *RacketUsecase) UpdateRacket(racketID string, req *dto.UpdateRacketRequest) (*models.Racket, error) {
	var racket models.Racket
	err := u.db.Where("id = ? AND deleted_at IS NULL", racketID).First(&racket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("racket not found")
		}
		return nil, fmt.Errorf("failed to find racket: %w", err)
	}

	// 檢查品牌和型號是否與其他球拍衝突
	if req.Brand != nil || req.Model != nil {
		brand := racket.Brand
		model := racket.Model
		if req.Brand != nil {
			brand = *req.Brand
		}
		if req.Model != nil {
			model = *req.Model
		}

		var existingRacket models.Racket
		err := u.db.Where("brand = ? AND model = ? AND id != ? AND deleted_at IS NULL", brand, model, racketID).First(&existingRacket).Error
		if err == nil {
			return nil, errors.New("racket with same brand and model already exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing racket: %w", err)
		}
	}

	// 更新欄位
	updates := make(map[string]interface{})
	if req.Brand != nil {
		updates["brand"] = *req.Brand
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.Year != nil {
		updates["year"] = *req.Year
	}
	if req.HeadSize != nil {
		updates["head_size"] = *req.HeadSize
	}
	if req.Weight != nil {
		updates["weight"] = *req.Weight
	}
	if req.Balance != nil {
		updates["balance"] = *req.Balance
	}
	if req.StringPattern != nil {
		updates["string_pattern"] = *req.StringPattern
	}
	if req.BeamWidth != nil {
		updates["beam_width"] = *req.BeamWidth
	}
	if req.Length != nil {
		updates["length"] = *req.Length
	}
	if req.Stiffness != nil {
		updates["stiffness"] = *req.Stiffness
	}
	if req.SwingWeight != nil {
		updates["swing_weight"] = *req.SwingWeight
	}
	if req.PowerLevel != nil {
		updates["power_level"] = *req.PowerLevel
	}
	if req.ControlLevel != nil {
		updates["control_level"] = *req.ControlLevel
	}
	if req.ManeuverLevel != nil {
		updates["maneuver_level"] = *req.ManeuverLevel
	}
	if req.StabilityLevel != nil {
		updates["stability_level"] = *req.StabilityLevel
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Images != nil {
		updates["images"] = req.Images
	}
	if req.MSRP != nil {
		updates["msrp"] = *req.MSRP
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := u.db.Model(&racket).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update racket: %w", err)
	}

	// 重新載入更新後的球拍
	if err := u.db.Preload("Reviews").Preload("Prices").Where("id = ?", racketID).First(&racket).Error; err != nil {
		return nil, fmt.Errorf("failed to reload racket: %w", err)
	}

	return &racket, nil
}

// DeleteRacket 刪除球拍
func (u *RacketUsecase) DeleteRacket(racketID string) error {
	var racket models.Racket
	err := u.db.Where("id = ? AND deleted_at IS NULL", racketID).First(&racket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("racket not found")
		}
		return fmt.Errorf("failed to find racket: %w", err)
	}

	if err := u.db.Delete(&racket).Error; err != nil {
		return fmt.Errorf("failed to delete racket: %w", err)
	}

	return nil
}

// SearchRackets 搜尋球拍
func (u *RacketUsecase) SearchRackets(req *dto.RacketSearchRequest) (*dto.RacketSearchResponse, error) {
	// 設置默認值
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	query := u.db.Model(&models.Racket{}).Where("deleted_at IS NULL AND is_active = true")

	// 應用篩選條件
	if req.Query != nil && *req.Query != "" {
		searchTerm := "%" + *req.Query + "%"
		query = query.Where("brand ILIKE ? OR model ILIKE ? OR description ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	if req.Brand != nil && *req.Brand != "" {
		query = query.Where("brand ILIKE ?", "%"+*req.Brand+"%")
	}

	if req.MinHeadSize != nil {
		query = query.Where("head_size >= ?", *req.MinHeadSize)
	}
	if req.MaxHeadSize != nil {
		query = query.Where("head_size <= ?", *req.MaxHeadSize)
	}

	if req.MinWeight != nil {
		query = query.Where("weight >= ?", *req.MinWeight)
	}
	if req.MaxWeight != nil {
		query = query.Where("weight <= ?", *req.MaxWeight)
	}

	if req.PowerLevel != nil {
		query = query.Where("power_level = ?", *req.PowerLevel)
	}
	if req.ControlLevel != nil {
		query = query.Where("control_level = ?", *req.ControlLevel)
	}
	if req.ManeuverLevel != nil {
		query = query.Where("maneuver_level = ?", *req.ManeuverLevel)
	}
	if req.StabilityLevel != nil {
		query = query.Where("stability_level = ?", *req.StabilityLevel)
	}

	if req.MinRating != nil {
		query = query.Where("average_rating >= ?", *req.MinRating)
	}

	// 價格篩選需要子查詢
	if req.MinPrice != nil || req.MaxPrice != nil {
		priceQuery := u.db.Model(&models.RacketPrice{}).
			Select("racket_id").
			Where("deleted_at IS NULL AND is_available = true")

		if req.MinPrice != nil {
			priceQuery = priceQuery.Where("price >= ?", *req.MinPrice)
		}
		if req.MaxPrice != nil {
			priceQuery = priceQuery.Where("price <= ?", *req.MaxPrice)
		}

		query = query.Where("id IN (?)", priceQuery)
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count rackets: %w", err)
	}

	// 應用排序
	sortBy := "brand"
	sortOrder := "asc"
	if req.SortBy != nil {
		sortBy = *req.SortBy
	}
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	switch sortBy {
	case "brand":
		query = query.Order("brand " + sortOrder + ", model " + sortOrder)
	case "model":
		query = query.Order("model " + sortOrder)
	case "price":
		// 按最低價格排序
		query = query.Joins("LEFT JOIN racket_prices ON rackets.id = racket_prices.racket_id AND racket_prices.deleted_at IS NULL AND racket_prices.is_available = true").
			Group("rackets.id").
			Order("MIN(racket_prices.price) " + sortOrder)
	case "rating":
		query = query.Order("average_rating " + sortOrder)
	case "popularity":
		query = query.Order("total_reviews " + sortOrder)
	default:
		query = query.Order("brand " + sortOrder + ", model " + sortOrder)
	}

	// 應用分頁
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// 預載入相關數據
	query = query.Preload("Prices", "deleted_at IS NULL AND is_available = true")

	var rackets []models.Racket
	if err := query.Find(&rackets).Error; err != nil {
		return nil, fmt.Errorf("failed to search rackets: %w", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.RacketSearchResponse{
		Rackets:    rackets,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAvailableBrands 獲取可用品牌列表
func (u *RacketUsecase) GetAvailableBrands() ([]string, error) {
	var brands []string
	err := u.db.Model(&models.Racket{}).
		Where("deleted_at IS NULL AND is_active = true").
		Distinct("brand").
		Order("brand").
		Pluck("brand", &brands).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get available brands: %w", err)
	}

	return brands, nil
}

// GetRacketSpecifications 獲取球拍規格選項
func (u *RacketUsecase) GetRacketSpecifications() map[string]interface{} {
	return map[string]interface{}{
		"headSizeRanges": []map[string]interface{}{
			{"label": "Midsize (85-97 sq in)", "min": 85, "max": 97},
			{"label": "Midplus (98-105 sq in)", "min": 98, "max": 105},
			{"label": "Oversize (106+ sq in)", "min": 106, "max": 140},
		},
		"weightRanges": []map[string]interface{}{
			{"label": "Light (250-280g)", "min": 250, "max": 280},
			{"label": "Medium (281-310g)", "min": 281, "max": 310},
			{"label": "Heavy (311g+)", "min": 311, "max": 400},
		},
		"stringPatterns": []string{
			"16x19", "16x20", "18x20", "16x18", "14x18", "12x18",
		},
		"currencies": []string{"TWD", "USD", "EUR"},
		"levels": []map[string]interface{}{
			{"value": 1, "label": "Very Low"},
			{"value": 2, "label": "Low"},
			{"value": 3, "label": "Low-Medium"},
			{"value": 4, "label": "Medium"},
			{"value": 5, "label": "Medium"},
			{"value": 6, "label": "Medium-High"},
			{"value": 7, "label": "High"},
			{"value": 8, "label": "High"},
			{"value": 9, "label": "Very High"},
			{"value": 10, "label": "Maximum"},
		},
	}
}
