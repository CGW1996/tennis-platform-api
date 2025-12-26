# 網球平台後端 API

基於 Go 和 Gin 框架的 RESTful API 服務。

## 專案結構

```
backend/
├── cmd/                    # 應用程式入口點
│   ├── server/            # API 服務器
│   ├── migrate/           # 數據庫遷移工具
│   └── seed/              # 數據種子工具
├── internal/              # 私有應用程式代碼
│   ├── api/               # API 處理器和路由
│   ├── auth/              # 認證和授權
│   ├── config/            # 配置管理
│   ├── db/                # 數據庫連接和遷移
│   ├── middleware/        # HTTP 中間件
│   ├── models/            # 數據模型
│   ├── repository/        # 數據訪問層
│   ├── service/           # 業務邏輯層
│   └── utils/             # 工具函數
├── pkg/                   # 可重用的庫代碼
│   ├── logger/            # 日誌工具
│   ├── validator/         # 驗證工具
│   └── response/          # API 響應格式
├── migrations/            # 數據庫遷移文件
├── docs/                  # API 文檔
├── Dockerfile             # 生產環境 Docker 文件
├── Dockerfile.dev         # 開發環境 Docker 文件
├── go.mod                 # Go 模組定義
├── go.sum                 # Go 模組校驗和
└── README.md
```

## 開發指南

### 本地開發

1. 安裝依賴：
```bash
go mod download
```

2. 啟動開發服務器：
```bash
go run cmd/server/main.go
```

3. 運行測試：
```bash
go test -v ./...
```

### API 文檔

API 文檔使用 Swagger 生成，啟動服務器後訪問：
- Swagger UI: http://localhost:8080/swagger/index.html

### 環境變量

| 變量名 | 描述 | 默認值 |
|--------|------|--------|
| PORT | 服務器端口 | 8080 |
| DB_HOST | 數據庫主機 | localhost |
| DB_PORT | 數據庫端口 | 5432 |
| DB_NAME | 數據庫名稱 | tennis_platform |
| DB_USER | 數據庫用戶 | tennis_user |
| DB_PASSWORD | 數據庫密碼 | tennis_password |
| REDIS_HOST | Redis 主機 | localhost |
| REDIS_PORT | Redis 端口 | 6379 |
| ELASTICSEARCH_URL | Elasticsearch URL | http://localhost:9200 |
| JWT_SECRET | JWT 密鑰 | your-jwt-secret-key |

### 代碼規範

- 使用 `gofmt` 格式化代碼
- 使用 `golangci-lint` 進行代碼檢查
- 遵循 Go 官方編碼規範
- 為公共函數和結構體添加註釋