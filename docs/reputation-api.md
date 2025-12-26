# 信譽評分系統 API 文檔

## 概述

信譽評分系統用於追蹤和評估用戶在網球平台上的行為表現，包括出席率、準時度、技術等級準確度和行為評分。

## API 端點

### 1. 獲取用戶信譽分數

**GET** `/api/v1/reputation/users/{userId}/score`

獲取指定用戶的信譽分數和統計信息。

#### 路徑參數
- `userId` (string, required): 用戶ID

#### 響應示例
```json
{
  "id": "uuid",
  "userId": "user-uuid",
  "attendanceRate": 95.5,
  "punctualityScore": 88.2,
  "skillAccuracy": 92.0,
  "behaviorRating": 4.3,
  "totalMatches": 25,
  "completedMatches": 24,
  "cancelledMatches": 1,
  "overallScore": 91.8,
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

### 2. 獲取用戶信譽歷史記錄

**GET** `/api/v1/reputation/users/{userId}/history`

獲取指定用戶的詳細信譽歷史記錄。

#### 路徑參數
- `userId` (string, required): 用戶ID

#### 認證
需要 JWT Token

#### 響應示例
```json
{
  "userId": "user-uuid",
  "punctualityRecords": [
    {
      "id": "record-uuid",
      "userId": "user-uuid",
      "matchId": "match-uuid",
      "isOnTime": true,
      "delayMinutes": 0,
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ],
  "skillRecords": [
    {
      "id": "record-uuid",
      "userId": "user-uuid",
      "matchId": "match-uuid",
      "reportedLevel": 4.0,
      "actualLevel": 3.8,
      "accuracy": 95.0,
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ],
  "behaviorReviews": [
    {
      "id": "review-uuid",
      "userId": "user-uuid",
      "reviewerId": "reviewer-uuid",
      "matchId": "match-uuid",
      "rating": 4.5,
      "comment": "很好的球友，技術不錯且態度友善",
      "tags": ["friendly", "skilled", "punctual"],
      "createdAt": "2024-01-15T10:00:00Z"
    }
  ]
}
```

### 3. 記錄比賽出席情況

**POST** `/api/v1/reputation/users/{userId}/attendance`

記錄用戶的比賽出席情況，影響出席率評分。

#### 路徑參數
- `userId` (string, required): 用戶ID

#### 認證
需要 JWT Token

#### 請求體
```json
{
  "matchId": "match-uuid",
  "status": "completed"
}
```

#### 請求參數
- `matchId` (string, required): 比賽ID
- `status` (string, required): 出席狀態，可選值：`completed`, `cancelled`, `no_show`

#### 響應示例
```json
{
  "message": "出席記錄已更新"
}
```

### 4. 記錄比賽準時情況

**POST** `/api/v1/reputation/users/{userId}/punctuality`

記錄用戶的比賽準時情況，影響準時度評分。

#### 路徑參數
- `userId` (string, required): 用戶ID

#### 認證
需要 JWT Token

#### 請求體
```json
{
  "matchId": "match-uuid",
  "arrivalTime": "2024-01-15T10:05:00Z"
}
```

#### 請求參數
- `matchId` (string, required): 比賽ID
- `arrivalTime` (string, required): 到達時間 (ISO 8601 格式)

#### 響應示例
```json
{
  "message": "準時記錄已更新"
}
```

### 5. 記錄技術等級準確度

**POST** `/api/v1/reputation/users/{userId}/skill-accuracy`

記錄用戶的技術等級準確度，影響技術準確度評分。

#### 路徑參數
- `userId` (string, required): 用戶ID

#### 認證
需要 JWT Token

#### 請求體
```json
{
  "matchId": "match-uuid",
  "reportedLevel": 4.0,
  "observedLevel": 3.8
}
```

#### 請求參數
- `matchId` (string, required): 比賽ID
- `reportedLevel` (number, required): 用戶自報的NTRP等級 (1.0-7.0)
- `observedLevel` (number, required): 實際觀察到的等級 (1.0-7.0)

#### 響應示例
```json
{
  "message": "技術準確度記錄已更新"
}
```

### 6. 提交行為評價

**POST** `/api/v1/reputation/users/{userId}/behavior-review`

對其他用戶提交行為評價，影響其行為評分。

#### 路徑參數
- `userId` (string, required): 被評價用戶ID

#### 認證
需要 JWT Token

#### 請求體
```json
{
  "matchId": "match-uuid",
  "rating": 4.5,
  "comment": "很好的球友，技術不錯且態度友善",
  "tags": ["friendly", "skilled", "punctual"]
}
```

#### 請求參數
- `matchId` (string, optional): 比賽ID
- `rating` (number, required): 評分 (1.0-5.0)
- `comment` (string, optional): 評價內容
- `tags` (array, optional): 評價標籤

#### 響應示例
```json
{
  "message": "行為評價已提交"
}
```

### 7. 獲取信譽排行榜

**GET** `/api/v1/reputation/leaderboard`

獲取信譽分數排行榜。

#### 查詢參數
- `limit` (integer, optional): 返回數量限制，默認50，最大100

#### 響應示例
```json
[
  {
    "id": "uuid",
    "userId": "user-uuid",
    "attendanceRate": 98.5,
    "punctualityScore": 95.2,
    "skillAccuracy": 94.0,
    "behaviorRating": 4.8,
    "totalMatches": 50,
    "completedMatches": 49,
    "cancelledMatches": 1,
    "overallScore": 96.2,
    "updatedAt": "2024-01-15T10:30:00Z",
    "user": {
      "id": "user-uuid",
      "profile": {
        "firstName": "張",
        "lastName": "三",
        "avatarUrl": "https://example.com/avatar.jpg",
        "ntrpLevel": 4.5
      }
    }
  }
]
```

### 8. 獲取信譽統計信息

**GET** `/api/v1/reputation/stats`

獲取平台信譽系統的統計信息。

#### 響應示例
```json
{
  "totalUsers": 1250,
  "averageScore": 82.5,
  "highReputationUsers": 320,
  "activeUsers": 890
}
```

### 9. 更新用戶NTRP等級

**POST** `/api/v1/reputation/users/{userId}/update-ntrp`

基於信譽系統數據自動調整用戶NTRP等級。

#### 路徑參數
- `userId` (string, required): 用戶ID

#### 認證
需要 JWT Token

#### 響應示例
```json
{
  "message": "NTRP等級已更新"
}
```

## 信譽評分計算

### 綜合分數計算公式

```
綜合分數 = 出席率 × 0.3 + 準時度 × 0.2 + 技術準確度 × 0.2 + 行為評分 × 0.3
```

### 各項評分說明

#### 1. 出席率 (0-100分)
- 計算公式：完成的比賽數 / 總比賽數 × 100
- 完成比賽：狀態為 `completed` 的比賽
- 取消比賽：狀態為 `cancelled` 或 `no_show` 的比賽

#### 2. 準時度 (0-100分)
- 準時到達：100分
- 遲到：100 - (遲到分鐘數 × 2)，最低0分
- 取最近10次記錄的平均值

#### 3. 技術準確度 (0-100分)
- 計算公式：100 - (|自報等級 - 實際等級| × 50)，最低0分
- 等級差距0：100分
- 等級差距1.0：50分
- 等級差距2.0：0分
- 取最近10次記錄的平均值

#### 4. 行為評分 (1-5分，轉換為0-100分)
- 轉換公式：(評分 - 1) / 4 × 100
- 1分 → 0分
- 3分 → 50分
- 5分 → 100分
- 取最近20次評價的平均值

## 錯誤碼

| 錯誤碼 | HTTP狀態碼 | 描述 |
|--------|------------|------|
| INVALID_USER_ID | 400 | 用戶ID無效 |
| INVALID_REQUEST | 400 | 請求參數無效 |
| UNAUTHORIZED | 401 | 未授權的請求 |
| GET_REPUTATION_FAILED | 500 | 獲取信譽分數失敗 |
| GET_HISTORY_FAILED | 500 | 獲取信譽歷史失敗 |
| RECORD_ATTENDANCE_FAILED | 500 | 記錄出席情況失敗 |
| RECORD_PUNCTUALITY_FAILED | 500 | 記錄準時情況失敗 |
| RECORD_SKILL_ACCURACY_FAILED | 500 | 記錄技術準確度失敗 |
| SUBMIT_REVIEW_FAILED | 500 | 提交評價失敗 |
| GET_LEADERBOARD_FAILED | 500 | 獲取排行榜失敗 |
| GET_STATS_FAILED | 500 | 獲取統計信息失敗 |
| UPDATE_NTRP_FAILED | 500 | 更新NTRP等級失敗 |

## 使用示例

### 記錄比賽完成
```bash
curl -X POST "http://localhost:8080/api/v1/reputation/users/user-123/attendance" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "matchId": "match-456",
    "status": "completed"
  }'
```

### 提交行為評價
```bash
curl -X POST "http://localhost:8080/api/v1/reputation/users/user-123/behavior-review" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "matchId": "match-456",
    "rating": 4.5,
    "comment": "很好的球友，技術不錯且態度友善",
    "tags": ["friendly", "skilled", "punctual"]
  }'
```

### 獲取排行榜
```bash
curl -X GET "http://localhost:8080/api/v1/reputation/leaderboard?limit=10"
```