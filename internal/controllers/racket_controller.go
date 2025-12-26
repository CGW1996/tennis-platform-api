package controllers

import (
	"net/http"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RacketUsecaseInterface 球拍用例接口
type RacketUsecaseInterface interface {
	CreateRacket(req *dto.CreateRacketRequest) (*models.Racket, error)
	GetRacketByID(racketID string) (*models.Racket, error)
	UpdateRacket(racketID string, req *dto.UpdateRacketRequest) (*models.Racket, error)
	DeleteRacket(racketID string) error
	SearchRackets(req *dto.RacketSearchRequest) (*dto.RacketSearchResponse, error)
	GetAvailableBrands() ([]string, error)
	GetRacketSpecifications() map[string]interface{}
}

// RacketPriceUsecaseInterface 球拍價格用例接口
type RacketPriceUsecaseInterface interface {
	CreateRacketPrice(req *dto.CreateRacketPriceRequest) (*models.RacketPrice, error)
	UpdateRacketPrice(priceID string, req *dto.UpdateRacketPriceRequest) (*models.RacketPrice, error)
	DeleteRacketPrice(priceID string) error
	GetRacketPrices(racketID string) ([]models.RacketPrice, error)
	GetLowestPrice(racketID string) (*models.RacketPrice, error)
	UpdatePriceAvailability(priceID string, isAvailable bool) error
}

// RacketReviewUsecaseInterface 球拍評價用例接口
type RacketReviewUsecaseInterface interface {
	CreateRacketReview(userID string, req *dto.CreateRacketReviewRequest) (*models.RacketReview, error)
	GetRacketReview(reviewID string) (*models.RacketReview, error)
	UpdateRacketReview(reviewID, userID string, req *dto.UpdateRacketReviewRequest) (*models.RacketReview, error)
	DeleteRacketReview(reviewID, userID string) error
	GetRacketReviews(req *dto.RacketReviewListRequest) (*dto.RacketReviewListResponse, error)
	MarkReviewHelpful(reviewID, userID string, helpful bool) error
	GetRacketReviewStatistics(racketID string) (*dto.RacketReviewStatistics, error)
}

// RacketController 球拍控制器
type RacketController struct {
	racketUsecase RacketUsecaseInterface
	priceUsecase  RacketPriceUsecaseInterface
	reviewUsecase RacketReviewUsecaseInterface
	uploadService *services.UploadService
}

// NewRacketController 創建新的球拍控制器
func NewRacketController(
	racketUsecase RacketUsecaseInterface,
	priceUsecase RacketPriceUsecaseInterface,
	reviewUsecase RacketReviewUsecaseInterface,
	uploadService *services.UploadService,
) *RacketController {
	return &RacketController{
		racketUsecase: racketUsecase,
		priceUsecase:  priceUsecase,
		reviewUsecase: reviewUsecase,
		uploadService: uploadService,
	}
}

// CreateRacket 創建球拍
// @Summary 創建球拍
// @Description 創建新的球拍記錄
// @Tags rackets
// @Accept json
// @Produce json
// @Param racket body dto.CreateRacketRequest true "球拍資訊"
// @Success 201 {object} models.Racket
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets [post]
func (c *RacketController) CreateRacket(ctx *gin.Context) {
	var req dto.CreateRacketRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	racket, err := c.racketUsecase.CreateRacket(&req)
	if err != nil {
		if err.Error() == "racket with same brand and model already exists" {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "Racket already exists",
				"message": "A racket with the same brand and model already exists",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create racket",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, racket)
}

// GetRacket 獲取球拍詳情
// @Summary 獲取球拍詳情
// @Description 根據ID獲取球拍詳細資訊
// @Tags rackets
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Success 200 {object} models.Racket
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id} [get]
func (c *RacketController) GetRacket(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	racket, err := c.racketUsecase.GetRacketByID(racketID)
	if err != nil {
		if err.Error() == "racket not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Racket not found",
				"message": "The requested racket does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get racket",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, racket)
}

// UpdateRacket 更新球拍
// @Summary 更新球拍
// @Description 更新球拍資訊
// @Tags rackets
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Param racket body dto.UpdateRacketRequest true "更新的球拍資訊"
// @Success 200 {object} models.Racket
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id} [put]
func (c *RacketController) UpdateRacket(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	var req dto.UpdateRacketRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	racket, err := c.racketUsecase.UpdateRacket(racketID, &req)
	if err != nil {
		if err.Error() == "racket not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Racket not found",
				"message": "The requested racket does not exist",
			})
			return
		}
		if err.Error() == "racket with same brand and model already exists" {
			ctx.JSON(http.StatusConflict, gin.H{
				"error":   "Racket already exists",
				"message": "A racket with the same brand and model already exists",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update racket",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, racket)
}

// DeleteRacket 刪除球拍
// @Summary 刪除球拍
// @Description 軟刪除球拍記錄
// @Tags rackets
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id} [delete]
func (c *RacketController) DeleteRacket(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	err := c.racketUsecase.DeleteRacket(racketID)
	if err != nil {
		if err.Error() == "racket not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Racket not found",
				"message": "The requested racket does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete racket",
			"message": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// SearchRackets 搜尋球拍
// @Summary 搜尋球拍
// @Description 根據條件搜尋球拍
// @Tags rackets
// @Accept json
// @Produce json
// @Param query query string false "搜尋關鍵字"
// @Param brand query string false "品牌"
// @Param minHeadSize query int false "最小拍面大小"
// @Param maxHeadSize query int false "最大拍面大小"
// @Param minWeight query int false "最小重量"
// @Param maxWeight query int false "最大重量"
// @Param minPrice query number false "最低價格"
// @Param maxPrice query number false "最高價格"
// @Param powerLevel query int false "力量等級"
// @Param controlLevel query int false "控制等級"
// @Param maneuverLevel query int false "操控等級"
// @Param stabilityLevel query int false "穩定等級"
// @Param minRating query number false "最低評分"
// @Param sortBy query string false "排序欄位" Enums(brand, model, price, rating, popularity)
// @Param sortOrder query string false "排序順序" Enums(asc, desc)
// @Param page query int false "頁碼"
// @Param pageSize query int false "每頁數量"
// @Success 200 {object} dto.RacketSearchResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets [get]
func (c *RacketController) SearchRackets(ctx *gin.Context) {
	var req dto.RacketSearchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	response, err := c.racketUsecase.SearchRackets(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search rackets",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetAvailableBrands 獲取可用品牌
// @Summary 獲取可用品牌
// @Description 獲取所有可用的球拍品牌列表
// @Tags rackets
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/brands [get]
func (c *RacketController) GetAvailableBrands(ctx *gin.Context) {
	brands, err := c.racketUsecase.GetAvailableBrands()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get available brands",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"brands": brands,
	})
}

// GetRacketSpecifications 獲取球拍規格選項
// @Summary 獲取球拍規格選項
// @Description 獲取球拍規格的可選項目和範圍
// @Tags rackets
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/rackets/specifications [get]
func (c *RacketController) GetRacketSpecifications(ctx *gin.Context) {
	specifications := c.racketUsecase.GetRacketSpecifications()
	ctx.JSON(http.StatusOK, specifications)
}

// UploadRacketImages 上傳球拍圖片
// @Summary 上傳球拍圖片
// @Description 上傳球拍相關圖片
// @Tags rackets
// @Accept multipart/form-data
// @Produce json
// @Param images formData file true "圖片文件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/images [post]
func (c *RacketController) UploadRacketImages(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to parse multipart form",
			"message": err.Error(),
		})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "No images provided",
		})
		return
	}

	var imageURLs []string
	for _, file := range files {
		result, err := c.uploadService.UploadFile(file, "rackets")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to upload image",
				"message": err.Error(),
			})
			return
		}
		imageURLs = append(imageURLs, result.URL)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"images": imageURLs,
	})
}

// GetRacketPrices 獲取球拍價格
// @Summary 獲取球拍價格
// @Description 獲取指定球拍的所有價格資訊
// @Tags racket-prices
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id}/prices [get]
func (c *RacketController) GetRacketPrices(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	// 檢查球拍是否存在
	_, err := c.racketUsecase.GetRacketByID(racketID)
	if err != nil {
		if err.Error() == "racket not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Racket not found",
				"message": "The requested racket does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to verify racket",
			"message": err.Error(),
		})
		return
	}

	prices, err := c.priceUsecase.GetRacketPrices(racketID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get racket prices",
			"message": err.Error(),
		})
		return
	}

	// 獲取最低價格
	lowestPrice, _ := c.priceUsecase.GetLowestPrice(racketID)

	ctx.JSON(http.StatusOK, gin.H{
		"prices":      prices,
		"lowestPrice": lowestPrice,
	})
}

// CreateRacketPrice 創建球拍價格
// @Summary 創建球拍價格
// @Description 為球拍添加新的價格資訊
// @Tags racket-prices
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Param price body dto.CreateRacketPriceRequest true "價格資訊"
// @Success 201 {object} models.RacketPrice
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id}/prices [post]
func (c *RacketController) CreateRacketPrice(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	var req dto.CreateRacketPriceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// 設置球拍ID
	req.RacketID = racketID

	price, err := c.priceUsecase.CreateRacketPrice(&req)
	if err != nil {
		if err.Error() == "racket not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Racket not found",
				"message": "The requested racket does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create racket price",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, price)
}

// UpdateRacketPrice 更新球拍價格
// @Summary 更新球拍價格
// @Description 更新球拍價格資訊
// @Tags racket-prices
// @Accept json
// @Produce json
// @Param priceId path string true "價格ID"
// @Param price body dto.UpdateRacketPriceRequest true "更新的價格資訊"
// @Success 200 {object} models.RacketPrice
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/racket-prices/{priceId} [put]
func (c *RacketController) UpdateRacketPrice(ctx *gin.Context) {
	priceID := ctx.Param("priceId")
	if priceID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Price ID is required",
		})
		return
	}

	var req dto.UpdateRacketPriceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	price, err := c.priceUsecase.UpdateRacketPrice(priceID, &req)
	if err != nil {
		if err.Error() == "price not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Price not found",
				"message": "The requested price does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update racket price",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, price)
}

// DeleteRacketPrice 刪除球拍價格
// @Summary 刪除球拍價格
// @Description 刪除球拍價格記錄
// @Tags racket-prices
// @Accept json
// @Produce json
// @Param priceId path string true "價格ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/racket-prices/{priceId} [delete]
func (c *RacketController) DeleteRacketPrice(ctx *gin.Context) {
	priceID := ctx.Param("priceId")
	if priceID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Price ID is required",
		})
		return
	}

	err := c.priceUsecase.DeleteRacketPrice(priceID)
	if err != nil {
		if err.Error() == "price not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Price not found",
				"message": "The requested price does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete racket price",
			"message": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// UpdatePriceAvailability 更新價格可用性
// @Summary 更新價格可用性
// @Description 更新球拍價格的可用性狀態
// @Tags racket-prices
// @Accept json
// @Produce json
// @Param priceId path string true "價格ID"
// @Param availability body map[string]bool true "可用性狀態"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/racket-prices/{priceId}/availability [put]
func (c *RacketController) UpdatePriceAvailability(ctx *gin.Context) {
	priceID := ctx.Param("priceId")
	if priceID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Price ID is required",
		})
		return
	}

	var req struct {
		IsAvailable bool `json:"isAvailable"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := c.priceUsecase.UpdatePriceAvailability(priceID, req.IsAvailable)
	if err != nil {
		if err.Error() == "price not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Price not found",
				"message": "The requested price does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update price availability",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Price availability updated successfully",
		"isAvailable": req.IsAvailable,
	})
}

// GetRacketReviews 獲取球拍評價
// @Summary 獲取球拍評價
// @Description 獲取指定球拍的評價列表
// @Tags racket-reviews
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Param page query int false "頁碼"
// @Param pageSize query int false "每頁數量"
// @Param sortBy query string false "排序欄位" Enums(rating, date, helpful)
// @Param sortOrder query string false "排序順序" Enums(asc, desc)
// @Success 200 {object} dto.RacketReviewListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id}/reviews [get]
func (c *RacketController) GetRacketReviews(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	var req dto.RacketReviewListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// 設置球拍ID
	req.RacketID = &racketID

	response, err := c.reviewUsecase.GetRacketReviews(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get racket reviews",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateRacketReview 創建球拍評價
// @Summary 創建球拍評價
// @Description 為球拍創建新的評價
// @Tags racket-reviews
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Param review body dto.CreateRacketReviewRequest true "評價資訊"
// @Success 201 {object} models.RacketReview
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id}/reviews [post]
func (c *RacketController) CreateRacketReview(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req dto.CreateRacketReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// 設置球拍ID
	req.RacketID = racketID

	review, err := c.reviewUsecase.CreateRacketReview(userID.(string), &req)
	if err != nil {
		if err.Error() == "racket not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Racket not found",
				"message": "The requested racket does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create racket review",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, review)
}

// GetRacketReviewStatistics 獲取球拍評價統計
// @Summary 獲取球拍評價統計
// @Description 獲取指定球拍的評價統計資訊
// @Tags racket-reviews
// @Accept json
// @Produce json
// @Param id path string true "球拍ID"
// @Success 200 {object} dto.RacketReviewStatistics
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rackets/{id}/reviews/statistics [get]
func (c *RacketController) GetRacketReviewStatistics(ctx *gin.Context) {
	racketID := ctx.Param("id")
	if racketID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Racket ID is required",
		})
		return
	}

	statistics, err := c.reviewUsecase.GetRacketReviewStatistics(racketID)
	if err != nil {
		if err.Error() == "racket not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Racket not found",
				"message": "The requested racket does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get review statistics",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, statistics)
}

// MarkRacketReviewHelpful 標記評價有用
// @Summary 標記評價有用
// @Description 標記球拍評價為有用或無用
// @Tags racket-reviews
// @Accept json
// @Produce json
// @Param reviewId path string true "評價ID"
// @Param helpful body map[string]bool true "是否有用"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/racket-reviews/{reviewId}/helpful [post]
func (c *RacketController) MarkRacketReviewHelpful(ctx *gin.Context) {
	reviewID := ctx.Param("reviewId")
	if reviewID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Review ID is required",
		})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req struct {
		Helpful bool `json:"helpful"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := c.reviewUsecase.MarkReviewHelpful(reviewID, userID.(string), req.Helpful)
	if err != nil {
		if err.Error() == "review not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "Review not found",
				"message": "The requested review does not exist",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark review helpful",
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Review marked successfully",
		"helpful": req.Helpful,
	})
}
