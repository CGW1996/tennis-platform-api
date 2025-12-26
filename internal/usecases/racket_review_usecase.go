package usecases

import (
	"errors"
	"fmt"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"

	"gorm.io/gorm"
)

// RacketReviewUsecase 球拍評價用例
type RacketReviewUsecase struct {
	db *gorm.DB
}

// NewRacketReviewUsecase 創建新的球拍評價用例
func NewRacketReviewUsecase(db *gorm.DB) *RacketReviewUsecase {
	return &RacketReviewUsecase{
		db: db,
	}
}

// CreateRacketReview 創建球拍評價
func (u *RacketReviewUsecase) CreateRacketReview(userID string, req *dto.CreateRacketReviewRequest) (*models.RacketReview, error) {
	// 檢查球拍是否存在
	var racket models.Racket
	err := u.db.Where("id = ? AND deleted_at IS NULL", req.RacketID).First(&racket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("racket not found")
		}
		return nil, fmt.Errorf("failed to check racket: %w", err)
	}

	// 檢查用戶是否已經評價過這個球拍
	var existingReview models.RacketReview
	err = u.db.Where("racket_id = ? AND user_id = ? AND deleted_at IS NULL", req.RacketID, userID).First(&existingReview).Error
	if err == nil {
		return nil, errors.New("user has already reviewed this racket")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing review: %w", err)
	}

	review := &models.RacketReview{
		RacketID:      req.RacketID,
		UserID:        userID,
		Rating:        req.Rating,
		PowerRating:   req.PowerRating,
		ControlRating: req.ControlRating,
		ComfortRating: req.ComfortRating,
		Comment:       req.Comment,
		PlayingStyle:  req.PlayingStyle,
		UsageDuration: req.UsageDuration,
	}

	if err := u.db.Create(review).Error; err != nil {
		return nil, fmt.Errorf("failed to create racket review: %w", err)
	}

	// 預載入關聯數據
	if err := u.db.Preload("User").Preload("Racket").Where("id = ?", review.ID).First(review).Error; err != nil {
		return nil, fmt.Errorf("failed to reload review: %w", err)
	}

	return review, nil
}

// GetRacketReview 獲取球拍評價
func (u *RacketReviewUsecase) GetRacketReview(reviewID string) (*models.RacketReview, error) {
	var review models.RacketReview
	err := u.db.Preload("User").Preload("Racket").Where("id = ? AND deleted_at IS NULL", reviewID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("review not found")
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return &review, nil
}

// UpdateRacketReview 更新球拍評價
func (u *RacketReviewUsecase) UpdateRacketReview(reviewID, userID string, req *dto.UpdateRacketReviewRequest) (*models.RacketReview, error) {
	var review models.RacketReview
	err := u.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", reviewID, userID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("review not found or not owned by user")
		}
		return nil, fmt.Errorf("failed to find review: %w", err)
	}

	// 更新欄位
	updates := make(map[string]interface{})
	if req.Rating != nil {
		updates["rating"] = *req.Rating
	}
	if req.PowerRating != nil {
		updates["power_rating"] = *req.PowerRating
	}
	if req.ControlRating != nil {
		updates["control_rating"] = *req.ControlRating
	}
	if req.ComfortRating != nil {
		updates["comfort_rating"] = *req.ComfortRating
	}
	if req.Comment != nil {
		updates["comment"] = *req.Comment
	}
	if req.PlayingStyle != nil {
		updates["playing_style"] = *req.PlayingStyle
	}
	if req.UsageDuration != nil {
		updates["usage_duration"] = *req.UsageDuration
	}

	if err := u.db.Model(&review).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	// 重新載入更新後的評價
	if err := u.db.Preload("User").Preload("Racket").Where("id = ?", reviewID).First(&review).Error; err != nil {
		return nil, fmt.Errorf("failed to reload review: %w", err)
	}

	return &review, nil
}

// DeleteRacketReview 刪除球拍評價
func (u *RacketReviewUsecase) DeleteRacketReview(reviewID, userID string) error {
	var review models.RacketReview
	err := u.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", reviewID, userID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("review not found or not owned by user")
		}
		return fmt.Errorf("failed to find review: %w", err)
	}

	if err := u.db.Delete(&review).Error; err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	return nil
}

// GetRacketReviews 獲取球拍評價列表
func (u *RacketReviewUsecase) GetRacketReviews(req *dto.RacketReviewListRequest) (*dto.RacketReviewListResponse, error) {
	// 設置默認值
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	query := u.db.Model(&models.RacketReview{}).Where("deleted_at IS NULL")

	// 應用篩選條件
	if req.RacketID != nil {
		query = query.Where("racket_id = ?", *req.RacketID)
	}

	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}

	if req.MinRating != nil {
		query = query.Where("rating >= ?", *req.MinRating)
	}

	if req.MaxRating != nil {
		query = query.Where("rating <= ?", *req.MaxRating)
	}

	if req.PlayingStyle != nil {
		query = query.Where("playing_style = ?", *req.PlayingStyle)
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count reviews: %w", err)
	}

	// 應用排序
	sortBy := "created_at"
	sortOrder := "desc"
	if req.SortBy != nil {
		sortBy = *req.SortBy
	}
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	switch sortBy {
	case "rating":
		query = query.Order("rating " + sortOrder)
	case "date":
		query = query.Order("created_at " + sortOrder)
	case "helpful":
		query = query.Order("is_helpful " + sortOrder)
	default:
		query = query.Order("created_at " + sortOrder)
	}

	// 應用分頁
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// 預載入相關數據
	query = query.Preload("User").Preload("Racket")

	var reviews []models.RacketReview
	if err := query.Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.RacketReviewListResponse{
		Reviews:    reviews,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// MarkReviewHelpful 標記評價有用
func (u *RacketReviewUsecase) MarkReviewHelpful(reviewID, userID string, helpful bool) error {
	var review models.RacketReview
	err := u.db.Where("id = ? AND deleted_at IS NULL", reviewID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("review not found")
		}
		return fmt.Errorf("failed to find review: %w", err)
	}

	// 檢查用戶是否是評價的作者（不能標記自己的評價）
	if review.UserID == userID {
		return errors.New("cannot mark own review as helpful")
	}

	// 更新有用計數
	increment := 1
	if !helpful {
		increment = -1
	}

	if err := u.db.Model(&review).Update("is_helpful", gorm.Expr("is_helpful + ?", increment)).Error; err != nil {
		return fmt.Errorf("failed to update helpful count: %w", err)
	}

	return nil
}

// GetRacketReviewStatistics 獲取球拍評價統計
func (u *RacketReviewUsecase) GetRacketReviewStatistics(racketID string) (*dto.RacketReviewStatistics, error) {
	// 檢查球拍是否存在
	var racket models.Racket
	err := u.db.Where("id = ? AND deleted_at IS NULL", racketID).First(&racket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("racket not found")
		}
		return nil, fmt.Errorf("failed to check racket: %w", err)
	}

	stats := &dto.RacketReviewStatistics{
		RacketID:           racketID,
		RatingDistribution: make(map[string]int),
		PlayingStyleStats:  make(map[string]dto.PlayingStyleStat),
	}

	// 基本統計
	var basicStats struct {
		TotalReviews  int     `json:"total_reviews"`
		AverageRating float64 `json:"average_rating"`
	}

	err = u.db.Model(&models.RacketReview{}).
		Select("COUNT(*) as total_reviews, AVG(rating::numeric) as average_rating").
		Where("racket_id = ? AND deleted_at IS NULL", racketID).
		Scan(&basicStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get basic statistics: %w", err)
	}

	stats.TotalReviews = basicStats.TotalReviews
	stats.AverageRating = basicStats.AverageRating

	// 評分分佈
	var ratingDist []struct {
		Rating int `json:"rating"`
		Count  int `json:"count"`
	}

	err = u.db.Model(&models.RacketReview{}).
		Select("rating, COUNT(*) as count").
		Where("racket_id = ? AND deleted_at IS NULL", racketID).
		Group("rating").
		Scan(&ratingDist).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}

	for _, dist := range ratingDist {
		stats.RatingDistribution[fmt.Sprintf("%d", dist.Rating)] = dist.Count
	}

	// 詳細評分統計
	var detailedRatings struct {
		PowerRating   *float64 `json:"power_rating"`
		ControlRating *float64 `json:"control_rating"`
		ComfortRating *float64 `json:"comfort_rating"`
	}

	err = u.db.Model(&models.RacketReview{}).
		Select("AVG(power_rating::numeric) as power_rating, AVG(control_rating::numeric) as control_rating, AVG(comfort_rating::numeric) as comfort_rating").
		Where("racket_id = ? AND deleted_at IS NULL", racketID).
		Scan(&detailedRatings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get detailed ratings: %w", err)
	}

	stats.PowerRating = detailedRatings.PowerRating
	stats.ControlRating = detailedRatings.ControlRating
	stats.ComfortRating = detailedRatings.ComfortRating

	// 打法統計
	var playingStyleStats []struct {
		PlayingStyle  string  `json:"playing_style"`
		Count         int     `json:"count"`
		AverageRating float64 `json:"average_rating"`
	}

	err = u.db.Model(&models.RacketReview{}).
		Select("playing_style, COUNT(*) as count, AVG(rating::numeric) as average_rating").
		Where("racket_id = ? AND deleted_at IS NULL", racketID).
		Group("playing_style").
		Scan(&playingStyleStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get playing style statistics: %w", err)
	}

	for _, stat := range playingStyleStats {
		stats.PlayingStyleStats[stat.PlayingStyle] = dto.PlayingStyleStat{
			Count:         stat.Count,
			AverageRating: stat.AverageRating,
		}
	}

	// 使用時長統計
	var usageStats struct {
		AverageDuration *float64 `json:"average_duration"`
		MinDuration     *int     `json:"min_duration"`
		MaxDuration     *int     `json:"max_duration"`
	}

	err = u.db.Model(&models.RacketReview{}).
		Select("AVG(usage_duration::numeric) as average_duration, MIN(usage_duration) as min_duration, MAX(usage_duration) as max_duration").
		Where("racket_id = ? AND deleted_at IS NULL AND usage_duration IS NOT NULL", racketID).
		Scan(&usageStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get usage statistics: %w", err)
	}

	if usageStats.AverageDuration != nil {
		stats.UsageDurationStats = &dto.UsageDurationStat{
			AverageDuration: *usageStats.AverageDuration,
			MinDuration:     *usageStats.MinDuration,
			MaxDuration:     *usageStats.MaxDuration,
		}
	}

	return stats, nil
}

// GetUserRacketReviews 獲取用戶的球拍評價
func (u *RacketReviewUsecase) GetUserRacketReviews(userID string, page, pageSize int) (*dto.RacketReviewListResponse, error) {
	sortBy := "date"
	sortOrder := "desc"
	req := &dto.RacketReviewListRequest{
		UserID:    &userID,
		Page:      page,
		PageSize:  pageSize,
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
	}

	return u.GetRacketReviews(req)
}

// GetReviewsByPlayingStyle 根據打法獲取評價
func (u *RacketReviewUsecase) GetReviewsByPlayingStyle(racketID, playingStyle string, page, pageSize int) (*dto.RacketReviewListResponse, error) {
	sortBy := "rating"
	sortOrder := "desc"
	req := &dto.RacketReviewListRequest{
		RacketID:     &racketID,
		PlayingStyle: &playingStyle,
		Page:         page,
		PageSize:     pageSize,
		SortBy:       &sortBy,
		SortOrder:    &sortOrder,
	}

	return u.GetRacketReviews(req)
}
