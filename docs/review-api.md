# 場地評價系統 API 文檔

## 概述

場地評價系統提供完整的評價功能，包括五星評分、文字評價、圖片上傳、評價統計、舉報機制和審核功能。

## API 端點

### 評價管理

#### 創建評價

```http
POST /api/v1/reviews
```

**需要認證**: 是

**請求體**:
```json
{
  "courtId": "uuid",
  "rating": 4,
  "comment": "場地很不錯，設施齊全。",
  "images": ["https://example.com/image1.jpg", "https://example.com/image2.jpg"]
}
```

**響應**:
```json
{
  "id": "uuid",
  "courtId": "uuid",
  "userId": "uuid",
  "rating": 4,
  "comment": "場地很不錯，設施齊全。",
  "images": ["https://example.com/image1.jpg", "https://example.com/image2.jpg"],
  "isHelpful": 0,
  "isReported": false,
  "reportCount": 0,
  "status": "active",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "user": {
    "id": "uuid",
    "profile": {
      "firstName": "張",
      "lastName": "三",
      "avatarUrl": "https://example.com/avatar.jpg"
    }
  }
}
```

#### 獲取評價詳情

```http
GET /api/v1/reviews/{id}
```

**需要認證**: 否

**響應**: 同創建評價響應

#### 更新評價

```http
PUT /api/v1/reviews/{id}
```

**需要認證**: 是（僅評價作者）

**請求體**:
```json
{
  "rating": 5,
  "comment": "重新評估後覺得這個場地非常棒！",
  "images": ["https://example.com/updated_image.jpg"]
}
```

**響應**: 同創建評價響應

#### 刪除評價

```http
DELETE /api/v1/reviews/{id}
```

**需要認證**: 是（僅評價作者）

**響應**:
```json
{
  "message": "評價刪除成功"
}
```

### 評價列表

#### 獲取評價列表

```http
GET /api/v1/reviews
```

**需要認證**: 否

**查詢參數**:
- `courtId` (string, optional): 場地ID篩選
- `userId` (string, optional): 用戶ID篩選
- `rating` (int, optional): 評分篩選 (1-5)
- `sortBy` (string, optional): 排序欄位 (`rating`, `created_at`, `helpful`)
- `sortOrder` (string, optional): 排序順序 (`asc`, `desc`)
- `page` (int, optional): 頁碼，默認 1
- `pageSize` (int, optional): 每頁數量，默認 20，最大 50

**響應**:
```json
{
  "reviews": [
    {
      "id": "uuid",
      "courtId": "uuid",
      "userId": "uuid",
      "rating": 4,
      "comment": "場地很不錯",
      "images": ["https://example.com/image.jpg"],
      "isHelpful": 5,
      "createdAt": "2024-01-01T00:00:00Z",
      "user": {
        "profile": {
          "firstName": "張",
          "lastName": "三",
          "avatarUrl": "https://example.com/avatar.jpg"
        }
      },
      "court": {
        "id": "uuid",
        "name": "測試網球場"
      }
    }
  ],
  "total": 100,
  "page": 1,
  "pageSize": 20,
  "totalPages": 5
}
```

### 評價統計

#### 獲取場地評價統計

```http
GET /api/v1/courts/{courtId}/reviews/statistics
```

**需要認證**: 否

**響應**:
```json
{
  "totalReviews": 25,
  "averageRating": 4.2,
  "ratingBreakdown": {
    "5": 10,
    "4": 8,
    "3": 5,
    "2": 1,
    "1": 1
  },
  "recentReviews": [
    {
      "id": "uuid",
      "rating": 5,
      "comment": "非常棒的場地！",
      "createdAt": "2024-01-01T00:00:00Z",
      "user": {
        "profile": {
          "firstName": "李",
          "lastName": "四"
        }
      }
    }
  ]
}
```

### 評價互動

#### 標記評價為有用

```http
POST /api/v1/reviews/{id}/helpful?helpful=true
```

**需要認證**: 是

**查詢參數**:
- `helpful` (boolean): true 表示有用，false 表示取消有用

**響應**:
```json
{
  "message": "標記成功"
}
```

#### 舉報評價

```http
POST /api/v1/reviews/{id}/report
```

**需要認證**: 是

**請求體**:
```json
{
  "reason": "inappropriate",
  "comment": "評價內容不當"
}
```

**舉報原因**:
- `spam`: 垃圾信息
- `inappropriate`: 不當內容
- `fake`: 虛假評價
- `offensive`: 冒犯性內容
- `other`: 其他原因

**響應**:
```json
{
  "message": "舉報提交成功"
}
```

### 圖片上傳

#### 上傳評價圖片

```http
POST /api/v1/reviews/images
```

**需要認證**: 是

**請求類型**: `multipart/form-data`

**表單欄位**:
- `images`: 圖片文件（支援多個文件）

**響應**:
```json
{
  "message": "圖片上傳成功",
  "uploads": [
    {
      "filename": "image1.jpg",
      "url": "https://example.com/uploads/reviews/image1.jpg",
      "size": 1024000
    }
  ],
  "imageUrls": [
    "https://example.com/uploads/reviews/image1.jpg"
  ]
}
```

## 錯誤響應

### 400 Bad Request

```json
{
  "error": "請求參數錯誤",
  "details": "rating must be between 1 and 5"
}
```

### 401 Unauthorized

```json
{
  "error": "用戶未認證"
}
```

### 403 Forbidden

```json
{
  "error": "評價不存在或無權限修改"
}
```

### 404 Not Found

```json
{
  "error": "評價不存在"
}
```

## 業務規則

### 評價創建規則

1. **唯一性**: 每個用戶對每個場地只能創建一個評價
2. **評分範圍**: 評分必須在 1-5 之間
3. **評論長度**: 評論最多 1000 字符
4. **圖片限制**: 最多上傳 5 張圖片

### 評價更新規則

1. **權限控制**: 只有評價作者可以更新自己的評價
2. **狀態限制**: 只有狀態為 `active` 的評價可以更新
3. **時間限制**: 評價創建後 30 天內可以修改

### 評價刪除規則

1. **軟刪除**: 評價刪除採用軟刪除方式
2. **統計更新**: 刪除評價後自動更新場地評分統計
3. **權限控制**: 只有評價作者可以刪除自己的評價

### 舉報機制

1. **重複舉報**: 同一用戶不能重複舉報同一評價
2. **自動隱藏**: 當舉報次數達到 5 次時，評價自動隱藏
3. **審核流程**: 被舉報的評價需要管理員審核

### 有用性標記

1. **自評限制**: 用戶不能對自己的評價標記有用
2. **計數更新**: 有用性計數實時更新
3. **取消標記**: 用戶可以取消之前的有用標記

## 數據模型

### CourtReview 評價模型

```go
type CourtReview struct {
    ID           string    `json:"id"`
    CourtID      string    `json:"courtId"`
    UserID       string    `json:"userId"`
    Rating       int       `json:"rating"`
    Comment      *string   `json:"comment"`
    Images       []string  `json:"images"`
    IsHelpful    int       `json:"isHelpful"`
    IsReported   bool      `json:"isReported"`
    ReportCount  int       `json:"reportCount"`
    Status       string    `json:"status"`
    ModeratedAt  *time.Time `json:"moderatedAt"`
    ModeratedBy  *string   `json:"moderatedBy"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}
```

### ReviewReport 舉報模型

```go
type ReviewReport struct {
    ID        string    `json:"id"`
    ReviewID  string    `json:"reviewId"`
    UserID    string    `json:"userId"`
    Reason    string    `json:"reason"`
    Comment   *string   `json:"comment"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

## 使用示例

### JavaScript 客戶端示例

```javascript
// 創建評價
const createReview = async (courtId, rating, comment, images) => {
  const response = await fetch('/api/v1/reviews', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      courtId,
      rating,
      comment,
      images
    })
  });
  
  return response.json();
};

// 獲取場地評價
const getCourtReviews = async (courtId, page = 1) => {
  const response = await fetch(
    `/api/v1/reviews?courtId=${courtId}&page=${page}&sortBy=created_at&sortOrder=desc`
  );
  
  return response.json();
};

// 標記評價有用
const markReviewHelpful = async (reviewId, helpful) => {
  const response = await fetch(
    `/api/v1/reviews/${reviewId}/helpful?helpful=${helpful}`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  return response.json();
};

// 舉報評價
const reportReview = async (reviewId, reason, comment) => {
  const response = await fetch(`/api/v1/reviews/${reviewId}/report`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      reason,
      comment
    })
  });
  
  return response.json();
};
```

### cURL 示例

```bash
# 創建評價
curl -X POST "http://localhost:8080/api/v1/reviews" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "courtId": "court-uuid",
    "rating": 4,
    "comment": "場地很不錯，設施齊全。",
    "images": ["https://example.com/image.jpg"]
  }'

# 獲取評價列表
curl "http://localhost:8080/api/v1/reviews?courtId=court-uuid&page=1&pageSize=10"

# 獲取評價統計
curl "http://localhost:8080/api/v1/courts/court-uuid/reviews/statistics"

# 標記評價有用
curl -X POST "http://localhost:8080/api/v1/reviews/review-uuid/helpful?helpful=true" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 舉報評價
curl -X POST "http://localhost:8080/api/v1/reviews/review-uuid/report" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "reason": "inappropriate",
    "comment": "評價內容不當"
  }'
```

## 性能考慮

### 索引優化

- `court_reviews(court_id, created_at)`: 場地評價列表查詢
- `court_reviews(user_id)`: 用戶評價查詢
- `court_reviews(rating)`: 評分篩選
- `court_reviews(status)`: 狀態篩選
- `review_reports(review_id, user_id)`: 舉報查詢

### 緩存策略

- 場地評價統計緩存 5 分鐘
- 評價列表緩存 1 分鐘
- 用戶評價緩存 10 分鐘

### 分頁限制

- 最大每頁 50 條記錄
- 默認每頁 20 條記錄
- 支援游標分頁以提升大數據集性能

## 安全考慮

### 輸入驗證

- 評分範圍驗證 (1-5)
- 評論長度限制 (1000 字符)
- 圖片格式和大小驗證
- UUID 格式驗證

### 權限控制

- 評價 CRUD 操作權限檢查
- 舉報功能防濫用
- 管理員審核權限

### 防濫用機制

- 評價創建頻率限制
- 舉報功能冷卻時間
- 圖片上傳大小和數量限制