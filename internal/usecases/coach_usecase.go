package usecases

import (
	"errors"
	"fmt"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// CoachUsecase 教練用例
type CoachUsecase struct {
	db *gorm.DB
}

// NewCoachUsecase 創建新的教練用例
func NewCoachUsecase(db *gorm.DB) *CoachUsecase {
	return &CoachUsecase{
		db: db,
	}
}

// CreateCoachProfile 創建教練檔案
func (cu *CoachUsecase) CreateCoachProfile(userID string, req *dto.CreateCoachProfileRequest) (*models.Coach, error) {
	// 檢查用戶是否存在
	var user models.User
	if err := cu.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 檢查是否已有教練檔案
	var existingCoach models.Coach
	if err := cu.db.Where("user_id = ?", userID).First(&existingCoach).Error; err == nil {
		return nil, errors.New("教練檔案已存在")
	}

	// 驗證可用時間格式
	if err := cu.validateAvailableHours(req.AvailableHours); err != nil {
		return nil, err
	}

	// 設置默認貨幣
	currency := req.Currency
	if currency == "" {
		currency = "TWD"
	}

	// 創建教練檔案
	coach := models.Coach{
		UserID:         userID,
		LicenseNumber:  req.LicenseNumber,
		Certifications: models.StringArray(req.Certifications),
		Experience:     req.Experience,
		Specialties:    models.StringArray(req.Specialties),
		Biography:      req.Biography,
		HourlyRate:     req.HourlyRate,
		Currency:       currency,
		Languages:      models.StringArray(req.Languages),
		AvailableHours: req.AvailableHours,
		IsVerified:     false, // 新教練需要認證
		IsActive:       true,
	}

	if err := cu.db.Create(&coach).Error; err != nil {
		return nil, errors.New("創建教練檔案失敗")
	}

	// 重新載入教練數據（包含關聯）
	if err := cu.db.Preload("User").Preload("User.Profile").Where("id = ?", coach.ID).First(&coach).Error; err != nil {
		return nil, errors.New("載入教練數據失敗")
	}

	return &coach, nil
}

// GetCoachByID 根據ID獲取教練
func (cu *CoachUsecase) GetCoachByID(coachID string) (*models.Coach, error) {
	var coach models.Coach
	if err := cu.db.Preload("User").Preload("User.Profile").Where("id = ? AND is_active = ?", coachID, true).First(&coach).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("教練不存在")
		}
		return nil, errors.New("獲取教練信息失敗")
	}
	return &coach, nil
}

// GetCoachByUserID 根據用戶ID獲取教練
func (cu *CoachUsecase) GetCoachByUserID(userID string) (*models.Coach, error) {
	var coach models.Coach
	if err := cu.db.Preload("User").Preload("User.Profile").Where("user_id = ? AND is_active = ?", userID, true).First(&coach).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("教練檔案不存在")
		}
		return nil, errors.New("獲取教練信息失敗")
	}
	return &coach, nil
}

// UpdateCoachProfile 更新教練檔案
func (cu *CoachUsecase) UpdateCoachProfile(coachID string, req *dto.UpdateCoachProfileRequest) (*models.Coach, error) {
	// 查找教練
	var coach models.Coach
	if err := cu.db.Where("id = ?", coachID).First(&coach).Error; err != nil {
		return nil, errors.New("教練不存在")
	}

	// 驗證可用時間格式
	if req.AvailableHours != nil {
		if err := cu.validateAvailableHours(*req.AvailableHours); err != nil {
			return nil, err
		}
	}

	// 更新檔案
	updates := make(map[string]interface{})

	if req.LicenseNumber != nil {
		updates["license_number"] = *req.LicenseNumber
	}
	if req.Certifications != nil {
		updates["certifications"] = models.StringArray(req.Certifications)
	}
	if req.Experience != nil {
		updates["experience"] = *req.Experience
	}
	if req.Specialties != nil {
		updates["specialties"] = models.StringArray(req.Specialties)
	}
	if req.Biography != nil {
		updates["biography"] = *req.Biography
	}
	if req.HourlyRate != nil {
		updates["hourly_rate"] = *req.HourlyRate
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.Languages != nil {
		updates["languages"] = models.StringArray(req.Languages)
	}
	if req.AvailableHours != nil {
		updates["available_hours"] = *req.AvailableHours
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		if err := cu.db.Model(&coach).Updates(updates).Error; err != nil {
			return nil, errors.New("更新教練檔案失敗")
		}
	}

	// 重新載入教練數據
	if err := cu.db.Preload("User").Preload("User.Profile").Where("id = ?", coachID).First(&coach).Error; err != nil {
		return nil, errors.New("載入教練數據失敗")
	}

	return &coach, nil
}

// SearchCoaches 搜尋教練
func (cu *CoachUsecase) SearchCoaches(req *dto.CoachSearchRequest) ([]models.Coach, int64, error) {
	query := cu.db.Model(&models.Coach{}).Preload("User").Preload("User.Profile")

	// 基本篩選條件
	query = query.Where("is_active = ?", true)

	// 專長篩選
	if len(req.Specialties) > 0 {
		query = query.Where("specialties && ?", pq.StringArray(req.Specialties))
	}

	// 經驗篩選
	if req.MinExperience != nil {
		query = query.Where("experience >= ?", *req.MinExperience)
	}
	if req.MaxExperience != nil {
		query = query.Where("experience <= ?", *req.MaxExperience)
	}

	// 價格篩選
	if req.MinHourlyRate != nil {
		query = query.Where("hourly_rate >= ?", *req.MinHourlyRate)
	}
	if req.MaxHourlyRate != nil {
		query = query.Where("hourly_rate <= ?", *req.MaxHourlyRate)
	}

	// 語言篩選
	if len(req.Languages) > 0 {
		query = query.Where("languages && ?", pq.StringArray(req.Languages))
	}

	// 評分篩選
	if req.MinRating != nil {
		query = query.Where("average_rating >= ?", *req.MinRating)
	}

	// 認證狀態篩選
	if req.IsVerified != nil {
		query = query.Where("is_verified = ?", *req.IsVerified)
	}

	// 活躍狀態篩選
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("計算教練總數失敗")
	}

	// 排序
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "average_rating"
	} else {
		// 映射排序欄位到資料庫欄位名稱
		fieldMap := map[string]string{
			"rating":     "average_rating",
			"experience": "experience",
			"hourlyRate": "hourly_rate",
			"createdAt":  "created_at",
		}
		if dbField, exists := fieldMap[sortBy]; exists {
			sortBy = dbField
		}
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 分頁
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query = query.Offset(offset).Limit(limit)

	// 執行查詢
	var coaches []models.Coach
	if err := query.Find(&coaches).Error; err != nil {
		return nil, 0, errors.New("搜尋教練失敗")
	}

	return coaches, total, nil
}

// VerifyCoach 認證教練
func (cu *CoachUsecase) VerifyCoach(req *dto.CoachVerificationRequest) (*models.Coach, error) {
	// 查找教練
	var coach models.Coach
	if err := cu.db.Where("id = ?", req.CoachID).First(&coach).Error; err != nil {
		return nil, errors.New("教練不存在")
	}

	// 更新認證狀態
	updates := map[string]interface{}{
		"is_verified": req.IsVerified,
	}

	if err := cu.db.Model(&coach).Updates(updates).Error; err != nil {
		return nil, errors.New("更新教練認證狀態失敗")
	}

	// 重新載入教練數據
	if err := cu.db.Preload("User").Preload("User.Profile").Where("id = ?", req.CoachID).First(&coach).Error; err != nil {
		return nil, errors.New("載入教練數據失敗")
	}

	return &coach, nil
}

// GetCoachSpecialties 獲取教練專長選項
func (cu *CoachUsecase) GetCoachSpecialties() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "beginner", "label": "初學者教學", "description": "專門教導網球初學者基本技巧"},
		{"value": "intermediate", "label": "中級教學", "description": "提升中級球員的技術和戰術"},
		{"value": "advanced", "label": "高級教學", "description": "訓練高級球員的競技技巧"},
		{"value": "junior", "label": "青少年教學", "description": "專門教導青少年和兒童網球"},
	}
}

// GetCoachCertifications 獲取常見教練認證
func (cu *CoachUsecase) GetCoachCertifications() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "PTR", "label": "PTR 職業網球教練認證", "description": "Professional Tennis Registry"},
		{"value": "USPTA", "label": "USPTA 美國職業網球教練協會", "description": "United States Professional Tennis Association"},
		{"value": "ITF", "label": "ITF 國際網球總會教練認證", "description": "International Tennis Federation"},
		{"value": "CTCA", "label": "中華民國網球協會教練證", "description": "Chinese Taipei Tennis Association"},
		{"value": "RPT", "label": "註冊職業網球教練", "description": "Registered Professional Tennis Coach"},
	}
}

// validateAvailableHours 驗證可用時間格式
func (cu *CoachUsecase) validateAvailableHours(availableHours models.AvailableHours) error {
	if availableHours == nil {
		return nil
	}

	validDays := map[string]bool{
		"monday":    true,
		"tuesday":   true,
		"wednesday": true,
		"thursday":  true,
		"friday":    true,
		"saturday":  true,
		"sunday":    true,
	}

	for day, timeSlots := range availableHours {
		if !validDays[day] {
			return fmt.Errorf("無效的星期: %s", day)
		}

		for _, timeSlot := range timeSlots {
			if err := cu.validateTimeSlot(timeSlot); err != nil {
				return fmt.Errorf("無效的時間段 %s: %v", timeSlot, err)
			}
		}
	}

	return nil
}

// validateTimeSlot 驗證時間段格式 (例如: "09:00-12:00")
func (cu *CoachUsecase) validateTimeSlot(timeSlot string) error {
	// 簡單的時間格式驗證
	if len(timeSlot) != 11 {
		return errors.New("時間段格式應為 HH:MM-HH:MM")
	}

	if timeSlot[5] != '-' {
		return errors.New("時間段格式應為 HH:MM-HH:MM")
	}

	startTime := timeSlot[:5]
	endTime := timeSlot[6:]

	if err := cu.validateTime(startTime); err != nil {
		return fmt.Errorf("開始時間格式錯誤: %v", err)
	}

	if err := cu.validateTime(endTime); err != nil {
		return fmt.Errorf("結束時間格式錯誤: %v", err)
	}

	// 驗證開始時間小於結束時間
	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)
	if start.After(end) || start.Equal(end) {
		return errors.New("開始時間必須早於結束時間")
	}

	return nil
}

// validateTime 驗證時間格式 (例如: "09:00")
func (cu *CoachUsecase) validateTime(timeStr string) error {
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		return errors.New("時間格式應為 HH:MM")
	}
	return nil
}

// ===== 課程管理相關方法 =====

// ScheduleItem 時間表項目
type ScheduleItem struct {
	DayOfWeek int    `json:"dayOfWeek" binding:"required,min=0,max=6"`
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
	IsActive  bool   `json:"isActive"`
}

// CreateLessonType 創建課程類型
func (cu *CoachUsecase) CreateLessonType(coachID string, req *dto.CreateLessonTypeRequest) (*models.LessonType, error) {
	// 檢查教練是否存在
	var coach models.Coach
	if err := cu.db.Where("id = ? AND is_active = ?", coachID, true).First(&coach).Error; err != nil {
		return nil, errors.New("教練不存在")
	}

	// 驗證團體課程參數
	if req.Type == "group" || req.Type == "clinic" {
		if req.MaxParticipants == nil || *req.MaxParticipants < 2 {
			return nil, errors.New("團體課程必須設定最大參與人數（至少2人）")
		}
		if req.MinParticipants != nil && *req.MinParticipants > *req.MaxParticipants {
			return nil, errors.New("最小參與人數不能大於最大參與人數")
		}
	}

	// 設置默認貨幣
	currency := req.Currency
	if currency == "" {
		currency = "TWD"
	}

	// 創建課程類型
	lessonType := models.LessonType{
		CoachID:         coachID,
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		Level:           req.Level,
		Duration:        req.Duration,
		Price:           req.Price,
		Currency:        currency,
		MaxParticipants: req.MaxParticipants,
		MinParticipants: req.MinParticipants,
		Equipment:       models.StringArray(req.Equipment),
		Prerequisites:   req.Prerequisites,
		IsActive:        true,
	}

	if err := cu.db.Create(&lessonType).Error; err != nil {
		return nil, errors.New("創建課程類型失敗")
	}

	// 重新載入數據
	if err := cu.db.Preload("Coach").Where("id = ?", lessonType.ID).First(&lessonType).Error; err != nil {
		return nil, errors.New("載入課程類型數據失敗")
	}

	return &lessonType, nil
}

// GetLessonTypes 獲取課程類型列表
func (cu *CoachUsecase) GetLessonTypes(coachID string) ([]models.LessonType, error) {
	var lessonTypes []models.LessonType
	if err := cu.db.Where("coach_id = ? AND is_active = ?", coachID, true).Find(&lessonTypes).Error; err != nil {
		return nil, errors.New("獲取課程類型失敗")
	}
	return lessonTypes, nil
}

// UpdateLessonType 更新課程類型
func (cu *CoachUsecase) UpdateLessonType(lessonTypeID string, req *dto.UpdateLessonTypeRequest) (*models.LessonType, error) {
	// 查找課程類型
	var lessonType models.LessonType
	if err := cu.db.Where("id = ?", lessonTypeID).First(&lessonType).Error; err != nil {
		return nil, errors.New("課程類型不存在")
	}

	// 更新欄位
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.MaxParticipants != nil {
		updates["max_participants"] = *req.MaxParticipants
	}
	if req.MinParticipants != nil {
		updates["min_participants"] = *req.MinParticipants
	}
	if req.Equipment != nil {
		updates["equipment"] = models.StringArray(req.Equipment)
	}
	if req.Prerequisites != nil {
		updates["prerequisites"] = *req.Prerequisites
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		if err := cu.db.Model(&lessonType).Updates(updates).Error; err != nil {
			return nil, errors.New("更新課程類型失敗")
		}
	}

	// 重新載入數據
	if err := cu.db.Preload("Coach").Where("id = ?", lessonTypeID).First(&lessonType).Error; err != nil {
		return nil, errors.New("載入課程類型數據失敗")
	}

	return &lessonType, nil
}

// DeleteLessonType 刪除課程類型
func (cu *CoachUsecase) DeleteLessonType(lessonTypeID string) error {
	// 檢查是否有關聯的課程
	var count int64
	if err := cu.db.Model(&models.Lesson{}).Where("lesson_type_id = ?", lessonTypeID).Count(&count).Error; err != nil {
		return errors.New("檢查關聯課程失敗")
	}

	if count > 0 {
		return errors.New("無法刪除：存在關聯的課程")
	}

	// 軟刪除課程類型
	if err := cu.db.Where("id = ?", lessonTypeID).Delete(&models.LessonType{}).Error; err != nil {
		return errors.New("刪除課程類型失敗")
	}

	return nil
}

// CreateLesson 創建課程
func (cu *CoachUsecase) CreateLesson(req *dto.CreateLessonRequest) (*models.Lesson, error) {
	// 檢查教練是否存在
	var coach models.Coach
	if err := cu.db.Where("id = ? AND is_active = ?", req.CoachID, true).First(&coach).Error; err != nil {
		return nil, errors.New("教練不存在")
	}

	// 檢查學生是否存在
	var student models.User
	if err := cu.db.Where("id = ?", req.StudentID).First(&student).Error; err != nil {
		return nil, errors.New("學生不存在")
	}

	// 檢查時間衝突
	if err := cu.checkTimeConflict(req.CoachID, req.ScheduledAt, req.Duration); err != nil {
		return nil, err
	}

	// 設置默認貨幣
	currency := req.Currency
	if currency == "" {
		currency = "TWD"
	}

	// 創建課程
	lesson := models.Lesson{
		CoachID:      req.CoachID,
		StudentID:    req.StudentID,
		LessonTypeID: req.LessonTypeID,
		CourtID:      req.CourtID,
		Type:         req.Type,
		Level:        req.Level,
		Duration:     req.Duration,
		Price:        req.Price,
		Currency:     currency,
		ScheduledAt:  req.ScheduledAt,
		Status:       "scheduled",
		Notes:        req.Notes,
	}

	if err := cu.db.Create(&lesson).Error; err != nil {
		return nil, errors.New("創建課程失敗")
	}

	// 重新載入數據
	if err := cu.db.Preload("Coach").Preload("Student").Preload("LessonType").Preload("Court").Where("id = ?", lesson.ID).First(&lesson).Error; err != nil {
		return nil, errors.New("載入課程數據失敗")
	}

	return &lesson, nil
}

// GetLesson 獲取課程詳情
func (cu *CoachUsecase) GetLesson(lessonID string) (*models.Lesson, error) {
	var lesson models.Lesson
	if err := cu.db.Preload("Coach").Preload("Student").Preload("LessonType").Preload("Court").Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("課程不存在")
		}
		return nil, errors.New("獲取課程信息失敗")
	}
	return &lesson, nil
}

// GetLessons 獲取課程列表
func (cu *CoachUsecase) GetLessons(req *dto.GetLessonsRequest) ([]models.Lesson, int64, error) {
	query := cu.db.Model(&models.Lesson{}).Preload("Coach").Preload("Student").Preload("LessonType").Preload("Court")

	// 篩選條件
	if req.CoachID != nil {
		query = query.Where("coach_id = ?", *req.CoachID)
	}
	if req.StudentID != nil {
		query = query.Where("student_id = ?", *req.StudentID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	// 日期範圍篩選
	if req.StartDate != nil {
		if startDate, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			query = query.Where("scheduled_at >= ?", startDate)
		}
	}
	if req.EndDate != nil {
		if endDate, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			endDate = endDate.Add(24 * time.Hour) // 包含結束日期
			query = query.Where("scheduled_at < ?", endDate)
		}
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("計算課程總數失敗")
	}

	// 排序和分頁
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query = query.Order("scheduled_at DESC").Offset(offset).Limit(limit)

	// 執行查詢
	var lessons []models.Lesson
	if err := query.Find(&lessons).Error; err != nil {
		return nil, 0, errors.New("獲取課程列表失敗")
	}

	return lessons, total, nil
}

// UpdateLesson 更新課程
func (cu *CoachUsecase) UpdateLesson(lessonID string, req *dto.UpdateLessonRequest) (*models.Lesson, error) {
	// 查找課程
	var lesson models.Lesson
	if err := cu.db.Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		return nil, errors.New("課程不存在")
	}

	// 檢查課程狀態
	if lesson.Status == "completed" || lesson.Status == "cancelled" {
		return nil, errors.New("已完成或已取消的課程無法修改")
	}

	// 檢查時間衝突（如果更新時間）
	if req.ScheduledAt != nil {
		if err := cu.checkTimeConflictExcluding(lesson.CoachID, *req.ScheduledAt, lesson.Duration, lessonID); err != nil {
			return nil, err
		}
	}

	// 更新欄位
	updates := make(map[string]interface{})

	if req.CourtID != nil {
		updates["court_id"] = *req.CourtID
	}
	if req.ScheduledAt != nil {
		updates["scheduled_at"] = *req.ScheduledAt
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := cu.db.Model(&lesson).Updates(updates).Error; err != nil {
			return nil, errors.New("更新課程失敗")
		}
	}

	// 重新載入數據
	if err := cu.db.Preload("Coach").Preload("Student").Preload("LessonType").Preload("Court").Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		return nil, errors.New("載入課程數據失敗")
	}

	return &lesson, nil
}

// CancelLesson 取消課程
func (cu *CoachUsecase) CancelLesson(lessonID string, req *dto.CancelLessonRequest) (*models.Lesson, error) {
	// 查找課程
	var lesson models.Lesson
	if err := cu.db.Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		return nil, errors.New("課程不存在")
	}

	// 檢查課程狀態
	if lesson.Status == "completed" || lesson.Status == "cancelled" {
		return nil, errors.New("課程已完成或已取消")
	}

	// 更新狀態
	updates := map[string]interface{}{
		"status":        "cancelled",
		"cancel_reason": req.Reason,
	}

	if err := cu.db.Model(&lesson).Updates(updates).Error; err != nil {
		return nil, errors.New("取消課程失敗")
	}

	// 重新載入數據
	if err := cu.db.Preload("Coach").Preload("Student").Preload("LessonType").Preload("Court").Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		return nil, errors.New("載入課程數據失敗")
	}

	return &lesson, nil
}

// GetCoachAvailability 獲取教練可用時間
func (cu *CoachUsecase) GetCoachAvailability(coachID string, date string) ([]models.TimeSlot, error) {
	// 解析日期
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("日期格式錯誤")
	}

	// 獲取星期幾 (0=Sunday, 6=Saturday)
	dayOfWeek := int(targetDate.Weekday())

	// 獲取教練的時間表
	var schedules []models.LessonSchedule
	if err := cu.db.Where("coach_id = ? AND day_of_week = ? AND is_active = ?", coachID, dayOfWeek, true).Find(&schedules).Error; err != nil {
		return nil, errors.New("獲取教練時間表失敗")
	}

	if len(schedules) == 0 {
		return []models.TimeSlot{}, nil
	}

	// 獲取當天已預訂的課程
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var bookedLessons []models.Lesson
	if err := cu.db.Where("coach_id = ? AND scheduled_at >= ? AND scheduled_at < ? AND status IN ?",
		coachID, startOfDay, endOfDay, []string{"scheduled", "in_progress"}).Find(&bookedLessons).Error; err != nil {
		return nil, errors.New("獲取已預訂課程失敗")
	}

	// 生成可用時間段
	var availableSlots []models.TimeSlot
	for _, schedule := range schedules {
		slots := cu.generateTimeSlots(schedule.StartTime, schedule.EndTime, 60) // 60分鐘間隔
		for _, slot := range slots {
			isBooked := cu.isTimeSlotBooked(slot, bookedLessons, targetDate)
			availableSlots = append(availableSlots, models.TimeSlot{
				StartTime: slot.StartTime,
				EndTime:   slot.EndTime,
				IsBooked:  isBooked,
			})
		}
	}

	return availableSlots, nil
}

// UpdateCoachSchedule 更新教練時間表
func (cu *CoachUsecase) UpdateCoachSchedule(coachID string, req *dto.UpdateScheduleRequest) error {
	// 檢查教練是否存在
	var coach models.Coach
	if err := cu.db.Where("id = ? AND is_active = ?", coachID, true).First(&coach).Error; err != nil {
		return errors.New("教練不存在")
	}

	// 驗證時間格式
	for _, schedule := range req.Schedules {
		if err := cu.validateTime(schedule.StartTime); err != nil {
			return fmt.Errorf("開始時間格式錯誤: %v", err)
		}
		if err := cu.validateTime(schedule.EndTime); err != nil {
			return fmt.Errorf("結束時間格式錯誤: %v", err)
		}

		// 驗證開始時間小於結束時間
		start, _ := time.Parse("15:04", schedule.StartTime)
		end, _ := time.Parse("15:04", schedule.EndTime)
		if start.After(end) || start.Equal(end) {
			return errors.New("開始時間必須早於結束時間")
		}
	}

	// 開始事務
	tx := cu.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 刪除現有時間表
	if err := tx.Where("coach_id = ?", coachID).Delete(&models.LessonSchedule{}).Error; err != nil {
		tx.Rollback()
		return errors.New("刪除現有時間表失敗")
	}

	// 創建新時間表
	for _, schedule := range req.Schedules {
		lessonSchedule := models.LessonSchedule{
			CoachID:   coachID,
			DayOfWeek: schedule.DayOfWeek,
			StartTime: schedule.StartTime,
			EndTime:   schedule.EndTime,
			IsActive:  schedule.IsActive,
		}

		if err := tx.Create(&lessonSchedule).Error; err != nil {
			tx.Rollback()
			return errors.New("創建時間表失敗")
		}
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return errors.New("更新時間表失敗")
	}

	return nil
}

// GetCoachSchedule 獲取教練時間表
func (cu *CoachUsecase) GetCoachSchedule(coachID string) ([]models.LessonSchedule, error) {
	var schedules []models.LessonSchedule
	if err := cu.db.Where("coach_id = ?", coachID).Order("day_of_week, start_time").Find(&schedules).Error; err != nil {
		return nil, errors.New("獲取教練時間表失敗")
	}
	return schedules, nil
}

// checkTimeConflict 檢查時間衝突
func (cu *CoachUsecase) checkTimeConflict(coachID string, scheduledAt time.Time, duration int) error {
	endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)

	var count int64
	if err := cu.db.Model(&models.Lesson{}).Where(
		"coach_id = ? AND status IN ? AND ((scheduled_at < ? AND scheduled_at + INTERVAL '1 minute' * duration > ?) OR (scheduled_at < ? AND scheduled_at + INTERVAL '1 minute' * duration > ?))",
		coachID, []string{"scheduled", "in_progress"}, endTime, scheduledAt, scheduledAt, endTime,
	).Count(&count).Error; err != nil {
		return errors.New("檢查時間衝突失敗")
	}

	if count > 0 {
		return errors.New("時間衝突：該時段已有其他課程")
	}

	return nil
}

// checkTimeConflictExcluding 檢查時間衝突（排除指定課程）
func (cu *CoachUsecase) checkTimeConflictExcluding(coachID string, scheduledAt time.Time, duration int, excludeLessonID string) error {
	endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)

	var count int64
	if err := cu.db.Model(&models.Lesson{}).Where(
		"coach_id = ? AND id != ? AND status IN ? AND ((scheduled_at < ? AND scheduled_at + INTERVAL '1 minute' * duration > ?) OR (scheduled_at < ? AND scheduled_at + INTERVAL '1 minute' * duration > ?))",
		coachID, excludeLessonID, []string{"scheduled", "in_progress"}, endTime, scheduledAt, scheduledAt, endTime,
	).Count(&count).Error; err != nil {
		return errors.New("檢查時間衝突失敗")
	}

	if count > 0 {
		return errors.New("時間衝突：該時段已有其他課程")
	}

	return nil
}

// generateTimeSlots 生成時間段
func (cu *CoachUsecase) generateTimeSlots(startTime, endTime string, intervalMinutes int) []models.TimeSlot {
	var slots []models.TimeSlot

	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)

	current := start
	for current.Before(end) {
		next := current.Add(time.Duration(intervalMinutes) * time.Minute)
		if next.After(end) {
			break
		}

		slots = append(slots, models.TimeSlot{
			StartTime: current.Format("15:04"),
			EndTime:   next.Format("15:04"),
			IsBooked:  false,
		})

		current = next
	}

	return slots
}

// isTimeSlotBooked 檢查時間段是否已預訂
func (cu *CoachUsecase) isTimeSlotBooked(slot models.TimeSlot, bookedLessons []models.Lesson, targetDate time.Time) bool {
	slotStart, _ := time.Parse("15:04", slot.StartTime)
	slotEnd, _ := time.Parse("15:04", slot.EndTime)

	// 將時間段轉換為目標日期的完整時間
	slotStartTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		slotStart.Hour(), slotStart.Minute(), 0, 0, targetDate.Location())
	slotEndTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
		slotEnd.Hour(), slotEnd.Minute(), 0, 0, targetDate.Location())

	for _, lesson := range bookedLessons {
		lessonEnd := lesson.ScheduledAt.Add(time.Duration(lesson.Duration) * time.Minute)

		// 檢查時間重疊
		if lesson.ScheduledAt.Before(slotEndTime) && lessonEnd.After(slotStartTime) {
			return true
		}
	}

	return false
}

// GetIntelligentRecommendations 獲取智能推薦
func (cu *CoachUsecase) GetIntelligentRecommendations(req *dto.IntelligentSchedulingRequest) ([]interface{}, error) {
	// 創建智能排課服務
	schedulingService := services.NewIntelligentSchedulingService(cu.db)

	// 轉換請求為服務所需的格式
	studentPrefs := &services.StudentPreferences{
		UserID:              req.StudentID,
		NTRPLevel:           req.NTRPLevel,
		PreferredTimes:      req.PreferredTimes,
		PreferredDays:       req.PreferredDays,
		MaxDistance:         req.MaxDistance,
		MinPrice:            req.MinPrice,
		MaxPrice:            req.MaxPrice,
		PreferredLessonType: req.PreferredLessonType,
	}

	if req.Location != nil {
		studentPrefs.Location = &services.Location{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Address:   req.Location.Address,
		}
	}

	// 獲取推薦
	recommendations, err := schedulingService.RecommendCoaches(studentPrefs, req.DateRange)
	if err != nil {
		return nil, err
	}

	// 轉換為通用接口類型
	result := make([]interface{}, len(recommendations))
	for i, rec := range recommendations {
		result[i] = rec
	}

	return result, nil
}

// FindOptimalLessonTime 尋找最佳課程時間
func (cu *CoachUsecase) FindOptimalLessonTime(req *dto.OptimalTimeRequest) (interface{}, error) {
	// 創建智能排課服務
	schedulingService := services.NewIntelligentSchedulingService(cu.db)

	// 轉換請求為服務所需的格式
	studentPrefs := &services.StudentPreferences{
		UserID:              req.StudentID,
		NTRPLevel:           req.NTRPLevel,
		PreferredTimes:      req.PreferredTimes,
		PreferredDays:       req.PreferredDays,
		MaxDistance:         req.MaxDistance,
		MinPrice:            req.MinPrice,
		MaxPrice:            req.MaxPrice,
		PreferredLessonType: req.PreferredLessonType,
	}

	if req.Location != nil {
		studentPrefs.Location = &services.Location{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Address:   req.Location.Address,
		}
	}

	// 尋找最佳時間
	recommendation, err := schedulingService.FindOptimalLessonTime(req.CoachID, req.StudentID, studentPrefs, req.DateRange)
	if err != nil {
		return nil, err
	}

	return recommendation, nil
}

// DetectSchedulingConflicts 檢測排課衝突
func (cu *CoachUsecase) DetectSchedulingConflicts(req *dto.ConflictDetectionRequest) ([]models.Lesson, error) {
	// 創建智能排課服務
	schedulingService := services.NewIntelligentSchedulingService(cu.db)

	// 檢測衝突
	conflicts, err := schedulingService.DetectSchedulingConflicts(req.CoachID, req.ScheduledAt, req.Duration, req.ExcludeLessonID)
	if err != nil {
		return nil, err
	}

	return conflicts, nil
}

// ResolveSchedulingConflict 解決排課衝突
func (cu *CoachUsecase) ResolveSchedulingConflict(req *dto.ConflictResolutionRequest) error {
	// 創建智能排課服務
	schedulingService := services.NewIntelligentSchedulingService(cu.db)

	// 解決衝突
	return schedulingService.ResolveSchedulingConflict(req.ConflictingLessonID, req.NewScheduledAt)
}

// GetCoachRecommendationFactors 獲取教練推薦因子（用於調試和優化）
func (cu *CoachUsecase) GetCoachRecommendationFactors(coachID string, studentPrefs *dto.IntelligentSchedulingRequest) (map[string]interface{}, error) {
	// 獲取教練信息
	coach, err := cu.GetCoachByID(coachID)
	if err != nil {
		return nil, err
	}

	// 計算匹配因子（這裡需要訪問私有方法，所以簡化實現）
	factors := map[string]interface{}{
		"coachId":       coach.ID,
		"coachName":     coach.User.Profile.FirstName + " " + coach.User.Profile.LastName,
		"experience":    coach.Experience,
		"averageRating": coach.AverageRating,
		"specialties":   coach.Specialties,
		"hourlyRate":    coach.HourlyRate,
		"isVerified":    coach.IsVerified,
		"totalLessons":  coach.TotalLessons,
		"totalReviews":  coach.TotalReviews,
	}

	// 添加技術等級匹配信息
	if studentPrefs.NTRPLevel > 0 {
		studentLevel := cu.getNTRPLevelCategory(studentPrefs.NTRPLevel)
		factors["studentLevel"] = studentLevel
		factors["levelMatch"] = cu.checkLevelMatch(coach.Specialties, studentLevel)
	}

	return factors, nil
}

// getNTRPLevelCategory 獲取NTRP等級分類（輔助方法）
func (cu *CoachUsecase) getNTRPLevelCategory(ntrp float64) string {
	if ntrp <= 2.5 {
		return "beginner"
	} else if ntrp <= 4.0 {
		return "intermediate"
	} else {
		return "advanced"
	}
}

// checkLevelMatch 檢查等級匹配（輔助方法）
func (cu *CoachUsecase) checkLevelMatch(specialties []string, studentLevel string) bool {
	for _, specialty := range specialties {
		if specialty == studentLevel {
			return true
		}
	}
	return false
}

// CreateCoachReview 創建教練評價
func (cu *CoachUsecase) CreateCoachReview(userID string, req *dto.CreateCoachReviewRequest) (*models.CoachReview, error) {
	// 檢查教練是否存在
	var coach models.Coach
	if err := cu.db.Where("id = ? AND is_active = ?", req.CoachID, true).First(&coach).Error; err != nil {
		return nil, errors.New("教練不存在")
	}

	// 檢查用戶是否存在
	var user models.User
	if err := cu.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 如果提供了課程ID，檢查課程是否存在且已完成
	if req.LessonID != nil {
		var lesson models.Lesson
		if err := cu.db.Where("id = ? AND student_id = ? AND coach_id = ?", *req.LessonID, userID, req.CoachID).First(&lesson).Error; err != nil {
			return nil, errors.New("課程不存在或無權限評價")
		}

		if lesson.Status != "completed" {
			return nil, errors.New("只能評價已完成的課程")
		}

		// 檢查是否已經評價過該課程
		var existingReview models.CoachReview
		if err := cu.db.Where("lesson_id = ? AND user_id = ?", *req.LessonID, userID).First(&existingReview).Error; err == nil {
			return nil, errors.New("該課程已經評價過")
		}
	}

	// 檢查是否已經評價過該教練（如果沒有指定課程ID）
	if req.LessonID == nil {
		var existingReview models.CoachReview
		if err := cu.db.Where("coach_id = ? AND user_id = ? AND lesson_id IS NULL", req.CoachID, userID).First(&existingReview).Error; err == nil {
			return nil, errors.New("已經評價過該教練")
		}
	}

	// 創建評價
	review := models.CoachReview{
		CoachID:  req.CoachID,
		UserID:   userID,
		LessonID: req.LessonID,
		Rating:   req.Rating,
		Comment:  req.Comment,
		Tags:     req.Tags,
	}

	if err := cu.db.Create(&review).Error; err != nil {
		return nil, errors.New("創建評價失敗")
	}

	// 重新載入數據（包含關聯）
	if err := cu.db.Preload("Coach").Preload("User").Preload("User.Profile").Preload("Lesson").Where("id = ?", review.ID).First(&review).Error; err != nil {
		return nil, errors.New("載入評價數據失敗")
	}

	return &review, nil
}

// GetCoachReview 獲取教練評價詳情
func (cu *CoachUsecase) GetCoachReview(reviewID string) (*models.CoachReview, error) {
	var review models.CoachReview
	if err := cu.db.Preload("Coach").Preload("User").Preload("User.Profile").Preload("Lesson").Where("coach_id = ?", reviewID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("評價不存在")
		}
		return nil, errors.New("獲取評價失敗")
	}
	return &review, nil
}

// GetCoachReviews 獲取教練評價列表
func (cu *CoachUsecase) GetCoachReviews(req *dto.CoachReviewSearchRequest) ([]models.CoachReview, int64, error) {
	query := cu.db.Model(&models.CoachReview{}).Preload("User").Preload("User.Profile").Preload("Lesson")

	// 基本篩選條件
	query = query.Where("coach_id = ?", req.CoachID)

	// 評分篩選
	if req.Rating != nil {
		query = query.Where("rating = ?", *req.Rating)
	}

	// 是否有評論篩選
	if req.HasComment != nil {
		if *req.HasComment {
			query = query.Where("comment IS NOT NULL AND comment != ''")
		} else {
			query = query.Where("comment IS NULL OR comment = ''")
		}
	}

	// 標籤篩選
	if len(req.Tags) > 0 {
		query = query.Where("tags && ?", pq.StringArray(req.Tags))
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("計算評價總數失敗")
	}

	// 排序
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 分頁
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query = query.Offset(offset).Limit(limit)

	// 執行查詢
	var reviews []models.CoachReview
	if err := query.Find(&reviews).Error; err != nil {
		return nil, 0, errors.New("獲取評價列表失敗")
	}

	return reviews, total, nil
}

// UpdateCoachReview 更新教練評價
func (cu *CoachUsecase) UpdateCoachReview(reviewID string, userID string, req *dto.UpdateCoachReviewRequest) (*models.CoachReview, error) {
	// 查找評價
	var review models.CoachReview
	if err := cu.db.Where("id = ?", reviewID).First(&review).Error; err != nil {
		return nil, errors.New("評價不存在")
	}

	// 檢查權限：只有評價者本人可以更新
	if review.UserID != userID {
		return nil, errors.New("無權限更新此評價")
	}

	// 檢查評價是否在可編輯時間內（例如：創建後24小時內）
	if time.Since(review.CreatedAt) > 24*time.Hour {
		return nil, errors.New("評價創建超過24小時，無法修改")
	}

	// 更新欄位
	updates := make(map[string]interface{})

	if req.Rating != nil {
		updates["rating"] = *req.Rating
	}
	if req.Comment != nil {
		updates["comment"] = *req.Comment
	}
	if req.Tags != nil {
		updates["tags"] = pq.StringArray(req.Tags)
	}

	if len(updates) > 0 {
		if err := cu.db.Model(&review).Updates(updates).Error; err != nil {
			return nil, errors.New("更新評價失敗")
		}
	}

	// 重新載入數據
	if err := cu.db.Preload("Coach").Preload("User").Preload("User.Profile").Preload("Lesson").Where("id = ?", reviewID).First(&review).Error; err != nil {
		return nil, errors.New("載入評價數據失敗")
	}

	return &review, nil
}

// DeleteCoachReview 刪除教練評價
func (cu *CoachUsecase) DeleteCoachReview(reviewID string, userID string) error {
	// 查找評價
	var review models.CoachReview
	if err := cu.db.Where("id = ?", reviewID).First(&review).Error; err != nil {
		return errors.New("評價不存在")
	}

	// 檢查權限：只有評價者本人可以刪除
	if review.UserID != userID {
		return errors.New("無權限刪除此評價")
	}

	// 檢查評價是否在可刪除時間內（例如：創建後24小時內）
	if time.Since(review.CreatedAt) > 24*time.Hour {
		return errors.New("評價創建超過24小時，無法刪除")
	}

	// 軟刪除評價
	if err := cu.db.Where("id = ?", reviewID).Delete(&models.CoachReview{}).Error; err != nil {
		return errors.New("刪除評價失敗")
	}

	return nil
}

// MarkReviewHelpful 標記評價有用
func (cu *CoachUsecase) MarkReviewHelpful(userID string, req *dto.MarkReviewHelpfulRequest) (*models.CoachReview, error) {
	// 查找評價
	var review models.CoachReview
	if err := cu.db.Where("id = ?", req.ReviewID).First(&review).Error; err != nil {
		return nil, errors.New("評價不存在")
	}

	// 檢查用戶不能標記自己的評價
	if review.UserID == userID {
		return nil, errors.New("不能標記自己的評價")
	}

	// 更新有用計數
	if req.IsHelpful {
		if err := cu.db.Model(&review).UpdateColumn("is_helpful", gorm.Expr("is_helpful + ?", 1)).Error; err != nil {
			return nil, errors.New("標記評價有用失敗")
		}
	} else {
		if err := cu.db.Model(&review).UpdateColumn("is_helpful", gorm.Expr("GREATEST(is_helpful - ?, 0)", 1)).Error; err != nil {
			return nil, errors.New("取消標記評價有用失敗")
		}
	}

	// 重新載入數據
	if err := cu.db.Preload("Coach").Preload("User").Preload("User.Profile").Preload("Lesson").Where("id = ?", req.ReviewID).First(&review).Error; err != nil {
		return nil, errors.New("載入評價數據失敗")
	}

	return &review, nil
}

// GetCoachReviewStatistics 獲取教練評價統計
func (cu *CoachUsecase) GetCoachReviewStatistics(coachID string) (map[string]interface{}, error) {
	// 檢查教練是否存在
	var coach models.Coach
	if err := cu.db.Where("id = ? AND is_active = ?", coachID, true).First(&coach).Error; err != nil {
		return nil, errors.New("教練不存在")
	}

	// 獲取評分分佈
	var ratingDistribution []struct {
		Rating int   `json:"rating"`
		Count  int64 `json:"count"`
	}

	if err := cu.db.Model(&models.CoachReview{}).
		Select("rating, COUNT(*) as count").
		Where("coach_id = ?", coachID).
		Group("rating").
		Order("rating DESC").
		Scan(&ratingDistribution).Error; err != nil {
		return nil, errors.New("獲取評分分佈失敗")
	}

	// 獲取標籤統計
	var tagStats []struct {
		Tag   string `json:"tag"`
		Count int64  `json:"count"`
	}

	if err := cu.db.Raw(`
		SELECT unnest(tags) as tag, COUNT(*) as count
		FROM coach_reviews 
		WHERE coach_id = ? AND deleted_at IS NULL AND tags IS NOT NULL
		GROUP BY tag
		ORDER BY count DESC
		LIMIT 10
	`, coachID).Scan(&tagStats).Error; err != nil {
		return nil, errors.New("獲取標籤統計失敗")
	}

	// 獲取最近評價
	var recentReviews []models.CoachReview
	if err := cu.db.Preload("User").Preload("User.Profile").
		Where("coach_id = ?", coachID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentReviews).Error; err != nil {
		return nil, errors.New("獲取最近評價失敗")
	}

	// 獲取月度評價趨勢（最近12個月）
	var monthlyTrend []struct {
		Month     string  `json:"month"`
		Count     int64   `json:"count"`
		AvgRating float64 `json:"avgRating"`
	}

	if err := cu.db.Raw(`
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM') as month,
			COUNT(*) as count,
			AVG(rating::numeric) as avg_rating
		FROM coach_reviews 
		WHERE coach_id = ? AND deleted_at IS NULL 
			AND created_at >= NOW() - INTERVAL '12 months'
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month DESC
	`, coachID).Scan(&monthlyTrend).Error; err != nil {
		return nil, errors.New("獲取月度趨勢失敗")
	}

	statistics := map[string]interface{}{
		"totalReviews":       coach.TotalReviews,
		"averageRating":      coach.AverageRating,
		"ratingDistribution": ratingDistribution,
		"topTags":            tagStats,
		"recentReviews":      recentReviews,
		"monthlyTrend":       monthlyTrend,
	}

	return statistics, nil
}

// GetAvailableReviewTags 獲取可用的評價標籤
func (cu *CoachUsecase) GetAvailableReviewTags() []map[string]interface{} {
	return []map[string]interface{}{
		{"value": "patient", "label": "耐心", "category": "teaching"},
		{"value": "professional", "label": "專業", "category": "teaching"},
		{"value": "knowledgeable", "label": "知識豐富", "category": "teaching"},
		{"value": "encouraging", "label": "鼓勵", "category": "teaching"},
		{"value": "clear_instruction", "label": "指導清晰", "category": "teaching"},
		{"value": "punctual", "label": "準時", "category": "behavior"},
		{"value": "friendly", "label": "友善", "category": "behavior"},
		{"value": "responsive", "label": "回應迅速", "category": "behavior"},
		{"value": "flexible", "label": "彈性", "category": "behavior"},
		{"value": "organized", "label": "有組織", "category": "behavior"},
		{"value": "technique_focused", "label": "技術導向", "category": "style"},
		{"value": "fitness_focused", "label": "體能導向", "category": "style"},
		{"value": "strategy_focused", "label": "戰術導向", "category": "style"},
		{"value": "fun_approach", "label": "趣味教學", "category": "style"},
		{"value": "competitive_training", "label": "競技訓練", "category": "style"},
	}
}

// CheckCanReviewCoach 檢查是否可以評價教練
func (cu *CoachUsecase) CheckCanReviewCoach(userID string, coachID string, lessonID *string) (bool, string, error) {
	// 檢查教練是否存在
	var coach models.Coach
	if err := cu.db.Where("id = ? AND is_active = ?", coachID, true).First(&coach).Error; err != nil {
		return false, "教練不存在", nil
	}

	// 檢查用戶不能評價自己
	if coach.UserID == userID {
		return false, "不能評價自己", nil
	}

	// 如果指定了課程ID
	if lessonID != nil {
		var lesson models.Lesson
		if err := cu.db.Where("id = ? AND student_id = ? AND coach_id = ?", *lessonID, userID, coachID).First(&lesson).Error; err != nil {
			return false, "課程不存在或無權限評價", nil
		}

		if lesson.Status != "completed" {
			return false, "只能評價已完成的課程", nil
		}

		// 檢查是否已經評價過該課程
		var existingReview models.CoachReview
		if err := cu.db.Where("lesson_id = ? AND user_id = ?", *lessonID, userID).First(&existingReview).Error; err == nil {
			return false, "該課程已經評價過", nil
		}

		return true, "可以評價", nil
	}

	// 如果沒有指定課程ID，檢查是否有已完成的課程
	var completedLessonsCount int64
	if err := cu.db.Model(&models.Lesson{}).
		Where("coach_id = ? AND student_id = ? AND status = ?", coachID, userID, "completed").
		Count(&completedLessonsCount).Error; err != nil {
		return false, "檢查課程記錄失敗", err
	}

	if completedLessonsCount == 0 {
		return false, "需要先完成課程才能評價教練", nil
	}

	// 檢查是否已經評價過該教練（沒有指定課程的一般評價）
	var existingReview models.CoachReview
	if err := cu.db.Where("coach_id = ? AND user_id = ? AND lesson_id IS NULL", coachID, userID).First(&existingReview).Error; err == nil {
		return false, "已經評價過該教練", nil
	}

	return true, "可以評價", nil
}
