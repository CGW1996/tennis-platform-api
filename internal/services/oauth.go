package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"tennis-platform/backend/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthService OAuth 服務
type OAuthService struct {
	config *config.Config
}

// OAuthUserInfo OAuth 用戶資訊
type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
	Provider      string `json:"provider"`
}

// GoogleUserInfo Google 用戶資訊結構
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

// FacebookUserInfo Facebook 用戶資訊結構
type FacebookUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Picture   struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// AppleUserInfo Apple 用戶資訊結構
type AppleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
}

// NewOAuthService 創建新的 OAuth 服務
func NewOAuthService(cfg *config.Config) *OAuthService {
	return &OAuthService{
		config: cfg,
	}
}

// GetGoogleOAuthConfig 獲取 Google OAuth 配置
func (o *OAuthService) GetGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     o.config.OAuth.Google.ClientID,
		ClientSecret: o.config.OAuth.Google.ClientSecret,
		RedirectURL:  o.config.OAuth.Google.RedirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
}

// GetFacebookOAuthConfig 獲取 Facebook OAuth 配置
func (o *OAuthService) GetFacebookOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     o.config.OAuth.Facebook.ClientID,
		ClientSecret: o.config.OAuth.Facebook.ClientSecret,
		RedirectURL:  o.config.OAuth.Facebook.RedirectURL,
		Scopes:       []string{"email", "public_profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/v18.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v18.0/oauth/access_token",
		},
	}
}

// GetAppleOAuthConfig 獲取 Apple OAuth 配置
func (o *OAuthService) GetAppleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     o.config.OAuth.Apple.ClientID,
		ClientSecret: o.config.OAuth.Apple.ClientSecret,
		RedirectURL:  o.config.OAuth.Apple.RedirectURL,
		Scopes:       []string{"name", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
		},
	}
}

// GetAuthURL 獲取授權 URL
func (o *OAuthService) GetAuthURL(provider, state string) (string, error) {
	var config *oauth2.Config

	switch provider {
	case "google":
		config = o.GetGoogleOAuthConfig()
	case "facebook":
		config = o.GetFacebookOAuthConfig()
	case "apple":
		config = o.GetAppleOAuthConfig()
	default:
		return "", errors.New("不支援的 OAuth 提供商")
	}

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// ExchangeCodeForToken 交換授權碼獲取令牌
func (o *OAuthService) ExchangeCodeForToken(provider, code string) (*oauth2.Token, error) {
	var config *oauth2.Config

	switch provider {
	case "google":
		config = o.GetGoogleOAuthConfig()
	case "facebook":
		config = o.GetFacebookOAuthConfig()
	case "apple":
		config = o.GetAppleOAuthConfig()
	default:
		return nil, errors.New("不支援的 OAuth 提供商")
	}

	return config.Exchange(context.Background(), code)
}

// GetUserInfo 獲取用戶資訊
func (o *OAuthService) GetUserInfo(provider string, token *oauth2.Token) (*OAuthUserInfo, error) {
	switch provider {
	case "google":
		return o.getGoogleUserInfo(token)
	case "facebook":
		return o.getFacebookUserInfo(token)
	case "apple":
		return o.getAppleUserInfo(token)
	default:
		return nil, errors.New("不支援的 OAuth 提供商")
	}
}

// getGoogleUserInfo 獲取 Google 用戶資訊
func (o *OAuthService) getGoogleUserInfo(token *oauth2.Token) (*OAuthUserInfo, error) {
	client := o.GetGoogleOAuthConfig().Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("獲取 Google 用戶資訊失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取 Google 用戶資訊失敗: %v", err)
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("解析 Google 用戶資訊失敗: %v", err)
	}

	return &OAuthUserInfo{
		ID:            googleUser.ID,
		Email:         googleUser.Email,
		Name:          googleUser.Name,
		FirstName:     googleUser.GivenName,
		LastName:      googleUser.FamilyName,
		Picture:       googleUser.Picture,
		EmailVerified: googleUser.EmailVerified,
		Provider:      "google",
	}, nil
}

// getFacebookUserInfo 獲取 Facebook 用戶資訊
func (o *OAuthService) getFacebookUserInfo(token *oauth2.Token) (*OAuthUserInfo, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://graph.facebook.com/me?fields=id,email,name,first_name,last_name,picture&access_token=%s", token.AccessToken)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("獲取 Facebook 用戶資訊失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取 Facebook 用戶資訊失敗: %v", err)
	}

	var facebookUser FacebookUserInfo
	if err := json.Unmarshal(body, &facebookUser); err != nil {
		return nil, fmt.Errorf("解析 Facebook 用戶資訊失敗: %v", err)
	}

	return &OAuthUserInfo{
		ID:            facebookUser.ID,
		Email:         facebookUser.Email,
		Name:          facebookUser.Name,
		FirstName:     facebookUser.FirstName,
		LastName:      facebookUser.LastName,
		Picture:       facebookUser.Picture.Data.URL,
		EmailVerified: true, // Facebook 不提供此資訊，假設已驗證
		Provider:      "facebook",
	}, nil
}

// getAppleUserInfo 獲取 Apple 用戶資訊
func (o *OAuthService) getAppleUserInfo(token *oauth2.Token) (*OAuthUserInfo, error) {
	// Apple 的用戶資訊通常在 ID Token 中，需要解析 JWT
	// 這裡簡化處理，實際應用中需要驗證和解析 ID Token

	// 注意：Apple 只在首次授權時提供用戶資訊，後續需要從 ID Token 中獲取
	return &OAuthUserInfo{
		ID:            token.Extra("sub").(string),
		Email:         token.Extra("email").(string),
		EmailVerified: token.Extra("email_verified").(bool),
		Provider:      "apple",
	}, nil
}

// GenerateState 生成 OAuth 狀態參數
func (o *OAuthService) GenerateState() string {
	// 在實際應用中，應該生成隨機的狀態參數並存儲在會話中
	return "random-state-string"
}

// ValidateState 驗證 OAuth 狀態參數
func (o *OAuthService) ValidateState(state string) bool {
	// 在實際應用中，應該驗證狀態參數是否與會話中存儲的一致
	return state == "random-state-string"
}
