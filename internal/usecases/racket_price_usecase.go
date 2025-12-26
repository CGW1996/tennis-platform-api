package usecases

import (
	"errors"
	"fmt"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// RacketPriceUsecase 球拍價格用例
type RacketPriceUsecase struct {
	db *gorm.DB
}

// NewRacketPriceUsecase 創建新的球拍價格用例
func NewRacketPriceUsecase(db *gorm.DB) *RacketPriceUsecase {
	return &RacketPriceUsecase{
		db: db,
	}
}

// CreateRacketPrice 創建球拍價格
func (u *RacketPriceUsecase) CreateRacketPrice(req *dto.CreateRacketPriceRequest) (*models.RacketPrice, error) {
	// 檢查球拍是否存在
	var racket models.Racket
	err := u.db.Where("id = ? AND deleted_at IS NULL", req.RacketID).First(&racket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("racket not found")
		}
		return nil, fmt.Errorf("failed to check racket: %w", err)
	}

	// 檢查是否已存在相同零售商的價格
	var existingPrice models.RacketPrice
	err = u.db.Where("racket_id = ? AND retailer = ? AND deleted_at IS NULL", req.RacketID, req.Retailer).First(&existingPrice).Error
	if err == nil {
		return nil, errors.New("price for this retailer already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing price: %w", err)
	}

	// 設置默認值
	currency := req.Currency
	if currency == "" {
		currency = "TWD"
	}

	price := &models.RacketPrice{
		RacketID:    req.RacketID,
		Retailer:    req.Retailer,
		Price:       req.Price,
		Currency:    currency,
		URL:         req.URL,
		IsAvailable: req.IsAvailable,
		LastChecked: time.Now(),
	}

	if err := u.db.Create(price).Error; err != nil {
		return nil, fmt.Errorf("failed to create racket price: %w", err)
	}

	return price, nil
}

// UpdateRacketPrice 更新球拍價格
func (u *RacketPriceUsecase) UpdateRacketPrice(priceID string, req *dto.UpdateRacketPriceRequest) (*models.RacketPrice, error) {
	var price models.RacketPrice
	err := u.db.Where("id = ? AND deleted_at IS NULL", priceID).First(&price).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("price not found")
		}
		return nil, fmt.Errorf("failed to find price: %w", err)
	}

	// 檢查零售商是否與其他價格衝突
	if req.Retailer != nil {
		var existingPrice models.RacketPrice
		err := u.db.Where("racket_id = ? AND retailer = ? AND id != ? AND deleted_at IS NULL", price.RacketID, *req.Retailer, priceID).First(&existingPrice).Error
		if err == nil {
			return nil, errors.New("price for this retailer already exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing price: %w", err)
		}
	}

	// 更新欄位
	updates := make(map[string]interface{})
	if req.Retailer != nil {
		updates["retailer"] = *req.Retailer
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.IsAvailable != nil {
		updates["is_available"] = *req.IsAvailable
	}
	updates["last_checked"] = time.Now()

	if err := u.db.Model(&price).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update racket price: %w", err)
	}

	// 重新載入更新後的價格
	if err := u.db.Where("id = ?", priceID).First(&price).Error; err != nil {
		return nil, fmt.Errorf("failed to reload price: %w", err)
	}

	return &price, nil
}

// DeleteRacketPrice 刪除球拍價格
func (u *RacketPriceUsecase) DeleteRacketPrice(priceID string) error {
	var price models.RacketPrice
	err := u.db.Where("id = ? AND deleted_at IS NULL", priceID).First(&price).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("price not found")
		}
		return fmt.Errorf("failed to find price: %w", err)
	}

	if err := u.db.Delete(&price).Error; err != nil {
		return fmt.Errorf("failed to delete racket price: %w", err)
	}

	return nil
}

// GetRacketPrices 獲取球拍價格列表
func (u *RacketPriceUsecase) GetRacketPrices(racketID string) ([]models.RacketPrice, error) {
	var prices []models.RacketPrice
	err := u.db.Where("racket_id = ? AND deleted_at IS NULL", racketID).
		Order("price ASC").
		Find(&prices).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get racket prices: %w", err)
	}

	return prices, nil
}

// GetLowestPrice 獲取最低價格
func (u *RacketPriceUsecase) GetLowestPrice(racketID string) (*models.RacketPrice, error) {
	var price models.RacketPrice
	err := u.db.Where("racket_id = ? AND deleted_at IS NULL AND is_available = true", racketID).
		Order("price ASC").
		First(&price).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 沒有可用價格
		}
		return nil, fmt.Errorf("failed to get lowest price: %w", err)
	}

	return &price, nil
}

// UpdatePriceAvailability 更新價格可用性
func (u *RacketPriceUsecase) UpdatePriceAvailability(priceID string, isAvailable bool) error {
	var price models.RacketPrice
	err := u.db.Where("id = ? AND deleted_at IS NULL", priceID).First(&price).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("price not found")
		}
		return fmt.Errorf("failed to find price: %w", err)
	}

	updates := map[string]interface{}{
		"is_available": isAvailable,
		"last_checked": time.Now(),
	}

	if err := u.db.Model(&price).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update price availability: %w", err)
	}

	return nil
}

// GetPriceComparison 獲取價格比較
func (u *RacketPriceUsecase) GetPriceComparison(racketID string) (*PriceComparisonResponse, error) {
	prices, err := u.GetRacketPrices(racketID)
	if err != nil {
		return nil, err
	}

	if len(prices) == 0 {
		return &PriceComparisonResponse{
			RacketID:       racketID,
			Prices:         []models.RacketPrice{},
			LowestPrice:    nil,
			HighestPrice:   nil,
			AveragePrice:   0,
			PriceRange:     0,
			AvailableCount: 0,
		}, nil
	}

	// 只考慮可用的價格
	var availablePrices []models.RacketPrice
	var totalPrice float64
	var lowestPrice, highestPrice *models.RacketPrice

	for i, price := range prices {
		if price.IsAvailable {
			availablePrices = append(availablePrices, price)
			totalPrice += price.Price

			if lowestPrice == nil || price.Price < lowestPrice.Price {
				lowestPrice = &prices[i]
			}
			if highestPrice == nil || price.Price > highestPrice.Price {
				highestPrice = &prices[i]
			}
		}
	}

	var averagePrice, priceRange float64
	if len(availablePrices) > 0 {
		averagePrice = totalPrice / float64(len(availablePrices))
		if lowestPrice != nil && highestPrice != nil {
			priceRange = highestPrice.Price - lowestPrice.Price
		}
	}

	return &PriceComparisonResponse{
		RacketID:       racketID,
		Prices:         prices,
		LowestPrice:    lowestPrice,
		HighestPrice:   highestPrice,
		AveragePrice:   averagePrice,
		PriceRange:     priceRange,
		AvailableCount: len(availablePrices),
	}, nil
}

// PriceComparisonResponse 價格比較回應
type PriceComparisonResponse struct {
	RacketID       string               `json:"racketId"`
	Prices         []models.RacketPrice `json:"prices"`
	LowestPrice    *models.RacketPrice  `json:"lowestPrice"`
	HighestPrice   *models.RacketPrice  `json:"highestPrice"`
	AveragePrice   float64              `json:"averagePrice"`
	PriceRange     float64              `json:"priceRange"`
	AvailableCount int                  `json:"availableCount"`
}

// BatchUpdatePrices 批量更新價格
func (u *RacketPriceUsecase) BatchUpdatePrices(updates []BatchPriceUpdate) error {
	tx := u.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	for _, update := range updates {
		var price models.RacketPrice
		err := tx.Where("id = ? AND deleted_at IS NULL", update.PriceID).First(&price).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue // 跳過不存在的價格
			}
			return fmt.Errorf("failed to find price %s: %w", update.PriceID, err)
		}

		updateFields := map[string]interface{}{
			"last_checked": time.Now(),
		}

		if update.Price != nil {
			updateFields["price"] = *update.Price
		}
		if update.IsAvailable != nil {
			updateFields["is_available"] = *update.IsAvailable
		}

		if err := tx.Model(&price).Updates(updateFields).Error; err != nil {
			return fmt.Errorf("failed to update price %s: %w", update.PriceID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BatchPriceUpdate 批量價格更新
type BatchPriceUpdate struct {
	PriceID     string   `json:"priceId"`
	Price       *float64 `json:"price,omitempty"`
	IsAvailable *bool    `json:"isAvailable,omitempty"`
}

// GetRetailerStatistics 獲取零售商統計
func (u *RacketPriceUsecase) GetRetailerStatistics() (*RetailerStatistics, error) {
	var stats []RetailerStat
	err := u.db.Model(&models.RacketPrice{}).
		Select("retailer, COUNT(*) as total_prices, AVG(price) as average_price, MIN(price) as min_price, MAX(price) as max_price, COUNT(CASE WHEN is_available = true THEN 1 END) as available_count").
		Where("deleted_at IS NULL").
		Group("retailer").
		Order("total_prices DESC").
		Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get retailer statistics: %w", err)
	}

	var totalRetailers int64
	err = u.db.Model(&models.RacketPrice{}).
		Where("deleted_at IS NULL").
		Distinct("retailer").
		Count(&totalRetailers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count retailers: %w", err)
	}

	return &RetailerStatistics{
		TotalRetailers: int(totalRetailers),
		Retailers:      stats,
	}, nil
}

// RetailerStatistics 零售商統計
type RetailerStatistics struct {
	TotalRetailers int            `json:"totalRetailers"`
	Retailers      []RetailerStat `json:"retailers"`
}

// RetailerStat 零售商統計
type RetailerStat struct {
	Retailer       string  `json:"retailer"`
	TotalPrices    int     `json:"totalPrices"`
	AveragePrice   float64 `json:"averagePrice"`
	MinPrice       float64 `json:"minPrice"`
	MaxPrice       float64 `json:"maxPrice"`
	AvailableCount int     `json:"availableCount"`
}
