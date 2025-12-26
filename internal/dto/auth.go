package dto

import "tennis-platform/backend/internal/models"

// RegisterRequest 註冊請求
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

// LoginRequest 登入請求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 認證響應
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
}

// RefreshTokenRequest 刷新令牌請求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ForgotPasswordRequest 忘記密碼請求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重設密碼請求
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// OAuthLoginRequest OAuth 登入請求
type OAuthLoginRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google facebook apple"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

// LinkOAuthAccountRequest 關聯 OAuth 帳號請求
type LinkOAuthAccountRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google facebook apple"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
}

// UnlinkOAuthAccountRequest 解除關聯 OAuth 帳號請求
type UnlinkOAuthAccountRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google facebook apple"`
}
