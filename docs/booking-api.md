# 場地預訂 API 文檔

## 概述

場地預訂 API 提供完整的網球場地預訂功能，包括創建預訂、查詢可用時間、管理預訂狀態等。

## 基本信息

- **Base URL**: `/api/v1`
- **認證方式**: Bearer Token (JWT)
- **內容類型**: `application/json`

## API 端點

### 1. 創建預訂

創建新的場地預訂。

**端點**: `POST /bookings`

**請求頭**:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**請求體**:
```json
{
  "courtId": "uuid",
  "startTime": "2024-01-15T10:00:00Z",
  "endTime": "2024-01-15T12:00:00Z",
  "notes": "與朋友練習"
}
```

**請求參數說明**:
- `courtId` (string, required): 場地ID
- `startTime` (string, required): 預訂開始時間 (ISO 8601格式)
- `endTime` (string, required): 預訂結束時間 (ISO 8601格式)
- `notes` (string, optional): 預訂備註，最多500字符

**成功回應** (201 Created):
```json
{
  "id": "booking-uuid",
  "courtId": "court-uuid",
  "userId": "user-uuid",
  "startTime": "2024-01-15T10:00:00Z",
  "endTime": "2024-01-15T12:00:00Z",
  "totalPrice": 200.0,
  "status": "pending",
  "notes": "與朋友練習",
  "createdAt": "2024-01-10T08:00:00Z",
  "updatedAt": "2024-01-10T08:00:00Z",
  "court": {
    "id": "court-uuid",
    "name": "中央球場",
    "address": "台北市信義區信義路五段7號",
    "pricePerHour": 100.0,
    "currency": "TWD"
  },
  "user": {
    "id": "user-uuid",
    "email": "user@example.com"
  }
}
```

**錯誤回應**:
- `400 Bad Request`: 請求參數錯誤
- `401 Unauthorized`: 未認證
- `409 Conflict`: 時間衝突

### 2. 獲取預訂詳情

根據預訂ID獲取預訂的詳細信息。

**端點**: `GET /bookings/{id}`

**請求頭**:
```
Authorization: Bearer <access_token>
```

**路徑參數**:
- `id` (string, required): 預訂ID

**成功回應** (200 OK):
```json
{
  "id": "booking-uuid",
  "courtId": "court-uuid",
  "userId": "user-uuid",
  "startTime": "2024-01-15T10:00:00Z",
  "endTime": "2024-01-15T12:00:00Z",
  "totalPrice": 200.0,
  "status": "confirmed",
  "notes": "與朋友練習",
  "createdAt": "2024-01-10T08:00:00Z",
  "updatedAt": "2024-01-10T08:00:00Z",
  "court": {
    "id": "court-uuid",
    "name": "中央球場",
    "address": "台北市信義區信義路五段7號"
  },
  "user": {
    "id": "user-uuid",
    "email": "user@example.com"
  }
}
```

**錯誤回應**:
- `401 Unauthorized`: 未認證
- `404 Not Found`: 預訂不存在

### 3. 更新預訂

更新現有預訂的信息。

**端點**: `PUT /bookings/{id}`

**請求頭**:
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**路徑參數**:
- `id` (string, required): 預訂ID

**請求體**:
```json
{
  "startTime": "2024-01-15T11:00:00Z",
  "endTime": "2024-01-15T13:00:00Z",
  "notes": "更新後的備註",
  "status": "confirmed"
}
```

**請求參數說明**:
- `startTime` (string, optional): 新的開始時間
- `endTime` (string, optional): 新的結束時間
- `notes` (string, optional): 更新備註
- `status` (string, optional): 預訂狀態 (pending, confirmed, cancelled, completed)

**成功回應** (200 OK):
```json
{
  "id": "booking-uuid",
  "courtId": "court-uuid",
  "userId": "user-uuid",
  "startTime": "2024-01-15T11:00:00Z",
  "endTime": "2024-01-15T13:00:00Z",
  "totalPrice": 200.0,
  "status": "confirmed",
  "notes": "更新後的備註",
  "updatedAt": "2024-01-10T09:00:00Z"
}
```

**錯誤回應**:
- `400 Bad Request`: 請求參數錯誤或業務邏輯錯誤
- `401 Unauthorized`: 未認證
- `403 Forbidden`: 無權限修改此預訂
- `404 Not Found`: 預訂不存在

### 4. 取消預訂

取消指定的預訂。

**端點**: `POST /bookings/{id}/cancel`

**請求頭**:
```
Authorization: Bearer <access_token>
```

**路徑參數**:
- `id` (string, required): 預訂ID

**成功回應** (200 OK):
```json
{
  "message": "預訂取消成功"
}
```

**錯誤回應**:
- `400 Bad Request`: 無法取消（如時間限制）
- `401 Unauthorized`: 未認證
- `403 Forbidden`: 無權限取消此預訂
- `404 Not Found`: 預訂不存在

### 5. 獲取預訂列表

根據條件獲取預訂列表。

**端點**: `GET /bookings`

**請求頭**:
```
Authorization: Bearer <access_token>
```

**查詢參數**:
- `courtId` (string, optional): 場地ID篩選
- `userId` (string, optional): 用戶ID篩選
- `status` (string, optional): 預訂狀態篩選 (pending, confirmed, cancelled, completed)
- `startDate` (string, optional): 開始日期篩選 (YYYY-MM-DD)
- `endDate` (string, optional): 結束日期篩選 (YYYY-MM-DD)
- `page` (integer, optional): 頁碼，默認1
- `pageSize` (integer, optional): 每頁數量，默認20，最大100

**成功回應** (200 OK):
```json
{
  "bookings": [
    {
      "id": "booking-uuid-1",
      "courtId": "court-uuid",
      "userId": "user-uuid",
      "startTime": "2024-01-15T10:00:00Z",
      "endTime": "2024-01-15T12:00:00Z",
      "totalPrice": 200.0,
      "status": "confirmed",
      "court": {
        "id": "court-uuid",
        "name": "中央球場"
      },
      "user": {
        "id": "user-uuid",
        "email": "user@example.com"
      }
    }
  ],
  "total": 25,
  "page": 1,
  "pageSize": 20,
  "totalPages": 2
}
```

### 6. 查詢場地可用時間

查詢指定場地在指定日期的可用時間段。

**端點**: `GET /courts/availability`

**查詢參數**:
- `courtId` (string, required): 場地ID
- `date` (string, required): 查詢日期 (YYYY-MM-DD)
- `duration` (integer, optional): 預訂時長（分鐘），默認60分鐘

**成功回應** (200 OK):
```json
{
  "date": "2024-01-15",
  "courtId": "court-uuid",
  "timeSlots": [
    {
      "startTime": "2024-01-15T09:00:00Z",
      "endTime": "2024-01-15T10:00:00Z",
      "available": true,
      "price": 100.0
    },
    {
      "startTime": "2024-01-15T10:00:00Z",
      "endTime": "2024-01-15T11:00:00Z",
      "available": false,
      "price": 100.0
    },
    {
      "startTime": "2024-01-15T11:00:00Z",
      "endTime": "2024-01-15T12:00:00Z",
      "available": true,
      "price": 100.0
    }
  ]
}
```

## 預訂狀態說明

- `pending`: 待確認 - 剛創建的預訂，等待確認
- `confirmed`: 已確認 - 預訂已確認，可以使用
- `cancelled`: 已取消 - 預訂已被取消
- `completed`: 已完成 - 預訂時間已過，使用完成

## 業務規則

### 預訂時間限制
- 預訂時長最少30分鐘，最多8小時
- 不能預訂過去的時間
- 不能預訂30天後的時間
- 預訂開始前2小時內無法取消

### 時間衝突檢測
- 系統會自動檢測時間衝突
- 同一場地同一時間段只能有一個有效預訂
- 狀態為 `pending` 或 `confirmed` 的預訂會被視為有效預訂

### 營業時間檢查
- 預訂時間必須在場地營業時間內
- 如果場地在指定日期關閉，無法創建預訂

### 價格計算
- 總價格 = 預訂時長（小時）× 場地每小時價格
- 價格會根據時間變更自動重新計算

## 錯誤處理

所有API都遵循統一的錯誤格式：

```json
{
  "error": "錯誤描述",
  "details": "詳細錯誤信息（可選）"
}
```

常見錯誤碼：
- `400`: 請求參數錯誤或業務邏輯錯誤
- `401`: 未認證或認證失效
- `403`: 權限不足
- `404`: 資源不存在
- `409`: 資源衝突（如時間衝突）
- `500`: 服務器內部錯誤

## 使用示例

### 創建預訂流程

1. **查詢可用時間**:
```bash
curl -X GET "http://localhost:8080/api/v1/courts/availability?courtId=court-uuid&date=2024-01-15&duration=120"
```

2. **創建預訂**:
```bash
curl -X POST "http://localhost:8080/api/v1/bookings" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "courtId": "court-uuid",
    "startTime": "2024-01-15T10:00:00Z",
    "endTime": "2024-01-15T12:00:00Z",
    "notes": "週末練習"
  }'
```

3. **確認預訂**:
```bash
curl -X PUT "http://localhost:8080/api/v1/bookings/booking-uuid" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "confirmed"
  }'
```

### 管理預訂流程

1. **查看我的預訂**:
```bash
curl -X GET "http://localhost:8080/api/v1/bookings?userId=my-user-id&status=confirmed" \
  -H "Authorization: Bearer <access_token>"
```

2. **取消預訂**:
```bash
curl -X POST "http://localhost:8080/api/v1/bookings/booking-uuid/cancel" \
  -H "Authorization: Bearer <access_token>"
```

## 通知和提醒

預訂系統支持以下通知功能（需要配合通知服務實現）：

- **預訂確認通知**: 預訂創建成功後發送
- **預訂提醒**: 預訂開始前1小時發送提醒
- **取消通知**: 預訂被取消時發送通知
- **狀態變更通知**: 預訂狀態變更時發送通知

## 注意事項

1. **時區處理**: 所有時間都使用UTC時間，前端需要根據用戶時區進行轉換
2. **併發處理**: 系統使用數據庫事務確保併發安全
3. **數據一致性**: 預訂相關的所有操作都會保持數據一致性
4. **性能優化**: 可用時間查詢會緩存場地營業時間信息
5. **安全性**: 用戶只能操作自己的預訂，系統會進行權限檢查