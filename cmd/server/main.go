package main

import (
	"log"
	"tennis-platform/backend/internal/api"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/db"
)

// @title 網球平台 API
// @version 1.0
// @description 網球平台後端 API 服務
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化數據庫
	dbManager, err := db.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer dbManager.Close()

	// 初始化 API 服務器
	server := api.NewServer(cfg, dbManager.DB)

	// 啟動服務器
	log.Printf("Starting server on port %s", cfg.Port)
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
