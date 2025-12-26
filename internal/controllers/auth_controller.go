package controllers

import (
	"net/http"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/usecases"

	"github.com/gin-gonic/gin"
)

// AuthController 認證控制器
type AuthController struct {
	authUsecase *usecases.AuthUsecase
}

// NewAuthController 創建新的認證控制器
func NewAuthController(authUsecase *usecases.AuthUsecase) *AuthController {
	return &AuthController{
		authUsecase: authUsecase,
	}
}

// Register 用戶註冊
// @Summary 用戶註冊
// @Description 註冊新用戶帳號
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "註冊請求"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	response, err := ac.authUsecase.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login 用戶登入
// @Summary 用戶登入
// @Description 用戶登入系統
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登入請求"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	response, err := ac.authUsecase.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken 刷新訪問令牌
// @Summary 刷新訪問令牌
// @Description 使用刷新令牌獲取新的訪問令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新令牌請求"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	response, err := ac.authUsecase.RefreshToken(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout 用戶登出
// @Summary 用戶登出
// @Description 用戶登出系統，撤銷刷新令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "登出請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	if err := ac.authUsecase.Logout(req.RefreshToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "登出失敗",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
	})
}

// ForgotPassword 忘記密碼
// @Summary 忘記密碼
// @Description 發送密碼重設郵件
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "忘記密碼請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	if err := ac.authUsecase.ForgotPassword(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "如果該郵箱存在，我們已發送密碼重設連結",
	})
}

// ResetPassword 重設密碼
// @Summary 重設密碼
// @Description 使用重設令牌重設密碼
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "重設密碼請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/reset-password [post]
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	if err := ac.authUsecase.ResetPassword(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密碼重設成功",
	})
}

// GetOAuthAuthURL 獲取 OAuth 授權 URL
// @Summary 獲取 OAuth 授權 URL
// @Description 獲取指定提供商的 OAuth 授權 URL
// @Tags auth
// @Produce json
// @Param provider path string true "OAuth 提供商" Enums(google,facebook,apple)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/oauth/{provider} [get]
func (ac *AuthController) GetOAuthAuthURL(c *gin.Context) {
	provider := c.Param("provider")

	authURL, err := ac.authUsecase.GetOAuthAuthURL(provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authUrl": authURL,
	})
}

// OAuthCallback OAuth 回調處理
// @Summary OAuth 回調處理
// @Description 處理 OAuth 提供商的回調並完成登入
// @Tags auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth 提供商" Enums(google,facebook,apple)
// @Param request body dto.OAuthLoginRequest true "OAuth 登入請求"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/oauth/{provider}/callback [post]
func (ac *AuthController) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")

	var req dto.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 設置提供商
	req.Provider = provider

	response, err := ac.authUsecase.OAuthLogin(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// LinkOAuthAccount 關聯 OAuth 帳號
// @Summary 關聯 OAuth 帳號
// @Description 將 OAuth 帳號關聯到當前用戶
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "OAuth 提供商" Enums(google,facebook,apple)
// @Param request body dto.LinkOAuthAccountRequest true "關聯 OAuth 帳號請求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/oauth/{provider}/link [post]
func (ac *AuthController) LinkOAuthAccount(c *gin.Context) {
	provider := c.Param("provider")
	userID := c.GetString("userID") // 從中間件獲取用戶ID

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授權",
		})
		return
	}

	var req dto.LinkOAuthAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "請求參數錯誤",
			"details": err.Error(),
		})
		return
	}

	// 設置提供商
	req.Provider = provider

	if err := ac.authUsecase.LinkOAuthAccount(userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OAuth 帳號關聯成功",
	})
}

// UnlinkOAuthAccount 解除關聯 OAuth 帳號
// @Summary 解除關聯 OAuth 帳號
// @Description 解除當前用戶與指定 OAuth 提供商的關聯
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "OAuth 提供商" Enums(google,facebook,apple)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/oauth/{provider}/unlink [delete]
func (ac *AuthController) UnlinkOAuthAccount(c *gin.Context) {
	provider := c.Param("provider")
	userID := c.GetString("userID") // 從中間件獲取用戶ID

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授權",
		})
		return
	}

	req := dto.UnlinkOAuthAccountRequest{
		Provider: provider,
	}

	if err := ac.authUsecase.UnlinkOAuthAccount(userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OAuth 帳號解除關聯成功",
	})
}

// GetLinkedOAuthAccounts 獲取已關聯的 OAuth 帳號
// @Summary 獲取已關聯的 OAuth 帳號
// @Description 獲取當前用戶已關聯的所有 OAuth 帳號列表
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/oauth/accounts [get]
func (ac *AuthController) GetLinkedOAuthAccounts(c *gin.Context) {
	userID := c.GetString("userID") // 從中間件獲取用戶ID

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授權",
		})
		return
	}

	accounts, err := ac.authUsecase.GetLinkedOAuthAccounts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "獲取 OAuth 帳號列表失敗",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}
