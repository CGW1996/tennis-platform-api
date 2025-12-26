package usecases

import (
	"encoding/json"
	"errors"
	"fmt"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"time"

	"gorm.io/gorm"
)

// BookingUsecase 預訂用例
type BookingUsecase struct {
	db                  *gorm.DB
	notificationService services.NotificationService
}

// NewBookingUsecase 創建新的預訂用例
func NewBookingUsecase(db *gorm.DB, notificationService services.NotificationService) *BookingUsecase {
	return &BookingUsecase{
		db:                  db,
		notificationService: notificationService,
	}
}

// CreateBookingRequest 創建預訂請求
type CreateBookingRequest struct {
	CourtID   string    `json:"courtId" binding:"required,uuid"`
	StartTime time.Time `json:"startTime" binding:"required"`
	EndTime   time.Time `json:"endTime" binding:"required"`
	Notes     *string   `json:"notes" binding:"omitempty,max=500"`
}

// UpdateBookingRequest 更新預訂請求
type UpdateBookingRequest struct {
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	Notes     *string    `json:"notes" binding:"omitempty,max=500"`
	Status    *string    `json:"status" binding:"omitempty,oneof=pending confirmed cancelled completed"`
}

// BookingListRequest 預訂列表請求
type BookingListRequest struct {
	CourtID   *string    `form:"courtId" binding:"omitempty,uuid"`
	UserID    *string    `form:"userId" binding:"omitempty,uuid"`
	Status    *string    `form:"status" binding:"omitempty,oneof=pending confirmed cancelled completed"`
	StartDate *time.Time `form:"startDate"`
	EndDate   *time.Time `form:"endDate"`
	Page      int        `form:"page" binding:"omitempty,min=1"`
	PageSize  int        `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// BookingListResponse 預訂列表回應
type BookingListResponse struct {
	Bookings   []models.Booking `json:"bookings"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

// AvailabilityRequest 可用時間查詢請求
type AvailabilityRequest struct {
	CourtID  string    `form:"courtId" binding:"required,uuid"`
	Date     time.Time `form:"date" binding:"required"`
	Duration int       `form:"duration" binding:"omitempty,min=30,max=480"` // 分鐘，默認60分鐘
}

// dto.TimeSlot 時間段
type TimeSlot struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Available bool      `json:"available"`
	Price     float64   `json:"price"`
}

// AvailabilityResponse 可用時間回應
type AvailabilityResponse struct {
	Date      time.Time      `json:"date"`
	CourtID   string         `json:"courtId"`
	TimeSlots []dto.TimeSlot `json:"timeSlots"`
}

// CreateBooking 創建預訂
func (bu *BookingUsecase) CreateBooking(userID string, req *dto.CreateBookingRequest) (*models.Booking, error) {
	// 驗證時間
	if err := bu.validateBookingTime(req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// 檢查場地是否存在且活躍
	var court models.Court
	if err := bu.db.Where("id = ? AND deleted_at IS NULL AND is_active = true", req.CourtID).First(&court).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("場地不存在或不可用")
		}
		return nil, errors.New("查詢場地失敗")
	}

	// 檢查時間衝突
	if err := bu.checkTimeConflict(req.CourtID, req.StartTime, req.EndTime, ""); err != nil {
		return nil, err
	}

	// 檢查場地營業時間
	if err := bu.checkOperatingHours(&court, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// 計算總價格
	duration := req.EndTime.Sub(req.StartTime).Hours()
	totalPrice := duration * court.PricePerHour

	// 創建預訂
	booking := models.Booking{
		CourtID:    req.CourtID,
		UserID:     userID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		TotalPrice: totalPrice,
		Status:     "pending",
		Notes:      req.Notes,
	}

	if err := bu.db.Create(&booking).Error; err != nil {
		return nil, errors.New("創建預訂失敗")
	}

	// 載入關聯數據
	if err := bu.db.Preload("Court").Preload("User").First(&booking, booking.ID).Error; err != nil {
		return nil, errors.New("載入預訂數據失敗")
	}

	// 發送預訂確認通知
	if bu.notificationService != nil {
		if err := bu.notificationService.SendBookingConfirmation(&booking); err != nil {
			// 記錄錯誤但不影響主要流程
			fmt.Printf("Warning: Failed to send booking confirmation notification: %v\n", err)
		}
	}

	return &booking, nil
}

// GetBooking 獲取預訂詳情
func (bu *BookingUsecase) GetBooking(bookingID string) (*models.Booking, error) {
	var booking models.Booking
	if err := bu.db.Preload("Court").Preload("User").Where("id = ? AND deleted_at IS NULL", bookingID).First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("預訂不存在")
		}
		return nil, errors.New("獲取預訂失敗")
	}
	return &booking, nil
}

// UpdateBooking 更新預訂
func (bu *BookingUsecase) UpdateBooking(bookingID, userID string, req *dto.UpdateBookingRequest) (*models.Booking, error) {
	// 獲取現有預訂
	var booking models.Booking
	if err := bu.db.Where("id = ? AND deleted_at IS NULL", bookingID).First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("預訂不存在")
		}
		return nil, errors.New("獲取預訂失敗")
	}

	// 檢查權限（只有預訂者可以修改）
	if booking.UserID != userID {
		return nil, errors.New("無權限修改此預訂")
	}

	// 檢查預訂狀態（只有 pending 狀態可以修改時間）
	if (req.StartTime != nil || req.EndTime != nil) && booking.Status != "pending" {
		return nil, errors.New("只有待確認的預訂可以修改時間")
	}

	// 準備更新數據
	updates := make(map[string]interface{})

	// 如果要更新時間，需要重新驗證
	if req.StartTime != nil || req.EndTime != nil {
		startTime := booking.StartTime
		endTime := booking.EndTime

		if req.StartTime != nil {
			startTime = *req.StartTime
		}
		if req.EndTime != nil {
			endTime = *req.EndTime
		}

		// 驗證新時間
		if err := bu.validateBookingTime(startTime, endTime); err != nil {
			return nil, err
		}

		// 檢查時間衝突（排除當前預訂）
		if err := bu.checkTimeConflict(booking.CourtID, startTime, endTime, bookingID); err != nil {
			return nil, err
		}

		// 獲取場地信息檢查營業時間
		var court models.Court
		if err := bu.db.Where("id = ?", booking.CourtID).First(&court).Error; err != nil {
			return nil, errors.New("獲取場地信息失敗")
		}

		if err := bu.checkOperatingHours(&court, startTime, endTime); err != nil {
			return nil, err
		}

		// 重新計算價格
		duration := endTime.Sub(startTime).Hours()
		totalPrice := duration * court.PricePerHour

		updates["start_time"] = startTime
		updates["end_time"] = endTime
		updates["total_price"] = totalPrice
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	// 記錄舊狀態用於通知
	oldStatus := booking.Status

	// 執行更新
	if len(updates) > 0 {
		if err := bu.db.Model(&booking).Updates(updates).Error; err != nil {
			return nil, errors.New("更新預訂失敗")
		}
	}

	// 重新載入數據
	if err := bu.db.Preload("Court").Preload("User").First(&booking, bookingID).Error; err != nil {
		return nil, errors.New("載入預訂數據失敗")
	}

	// 如果狀態有變更，發送通知
	if bu.notificationService != nil && req.Status != nil && oldStatus != *req.Status {
		if err := bu.notificationService.SendBookingStatusUpdate(&booking, oldStatus); err != nil {
			// 記錄錯誤但不影響主要流程
			fmt.Printf("Warning: Failed to send booking status update notification: %v\n", err)
		}
	}

	return &booking, nil
}

// CancelBooking 取消預訂
func (bu *BookingUsecase) CancelBooking(bookingID, userID string) error {
	var booking models.Booking
	if err := bu.db.Where("id = ? AND deleted_at IS NULL", bookingID).First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("預訂不存在")
		}
		return errors.New("獲取預訂失敗")
	}

	// 檢查權限
	if booking.UserID != userID {
		return errors.New("無權限取消此預訂")
	}

	// 檢查預訂狀態
	if booking.Status == "cancelled" {
		return errors.New("預訂已經取消")
	}

	if booking.Status == "completed" {
		return errors.New("已完成的預訂無法取消")
	}

	// 檢查取消時間限制（例如：開始時間前2小時才能取消）
	if time.Now().Add(2 * time.Hour).After(booking.StartTime) {
		return errors.New("預訂開始前2小時內無法取消")
	}

	// 更新狀態為取消
	if err := bu.db.Model(&booking).Update("status", "cancelled").Error; err != nil {
		return errors.New("取消預訂失敗")
	}

	// 載入關聯數據用於通知
	if err := bu.db.Preload("Court").Preload("User").First(&booking, bookingID).Error; err == nil {
		// 發送取消通知
		if bu.notificationService != nil {
			if err := bu.notificationService.SendBookingCancellation(&booking); err != nil {
				// 記錄錯誤但不影響主要流程
				fmt.Printf("Warning: Failed to send booking cancellation notification: %v\n", err)
			}
		}
	}

	return nil
}

// GetBookings 獲取預訂列表
func (bu *BookingUsecase) GetBookings(req *dto.BookingListRequest) (*dto.BookingListResponse, error) {
	// 設置默認值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 構建查詢
	query := bu.db.Model(&models.Booking{}).Where("deleted_at IS NULL")

	if req.CourtID != nil {
		query = query.Where("court_id = ?", *req.CourtID)
	}

	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if req.StartDate != nil {
		query = query.Where("start_time >= ?", *req.StartDate)
	}

	if req.EndDate != nil {
		query = query.Where("end_time <= ?", *req.EndDate)
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.New("計算預訂總數失敗")
	}

	// 分頁查詢
	var bookings []models.Booking
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Court").Preload("User").
		Order("start_time DESC").
		Offset(offset).Limit(req.PageSize).
		Find(&bookings).Error; err != nil {
		return nil, errors.New("獲取預訂列表失敗")
	}

	// 計算總頁數
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &dto.BookingListResponse{
		Bookings:   bookings,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAvailability 獲取場地可用時間
func (bu *BookingUsecase) GetAvailability(req *dto.AvailabilityRequest) (*dto.AvailabilityResponse, error) {
	// 設置默認時長
	if req.Duration <= 0 {
		req.Duration = 60 // 默認1小時
	}

	// 檢查場地是否存在
	var court models.Court
	if err := bu.db.Where("id = ? AND deleted_at IS NULL AND is_active = true", req.CourtID).First(&court).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("場地不存在或不可用")
		}
		return nil, errors.New("查詢場地失敗")
	}

	// 獲取當天的營業時間
	weekday := req.Date.Weekday().String()

	// 解析 OperatingHours JSON 到 map
	var operatingHours map[string]string
	if err := json.Unmarshal(court.OperatingHours, &operatingHours); err != nil {
		return nil, fmt.Errorf("解析營業時間失敗: %v", err)
	}

	hours, exists := operatingHours[weekday]
	if !exists || hours == "closed" {
		return &dto.AvailabilityResponse{
			Date:      req.Date,
			CourtID:   req.CourtID,
			TimeSlots: []dto.TimeSlot{},
		}, nil
	}

	// 解析營業時間
	openTime, closeTime, err := bu.parseOperatingHours(hours)
	if err != nil {
		return nil, fmt.Errorf("解析營業時間失敗: %v", err)
	}

	// 獲取當天已有的預訂
	startOfDay := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var existingBookings []models.Booking
	if err := bu.db.Where("court_id = ? AND status IN (?, ?) AND start_time < ? AND end_time > ?",
		req.CourtID, "pending", "confirmed", endOfDay, startOfDay).Find(&existingBookings).Error; err != nil {
		return nil, errors.New("獲取現有預訂失敗")
	}

	// 生成時間段
	timeSlots := bu.generateTimeSlots(req.Date, openTime, closeTime, req.Duration, court.PricePerHour, existingBookings)

	return &dto.AvailabilityResponse{
		Date:      req.Date,
		CourtID:   req.CourtID,
		TimeSlots: timeSlots,
	}, nil
}

// validateBookingTime 驗證預訂時間
func (bu *BookingUsecase) validateBookingTime(startTime, endTime time.Time) error {
	// 檢查結束時間是否晚於開始時間
	if !endTime.After(startTime) {
		return errors.New("結束時間必須晚於開始時間")
	}

	// 檢查預訂時長（最少30分鐘，最多8小時）
	duration := endTime.Sub(startTime)
	if duration < 30*time.Minute {
		return errors.New("預訂時長不能少於30分鐘")
	}
	if duration > 8*time.Hour {
		return errors.New("預訂時長不能超過8小時")
	}

	// 檢查是否為未來時間
	if startTime.Before(time.Now()) {
		return errors.New("不能預訂過去的時間")
	}

	// 檢查預訂時間是否在合理範圍內（例如：不能超過30天）
	if startTime.After(time.Now().Add(30 * 24 * time.Hour)) {
		return errors.New("不能預訂30天後的時間")
	}

	return nil
}

// checkTimeConflict 檢查時間衝突
func (bu *BookingUsecase) checkTimeConflict(courtID string, startTime, endTime time.Time, excludeBookingID string) error {
	query := bu.db.Where("court_id = ? AND status IN (?, ?) AND deleted_at IS NULL", courtID, "pending", "confirmed")

	// 排除指定的預訂ID（用於更新時）
	if excludeBookingID != "" {
		query = query.Where("id != ?", excludeBookingID)
	}

	// 檢查時間重疊
	query = query.Where("(start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?)",
		endTime, startTime, startTime, startTime, startTime, endTime)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return errors.New("檢查時間衝突失敗")
	}

	if count > 0 {
		return errors.New("該時間段已被預訂")
	}

	return nil
}

// checkOperatingHours 檢查營業時間
func (bu *BookingUsecase) checkOperatingHours(court *models.Court, startTime, endTime time.Time) error {
	weekday := startTime.Weekday().String()

	// 解析 OperatingHours JSON 到 map
	var operatingHours map[string]string
	if err := json.Unmarshal(court.OperatingHours, &operatingHours); err != nil {
		return fmt.Errorf("解析營業時間失敗: %v", err)
	}

	hours, exists := operatingHours[weekday]
	if !exists || hours == "closed" {
		return errors.New("場地在該日期不營業")
	}

	openTime, closeTime, err := bu.parseOperatingHours(hours)
	if err != nil {
		return fmt.Errorf("解析營業時間失敗: %v", err)
	}

	// 將預訂時間轉換為當天的時間進行比較
	bookingStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, startTime.Location())
	bookingEnd := time.Date(endTime.Year(), endTime.Month(), endTime.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, endTime.Location())

	dayStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	openDateTime := dayStart.Add(openTime)
	closeDateTime := dayStart.Add(closeTime)

	if bookingStart.Before(openDateTime) || bookingEnd.After(closeDateTime) {
		return fmt.Errorf("預訂時間超出營業時間範圍 (%s)", hours)
	}

	return nil
}

// parseOperatingHours 解析營業時間
func (bu *BookingUsecase) parseOperatingHours(hours string) (time.Duration, time.Duration, error) {
	if hours == "closed" {
		return 0, 0, errors.New("場地關閉")
	}

	// 解析格式如 "09:00-18:00"
	if len(hours) != 11 || hours[5] != '-' {
		return 0, 0, errors.New("營業時間格式錯誤")
	}

	openStr := hours[:5]
	closeStr := hours[6:]

	openTime, err := time.ParseDuration(openStr[:2] + "h" + openStr[3:] + "m")
	if err != nil {
		return 0, 0, fmt.Errorf("解析開始時間失敗: %v", err)
	}

	closeTime, err := time.ParseDuration(closeStr[:2] + "h" + closeStr[3:] + "m")
	if err != nil {
		return 0, 0, fmt.Errorf("解析結束時間失敗: %v", err)
	}

	return openTime, closeTime, nil
}

// generateTimeSlots 生成時間段
func (bu *BookingUsecase) generateTimeSlots(date time.Time, openTime, closeTime time.Duration, slotDuration int, pricePerHour float64, existingBookings []models.Booking) []dto.TimeSlot {
	var timeSlots []dto.TimeSlot

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	currentTime := dayStart.Add(openTime)
	endTime := dayStart.Add(closeTime)
	slotDur := time.Duration(slotDuration) * time.Minute

	for currentTime.Add(slotDur).Before(endTime) || currentTime.Add(slotDur).Equal(endTime) {
		slotEnd := currentTime.Add(slotDur)

		// 檢查該時間段是否被預訂
		available := true
		for _, booking := range existingBookings {
			if currentTime.Before(booking.EndTime) && slotEnd.After(booking.StartTime) {
				available = false
				break
			}
		}

		// 計算該時間段的價格
		hours := slotDur.Hours()
		price := hours * pricePerHour

		timeSlots = append(timeSlots, dto.TimeSlot{
			StartTime: currentTime,
			EndTime:   slotEnd,
			Available: available,
			Price:     price,
		})

		currentTime = currentTime.Add(30 * time.Minute) // 每30分鐘一個時間段
	}

	return timeSlots
}
