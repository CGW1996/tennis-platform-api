package middleware

import (
	"net/http"
	"strings"
	"tennis-platform/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 認證中間件
func AuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 Authorization header 獲取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少認證令牌",
			})
			c.Abort()
			return
		}

		// 檢查 Bearer 前綴
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "無效的認證令牌格式",
			})
			c.Abort()
			return
		}

		// 驗證令牌
		claims, err := jwtService.ValidateToken(tokenParts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "無效的認證令牌",
			})
			c.Abort()
			return
		}

		// 將用戶信息存儲在上下文中
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// OptionalAuthMiddleware 可選的認證中間件
func OptionalAuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := jwtService.ValidateToken(tokenParts[1])
		if err != nil {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
