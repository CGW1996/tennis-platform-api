# 教練評價系統 API 文檔

## 概述

教練評價系統允許學生對教練進行評分和評價，支援課程完成後的評價流程、評價統計展示、評價互動機制以及評價品質控制。

## API 端點

### 1. 創建教練評價

**POST** `/api/v1/coach-reviews`

創建新的教練評價。

#### 請求頭
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

#### 請求體
```json
{
  "coachId": "string",           // 必填：教練ID
  "lessonId": "string",          // 可選：課程ID（如果是針對特定課程的評價）
  "rating": 5,                   // 必填：評分（1-5）
  "comment": "string",           // 可選：評價內容（最多1000字）
  "tags": ["patient", "professional"] // 可選：評價標籤
}
```

#### 響應
```json
{
  "id": "string",
  "coachId": "string",
  "userId": "string",
  "lessonId": "string",
  "rating": 5,
  "comment": "string",
  "tags": ["patient", "professional"],
  "isHelpful": 0,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": "2024-01-15T10:00:00Z",
  "coach": {
    "id": "string",
    "user": {
      "profile": {
        "firstName": "string",
        "lastName": "string"
      }
    }
  },
  "user": {
    "profile": {
      "firstName": "string",
      "lastName": "string"
    }
  },
  "lesson": {
    "id": "string",
    "type": "individual",
    "scheduledAt": "2024-01-15T10:00:00Z"
  }
}
```

#### 錯誤響應
- `400 Bad Request`: 請求參數錯誤
- `401 Unauthorized`: 未認證
- `404 Not Found`: 教練或課程不存在

---

### 2. 獲取教練評價列表

**GET** `/api/v1/coach-reviews`

獲取特定教練的評價列表，支援多種篩選和排序選項。

#### 查詢參數
- `coachId` (必填): 教練ID
- `rating` (可選): 按評分篩選（1-5）
- `hasComment` (可選): 是否有評論（true/false）
- `tags` (可選): 按標籤篩選，多個標籤用逗號分隔
- `page` (可選): 頁碼，默認1
- `limit` (可選): 每頁數量，默認20，最大100
- `sortBy` (可選): 排序欄位（rating, createdAt, isHelpful），默認createdAt
- `sortOrder` (可選): 排序順序（asc, desc），默認desc

#### 響應
```json
{
  "reviews": [
    {
      "id": "string",
      "rating": 5,
      "comment": "string",
      "tags": ["patient", "professional"],
      "isHelpful": 3,
      "createdAt": "2024-01-15T10:00:00Z",
      "user": {
        "profile": {
          "firstName": "string",
          "lastName": "string"
        }
      }
    }
  ],
  "pagination": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "totalPages": 3,
    "hasNext": true,
    "hasPrev": false
  }
}
```

---

### 3. 獲取評價詳情

**GET** `/api/v1/coach-reviews/{id}`

獲取特定評價的詳細信息。

#### 響應
```json
{
  "id": "string",
  "coachId": "string",
  "userId": "string",
  "lessonId": "string",
  "rating": 5,
  "comment": "string",
  "tags": ["patient", "professional"],
  "isHelpful": 3,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": "2024-01-15T10:00:00Z",
  "coach": {
    "id": "string",
    "user": {
      "profile": {
        "firstName": "string",
        "lastName": "string"
      }
    }
  },
  "user": {
    "profile": {
      "firstName": "string",
      "lastName": "string"
    }
  },
  "lesson": {
    "id": "string",
    "type": "individual",
    "scheduledAt": "2024-01-15T10:00:00Z"
  }
}
```

---

### 4. 更新教練評價

**PUT** `/api/v1/coach-reviews/{id}`

更新自己的教練評價（僅限創建後24小時內）。

#### 請求頭
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

#### 請求體
```json
{
  "rating": 4,                   // 可選：更新評分
  "comment": "string",           // 可選：更新評價內容
  "tags": ["professional", "knowledgeable"] // 可選：更新標籤
}
```

#### 響應
與創建評價相同的響應格式。

#### 錯誤響應
- `400 Bad Request`: 請求參數錯誤或超過編輯時限
- `401 Unauthorized`: 未認證
- `403 Forbidden`: 無權限編輯此評價

---

### 5. 刪除教練評價

**DELETE** `/api/v1/coach-reviews/{id}`

刪除自己的教練評價（僅限創建後24小時內）。

#### 請求頭
```
Authorization: Bearer <access_token>
```

#### 響應
- `204 No Content`: 刪除成功

#### 錯誤響應
- `400 Bad Request`: 超過刪除時限
- `401 Unauthorized`: 未認證
- `403 Forbidden`: 無權限刪除此評價

---

### 6. 標記評價有用

**POST** `/api/v1/coach-reviews/mark-helpful`

標記或取消標記評價為有用。

#### 請求頭
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

#### 請求體
```json
{
  "reviewId": "string",          // 必填：評價ID
  "isHelpful": true              // 必填：是否有用
}
```

#### 響應
與獲取評價詳情相同的響應格式。

#### 錯誤響應
- `400 Bad Request`: 不能標記自己的評價
- `401 Unauthorized`: 未認證

---

### 7. 獲取教練評價統計

**GET** `/api/v1/coaches/{id}/review-statistics`

獲取教練的評價統計信息。

#### 響應
```json
{
  "statistics": {
    "totalReviews": 25,
    "averageRating": 4.6,
    "ratingDistribution": [
      {"rating": 5, "count": 15},
      {"rating": 4, "count": 8},
      {"rating": 3, "count": 2},
      {"rating": 2, "count": 0},
      {"rating": 1, "count": 0}
    ],
    "topTags": [
      {"tag": "patient", "count": 12},
      {"tag": "professional", "count": 10},
      {"tag": "knowledgeable", "count": 8}
    ],
    "recentReviews": [
      {
        "id": "string",
        "rating": 5,
        "comment": "string",
        "createdAt": "2024-01-15T10:00:00Z",
        "user": {
          "profile": {
            "firstName": "string",
            "lastName": "string"
          }
        }
      }
    ],
    "monthlyTrend": [
      {
        "month": "2024-01",
        "count": 5,
        "avgRating": 4.8
      }
    ]
  }
}
```

---

### 8. 獲取可用評價標籤

**GET** `/api/v1/coach-reviews/available-tags`

獲取所有可用的評價標籤選項。

#### 響應
```json
{
  "tags": [
    {
      "value": "patient",
      "label": "耐心",
      "category": "teaching"
    },
    {
      "value": "professional",
      "label": "專業",
      "category": "teaching"
    },
    {
      "value": "punctual",
      "label": "準時",
      "category": "behavior"
    }
  ]
}
```

---

### 9. 檢查是否可以評價教練

**GET** `/api/v1/coach-reviews/can-review`

檢查當前用戶是否可以評價指定教練。

#### 請求頭
```
Authorization: Bearer <access_token>
```

#### 查詢參數
- `coachId` (必填): 教練ID
- `lessonId` (可選): 課程ID

#### 響應
```json
{
  "canReview": true,
  "message": "可以評價"
}
```

可能的訊息：
- "可以評價"
- "教練不存在"
- "不能評價自己"
- "只能評價已完成的課程"
- "該課程已經評價過"
- "已經評價過該教練"
- "需要先完成課程才能評價教練"

---

## 評價標籤分類

### 教學相關 (teaching)
- `patient`: 耐心
- `professional`: 專業
- `knowledgeable`: 知識豐富
- `encouraging`: 鼓勵
- `clear_instruction`: 指導清晰

### 行為相關 (behavior)
- `punctual`: 準時
- `friendly`: 友善
- `responsive`: 回應迅速
- `flexible`: 彈性
- `organized`: 有組織

### 教學風格 (style)
- `technique_focused`: 技術導向
- `fitness_focused`: 體能導向
- `strategy_focused`: 戰術導向
- `fun_approach`: 趣味教學
- `competitive_training`: 競技訓練

## 業務規則

### 評價權限
1. 只有學生可以評價教練
2. 不能評價自己
3. 需要有已完成的課程才能評價
4. 每個課程只能評價一次
5. 每個教練只能有一個一般評價（不指定課程）

### 編輯限制
1. 評價創建後24小時內可以編輯
2. 評價創建後24小時內可以刪除
3. 只有評價者本人可以編輯或刪除

### 有用標記
1. 不能標記自己的評價
2. 每個用戶對每個評價只能標記一次

### 評價統計
1. 評分分佈按1-5星統計
2. 標籤統計顯示前10個最常用標籤
3. 最近評價顯示最新5條
4. 月度趨勢顯示最近12個月數據

## 錯誤碼

| 狀態碼 | 說明 |
|--------|------|
| 200 | 成功 |
| 201 | 創建成功 |
| 204 | 刪除成功 |
| 400 | 請求參數錯誤 |
| 401 | 未認證 |
| 403 | 無權限 |
| 404 | 資源不存在 |
| 500 | 服務器內部錯誤 |

## 使用示例

### 創建評價
```bash
curl -X POST http://localhost:8080/api/v1/coach-reviews \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "coachId": "coach-uuid",
    "rating": 5,
    "comment": "非常棒的教練！",
    "tags": ["patient", "professional"]
  }'
```

### 獲取評價列表
```bash
curl "http://localhost:8080/api/v1/coach-reviews?coachId=coach-uuid&rating=5&page=1&limit=10"
```

### 標記評價有用
```bash
curl -X POST http://localhost:8080/api/v1/coach-reviews/mark-helpful \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reviewId": "review-uuid",
    "isHelpful": true
  }'
```