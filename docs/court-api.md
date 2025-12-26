# 場地管理 API 文檔

## 概述

場地管理 API 提供網球場地的完整 CRUD 操作，包括場地搜尋、詳情查看、創建、更新和刪除功能。支援地理位置搜尋、多維度篩選和圖片上傳。

## API 端點

### 公開端點（無需認證）

#### 1. 搜尋場地
```
GET /api/v1/courts
```

**查詢參數：**
- `latitude` (number, optional): 緯度
- `longitude` (number, optional): 經度  
- `radius` (number, optional): 搜尋半徑（公里）
- `minPrice` (number, optional): 最低價格
- `maxPrice` (number, optional): 最高價格
- `courtType` (string, optional): 場地類型 (hard, clay, grass, indoor, outdoor)
- `facilities` (array, optional): 設施列表
- `minRating` (number, optional): 最低評分
- `sortBy` (string, optional): 排序欄位 (distance, price, rating, name)
- `sortOrder` (string, optional): 排序順序 (asc, desc)
- `page` (int, optional): 頁碼，默認 1
- `pageSize` (int, optional): 每頁數量，默認 20

**回應範例：**
```json
{
  "courts": [
    {
      "id": "uuid",
      "name": "台北網球中心",
      "description": "專業網球場地，設備完善",
      "address": "台北市信義區松壽路20號",
      "latitude": 25.0330,
      "longitude": 121.5654,
      "facilities": ["parking", "restroom", "lighting"],
      "courtType": "hard",
      "pricePerHour": 800.0,
      "currency": "TWD",
      "images": ["/images/courts/court1.jpg"],
      "operatingHours": {
        "monday": "06:00-22:00",
        "tuesday": "06:00-22:00"
      },
      "contactPhone": "+886-2-2345-6789",
      "contactEmail": "info@court.com",
      "website": "https://court.com",
      "averageRating": 4.5,
      "totalReviews": 25,
      "isActive": true,
      "distance": 2.5
    }
  ],
  "total": 10,
  "page": 1,
  "pageSize": 20,
  "totalPages": 1
}
```

#### 2. 獲取場地詳情
```
GET /api/v1/courts/{id}
```

**路徑參數：**
- `id` (string, required): 場地ID

**回應範例：**
```json
{
  "id": "uuid",
  "name": "台北網球中心",
  "description": "專業網球場地，設備完善",
  "address": "台北市信義區松壽路20號",
  "latitude": 25.0330,
  "longitude": 121.5654,
  "facilities": ["parking", "restroom", "lighting"],
  "courtType": "hard",
  "pricePerHour": 800.0,
  "currency": "TWD",
  "images": ["/images/courts/court1.jpg"],
  "operatingHours": {
    "monday": "06:00-22:00",
    "tuesday": "06:00-22:00"
  },
  "contactPhone": "+886-2-2345-6789",
  "contactEmail": "info@court.com",
  "website": "https://court.com",
  "averageRating": 4.5,
  "totalReviews": 25,
  "isActive": true,
  "reviews": [
    {
      "id": "uuid",
      "userId": "uuid",
      "rating": 5,
      "comment": "場地很棒！",
      "images": [],
      "createdAt": "2023-12-01T10:00:00Z"
    }
  ]
}
```

#### 3. 獲取場地類型列表
```
GET /api/v1/courts/types
```

**回應範例：**
```json
{
  "types": [
    {
      "key": "hard",
      "name": "硬地球場",
      "description": "最常見的球場類型，適合各種打法"
    },
    {
      "key": "clay",
      "name": "紅土球場", 
      "description": "球速較慢，適合底線型球員"
    }
  ]
}
```

#### 4. 獲取可用設施列表
```
GET /api/v1/courts/facilities
```

**回應範例：**
```json
{
  "facilities": [
    {
      "key": "parking",
      "name": "停車場",
      "icon": "🅿️"
    },
    {
      "key": "restroom",
      "name": "洗手間",
      "icon": "🚻"
    }
  ]
}
```

### 需要認證的端點

#### 5. 創建場地
```
POST /api/v1/courts
```

**請求標頭：**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**請求體：**
```json
{
  "name": "新網球場",
  "description": "場地描述",
  "address": "台北市信義區測試路123號",
  "latitude": 25.0330,
  "longitude": 121.5654,
  "facilities": ["parking", "restroom", "lighting"],
  "courtType": "hard",
  "pricePerHour": 800.0,
  "currency": "TWD",
  "images": ["/images/courts/court1.jpg"],
  "operatingHours": {
    "monday": "06:00-22:00",
    "tuesday": "06:00-22:00",
    "wednesday": "06:00-22:00",
    "thursday": "06:00-22:00",
    "friday": "06:00-22:00",
    "saturday": "06:00-22:00",
    "sunday": "06:00-22:00"
  },
  "contactPhone": "+886-2-2345-6789",
  "contactEmail": "info@court.com",
  "website": "https://court.com"
}
```

#### 6. 更新場地
```
PUT /api/v1/courts/{id}
```

**請求標頭：**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**請求體：**（所有欄位都是可選的）
```json
{
  "name": "更新後的場地名稱",
  "pricePerHour": 900.0,
  "facilities": ["parking", "restroom", "lighting", "wifi"],
  "isActive": true
}
```

#### 7. 刪除場地
```
DELETE /api/v1/courts/{id}
```

**請求標頭：**
```
Authorization: Bearer {access_token}
```

**回應範例：**
```json
{
  "message": "場地刪除成功"
}
```

#### 8. 上傳場地圖片
```
POST /api/v1/courts/{id}/images
```

**請求標頭：**
```
Authorization: Bearer {access_token}
Content-Type: multipart/form-data
```

**請求體：**
- `images` (file[]): 圖片文件（支援多個文件）

**回應範例：**
```json
{
  "message": "圖片上傳成功",
  "court": {
    "id": "uuid",
    "name": "場地名稱",
    "images": ["/uploads/courts/image1.jpg", "/uploads/courts/image2.jpg"]
  },
  "uploads": [
    {
      "fileName": "image1.jpg",
      "originalName": "original1.jpg",
      "size": 1024000,
      "url": "/uploads/courts/image1.jpg",
      "path": "/path/to/image1.jpg"
    }
  ]
}
```

## 資料模型

### Court（場地）
```json
{
  "id": "string (UUID)",
  "name": "string (required)",
  "description": "string (optional)",
  "address": "string (required)",
  "latitude": "number (required)",
  "longitude": "number (required)",
  "facilities": "string[] (optional)",
  "courtType": "string (required, enum: hard|clay|grass|indoor|outdoor)",
  "pricePerHour": "number (required)",
  "currency": "string (default: TWD)",
  "images": "string[] (optional)",
  "operatingHours": "object (optional)",
  "contactPhone": "string (optional)",
  "contactEmail": "string (optional)",
  "website": "string (optional)",
  "averageRating": "number (readonly)",
  "totalReviews": "number (readonly)",
  "isActive": "boolean (default: true)",
  "ownerId": "string (UUID, optional)",
  "createdAt": "datetime (readonly)",
  "updatedAt": "datetime (readonly)"
}
```

### 營業時間格式
```json
{
  "monday": "09:00-18:00",
  "tuesday": "09:00-18:00",
  "wednesday": "09:00-18:00",
  "thursday": "09:00-18:00",
  "friday": "09:00-18:00",
  "saturday": "08:00-20:00",
  "sunday": "closed"
}
```

## 錯誤回應

所有 API 端點在發生錯誤時會返回統一格式的錯誤回應：

```json
{
  "error": "錯誤訊息",
  "details": "詳細錯誤信息（可選）"
}
```

### 常見錯誤碼

- `400 Bad Request`: 請求參數錯誤
- `401 Unauthorized`: 未授權（需要登入）
- `404 Not Found`: 資源不存在
- `500 Internal Server Error`: 服務器內部錯誤

## 使用範例

### 1. 搜尋附近的硬地球場
```bash
curl -X GET "http://localhost:8080/api/v1/courts?latitude=25.0330&longitude=121.5654&radius=5&courtType=hard&sortBy=distance"
```

### 2. 創建新場地
```bash
curl -X POST "http://localhost:8080/api/v1/courts" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的網球場",
    "address": "台北市信義區測試路123號",
    "latitude": 25.0330,
    "longitude": 121.5654,
    "courtType": "hard",
    "pricePerHour": 600,
    "facilities": ["parking", "restroom"]
  }'
```

### 3. 上傳場地圖片
```bash
curl -X POST "http://localhost:8080/api/v1/courts/COURT_ID/images" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "images=@court1.jpg" \
  -F "images=@court2.jpg"
```

## 注意事項

1. **地理搜尋**：當提供 `latitude`、`longitude` 和 `radius` 參數時，系統會使用 PostGIS 進行高效的地理搜尋
2. **圖片上傳**：支援 jpg、jpeg、png、gif 格式，單個文件最大 10MB
3. **營業時間**：使用 24 小時制格式（HH:MM-HH:MM），關閉日期使用 "closed"
4. **設施驗證**：只接受預定義的設施類型，可通過 `/courts/facilities` 端點查看
5. **軟刪除**：刪除場地使用軟刪除，不會真正從數據庫中移除記錄
6. **權限控制**：只有場地擁有者或管理員可以修改場地信息