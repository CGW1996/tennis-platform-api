# 數據庫架構文檔

## 概述

網球平台使用 PostgreSQL 作為主數據庫，Redis 作為緩存和會話存儲，並支援 PostGIS 擴展進行地理位置查詢。

## 數據庫架構

### 核心實體

1. **用戶系統 (Users)**
   - `users` - 用戶基本資訊
   - `user_profiles` - 用戶詳細檔案
   - `oauth_accounts` - OAuth 帳號關聯
   - `refresh_tokens` - JWT Refresh Token

2. **場地系統 (Courts)**
   - `courts` - 網球場地資訊
   - `court_reviews` - 場地評價
   - `bookings` - 場地預訂

3. **配對系統 (Matches)**
   - `matches` - 球友配對/比賽
   - `match_participants` - 比賽參與者
   - `match_results` - 比賽結果
   - `reputation_scores` - 用戶信譽分數

4. **聊天系統 (Chat)**
   - `chat_rooms` - 聊天室
   - `chat_messages` - 聊天訊息
   - `chat_participants` - 聊天室參與者

5. **教練系統 (Coaches)**
   - `coaches` - 教練資訊
   - `coach_reviews` - 教練評價
   - `lessons` - 課程
   - `lesson_schedules` - 課程時間表

6. **球拍系統 (Rackets)**
   - `rackets` - 球拍資訊
   - `racket_reviews` - 球拍評價
   - `racket_prices` - 球拍價格
   - `racket_recommendations` - 球拍推薦記錄

7. **俱樂部系統 (Clubs)**
   - `clubs` - 俱樂部資訊
   - `club_members` - 俱樂部會員
   - `club_events` - 俱樂部活動
   - `club_event_participants` - 活動參與者
   - `club_reviews` - 俱樂部評價

## 特殊功能

### PostGIS 地理位置支援

- 支援地理位置查詢和距離計算
- 使用 `ST_Point` 存儲經緯度座標
- 使用 `ST_Distance` 計算距離
- 創建 GIST 索引優化地理查詢性能

### 自動評分更新

使用 PostgreSQL 觸發器自動更新平均評分：
- 場地評分自動更新
- 教練評分自動更新
- 球拍評分自動更新
- 俱樂部評分自動更新

### 全文搜索

使用 PostgreSQL 的全文搜索功能：
- 場地名稱和地址搜索
- 俱樂部名稱搜索
- 球拍品牌和型號搜索

## 遷移系統

### 遷移管理

- 使用自定義遷移管理器
- 支援版本控制和回滾
- 自動執行數據庫架構更新

### 遷移文件

1. `001_initial_schema` - 初始數據庫架構
2. `002_add_indexes` - 添加性能索引
3. `003_add_constraints` - 添加約束和觸發器

## 種子數據

開發環境自動載入測試數據：
- 測試用戶和檔案
- 示例場地資訊
- 教練和課程數據
- 球拍和俱樂部資訊

## Redis 緩存

### 緩存策略

- 用戶會話管理
- 熱門場地緩存
- 搜尋結果緩存
- 即時聊天訊息

### 數據結構

- **String**: 簡單鍵值對緩存
- **Hash**: 用戶會話和檔案緩存
- **Set**: 用戶關係和標籤
- **Sorted Set**: 排行榜和推薦
- **List**: 訊息隊列
- **Pub/Sub**: 即時通知

## 性能優化

### 索引策略

- 主鍵和外鍵索引
- 地理位置 GIST 索引
- 全文搜索 GIN 索引
- 複合索引優化查詢
- 部分索引減少存儲

### 查詢優化

- 使用 EXPLAIN ANALYZE 分析查詢
- 避免 N+1 查詢問題
- 使用 GORM 預載入關聯
- 分頁查詢優化

## 安全考量

### 數據保護

- 密碼使用 bcrypt 加密
- 敏感數據欄位標記為 `json:"-"`
- 軟刪除保護數據完整性
- 外鍵約束保證參照完整性

### 存取控制

- 使用 GORM 鉤子驗證數據
- 數據庫層面的檢查約束
- 行級安全策略（未來實現）

## 監控和維護

### 健康檢查

- 數據庫連接狀態監控
- Redis 連接狀態監控
- 查詢性能監控

### 備份策略

- 定期數據庫備份
- 增量備份和恢復
- 災難恢復計劃

## 使用方法

### 初始化數據庫

```go
// 載入配置
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// 初始化數據庫
dbManager, err := db.Initialize(cfg)
if err != nil {
    log.Fatal(err)
}
defer dbManager.Close()
```

### 執行遷移

```go
// 手動執行遷移
migrationManager := db.NewMigrationManager(dbManager.DB.GetDB())
err := migrationManager.RunMigrations()
if err != nil {
    log.Fatal(err)
}
```

### 載入種子數據

```go
// 載入測試數據
seeder := db.NewSeeder(dbManager.DB.GetDB())
err := seeder.SeedAll()
if err != nil {
    log.Fatal(err)
}
```

## 環境配置

### 開發環境

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=tennis_platform
DB_USER=tennis_user
DB_PASSWORD=tennis_password
DB_SSL_MODE=disable

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 生產環境

```env
DB_HOST=your-postgres-host
DB_PORT=5432
DB_NAME=tennis_platform_prod
DB_USER=tennis_user
DB_PASSWORD=secure_password
DB_SSL_MODE=require

REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=secure_password
REDIS_DB=0
```

## 故障排除

### 常見問題

1. **連接失敗**: 檢查數據庫服務是否運行
2. **遷移失敗**: 檢查數據庫權限和擴展
3. **性能問題**: 檢查索引和查詢計劃
4. **數據不一致**: 檢查約束和觸發器

### 調試工具

- 使用 `EXPLAIN ANALYZE` 分析查詢
- 檢查 PostgreSQL 日誌
- 監控 Redis 記憶體使用
- 使用 GORM 日誌模式調試