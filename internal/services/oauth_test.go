package services

import (
	"tennis-platform/backend/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthService_GetAuthURL(t *testing.T) {
	cfg := &config.Config{
		OAuth: config.OAuthConfig{
			Google: config.OAuthProviderConfig{
				ClientID:     "test-google-client-id",
				ClientSecret: "test-google-client-secret",
				RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/google/callback",
			},
			Facebook: config.OAuthProviderConfig{
				ClientID:     "test-facebook-client-id",
				ClientSecret: "test-facebook-client-secret",
				RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/facebook/callback",
			},
			Apple: config.OAuthProviderConfig{
				ClientID:     "test-apple-client-id",
				ClientSecret: "test-apple-client-secret",
				RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/apple/callback",
			},
		},
	}

	oauthService := NewOAuthService(cfg)

	t.Run("Google OAuth URL", func(t *testing.T) {
		authURL, err := oauthService.GetAuthURL("google", "test-state")
		assert.NoError(t, err)
		assert.Contains(t, authURL, "accounts.google.com")
		assert.Contains(t, authURL, "test-google-client-id")
		assert.Contains(t, authURL, "test-state")
	})

	t.Run("Facebook OAuth URL", func(t *testing.T) {
		authURL, err := oauthService.GetAuthURL("facebook", "test-state")
		assert.NoError(t, err)
		assert.Contains(t, authURL, "facebook.com")
		assert.Contains(t, authURL, "test-facebook-client-id")
		assert.Contains(t, authURL, "test-state")
	})

	t.Run("Apple OAuth URL", func(t *testing.T) {
		authURL, err := oauthService.GetAuthURL("apple", "test-state")
		assert.NoError(t, err)
		assert.Contains(t, authURL, "appleid.apple.com")
		assert.Contains(t, authURL, "test-apple-client-id")
		assert.Contains(t, authURL, "test-state")
	})

	t.Run("Unsupported Provider", func(t *testing.T) {
		_, err := oauthService.GetAuthURL("unsupported", "test-state")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不支援的 OAuth 提供商")
	})
}

func TestOAuthService_ValidateState(t *testing.T) {
	cfg := &config.Config{}
	oauthService := NewOAuthService(cfg)

	t.Run("Valid State", func(t *testing.T) {
		// 使用服務生成的狀態
		state := oauthService.GenerateState()
		isValid := oauthService.ValidateState(state)
		assert.True(t, isValid)
	})

	t.Run("Invalid State", func(t *testing.T) {
		isValid := oauthService.ValidateState("invalid-state")
		assert.False(t, isValid)
	})
}
