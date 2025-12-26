package controllers

import (
	"net/http"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// CourtUsecaseInterface 場地用例接口
type CourtUsecaseInterface interface {
	CreateCourt(req *dto.CreateCourtRequest) (*models.Court, error)
	GetCourtByID(courtID string) (*models.Court, error)
	UpdateCourt(courtID string, req *dto.UpdateCourtRequest) (*models.Court, error)
	DeleteCourt(courtID string) error
	SearchCourts(req *dto.CourtSearchRequest) (*dto.CourtSearchResponse, error)
	GetAvailableFacilities() []map[string]interface{}
	GetCourtTypes() []map[string]interface{}
}

// BookingUsecaseInterface 預訂用例接口
type BookingUsecaseInterface interface {
	CreateBooking(userID string, req *dto.CreateBookingRequest) (*models.Booking, error)
	GetBooking(bookingID string) (*models.Booking, error)
	UpdateBooking(bookingID, userID string, req *dto.UpdateBookingRequest) (*models.Booking, error)
	CancelBooking(bookingID, userID string) error
	GetBookings(req *dto.BookingListRequest) (*dto.BookingListResponse, error)
	GetAvailability(req *dto.AvailabilityRequest) (*dto.AvailabilityResponse, error)
}

// ReviewUsecaseInterface 評價用例接口
type ReviewUsecaseInterface interface {
	CreateReview(userID string, req *dto.CreateReviewRequest) (*models.CourtReview, error)
	GetReview(reviewID string) (*models.CourtReview, error)
	UpdateReview(reviewID, userID string, req *dto.UpdateReviewRequest) (*models.CourtReview, error)
	DeleteReview(reviewID, userID string) error
	GetReviews(req *dto.ReviewListRequest) (*dto.ReviewListResponse, error)
	ReportReview(reviewID, userID string, req *dto.ReportReviewRequest) error
	MarkReviewHelpful(reviewID, userID string, helpful bool) error
	GetReviewStatistics(courtID string) (*dto.ReviewStatistics, error)
}

// CourtController 場地控制器
type CourtController struct {
	courtUsecase   CourtUsecaseInterface
	reviewUsecase  ReviewUsecaseInterface
	bookingUsecase BookingUsecaseInterface
	uploadService  *services.UploadService
}

// NewCourtController 創建新的場地控制器
func NewCourtController(courtUsecase CourtUsecaseInterface, reviewUsecase ReviewUsecaseInterface, bookingUsecase BookingUsecaseInterface, uploadService *services.UploadService) *CourtController {
	return &CourtController{
		courtUsecase:   courtUsecase,
		reviewUsecase:  reviewUsecase,
		bookingUsecase: bookingUsecase,
		uploadService:  uploadService,
	}
}

// CreateCourt 創建場地
// @Summary 創建場地
// @Description 創建新的網球場地
// @Tags courts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCourtRequest true "創建場地請求"
// @Success 201 {object} models.Court
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/courts [post]
func (cc *CourtController) CreateCourt(c *gin.Context) {
	var req dto.CreateCourtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 如果有用戶ID，設置為場地擁有者
	if userID, exists := c.Get("userID"); exists {
		ownerID := userID.(string)
		req.OwnerID = &ownerID
	}

	court, err := cc.courtUsecase.CreateCourt(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, court)
}

// GetCourt 獲取場地詳情
// @Summary 獲取場地詳情
// @Description 根據ID獲取場地的詳細信息
// @Tags courts
// @Accept json
// @Produce json
// @Param id path string true "場地ID"
// @Success 200 {object} models.Court
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/courts/{id} [get]
func (cc *CourtController) GetCourt(c *gin.Context) {
	courtID := c.Param("id")
	if courtID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "場地ID不能為空",
		})
		return
	}

	court, err := cc.courtUsecase.GetCourtByID(courtID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, court)
}

// UpdateCourt 更新場地
// @Summary 更新場地
// @Description 更新場地信息
// @Tags courts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "場地ID"
// @Param request body dto.UpdateCourtRequest true "更新場地請求"
// @Success 200 {object} models.Court
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/courts/{id} [put]
func (cc *CourtController) UpdateCourt(c *gin.Context) {
	courtID := c.Param("id")
	if courtID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "場地ID不能為空",
		})
		return
	}

	var req dto.UpdateCourtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	court, err := cc.courtUsecase.UpdateCourt(courtID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, court)
}

// DeleteCourt 刪除場地
// @Summary 刪除場地
// @Description 刪除場地（軟刪除）
// @Tags courts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "場地ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/courts/{id} [delete]
func (cc *CourtController) DeleteCourt(c *gin.Context) {
	courtID := c.Param("id")
	if courtID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "場地ID不能為空",
		})
		return
	}

	if err := cc.courtUsecase.DeleteCourt(courtID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "場地刪除成功",
	})
}

// SearchCourts 搜尋場地
// @Summary 搜尋場地
// @Description 根據條件搜尋場地，支援地理位置搜尋和文字搜尋
// @Tags courts
// @Accept json
// @Produce json
// @Param query query string false "文字搜尋（場地名稱、描述、地址）"
// @Param latitude query number false "緯度"
// @Param longitude query number false "經度"
// @Param radius query number false "搜尋半徑（公里）"
// @Param minPrice query number false "最低價格"
// @Param maxPrice query number false "最高價格"
// @Param courtType query string false "場地類型" Enums(hard,clay,grass,indoor,outdoor)
// @Param facilities query array false "設施列表"
// @Param minRating query number false "最低評分"
// @Param sortBy query string false "排序欄位" Enums(distance,price,rating,name)
// @Param sortOrder query string false "排序順序" Enums(asc,desc)
// @Param page query int false "頁碼"
// @Param pageSize query int false "每頁數量"
// @Success 200 {object} dto.CourtSearchResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/courts [get]
func (cc *CourtController) SearchCourts(c *gin.Context) {
	var req dto.CourtSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	response, err := cc.courtUsecase.SearchCourts(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UploadCourtImages 上傳場地圖片
// @Summary 上傳場地圖片
// @Description 上傳場地圖片
// @Tags courts
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "場地ID"
// @Param images formData file true "圖片文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/courts/{id}/images [post]
func (cc *CourtController) UploadCourtImages(c *gin.Context) {
	courtID := c.Param("id")
	if courtID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "場地ID不能為空",
		})
		return
	}

	// 檢查場地是否存在
	_, err := cc.courtUsecase.GetCourtByID(courtID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 獲取上傳的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "獲取上傳文件失敗",
		})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "未找到上傳文件",
		})
		return
	}

	var uploadResults []services.UploadResult
	var imageURLs []string

	// 上傳每個文件
	for _, file := range files {
		uploadResult, err := cc.uploadService.UploadFile(file, "courts")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		uploadResults = append(uploadResults, *uploadResult)
		imageURLs = append(imageURLs, uploadResult.URL)
	}

	// 獲取現有圖片
	court, err := cc.courtUsecase.GetCourtByID(courtID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 合併新舊圖片URL
	allImages := append(court.Images, imageURLs...)

	// 更新場地圖片
	updateReq := dto.UpdateCourtRequest{
		Images: allImages,
	}

	updatedCourt, err := cc.courtUsecase.UpdateCourt(courtID, &updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "圖片上傳成功",
		"court":   updatedCourt,
		"uploads": uploadResults,
	})
}

// GetAvailableFacilities 獲取可用設施列表
// @Summary 獲取可用設施列表
// @Description 獲取所有可用的場地設施選項
// @Tags courts
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/courts/facilities [get]
func (cc *CourtController) GetAvailableFacilities(c *gin.Context) {
	facilities := cc.courtUsecase.GetAvailableFacilities()

	c.JSON(http.StatusOK, gin.H{
		"facilities": facilities,
	})
}

// GetCourtTypes 獲取場地類型列表
// @Summary 獲取場地類型列表
// @Description 獲取所有可用的場地類型選項
// @Tags courts
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/courts/types [get]
func (cc *CourtController) GetCourtTypes(c *gin.Context) {
	types := cc.courtUsecase.GetCourtTypes()

	c.JSON(http.StatusOK, gin.H{
		"types": types,
	})
}

// CreateReview 創建場地評價
// @Summary 創建場地評價
// @Description 為場地創建評價和評分
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateReviewRequest true "創建評價請求"
// @Success 201 {object} models.CourtReview
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/reviews [post]
func (cc *CourtController) CreateReview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	review, err := cc.reviewUsecase.CreateReview(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GetReview 獲取評價詳情
// @Summary 獲取評價詳情
// @Description 根據ID獲取評價的詳細信息
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "評價ID"
// @Success 200 {object} models.CourtReview
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/reviews/{id} [get]
func (cc *CourtController) GetReview(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	review, err := cc.reviewUsecase.GetReview(reviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// UpdateReview 更新評價
// @Summary 更新評價
// @Description 更新評價內容
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "評價ID"
// @Param request body dto.UpdateReviewRequest true "更新評價請求"
// @Success 200 {object} models.CourtReview
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/reviews/{id} [put]
func (cc *CourtController) UpdateReview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	var req dto.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	review, err := cc.reviewUsecase.UpdateReview(reviewID, userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, review)
}

// DeleteReview 刪除評價
// @Summary 刪除評價
// @Description 刪除評價（軟刪除）
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "評價ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/reviews/{id} [delete]
func (cc *CourtController) DeleteReview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	if err := cc.reviewUsecase.DeleteReview(reviewID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "評價刪除成功",
	})
}

// GetReviews 獲取評價列表
// @Summary 獲取評價列表
// @Description 根據條件獲取評價列表
// @Tags reviews
// @Accept json
// @Produce json
// @Param courtId query string false "場地ID"
// @Param userId query string false "用戶ID"
// @Param rating query int false "評分篩選"
// @Param sortBy query string false "排序欄位" Enums(rating,created_at,helpful)
// @Param sortOrder query string false "排序順序" Enums(asc,desc)
// @Param page query int false "頁碼"
// @Param pageSize query int false "每頁數量"
// @Success 200 {object} dto.ReviewListResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/reviews [get]
func (cc *CourtController) GetReviews(c *gin.Context) {
	var req dto.ReviewListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	response, err := cc.reviewUsecase.GetReviews(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReportReview 舉報評價
// @Summary 舉報評價
// @Description 舉報不當評價
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "評價ID"
// @Param request body dto.ReportReviewRequest true "舉報評價請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/reviews/{id}/report [post]
func (cc *CourtController) ReportReview(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	var req dto.ReportReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	if err := cc.reviewUsecase.ReportReview(reviewID, userID.(string), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "舉報提交成功",
	})
}

// MarkReviewHelpful 標記評價為有用
// @Summary 標記評價為有用
// @Description 標記評價為有用或取消標記
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "評價ID"
// @Param helpful query bool true "是否有用"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/reviews/{id}/helpful [post]
func (cc *CourtController) MarkReviewHelpful(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "評價ID不能為空",
		})
		return
	}

	helpful := c.Query("helpful") == "true"

	if err := cc.reviewUsecase.MarkReviewHelpful(reviewID, userID.(string), helpful); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "標記成功",
	})
}

// GetReviewStatistics 獲取場地評價統計
// @Summary 獲取場地評價統計
// @Description 獲取場地的評價統計信息
// @Tags reviews
// @Accept json
// @Produce json
// @Param courtId path string true "場地ID"
// @Success 200 {object} dto.ReviewStatistics
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/courts/{courtId}/reviews/statistics [get]
func (cc *CourtController) GetReviewStatistics(c *gin.Context) {
	courtID := c.Param("courtId")
	if courtID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "場地ID不能為空",
		})
		return
	}

	stats, err := cc.reviewUsecase.GetReviewStatistics(courtID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UploadReviewImages 上傳評價圖片
// @Summary 上傳評價圖片
// @Description 上傳評價圖片
// @Tags reviews
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param images formData file true "圖片文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/reviews/images [post]
func (cc *CourtController) UploadReviewImages(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	// 獲取上傳的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "獲取上傳文件失敗",
		})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "未找到上傳文件",
		})
		return
	}

	var uploadResults []services.UploadResult
	var imageURLs []string

	// 上傳每個文件
	for _, file := range files {
		uploadResult, err := cc.uploadService.UploadFile(file, "reviews")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		uploadResults = append(uploadResults, *uploadResult)
		imageURLs = append(imageURLs, uploadResult.URL)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "圖片上傳成功",
		"uploads":   uploadResults,
		"imageUrls": imageURLs,
	})
}

// CreateBooking 創建場地預訂
// @Summary 創建場地預訂
// @Description 為指定場地創建預訂
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBookingRequest true "創建預訂請求"
// @Success 201 {object} models.Booking
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/bookings [post]
func (cc *CourtController) CreateBooking(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	booking, err := cc.bookingUsecase.CreateBooking(userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// GetBooking 獲取預訂詳情
// @Summary 獲取預訂詳情
// @Description 根據ID獲取預訂的詳細信息
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "預訂ID"
// @Success 200 {object} models.Booking
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/bookings/{id} [get]
func (cc *CourtController) GetBooking(c *gin.Context) {
	bookingID := c.Param("id")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "預訂ID不能為空",
		})
		return
	}

	booking, err := cc.bookingUsecase.GetBooking(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// UpdateBooking 更新預訂
// @Summary 更新預訂
// @Description 更新預訂信息
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "預訂ID"
// @Param request body dto.UpdateBookingRequest true "更新預訂請求"
// @Success 200 {object} models.Booking
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/bookings/{id} [put]
func (cc *CourtController) UpdateBooking(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	bookingID := c.Param("id")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "預訂ID不能為空",
		})
		return
	}

	var req dto.UpdateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	booking, err := cc.bookingUsecase.UpdateBooking(bookingID, userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// CancelBooking 取消預訂
// @Summary 取消預訂
// @Description 取消指定的預訂
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "預訂ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/bookings/{id}/cancel [post]
func (cc *CourtController) CancelBooking(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用戶未認證",
		})
		return
	}

	bookingID := c.Param("id")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "預訂ID不能為空",
		})
		return
	}

	if err := cc.bookingUsecase.CancelBooking(bookingID, userID.(string)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "預訂取消成功",
	})
}

// GetBookings 獲取預訂列表
// @Summary 獲取預訂列表
// @Description 根據條件獲取預訂列表
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param courtId query string false "場地ID"
// @Param userId query string false "用戶ID"
// @Param status query string false "預訂狀態" Enums(pending,confirmed,cancelled,completed)
// @Param startDate query string false "開始日期 (YYYY-MM-DD)"
// @Param endDate query string false "結束日期 (YYYY-MM-DD)"
// @Param page query int false "頁碼"
// @Param pageSize query int false "每頁數量"
// @Success 200 {object} dto.BookingListResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/bookings [get]
func (cc *CourtController) GetBookings(c *gin.Context) {
	var req dto.BookingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	response, err := cc.bookingUsecase.GetBookings(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetCourtAvailability 獲取場地可用時間
// @Summary 獲取場地可用時間
// @Description 查詢指定場地在指定日期的可用時間段
// @Tags bookings
// @Accept json
// @Produce json
// @Param courtId query string true "場地ID"
// @Param date query string true "查詢日期 (YYYY-MM-DD)"
// @Param duration query int false "預訂時長（分鐘），默認60分鐘"
// @Success 200 {object} dto.AvailabilityResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/courts/availability [get]
func (cc *CourtController) GetCourtAvailability(c *gin.Context) {
	var req dto.AvailabilityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	response, err := cc.bookingUsecase.GetAvailability(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
