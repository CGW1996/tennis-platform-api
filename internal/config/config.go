package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 應用程式配置
type Config struct {
	// 服務器配置
	Port        string
	Env         string
	FrontendURL string

	// 數據庫配置
	Database DatabaseConfig

	// Redis 配置
	Redis RedisConfig

	// JWT 配置
	JWT JWTConfig

	// OAuth 配置
	OAuth OAuthConfig

	// 文件上傳配置
	Upload UploadConfig
}

// DatabaseConfig 數據庫配置
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  int // 分鐘
	RefreshTokenTTL int // 天
}

// OAuthConfig OAuth 配置
type OAuthConfig struct {
	Google   OAuthProviderConfig
	Facebook OAuthProviderConfig
	Apple    OAuthProviderConfig
}

// OAuthProviderConfig OAuth 提供商配置
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// UploadConfig 文件上傳配置
type UploadConfig struct {
	MaxFileSize int64  // 字節
	AllowedExts string // 允許的文件擴展名，逗號分隔
	UploadPath  string // 上傳路徑
}

// Load 載入配置
func Load() (*Config, error) {
	// 載入 .env 文件（如果存在）
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Env:         getEnv("ENV", "development"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5433),
			Name:     getEnv("DB_NAME", "tennis_platform"),
			User:     getEnv("DB_USER", "tennis_user"),
			Password: getEnv("DB_PASSWORD", "tennis_password"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6380),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},

		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-jwt-secret-key"),
			AccessTokenTTL:  getEnvAsInt("JWT_ACCESS_TTL", 15), // 15 分鐘
			RefreshTokenTTL: getEnvAsInt("JWT_REFRESH_TTL", 7), // 7 天
		},

		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/oauth/google/callback"),
			},
			Facebook: OAuthProviderConfig{
				ClientID:     getEnv("FACEBOOK_CLIENT_ID", ""),
				ClientSecret: getEnv("FACEBOOK_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("FACEBOOK_REDIRECT_URL", "http://localhost:8080/api/v1/auth/oauth/facebook/callback"),
			},
			Apple: OAuthProviderConfig{
				ClientID:     getEnv("APPLE_CLIENT_ID", ""),
				ClientSecret: getEnv("APPLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("APPLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/oauth/apple/callback"),
			},
		},

		Upload: UploadConfig{
			MaxFileSize: getEnvAsInt64("UPLOAD_MAX_SIZE", 10*1024*1024), // 10MB
			AllowedExts: getEnv("UPLOAD_ALLOWED_EXTS", "jpg,jpeg,png,gif,pdf"),
			UploadPath:  getEnv("UPLOAD_PATH", "./uploads"),
		},
	}

	return cfg, nil
}

// getEnv 獲取環境變量，如果不存在則返回默認值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 獲取環境變量並轉換為整數
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 獲取環境變量並轉換為 int64
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
