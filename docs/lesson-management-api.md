# 課程管理系統 API 文檔

## 概述

課程管理系統提供完整的網球課程管理功能，包括課程類型設定、課程預訂、時間表管理和狀態追蹤。

## 功能特性

### 1. 課程類型管理
- 創建、更新、刪除課程類型
- 支援個人課程、團體課程和訓練營
- 價格設定和設備需求管理
- 參與人數限制（團體課程）

### 2. 課程預訂系統
- 學生預訂課程
- 時間衝突檢測
- 課程狀態管理（已預訂、進行中、已完成、已取消）
- 取消機制和原因記錄

### 3. 時間表管理
- 教練可用時間設定
- 每週時間表配置
- 實時可用性查詢
- 時間段生成和衝突檢查

### 4. 通知系統
- 課程預訂確認
- 時間變更通知
- 取消通知
- 狀態更新提醒

## API 端點

### 課程類型管理

#### 創建課程類型
```http
POST /api/v1/coaches/lesson-types
Authorization: Bearer {token}
Content-Type: application/json

{
    "name": "初級個人課程",
    "description": "適合初學者的一對一網球課程",
    "type": "individual",
    "level": "beginner",
    "duration": 60,
    "price": 1500,
    "currency": "TWD",
    "maxParticipants": null,
    "minParticipants": null,
    "equipment": ["網球拍", "網球"],
    "prerequisites": "無需經驗"
}
```

**響應:**
```json
{
    "id": "uuid",
    "coachId": "uuid",
    "name": "初級個人課程",
    "description": "適合初學者的一對一網球課程",
    "type": "individual",
    "level": "beginner",
    "duration": 60,
    "price": 1500,
    "currency": "TWD",
    "maxParticipants": null,
    "minParticipants": null,
    "equipment": ["網球拍", "網球"],
    "prerequisites": "無需經驗",
    "isActive": true,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### 獲取課程類型列表
```http
GET /api/v1/coaches/{coachId}/lesson-types
```

#### 更新課程類型
```http
PUT /api/v1/lesson-types/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
    "name": "初級個人課程（更新）",
    "price": 1600,
    "description": "更新後的課程描述"
}
```

#### 刪除課程類型
```http
DELETE /api/v1/lesson-types/{id}
Authorization: Bearer {token}
```

### 課程管理

#### 創建課程（預訂）
```http
POST /api/v1/lessons
Authorization: Bearer {token}
Content-Type: application/json

{
    "coachId": "uuid",
    "lessonTypeId": "uuid",
    "courtId": "uuid",
    "type": "individual",
    "level": "beginner",
    "duration": 60,
    "price": 1500,
    "currency": "TWD",
    "scheduledAt": "2024-12-01T10:00:00Z",
    "notes": "第一次課程"
}
```

#### 獲取課程詳情
```http
GET /api/v1/lessons/{id}
Authorization: Bearer {token}
```

#### 獲取課程列表
```http
GET /api/v1/lessons?coachId={coachId}&studentId={studentId}&status={status}&startDate={date}&endDate={date}&page=1&limit=20
Authorization: Bearer {token}
```

**查詢參數:**
- `coachId`: 教練ID（可選）
- `studentId`: 學生ID（可選）
- `status`: 課程狀態（可選）- scheduled, in_progress, completed, cancelled
- `startDate`: 開始日期（可選）- YYYY-MM-DD
- `endDate`: 結束日期（可選）- YYYY-MM-DD
- `page`: 頁碼（可選，默認1）
- `limit`: 每頁數量（可選，默認20）

#### 更新課程
```http
PUT /api/v1/lessons/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
    "courtId": "uuid",
    "scheduledAt": "2024-12-01T11:00:00Z",
    "notes": "時間調整",
    "status": "scheduled"
}
```

#### 取消課程
```http
POST /api/v1/lessons/{id}/cancel
Authorization: Bearer {token}
Content-Type: application/json

{
    "reason": "學生臨時有事"
}
```

### 時間表管理

#### 獲取教練可用時間
```http
GET /api/v1/coaches/{coachId}/availability?date=2024-12-01
```

**響應:**
```json
{
    "date": "2024-12-01",
    "availability": [
        {
            "startTime": "09:00",
            "endTime": "10:00",
            "isBooked": false
        },
        {
            "startTime": "10:00",
            "endTime": "11:00",
            "isBooked": true
        }
    ]
}
```

#### 更新教練時間表
```http
PUT /api/v1/coaches/schedule
Authorization: Bearer {token}
Content-Type: application/json

{
    "schedules": [
        {
            "dayOfWeek": 1,
            "startTime": "09:00",
            "endTime": "17:00",
            "isActive": true
        },
        {
            "dayOfWeek": 2,
            "startTime": "10:00",
            "endTime": "18:00",
            "isActive": true
        }
    ]
}
```

**說明:**
- `dayOfWeek`: 星期幾（0=星期日，1=星期一，...，6=星期六）
- `startTime`: 開始時間（HH:MM格式）
- `endTime`: 結束時間（HH:MM格式）
- `isActive`: 是否啟用

#### 獲取教練時間表
```http
GET /api/v1/coaches/{coachId}/schedule
```

## 數據模型

### LessonType（課程類型）
```go
type LessonType struct {
    ID               string    `json:"id"`
    CoachID          string    `json:"coachId"`
    Name             string    `json:"name"`
    Description      *string   `json:"description"`
    Type             string    `json:"type"`             // individual, group, clinic
    Level            string    `json:"level"`            // beginner, intermediate, advanced
    Duration         int       `json:"duration"`         // 分鐘
    Price            float64   `json:"price"`
    Currency         string    `json:"currency"`
    MaxParticipants  *int      `json:"maxParticipants"`  // 最大參與人數
    MinParticipants  *int      `json:"minParticipants"`  // 最小參與人數
    Equipment        []string  `json:"equipment"`        // 需要的設備
    Prerequisites    *string   `json:"prerequisites"`    // 先決條件
    IsActive         bool      `json:"isActive"`
    CreatedAt        time.Time `json:"createdAt"`
    UpdatedAt        time.Time `json:"updatedAt"`
}
```

### Lesson（課程）
```go
type Lesson struct {
    ID           string    `json:"id"`
    CoachID      string    `json:"coachId"`
    StudentID    string    `json:"studentId"`
    LessonTypeID *string   `json:"lessonTypeId"`
    CourtID      *string   `json:"courtId"`
    Type         string    `json:"type"`         // individual, group, clinic
    Level        string    `json:"level"`        // beginner, intermediate, advanced
    Duration     int       `json:"duration"`     // 分鐘
    Price        float64   `json:"price"`
    Currency     string    `json:"currency"`
    ScheduledAt  time.Time `json:"scheduledAt"`
    Status       string    `json:"status"`       // scheduled, in_progress, completed, cancelled
    Notes        *string   `json:"notes"`
    PaymentID    *string   `json:"paymentId"`
    CancelReason *string   `json:"cancelReason"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}
```

### LessonSchedule（課程時間表）
```go
type LessonSchedule struct {
    ID        string    `json:"id"`
    CoachID   string    `json:"coachId"`
    DayOfWeek int       `json:"dayOfWeek"`    // 0=Sunday, 6=Saturday
    StartTime string    `json:"startTime"`    // "09:00"
    EndTime   string    `json:"endTime"`      // "17:00"
    IsActive  bool      `json:"isActive"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

### TimeSlot（時間段）
```go
type TimeSlot struct {
    StartTime string `json:"startTime"`
    EndTime   string `json:"endTime"`
    IsBooked  bool   `json:"isBooked"`
}
```

## 業務邏輯

### 課程類型驗證
1. **個人課程**: 不需要設定參與人數限制
2. **團體課程**: 必須設定最大參與人數（至少2人）
3. **訓練營**: 支援大型團體，可設定最小和最大參與人數

### 時間衝突檢測
1. 檢查教練在指定時間是否已有其他課程
2. 考慮課程持續時間，確保沒有時間重疊
3. 支援排除特定課程的衝突檢查（用於更新課程時間）

### 可用時間計算
1. 根據教練的週時間表生成基礎時間段
2. 查詢當天已預訂的課程
3. 標記已被預訂的時間段
4. 返回完整的可用性信息

### 課程狀態管理
- **scheduled**: 已預訂，等待開始
- **in_progress**: 課程進行中
- **completed**: 課程已完成
- **cancelled**: 課程已取消

### 取消政策
1. 只有已預訂或進行中的課程可以取消
2. 已完成或已取消的課程無法修改
3. 取消時必須提供原因
4. 系統記錄取消時間和原因

## 錯誤處理

### 常見錯誤碼
- `400 Bad Request`: 請求參數錯誤或業務邏輯錯誤
- `401 Unauthorized`: 未提供有效的認證令牌
- `403 Forbidden`: 無權限執行操作
- `404 Not Found`: 資源不存在
- `409 Conflict`: 時間衝突或狀態衝突

### 錯誤響應格式
```json
{
    "error": "錯誤描述",
    "details": "詳細錯誤信息（可選）"
}
```

## 安全考慮

### 權限控制
1. **課程類型管理**: 只有教練本人可以管理自己的課程類型
2. **課程預訂**: 學生可以預訂課程，教練和學生可以查看相關課程
3. **時間表管理**: 只有教練本人可以管理自己的時間表
4. **課程修改**: 只有相關的教練或學生可以修改課程

### 數據驗證
1. 所有輸入數據都經過嚴格驗證
2. 時間格式驗證和邏輯檢查
3. 價格和參與人數的合理性檢查
4. 防止SQL注入和XSS攻擊

## 性能優化

### 數據庫索引
- 課程類型按教練ID和狀態索引
- 課程按教練ID、學生ID和時間索引
- 時間表按教練ID和星期索引

### 緩存策略
- 教練時間表可以緩存（更新頻率較低）
- 可用時間查詢結果可以短期緩存
- 課程類型信息可以緩存

### 查詢優化
- 使用複合索引提升查詢性能
- 分頁查詢避免大量數據傳輸
- 預載入關聯數據減少N+1查詢問題

## 測試

### 單元測試
- 業務邏輯驗證
- 時間衝突檢測
- 數據驗證規則

### 整合測試
- API端點測試
- 數據庫操作測試
- 權限控制測試

### 端到端測試
- 完整的課程預訂流程
- 時間表管理流程
- 錯誤處理場景

使用提供的測試腳本 `test_lesson_api.sh` 可以快速驗證所有API端點的功能。