# 球拍管理 API 文檔

## 概述

球拍管理 API 提供球拍資訊的完整管理功能，包括球拍基本資訊、價格追蹤、評價系統等。

## 基礎路徑

```
/api/v1/rackets
```

## 認證

部分端點需要 JWT 認證，在請求頭中包含：
```
Authorization: Bearer <token>
```

## 球拍管理端點

### 1. 搜尋球拍

**GET** `/api/v1/rackets`

搜尋和篩選球拍列表。

#### 查詢參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| query | string | 否 | 搜尋關鍵字（品牌、型號、描述） |
| brand | string | 否 | 品牌名稱 |
| minHeadSize | integer | 否 | 最小拍面大小（平方英寸） |
| maxHeadSize | integer | 否 | 最大拍面大小（平方英寸） |
| minWeight | integer | 否 | 最小重量（克） |
| maxWeight | integer | 否 | 最大重量（克） |
| minPrice | number | 否 | 最低價格 |
| maxPrice | number | 否 | 最高價格 |
| powerLevel | integer | 否 | 力量等級（1-10） |
| controlLevel | integer | 否 | 控制等級（1-10） |
| maneuverLevel | integer | 否 | 操控等級（1-10） |
| stabilityLevel | integer | 否 | 穩定等級（1-10） |
| minRating | number | 否 | 最低評分（0-5） |
| sortBy | string | 否 | 排序欄位：brand, model, price, rating, popularity |
| sortOrder | string | 否 | 排序順序：asc, desc |
| page | integer | 否 | 頁碼（默認：1） |
| pageSize | integer | 否 | 每頁數量（默認：20，最大：100） |

#### 回應範例

```json
{
  "rackets": [
    {
      "id": "uuid",
      "brand": "Wilson",
      "model": "Pro Staff 97",
      "year": 2023,
      "headSize": 97,
      "weight": 315,
      "balance": 315,
      "stringPattern": "16x19",
      "beamWidth": 21.5,
      "length": 27,
      "stiffness": 68,
      "swingWeight": 335,
      "powerLevel": 6,
      "controlLevel": 9,
      "maneuverLevel": 7,
      "stabilityLevel": 8,
      "description": "專業級網球拍，適合高級球員",
      "images": ["url1", "url2"],
      "msrp": 8500.0,
      "currency": "TWD",
      "averageRating": 4.5,
      "totalReviews": 25,
      "isActive": true,
      "createdAt": "2023-01-01T00:00:00Z",
      "updatedAt": "2023-01-01T00:00:00Z",
      "prices": [
        {
          "id": "uuid",
          "retailer": "網球專賣店",
          "price": 7500.0,
          "currency": "TWD",
          "url": "https://example.com/product",
          "isAvailable": true,
          "lastChecked": "2023-01-01T00:00:00Z"
        }
      ]
    }
  ],
  "total": 100,
  "page": 1,
  "pageSize": 20,
  "totalPages": 5
}
```

### 2. 獲取球拍詳情

**GET** `/api/v1/rackets/{id}`

獲取指定球拍的詳細資訊。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 回應範例

```json
{
  "id": "uuid",
  "brand": "Wilson",
  "model": "Pro Staff 97",
  "year": 2023,
  "headSize": 97,
  "weight": 315,
  "balance": 315,
  "stringPattern": "16x19",
  "beamWidth": 21.5,
  "length": 27,
  "stiffness": 68,
  "swingWeight": 335,
  "powerLevel": 6,
  "controlLevel": 9,
  "maneuverLevel": 7,
  "stabilityLevel": 8,
  "description": "專業級網球拍，適合高級球員",
  "images": ["url1", "url2"],
  "msrp": 8500.0,
  "currency": "TWD",
  "averageRating": 4.5,
  "totalReviews": 25,
  "isActive": true,
  "createdAt": "2023-01-01T00:00:00Z",
  "updatedAt": "2023-01-01T00:00:00Z",
  "reviews": [...],
  "prices": [...]
}
```

### 3. 創建球拍 🔒

**POST** `/api/v1/rackets`

創建新的球拍記錄。需要認證。

#### 請求體

```json
{
  "brand": "Wilson",
  "model": "Pro Staff 97",
  "year": 2023,
  "headSize": 97,
  "weight": 315,
  "balance": 315,
  "stringPattern": "16x19",
  "beamWidth": 21.5,
  "length": 27,
  "stiffness": 68,
  "swingWeight": 335,
  "powerLevel": 6,
  "controlLevel": 9,
  "maneuverLevel": 7,
  "stabilityLevel": 8,
  "description": "專業級網球拍，適合高級球員",
  "images": ["url1", "url2"],
  "msrp": 8500.0,
  "currency": "TWD"
}
```

#### 驗證規則

- `brand`: 必需，1-100 字符
- `model`: 必需，1-100 字符
- `year`: 可選，1900-2030
- `headSize`: 必需，80-140 平方英寸
- `weight`: 必需，200-400 克
- `balance`: 可選，280-380 毫米
- `stringPattern`: 必需
- `beamWidth`: 可選，15-35 毫米
- `length`: 可選，26-29 英寸
- `stiffness`: 可選，40-80 RA
- `swingWeight`: 可選，250-400
- `powerLevel`: 可選，1-10
- `controlLevel`: 可選，1-10
- `maneuverLevel`: 可選，1-10
- `stabilityLevel`: 可選，1-10
- `description`: 可選，最多 2000 字符
- `msrp`: 可選，≥ 0
- `currency`: 可選，TWD/USD/EUR

### 4. 更新球拍 🔒

**PUT** `/api/v1/rackets/{id}`

更新球拍資訊。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 請求體

所有欄位都是可選的，只更新提供的欄位。

```json
{
  "brand": "Wilson",
  "model": "Pro Staff 97 v13",
  "year": 2024,
  "description": "更新的描述",
  "isActive": true
}
```

### 5. 刪除球拍 🔒

**DELETE** `/api/v1/rackets/{id}`

軟刪除球拍記錄。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 回應

```
HTTP 204 No Content
```

### 6. 獲取可用品牌

**GET** `/api/v1/rackets/brands`

獲取所有可用的球拍品牌列表。

#### 回應範例

```json
{
  "brands": [
    "Wilson",
    "Babolat",
    "Head",
    "Yonex",
    "Prince",
    "Tecnifibre"
  ]
}
```

### 7. 獲取球拍規格選項

**GET** `/api/v1/rackets/specifications`

獲取球拍規格的可選項目和範圍。

#### 回應範例

```json
{
  "headSizeRanges": [
    {"label": "Midsize (85-97 sq in)", "min": 85, "max": 97},
    {"label": "Midplus (98-105 sq in)", "min": 98, "max": 105},
    {"label": "Oversize (106+ sq in)", "min": 106, "max": 140}
  ],
  "weightRanges": [
    {"label": "Light (250-280g)", "min": 250, "max": 280},
    {"label": "Medium (281-310g)", "min": 281, "max": 310},
    {"label": "Heavy (311g+)", "min": 311, "max": 400}
  ],
  "stringPatterns": [
    "16x19", "16x20", "18x20", "16x18", "14x18", "12x18"
  ],
  "currencies": ["TWD", "USD", "EUR"],
  "levels": [
    {"value": 1, "label": "Very Low"},
    {"value": 2, "label": "Low"},
    {"value": 3, "label": "Low-Medium"},
    {"value": 4, "label": "Medium"},
    {"value": 5, "label": "Medium"},
    {"value": 6, "label": "Medium-High"},
    {"value": 7, "label": "High"},
    {"value": 8, "label": "High"},
    {"value": 9, "label": "Very High"},
    {"value": 10, "label": "Maximum"}
  ]
}
```

### 8. 上傳球拍圖片 🔒

**POST** `/api/v1/rackets/images`

上傳球拍相關圖片。需要認證。

#### 請求

- Content-Type: `multipart/form-data`
- 欄位名稱: `images`
- 支援多個文件

#### 回應範例

```json
{
  "images": [
    "https://example.com/uploads/rackets/image1.jpg",
    "https://example.com/uploads/rackets/image2.jpg"
  ]
}
```

## 球拍價格管理端點

### 1. 獲取球拍價格

**GET** `/api/v1/rackets/{id}/prices`

獲取指定球拍的所有價格資訊。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 回應範例

```json
{
  "prices": [
    {
      "id": "uuid",
      "racketId": "uuid",
      "retailer": "網球專賣店",
      "price": 7500.0,
      "currency": "TWD",
      "url": "https://example.com/product",
      "isAvailable": true,
      "lastChecked": "2023-01-01T00:00:00Z",
      "createdAt": "2023-01-01T00:00:00Z",
      "updatedAt": "2023-01-01T00:00:00Z"
    }
  ],
  "lowestPrice": {
    "id": "uuid",
    "retailer": "網球專賣店",
    "price": 7500.0,
    "currency": "TWD",
    "url": "https://example.com/product",
    "isAvailable": true
  }
}
```

### 2. 創建球拍價格 🔒

**POST** `/api/v1/rackets/{id}/prices`

為球拍添加新的價格資訊。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 請求體

```json
{
  "retailer": "網球專賣店",
  "price": 7500.0,
  "currency": "TWD",
  "url": "https://example.com/product",
  "isAvailable": true
}
```

#### 驗證規則

- `retailer`: 必需，1-100 字符
- `price`: 必需，≥ 0
- `currency`: 可選，TWD/USD/EUR
- `url`: 可選，有效 URL
- `isAvailable`: 可選，布林值

### 3. 更新球拍價格 🔒

**PUT** `/api/v1/racket-prices/{priceId}`

更新球拍價格資訊。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| priceId | string | 是 | 價格 ID |

#### 請求體

所有欄位都是可選的。

```json
{
  "retailer": "新的零售商名稱",
  "price": 8000.0,
  "currency": "TWD",
  "url": "https://newurl.com/product",
  "isAvailable": false
}
```

### 4. 刪除球拍價格 🔒

**DELETE** `/api/v1/racket-prices/{priceId}`

刪除球拍價格記錄。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| priceId | string | 是 | 價格 ID |

### 5. 更新價格可用性 🔒

**PUT** `/api/v1/racket-prices/{priceId}/availability`

更新球拍價格的可用性狀態。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| priceId | string | 是 | 價格 ID |

#### 請求體

```json
{
  "isAvailable": false
}
```

#### 回應範例

```json
{
  "message": "Price availability updated successfully",
  "isAvailable": false
}
```

## 球拍評價管理端點

### 1. 獲取球拍評價

**GET** `/api/v1/rackets/{id}/reviews`

獲取指定球拍的評價列表。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 查詢參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| page | integer | 否 | 頁碼（默認：1） |
| pageSize | integer | 否 | 每頁數量（默認：20） |
| sortBy | string | 否 | 排序欄位：rating, date, helpful |
| sortOrder | string | 否 | 排序順序：asc, desc |

#### 回應範例

```json
{
  "reviews": [
    {
      "id": "uuid",
      "racketId": "uuid",
      "userId": "uuid",
      "rating": 5,
      "powerRating": 4,
      "controlRating": 5,
      "comfortRating": 4,
      "comment": "非常好的球拍，控制性極佳",
      "playingStyle": "all-court",
      "usageDuration": 6,
      "isHelpful": 3,
      "createdAt": "2023-01-01T00:00:00Z",
      "updatedAt": "2023-01-01T00:00:00Z",
      "user": {
        "id": "uuid",
        "firstName": "張",
        "lastName": "三"
      }
    }
  ],
  "total": 25,
  "page": 1,
  "pageSize": 20,
  "totalPages": 2
}
```

### 2. 創建球拍評價 🔒

**POST** `/api/v1/rackets/{id}/reviews`

為球拍創建新的評價。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 請求體

```json
{
  "rating": 5,
  "powerRating": 4,
  "controlRating": 5,
  "comfortRating": 4,
  "comment": "非常好的球拍，控制性極佳",
  "playingStyle": "all-court",
  "usageDuration": 6
}
```

#### 驗證規則

- `rating`: 必需，1-5
- `powerRating`: 可選，1-5
- `controlRating`: 可選，1-5
- `comfortRating`: 可選，1-5
- `comment`: 可選，最多 2000 字符
- `playingStyle`: 必需，aggressive/defensive/all-court
- `usageDuration`: 可選，0-120 月

### 3. 獲取球拍評價統計

**GET** `/api/v1/rackets/{id}/reviews/statistics`

獲取指定球拍的評價統計資訊。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| id | string | 是 | 球拍 ID |

#### 回應範例

```json
{
  "racketId": "uuid",
  "totalReviews": 25,
  "averageRating": 4.5,
  "ratingDistribution": {
    "1": 0,
    "2": 1,
    "3": 2,
    "4": 10,
    "5": 12
  },
  "powerRating": 4.2,
  "controlRating": 4.7,
  "comfortRating": 4.1,
  "playingStyleStats": {
    "aggressive": {
      "count": 8,
      "averageRating": 4.3
    },
    "defensive": {
      "count": 7,
      "averageRating": 4.6
    },
    "all-court": {
      "count": 10,
      "averageRating": 4.5
    }
  },
  "usageDurationStats": {
    "averageDuration": 8.5,
    "minDuration": 1,
    "maxDuration": 24
  }
}
```

### 4. 標記評價有用 🔒

**POST** `/api/v1/racket-reviews/{reviewId}/helpful`

標記球拍評價為有用或無用。需要認證。

#### 路徑參數

| 參數 | 類型 | 必需 | 描述 |
|------|------|------|------|
| reviewId | string | 是 | 評價 ID |

#### 請求體

```json
{
  "helpful": true
}
```

#### 回應範例

```json
{
  "message": "Review marked successfully",
  "helpful": true
}
```

## 錯誤回應

所有端點都可能返回以下錯誤：

### 400 Bad Request

```json
{
  "error": "Invalid request format",
  "details": "validation error details"
}
```

### 401 Unauthorized

```json
{
  "error": "User not authenticated"
}
```

### 404 Not Found

```json
{
  "error": "Racket not found",
  "message": "The requested racket does not exist"
}
```

### 409 Conflict

```json
{
  "error": "Racket already exists",
  "message": "A racket with the same brand and model already exists"
}
```

### 500 Internal Server Error

```json
{
  "error": "Failed to create racket",
  "message": "detailed error message"
}
```

## 使用範例

### 搜尋 Wilson 品牌的球拍

```bash
curl -X GET "http://localhost:8080/api/v1/rackets?brand=Wilson&sortBy=rating&sortOrder=desc"
```

### 創建新球拍

```bash
curl -X POST "http://localhost:8080/api/v1/rackets" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "brand": "Wilson",
    "model": "Pro Staff 97",
    "headSize": 97,
    "weight": 315,
    "stringPattern": "16x19"
  }'
```

### 添加價格資訊

```bash
curl -X POST "http://localhost:8080/api/v1/rackets/{id}/prices" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "retailer": "網球專賣店",
    "price": 7500.0,
    "currency": "TWD",
    "isAvailable": true
  }'
```

### 創建評價

```bash
curl -X POST "http://localhost:8080/api/v1/rackets/{id}/reviews" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "rating": 5,
    "powerRating": 4,
    "controlRating": 5,
    "comment": "非常好的球拍",
    "playingStyle": "all-court"
  }'
```