package usecases

import (
	"errors"
	"fmt"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"

	"gorm.io/gorm"
)

// ReviewUsecase 評價用例
type ReviewUsecase struct {
	db            *gorm.DB
	uploadService *services.UploadService
}

// NewReviewUsecase 創建新的評價用例
func NewReviewUsecase(db *gorm.DB, uploadService *services.UploadService) *ReviewUsecase {
	return &ReviewUsecase{
		db:            db,
		uploadService: uploadService,
	}
}

// CreateReviewRequest 創建評價請求
type CreateReviewRequest struct {
	CourtID string   `json:"courtId" binding:"required,uuid"`
	Rating  int      `json:"rating" binding:"required,min=1,max=5"`
	Comment *string  `json:"comment" binding:"omitempty,max=1000"`
	Images  []string `json:"images"`
}

// UpdateReviewRequest 更新評價請求
type UpdateReviewRequest struct {
	Rating  *int     `json:"rating" binding:"omitempty,min=1,max=5"`
	Comment *string  `json:"comment" binding:"omitempty,max=1000"`
	Images  []string `json:"images"`
}

// ReportReviewRequest 舉報評價請求
type ReportReviewRequest struct {
	Reason  string  `json:"reason" binding:"required,oneof=spam inappropriate fake offensive other"`
	Comment *string `json:"comment" binding:"omitempty,max=500"`
}

// ReviewListRequest 評價列表請求
type ReviewListRequest struct {
	CourtID   *string `form:"courtId" binding:"omitempty,uuid"`
	UserID    *string `form:"userId" binding:"omitempty,uuid"`
	Rating    *int    `form:"rating" binding:"omitempty,min=1,max=5"`
	SortBy    *string `form:"sortBy" binding:"omitempty,oneof=rating created_at helpful"`
	SortOrder *string `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PageSize  int     `form:"pageSize" binding:"omitempty,min=1,max=50"`
}

// ReviewListResponse 評價列表回應
type ReviewListResponse struct {
	Reviews    []models.CourtReview `json:"reviews"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	TotalPages int                  `json:"totalPages"`
}

// ReviewStatistics 評價統計
type ReviewStatistics struct {
	TotalReviews    int                  `json:"totalReviews"`
	AverageRating   float64              `json:"averageRating"`
	RatingBreakdown map[string]int       `json:"ratingBreakdown"`
	RecentReviews   []models.CourtReview `json:"recentReviews"`
}

// CreateReview 創建評價
func (ru *ReviewUsecase) CreateReview(userID string, req *dto.CreateReviewRequest) (*models.CourtReview, error) {
	// 檢查場地是否存在
	var court models.Court
	if err := ru.db.Where("id = ? AND deleted_at IS NULL", req.CourtID).First(&court).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("場地不存在")
		}
		return nil, errors.New("檢查場地失敗")
	}

	// 檢查用戶是否已經評價過該場地
	var existingReview models.CourtReview
	if err := ru.db.Where("court_id = ? AND user_id = ? AND deleted_at IS NULL", req.CourtID, userID).First(&existingReview).Error; err == nil {
		return nil, errors.New("您已經評價過該場地")
	}

	// 創建評價
	review := models.CourtReview{
		CourtID: req.CourtID,
		UserID:  userID,
		Rating:  req.Rating,
		Comment: req.Comment,
		Images:  req.Images,
		Status:  "active",
	}

	// 開始事務
	tx := ru.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 創建評價
	if err := tx.Create(&review).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("創建評價失敗")
	}

	// 更新場地評分統計
	if err := ru.updateCourtRatingStats(tx, req.CourtID); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新場地評分統計失敗: %v", err)
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交事務失敗")
	}

	// 載入關聯數據
	if err := ru.db.Preload("User").Preload("User.Profile").Where("id = ?", review.ID).First(&review).Error; err != nil {
		return nil, errors.New("載入評價數據失敗")
	}

	return &review, nil
}

// GetReview 獲取評價詳情
func (ru *ReviewUsecase) GetReview(reviewID string) (*models.CourtReview, error) {
	var review models.CourtReview
	if err := ru.db.Preload("User").Preload("User.Profile").Preload("Court").
		Where("id = ? AND deleted_at IS NULL AND status = 'active'", reviewID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("評價不存在")
		}
		return nil, errors.New("獲取評價失敗")
	}
	return &review, nil
}

// UpdateReview 更新評價
func (ru *ReviewUsecase) UpdateReview(reviewID, userID string, req *dto.UpdateReviewRequest) (*models.CourtReview, error) {
	// 檢查評價是否存在且屬於該用戶
	var review models.CourtReview
	if err := ru.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", reviewID, userID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("評價不存在或無權限修改")
		}
		return nil, errors.New("檢查評價失敗")
	}

	// 檢查評價狀態
	if review.Status != "active" {
		return nil, errors.New("該評價無法修改")
	}

	// 準備更新數據
	updates := make(map[string]interface{})
	ratingChanged := false

	if req.Rating != nil && *req.Rating != review.Rating {
		updates["rating"] = *req.Rating
		ratingChanged = true
	}
	if req.Comment != nil {
		updates["comment"] = *req.Comment
	}
	if req.Images != nil {
		updates["images"] = req.Images
	}

	if len(updates) == 0 {
		return &review, nil
	}

	// 開始事務
	tx := ru.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新評價
	if err := tx.Model(&review).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("更新評價失敗")
	}

	// 如果評分改變，更新場地評分統計
	if ratingChanged {
		if err := ru.updateCourtRatingStats(tx, review.CourtID); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新場地評分統計失敗: %v", err)
		}
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交事務失敗")
	}

	// 重新載入評價數據
	if err := ru.db.Preload("User").Preload("User.Profile").Where("id = ?", reviewID).First(&review).Error; err != nil {
		return nil, errors.New("載入評價數據失敗")
	}

	return &review, nil
}

// DeleteReview 刪除評價
func (ru *ReviewUsecase) DeleteReview(reviewID, userID string) error {
	// 檢查評價是否存在且屬於該用戶
	var review models.CourtReview
	if err := ru.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", reviewID, userID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("評價不存在或無權限刪除")
		}
		return errors.New("檢查評價失敗")
	}

	// 開始事務
	tx := ru.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 軟刪除評價
	if err := tx.Delete(&review).Error; err != nil {
		tx.Rollback()
		return errors.New("刪除評價失敗")
	}

	// 更新場地評分統計
	if err := ru.updateCourtRatingStats(tx, review.CourtID); err != nil {
		tx.Rollback()
		return fmt.Errorf("更新場地評分統計失敗: %v", err)
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return errors.New("提交事務失敗")
	}

	return nil
}

// GetReviews 獲取評價列表
func (ru *ReviewUsecase) GetReviews(req *dto.ReviewListRequest) (*dto.ReviewListResponse, error) {
	// 設置默認值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == nil {
		sortBy := "created_at"
		req.SortBy = &sortBy
	}
	if req.SortOrder == nil {
		sortOrder := "desc"
		req.SortOrder = &sortOrder
	}

	// 構建查詢
	query := ru.db.Model(&models.CourtReview{}).
		Preload("User").Preload("User.Profile").Preload("Court").
		Where("deleted_at IS NULL AND status = 'active'")

	// 篩選條件
	if req.CourtID != nil {
		query = query.Where("court_id = ?", *req.CourtID)
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Rating != nil {
		query = query.Where("rating = ?", *req.Rating)
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.New("計算評價總數失敗")
	}

	// 排序
	orderClause := fmt.Sprintf("%s %s", *req.SortBy, *req.SortOrder)
	query = query.Order(orderClause)

	// 分頁
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 執行查詢
	var reviews []models.CourtReview
	if err := query.Find(&reviews).Error; err != nil {
		return nil, errors.New("獲取評價列表失敗")
	}

	// 計算總頁數
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &dto.ReviewListResponse{
		Reviews:    reviews,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ReportReview 舉報評價
func (ru *ReviewUsecase) ReportReview(reviewID, userID string, req *dto.ReportReviewRequest) error {
	// 檢查評價是否存在
	var review models.CourtReview
	if err := ru.db.Where("id = ? AND deleted_at IS NULL", reviewID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("評價不存在")
		}
		return errors.New("檢查評價失敗")
	}

	// 檢查是否已經舉報過
	var existingReport models.ReviewReport
	if err := ru.db.Where("review_id = ? AND user_id = ? AND deleted_at IS NULL", reviewID, userID).First(&existingReport).Error; err == nil {
		return errors.New("您已經舉報過該評價")
	}

	// 創建舉報記錄
	report := models.ReviewReport{
		ReviewID: reviewID,
		UserID:   userID,
		Reason:   req.Reason,
		Comment:  req.Comment,
		Status:   "pending",
	}

	// 開始事務
	tx := ru.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 創建舉報
	if err := tx.Create(&report).Error; err != nil {
		tx.Rollback()
		return errors.New("創建舉報失敗")
	}

	// 更新評價的舉報狀態
	updates := map[string]interface{}{
		"is_reported":  true,
		"report_count": gorm.Expr("report_count + 1"),
	}

	if err := tx.Model(&review).Updates(updates).Error; err != nil {
		tx.Rollback()
		return errors.New("更新評價舉報狀態失敗")
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return errors.New("提交事務失敗")
	}

	return nil
}

// MarkReviewHelpful 標記評價為有用
func (ru *ReviewUsecase) MarkReviewHelpful(reviewID, userID string, helpful bool) error {
	// 檢查評價是否存在
	var review models.CourtReview
	if err := ru.db.Where("id = ? AND deleted_at IS NULL AND status = 'active'", reviewID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("評價不存在")
		}
		return errors.New("檢查評價失敗")
	}

	// 不能對自己的評價標記有用
	if review.UserID == userID {
		return errors.New("不能對自己的評價標記有用")
	}

	// 更新有用計數
	var increment int
	if helpful {
		increment = 1
	} else {
		increment = -1
	}

	if err := ru.db.Model(&review).Update("is_helpful", gorm.Expr("is_helpful + ?", increment)).Error; err != nil {
		return errors.New("更新評價有用性失敗")
	}

	return nil
}

// GetReviewStatistics 獲取評價統計
func (ru *ReviewUsecase) GetReviewStatistics(courtID string) (*dto.ReviewStatistics, error) {
	// 檢查場地是否存在
	var court models.Court
	if err := ru.db.Where("id = ? AND deleted_at IS NULL", courtID).First(&court).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("場地不存在")
		}
		return nil, errors.New("檢查場地失敗")
	}

	// 獲取評分分佈
	var ratingBreakdown []struct {
		Rating int   `json:"rating"`
		Count  int64 `json:"count"`
	}

	if err := ru.db.Model(&models.CourtReview{}).
		Select("rating, COUNT(*) as count").
		Where("court_id = ? AND deleted_at IS NULL AND status = 'active'", courtID).
		Group("rating").
		Order("rating DESC").
		Find(&ratingBreakdown).Error; err != nil {
		return nil, errors.New("獲取評分分佈失敗")
	}

	// 轉換為 map
	ratingMap := make(map[string]int)
	for _, item := range ratingBreakdown {
		ratingMap[fmt.Sprintf("%d", item.Rating)] = int(item.Count)
	}

	// 獲取最近的評價
	var recentReviews []models.CourtReview
	if err := ru.db.Preload("User").Preload("User.Profile").
		Where("court_id = ? AND deleted_at IS NULL AND status = 'active'", courtID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentReviews).Error; err != nil {
		return nil, errors.New("獲取最近評價失敗")
	}

	return &dto.ReviewStatistics{
		TotalReviews:    int(court.TotalReviews),
		AverageRating:   court.AverageRating,
		RatingBreakdown: ratingMap,
		RecentReviews:   recentReviews,
	}, nil
}

// updateCourtRatingStats 更新場地評分統計
func (ru *ReviewUsecase) updateCourtRatingStats(tx *gorm.DB, courtID string) error {
	// 計算平均評分和總評價數
	var stats struct {
		TotalReviews  int64   `json:"total_reviews"`
		AverageRating float64 `json:"average_rating"`
	}

	if err := tx.Model(&models.CourtReview{}).
		Select("COUNT(*) as total_reviews, COALESCE(AVG(rating), 0) as average_rating").
		Where("court_id = ? AND deleted_at IS NULL AND status = 'active'", courtID).
		Scan(&stats).Error; err != nil {
		return err
	}

	// 更新場地統計
	updates := map[string]interface{}{
		"total_reviews":  stats.TotalReviews,
		"average_rating": stats.AverageRating,
	}

	if err := tx.Model(&models.Court{}).Where("id = ?", courtID).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}
