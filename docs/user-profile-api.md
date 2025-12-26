# 用戶檔案管理 API 文檔

## 概述

用戶檔案管理功能提供完整的用戶個人資料管理，包括基本資訊、網球技能等級、偏好設定、位置資訊和隱私控制。

## API 端點

### 1. 獲取用戶檔案

**GET** `/api/v1/users/profile`

獲取當前用戶的完整檔案資訊。

**請求頭:**
```
Authorization: Bearer <access_token>
```

**響應:**
```json
{
  "id": "user-uuid",
  "email": "user@example.com",
  "emailVerified": true,
  "isActive": true,
  "profile": {
    "userId": "user-uuid",
    "firstName": "張",
    "lastName": "三",
    "avatarUrl": "/uploads/avatars/avatar_user-uuid_1234567890.jpg",
    "ntrpLevel": 3.5,
    "playingStyle": "aggressive",
    "preferredHand": "right",
    "latitude": 25.0330,
    "longitude": 121.5654,
    "locationPrivacy": false,
    "bio": "熱愛網球的玩家",
    "birthDate": "1990-01-01",
    "gender": "male",
    "playingFrequency": "regular",
    "preferredTimes": ["morning", "evening"],
    "maxTravelDistance": 10.0,
    "profilePrivacy": "public",
    "createdAt": "2023-01-01T00:00:00Z",
    "updatedAt": "2023-01-01T00:00:00Z"
  }
}
```

### 2. 創建用戶檔案

**POST** `/api/v1/users/profile`

為新用戶創建詳細檔案。

**請求頭:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**請求體:**
```json
{
  "firstName": "張",
  "lastName": "三",
  "ntrpLevel": 3.5,
  "playingStyle": "aggressive",
  "preferredHand": "right",
  "latitude": 25.0330,
  "longitude": 121.5654,
  "bio": "熱愛網球的玩家",
  "birthDate": "1990-01-01T00:00:00Z",
  "gender": "male",
  "playingFrequency": "regular",
  "preferredTimes": ["morning", "evening"],
  "maxTravelDistance": 10.0,
  "locationPrivacy": false,
  "profilePrivacy": "public"
}
```

**響應:** 201 Created - 返回完整的用戶物件

### 3. 更新用戶檔案

**PUT** `/api/v1/users/profile`

更新用戶檔案資訊（部分更新）。

**請求頭:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**請求體:**
```json
{
  "firstName": "更新的名字",
  "ntrpLevel": 4.0,
  "bio": "更新的個人簡介"
}
```

**響應:** 200 OK - 返回更新後的用戶物件

### 4. 更新用戶偏好設定

**PUT** `/api/v1/users/preferences`

更新用戶的打球偏好和隱私設定。

**請求頭:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**請求體:**
```json
{
  "playingStyle": "defensive",
  "playingFrequency": "competitive",
  "preferredTimes": ["afternoon", "evening"],
  "maxTravelDistance": 15.0,
  "locationPrivacy": true,
  "profilePrivacy": "friends"
}
```

**響應:** 200 OK - 返回更新後的用戶物件

### 5. 更新用戶位置

**PUT** `/api/v1/users/location`

更新用戶的地理位置資訊。

**請求頭:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**請求體:**
```json
{
  "latitude": 25.0478,
  "longitude": 121.5319,
  "locationPrivacy": true
}
```

**響應:** 200 OK - 返回更新後的用戶物件

### 6. 上傳用戶頭像

**POST** `/api/v1/users/avatar`

上傳並更新用戶頭像。

**請求頭:**
```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**請求體:**
```
avatar: <image_file>
```

**響應:**
```json
{
  "message": "頭像上傳成功",
  "user": {
    // 完整用戶物件
  },
  "upload": {
    "fileName": "avatar_user-uuid_1234567890.jpg",
    "originalName": "my-photo.jpg",
    "size": 102400,
    "url": "/uploads/avatars/avatar_user-uuid_1234567890.jpg",
    "path": "./uploads/avatars/avatar_user-uuid_1234567890.jpg"
  }
}
```

### 7. 獲取 NTRP 等級列表

**GET** `/api/v1/users/ntrp-levels`

獲取所有可用的 NTRP 等級和描述（公開端點，無需認證）。

**響應:**
```json
{
  "levels": [
    {
      "level": 1.0,
      "description": "新手：剛開始學習網球"
    },
    {
      "level": 1.5,
      "description": "新手：有限的網球經驗"
    },
    // ... 更多等級
    {
      "level": 7.0,
      "description": "職業級：世界級職業選手"
    }
  ]
}
```

## 資料驗證規則

### 基本資訊
- `firstName`: 必填，1-100 字符
- `lastName`: 必填，1-100 字符
- `bio`: 可選，最多 500 字符

### NTRP 等級
- `ntrpLevel`: 可選，1.0-7.0，必須為 0.5 的倍數
- 有效值：1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0, 5.5, 6.0, 6.5, 7.0

### 打球偏好
- `playingStyle`: 可選，枚舉值：`aggressive`, `defensive`, `all-court`
- `preferredHand`: 可選，枚舉值：`right`, `left`, `both`
- `playingFrequency`: 可選，枚舉值：`casual`, `regular`, `competitive`
- `preferredTimes`: 可選，字符串陣列
- `maxTravelDistance`: 可選，0-100 公里

### 地理位置
- `latitude`: 可選，-90 到 90
- `longitude`: 可選，-180 到 180
- `locationPrivacy`: 可選，布林值

### 隱私設定
- `profilePrivacy`: 可選，枚舉值：`public`, `friends`, `private`
- `locationPrivacy`: 可選，布林值（true = 隱藏精確位置）

### 其他
- `gender`: 可選，枚舉值：`male`, `female`, `other`
- `birthDate`: 可選，ISO 8601 日期格式

## 文件上傳限制

### 頭像上傳
- **支援格式**: JPG, JPEG, PNG, GIF
- **最大大小**: 10MB
- **命名規則**: `avatar_{userID}_{timestamp}.{ext}`
- **存儲路徑**: `./uploads/avatars/`
- **訪問 URL**: `/uploads/avatars/{filename}`

## 錯誤響應

### 400 Bad Request
```json
{
  "error": "請求參數錯誤",
  "details": "具體的驗證錯誤訊息"
}
```

### 401 Unauthorized
```json
{
  "error": "未找到用戶信息"
}
```

### 404 Not Found
```json
{
  "error": "用戶不存在"
}
```

## 隱私控制

### 檔案隱私等級
- **public**: 所有用戶可見
- **friends**: 僅朋友可見
- **private**: 僅自己可見

### 位置隱私
- `locationPrivacy: false`: 顯示精確位置
- `locationPrivacy: true`: 只顯示大概區域（例如：台北市）

## 使用範例

### JavaScript/Fetch API

```javascript
// 獲取用戶檔案
const getProfile = async () => {
  const response = await fetch('/api/v1/users/profile', {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
  return response.json();
};

// 更新用戶檔案
const updateProfile = async (profileData) => {
  const response = await fetch('/api/v1/users/profile', {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(profileData)
  });
  return response.json();
};

// 上傳頭像
const uploadAvatar = async (file) => {
  const formData = new FormData();
  formData.append('avatar', file);
  
  const response = await fetch('/api/v1/users/avatar', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    },
    body: formData
  });
  return response.json();
};
```

### cURL 範例

```bash
# 獲取用戶檔案
curl -X GET "http://localhost:8080/api/v1/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 更新用戶檔案
curl -X PUT "http://localhost:8080/api/v1/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"firstName": "新名字", "ntrpLevel": 4.0}'

# 上傳頭像
curl -X POST "http://localhost:8080/api/v1/users/avatar" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "avatar=@/path/to/image.jpg"
```

## 注意事項

1. **認證要求**: 除了 NTRP 等級列表端點外，所有端點都需要有效的 JWT token
2. **檔案存儲**: 上傳的文件存儲在本地文件系統，生產環境建議使用雲存儲服務
3. **隱私保護**: 位置資訊會根據隱私設定進行模糊化處理
4. **NTRP 驗證**: NTRP 等級必須符合官方標準（0.5 的倍數）
5. **部分更新**: 所有更新端點都支援部分更新，只需提供要更改的欄位