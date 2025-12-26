package services

import (
	"errors"
	"fmt"
	"time"

	"tennis-platform/backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService JWT 服務
type JWTService struct {
	config *config.Config
}

// Claims JWT 聲明
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// NewJWTService 創建新的 JWT 服務
func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

// GenerateAccessToken 生成訪問令牌
func (j *JWTService) GenerateAccessToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.config.JWT.AccessTokenTTL) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tennis-platform",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.JWT.Secret))
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTService) GenerateRefreshToken(userID string) (string, error) {
	// 生成唯一的 JTI 來確保令牌唯一性
	jti := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())

	claims := jwt.RegisteredClaims{
		ID:        jti,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.config.JWT.RefreshTokenTTL) * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "tennis-platform",
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.JWT.Secret))
}

// ValidateToken 驗證令牌
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken 驗證刷新令牌
func (j *JWTService) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.config.JWT.Secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	return "", errors.New("invalid refresh token")
}
