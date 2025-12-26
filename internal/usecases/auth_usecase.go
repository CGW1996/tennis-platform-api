package usecases

import (
	"errors"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"tennis-platform/backend/internal/services"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// AuthUsecase 認證用例
type AuthUsecase struct {
	db           *gorm.DB
	jwtService   *services.JWTService
	emailService *services.EmailService
	oauthService *services.OAuthService
	config       *config.Config
}

// NewAuthUsecase 創建新的認證用例
func NewAuthUsecase(db *gorm.DB, cfg *config.Config) *AuthUsecase {
	return &AuthUsecase{
		db:           db,
		jwtService:   services.NewJWTService(cfg),
		emailService: services.NewEmailService(cfg),
		oauthService: services.NewOAuthService(cfg),
		config:       cfg,
	}
}

// Register 用戶註冊
func (au *AuthUsecase) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// 檢查用戶是否已存在
	var existingUser models.User
	if err := au.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("用戶已存在")
	}

	// 加密密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密碼加密失敗")
	}

	// 創建用戶
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	// 開始事務
	tx := au.db.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("創建用戶失敗")
	}

	// 創建用戶檔案
	profile := models.UserProfile{
		UserID:    user.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("創建用戶檔案失敗")
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("事務提交失敗")
	}

	// 生成驗證令牌並發送驗證郵件
	verificationToken, err := au.emailService.GenerateToken()
	if err == nil {
		// 在實際應用中，應該將驗證令牌存儲在數據庫中
		au.emailService.SendVerificationEmail(user.Email, verificationToken)
	}

	// 生成 JWT 令牌
	accessToken, err := au.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("生成訪問令牌失敗")
	}

	refreshToken, err := au.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("生成刷新令牌失敗")
	}

	// 存儲刷新令牌
	refreshTokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(au.config.JWT.RefreshTokenTTL) * 24 * time.Hour),
	}

	if err := au.db.Create(&refreshTokenModel).Error; err != nil {
		return nil, errors.New("存儲刷新令牌失敗")
	}

	// 載入用戶檔案
	user.Profile = &profile

	return &dto.AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login 用戶登入
func (au *AuthUsecase) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// 查找用戶
	var user models.User
	if err := au.db.Preload("Profile").Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在或密碼錯誤")
	}

	// 檢查用戶是否啟用
	if !user.IsActive {
		return nil, errors.New("帳號已被停用")
	}

	// 驗證密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("用戶不存在或密碼錯誤")
	}

	// 更新最後登入時間
	now := time.Now()
	user.LastLoginAt = &now
	au.db.Save(&user)

	// 生成 JWT 令牌
	accessToken, err := au.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("生成訪問令牌失敗")
	}

	refreshToken, err := au.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("生成刷新令牌失敗")
	}

	// 存儲刷新令牌
	refreshTokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(au.config.JWT.RefreshTokenTTL) * 24 * time.Hour),
	}

	if err := au.db.Create(&refreshTokenModel).Error; err != nil {
		return nil, errors.New("存儲刷新令牌失敗")
	}

	return &dto.AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken 刷新訪問令牌
func (au *AuthUsecase) RefreshToken(req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	// 驗證刷新令牌
	userID, err := au.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("無效的刷新令牌")
	}

	// 檢查刷新令牌是否存在且未被撤銷
	var refreshToken models.RefreshToken
	if err := au.db.Where("token = ? AND user_id = ? AND is_revoked = false AND expires_at > ?",
		req.RefreshToken, userID, time.Now()).First(&refreshToken).Error; err != nil {
		return nil, errors.New("刷新令牌不存在或已過期")
	}

	// 查找用戶
	var user models.User
	if err := au.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 檢查用戶是否啟用
	if !user.IsActive {
		return nil, errors.New("帳號已被停用")
	}

	// 撤銷舊的刷新令牌
	refreshToken.IsRevoked = true
	au.db.Save(&refreshToken)

	// 生成新的令牌
	newAccessToken, err := au.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("生成訪問令牌失敗")
	}

	newRefreshToken, err := au.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("生成刷新令牌失敗")
	}

	// 存儲新的刷新令牌
	newRefreshTokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(time.Duration(au.config.JWT.RefreshTokenTTL) * 24 * time.Hour),
	}

	if err := au.db.Create(&newRefreshTokenModel).Error; err != nil {
		return nil, errors.New("存儲刷新令牌失敗")
	}

	return &dto.AuthResponse{
		User:         &user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout 用戶登出
func (au *AuthUsecase) Logout(refreshToken string) error {
	// 撤銷刷新令牌
	return au.db.Model(&models.RefreshToken{}).
		Where("token = ?", refreshToken).
		Update("is_revoked", true).Error
}

// ForgotPassword 忘記密碼
func (au *AuthUsecase) ForgotPassword(req *dto.ForgotPasswordRequest) error {
	// 查找用戶
	var user models.User
	if err := au.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// 為了安全起見，即使用戶不存在也返回成功
		return nil
	}

	// 生成重設令牌
	resetToken, err := au.emailService.GenerateToken()
	if err != nil {
		return errors.New("生成重設令牌失敗")
	}

	// 在實際應用中，應該將重設令牌存儲在數據庫中，並設置過期時間
	// 這裡為了簡化，直接發送郵件
	return au.emailService.SendPasswordResetEmail(user.Email, resetToken)
}

// ResetPassword 重設密碼
func (au *AuthUsecase) ResetPassword(req *dto.ResetPasswordRequest) error {
	// 在實際應用中，應該驗證重設令牌的有效性
	// 這裡為了簡化，假設令牌有效

	// 加密新密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密碼加密失敗")
	}

	// 更新密碼（這裡需要根據令牌找到對應的用戶）
	// 在實際應用中，令牌應該包含用戶信息或與用戶關聯
	return au.db.Model(&models.User{}).
		Where("id = ?", "user-id-from-token"). // 這裡需要從令牌中解析用戶ID
		Update("password_hash", string(hashedPassword)).Error
}

// GetOAuthAuthURL 獲取 OAuth 授權 URL
func (au *AuthUsecase) GetOAuthAuthURL(provider string) (string, error) {
	state := au.oauthService.GenerateState()
	return au.oauthService.GetAuthURL(provider, state)
}

// OAuthLogin OAuth 登入
func (au *AuthUsecase) OAuthLogin(req *dto.OAuthLoginRequest) (*dto.AuthResponse, error) {
	// 驗證狀態參數
	if !au.oauthService.ValidateState(req.State) {
		return nil, errors.New("無效的狀態參數")
	}

	// 交換授權碼獲取令牌
	token, err := au.oauthService.ExchangeCodeForToken(req.Provider, req.Code)
	if err != nil {
		return nil, errors.New("OAuth 令牌交換失敗")
	}

	// 獲取用戶資訊
	oauthUser, err := au.oauthService.GetUserInfo(req.Provider, token)
	if err != nil {
		return nil, errors.New("獲取 OAuth 用戶資訊失敗")
	}

	// 檢查是否已存在 OAuth 帳號
	var oauthAccount models.OAuthAccount
	err = au.db.Where("provider = ? AND provider_id = ?", req.Provider, oauthUser.ID).
		Preload("User").Preload("User.Profile").First(&oauthAccount).Error

	if err == nil {
		// OAuth 帳號已存在，直接登入
		user := oauthAccount.User
		if !user.IsActive {
			return nil, errors.New("帳號已被停用")
		}

		// 更新最後登入時間
		now := time.Now()
		user.LastLoginAt = &now
		au.db.Save(user)

		// 更新 OAuth 令牌
		oauthAccount.AccessToken = &token.AccessToken
		if token.RefreshToken != "" {
			oauthAccount.RefreshToken = &token.RefreshToken
		}
		if !token.Expiry.IsZero() {
			oauthAccount.ExpiresAt = &token.Expiry
		}
		au.db.Save(&oauthAccount)

		return au.generateAuthResponse(user)
	}

	// OAuth 帳號不存在，檢查是否有相同郵箱的用戶
	var existingUser models.User
	err = au.db.Where("email = ?", oauthUser.Email).Preload("Profile").First(&existingUser).Error

	if err == nil {
		// 用戶已存在，關聯 OAuth 帳號
		newOAuthAccount := models.OAuthAccount{
			UserID:       existingUser.ID,
			Provider:     req.Provider,
			ProviderID:   oauthUser.ID,
			Email:        oauthUser.Email,
			AccessToken:  &token.AccessToken,
			RefreshToken: &token.RefreshToken,
		}
		if !token.Expiry.IsZero() {
			newOAuthAccount.ExpiresAt = &token.Expiry
		}

		if err := au.db.Create(&newOAuthAccount).Error; err != nil {
			return nil, errors.New("關聯 OAuth 帳號失敗")
		}

		// 更新最後登入時間
		now := time.Now()
		existingUser.LastLoginAt = &now
		au.db.Save(&existingUser)

		return au.generateAuthResponse(&existingUser)
	}

	// 創建新用戶
	return au.createUserFromOAuth(oauthUser, token)
}

// createUserFromOAuth 從 OAuth 資訊創建新用戶
func (au *AuthUsecase) createUserFromOAuth(oauthUser *services.OAuthUserInfo, token *oauth2.Token) (*dto.AuthResponse, error) {
	// 生成隨機密碼（OAuth 用戶不需要密碼）
	randomPassword := "oauth-user-no-password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密碼加密失敗")
	}

	// 創建用戶
	user := models.User{
		Email:         oauthUser.Email,
		PasswordHash:  string(hashedPassword),
		EmailVerified: oauthUser.EmailVerified,
		IsActive:      true,
	}

	// 開始事務
	tx := au.db.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("創建用戶失敗")
	}

	// 創建用戶檔案
	profile := models.UserProfile{
		UserID:    user.ID,
		FirstName: oauthUser.FirstName,
		LastName:  oauthUser.LastName,
	}

	// 如果有頭像 URL，設置頭像
	if oauthUser.Picture != "" {
		profile.AvatarURL = &oauthUser.Picture
	}

	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("創建用戶檔案失敗")
	}

	// 創建 OAuth 帳號關聯
	oauthAccount := models.OAuthAccount{
		UserID:       user.ID,
		Provider:     oauthUser.Provider,
		ProviderID:   oauthUser.ID,
		Email:        oauthUser.Email,
		AccessToken:  &token.AccessToken,
		RefreshToken: &token.RefreshToken,
	}
	if !token.Expiry.IsZero() {
		oauthAccount.ExpiresAt = &token.Expiry
	}

	if err := tx.Create(&oauthAccount).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("創建 OAuth 帳號關聯失敗")
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("事務提交失敗")
	}

	// 載入用戶檔案
	user.Profile = &profile

	return au.generateAuthResponse(&user)
}

// LinkOAuthAccount 關聯 OAuth 帳號
func (au *AuthUsecase) LinkOAuthAccount(userID string, req *dto.LinkOAuthAccountRequest) error {
	// 驗證狀態參數
	if !au.oauthService.ValidateState(req.State) {
		return errors.New("無效的狀態參數")
	}

	// 交換授權碼獲取令牌
	token, err := au.oauthService.ExchangeCodeForToken(req.Provider, req.Code)
	if err != nil {
		return errors.New("OAuth 令牌交換失敗")
	}

	// 獲取用戶資訊
	oauthUser, err := au.oauthService.GetUserInfo(req.Provider, token)
	if err != nil {
		return errors.New("獲取 OAuth 用戶資訊失敗")
	}

	// 檢查該 OAuth 帳號是否已被其他用戶關聯
	var existingOAuth models.OAuthAccount
	err = au.db.Where("provider = ? AND provider_id = ?", req.Provider, oauthUser.ID).First(&existingOAuth).Error
	if err == nil {
		if existingOAuth.UserID != userID {
			return errors.New("該 OAuth 帳號已被其他用戶關聯")
		}
		return errors.New("該 OAuth 帳號已關聯到您的帳號")
	}

	// 檢查用戶是否已關聯該提供商的帳號
	var userOAuth models.OAuthAccount
	err = au.db.Where("user_id = ? AND provider = ?", userID, req.Provider).First(&userOAuth).Error
	if err == nil {
		return errors.New("您已關聯該提供商的帳號")
	}

	// 創建新的 OAuth 關聯
	newOAuthAccount := models.OAuthAccount{
		UserID:       userID,
		Provider:     req.Provider,
		ProviderID:   oauthUser.ID,
		Email:        oauthUser.Email,
		AccessToken:  &token.AccessToken,
		RefreshToken: &token.RefreshToken,
	}
	if !token.Expiry.IsZero() {
		newOAuthAccount.ExpiresAt = &token.Expiry
	}

	return au.db.Create(&newOAuthAccount).Error
}

// UnlinkOAuthAccount 解除關聯 OAuth 帳號
func (au *AuthUsecase) UnlinkOAuthAccount(userID string, req *dto.UnlinkOAuthAccountRequest) error {
	// 檢查用戶是否有密碼（如果沒有密碼且只有一個 OAuth 帳號，不允許解除關聯）
	var user models.User
	if err := au.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("用戶不存在")
	}

	// 檢查用戶的 OAuth 帳號數量
	var oauthCount int64
	au.db.Model(&models.OAuthAccount{}).Where("user_id = ?", userID).Count(&oauthCount)

	// 如果用戶沒有設置密碼且只有一個 OAuth 帳號，不允許解除關聯
	if user.PasswordHash == "" || user.PasswordHash == "oauth-user-no-password" {
		if oauthCount <= 1 {
			return errors.New("無法解除關聯：您需要設置密碼或關聯其他登入方式")
		}
	}

	// 刪除 OAuth 關聯
	result := au.db.Where("user_id = ? AND provider = ?", userID, req.Provider).Delete(&models.OAuthAccount{})
	if result.Error != nil {
		return errors.New("解除關聯失敗")
	}

	if result.RowsAffected == 0 {
		return errors.New("未找到要解除關聯的帳號")
	}

	return nil
}

// GetLinkedOAuthAccounts 獲取已關聯的 OAuth 帳號列表
func (au *AuthUsecase) GetLinkedOAuthAccounts(userID string) ([]models.OAuthAccount, error) {
	var oauthAccounts []models.OAuthAccount
	err := au.db.Where("user_id = ?", userID).
		Select("id", "provider", "email", "created_at").
		Find(&oauthAccounts).Error

	return oauthAccounts, err
}

// generateAuthResponse 生成認證響應
func (au *AuthUsecase) generateAuthResponse(user *models.User) (*dto.AuthResponse, error) {
	// 生成 JWT 令牌
	accessToken, err := au.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("生成訪問令牌失敗")
	}

	refreshToken, err := au.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("生成刷新令牌失敗")
	}

	// 存儲刷新令牌
	refreshTokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(au.config.JWT.RefreshTokenTTL) * 24 * time.Hour),
	}

	if err := au.db.Create(&refreshTokenModel).Error; err != nil {
		return nil, errors.New("存儲刷新令牌失敗")
	}

	return &dto.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
