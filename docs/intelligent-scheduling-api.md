# 智能排課系統 API 文檔

## 概述

智能排課系統提供基於學生偏好和技術水平的教練推薦功能，包括時間匹配、技術等級匹配、地理位置偏好處理、最佳課程時間推薦和排課衝突檢測與解決。

## 功能特性

### 1. 智能教練推薦
- 基於學生NTRP等級匹配合適的教練
- 考慮時間偏好和可用性
- 地理位置距離計算和篩選
- 價格範圍篩選
- 課程類型偏好匹配

### 2. 最佳時間推薦
- 為特定教練和學生找到最佳課程時間
- 綜合考慮雙方時間偏好
- 避免時間衝突
- 提供匹配度評分

### 3. 排課衝突管理
- 實時檢測時間衝突
- 提供衝突解決方案
- 自動調整課程時間

### 4. 匹配因子分析
- 詳細的匹配度分析
- 技術等級相容性評估
- 經驗和評分權重計算

## API 端點

### 獲取智能排課選項

#### 獲取可用選項
```http
GET /api/v1/intelligent-scheduling/options
```

**響應:**
```json
{
    "options": {
        "lessonTypes": [
            {
                "value": "individual",
                "label": "個人課程",
                "description": "一對一教學"
            },
            {
                "value": "group",
                "label": "團體課程",
                "description": "小組教學"
            },
            {
                "value": "clinic",
                "label": "訓練營",
                "description": "大型團體訓練"
            }
        ],
        "skillLevels": [
            {
                "value": "beginner",
                "label": "初學者",
                "nterpRange": "1.0-2.5"
            },
            {
                "value": "intermediate",
                "label": "中級",
                "nterpRange": "2.5-4.0"
            },
            {
                "value": "advanced",
                "label": "高級",
                "nterpRange": "4.0-7.0"
            }
        ],
        "timeSlots": [
            {
                "value": "09:00-12:00",
                "label": "上午 (09:00-12:00)"
            },
            {
                "value": "12:00-14:00",
                "label": "中午 (12:00-14:00)"
            },
            {
                "value": "14:00-18:00",
                "label": "下午 (14:00-18:00)"
            },
            {
                "value": "18:00-21:00",
                "label": "晚上 (18:00-21:00)"
            }
        ],
        "daysOfWeek": [
            {"value": 0, "label": "星期日"},
            {"value": 1, "label": "星期一"},
            {"value": 2, "label": "星期二"},
            {"value": 3, "label": "星期三"},
            {"value": 4, "label": "星期四"},
            {"value": 5, "label": "星期五"},
            {"value": 6, "label": "星期六"}
        ],
        "maxDistance": {
            "min": 0,
            "max": 50,
            "default": 10,
            "unit": "公里"
        },
        "priceRange": {
            "min": 0,
            "max": 5000,
            "default": 1500,
            "currency": "TWD"
        }
    }
}
```

### 智能教練推薦

#### 獲取教練推薦
```http
POST /api/v1/intelligent-scheduling/recommendations
Authorization: Bearer {token}
Content-Type: application/json

{
    "ntrpLevel": 3.5,
    "preferredTimes": ["09:00-12:00", "14:00-18:00"],
    "preferredDays": [1, 2, 3, 4, 5],
    "maxDistance": 10,
    "minPrice": 1000,
    "maxPrice": 2000,
    "preferredLessonType": "individual",
    "dateRange": ["2024-12-01", "2024-12-02", "2024-12-03"],
    "location": {
        "latitude": 25.0330,
        "longitude": 121.5654,
        "address": "台北市信義區"
    }
}
```

**請求參數:**
- `ntrpLevel`: NTRP技術等級 (1.0-7.0)
- `preferredTimes`: 偏好時間段數組 (格式: "HH:MM-HH:MM")
- `preferredDays`: 偏好星期幾 (0=星期日, 6=星期六)
- `maxDistance`: 最大距離 (公里)
- `minPrice`: 最低價格 (可選)
- `maxPrice`: 最高價格 (可選)
- `preferredLessonType`: 偏好課程類型 (individual/group/clinic)
- `dateRange`: 日期範圍數組 (格式: "YYYY-MM-DD")
- `location`: 位置信息 (可選)

**響應:**
```json
{
    "recommendations": [
        {
            "coachId": "uuid",
            "coach": {
                "id": "uuid",
                "userId": "uuid",
                "experience": 5,
                "specialties": ["intermediate", "advanced"],
                "hourlyRate": 1500,
                "currency": "TWD",
                "averageRating": 4.8,
                "totalReviews": 25,
                "isVerified": true,
                "user": {
                    "profile": {
                        "firstName": "張",
                        "lastName": "教練"
                    }
                }
            },
            "timeSlot": {
                "start": "2024-12-01T10:00:00Z",
                "end": "2024-12-01T11:00:00Z",
                "dayOfWeek": 1
            },
            "matchScore": 0.85,
            "price": 1500,
            "location": "台北市信義區",
            "distance": 2.5,
            "matchFactors": {
                "skillLevel": 0.9,
                "timeCompatibility": 0.8,
                "locationScore": 0.9,
                "priceScore": 0.8,
                "experienceScore": 0.5,
                "ratingScore": 0.96
            },
            "lessonTypeId": "uuid",
            "lessonType": {
                "id": "uuid",
                "name": "中級個人課程",
                "type": "individual",
                "level": "intermediate",
                "duration": 60,
                "price": 1500
            }
        }
    ],
    "total": 1
}
```

### 最佳時間推薦

#### 尋找最佳課程時間
```http
POST /api/v1/intelligent-scheduling/optimal-time
Authorization: Bearer {token}
Content-Type: application/json

{
    "coachId": "uuid",
    "ntrpLevel": 3.5,
    "preferredTimes": ["09:00-12:00"],
    "preferredDays": [1, 2, 3],
    "maxDistance": 10,
    "maxPrice": 2000,
    "preferredLessonType": "individual",
    "dateRange": ["2024-12-01", "2024-12-02", "2024-12-03"],
    "location": {
        "latitude": 25.0330,
        "longitude": 121.5654,
        "address": "台北市信義區"
    }
}
```

**響應:**
```json
{
    "recommendation": {
        "coachId": "uuid",
        "coach": { /* 教練詳細信息 */ },
        "timeSlot": {
            "start": "2024-12-01T10:00:00Z",
            "end": "2024-12-01T11:00:00Z",
            "dayOfWeek": 1
        },
        "matchScore": 0.92,
        "price": 1500,
        "location": "台北市信義區",
        "distance": 2.5,
        "matchFactors": { /* 匹配因子詳情 */ },
        "lessonTypeId": "uuid",
        "lessonType": { /* 課程類型詳情 */ }
    }
}
```

### 排課衝突管理

#### 檢測排課衝突
```http
POST /api/v1/intelligent-scheduling/detect-conflicts
Authorization: Bearer {token}
Content-Type: application/json

{
    "coachId": "uuid",
    "scheduledAt": "2024-12-01T10:00:00Z",
    "duration": 60,
    "excludeLessonId": "uuid"
}
```

**請求參數:**
- `coachId`: 教練ID
- `scheduledAt`: 預定時間
- `duration`: 課程時長 (分鐘)
- `excludeLessonId`: 排除的課程ID (可選，用於更新課程時)

**響應:**
```json
{
    "conflicts": [
        {
            "id": "uuid",
            "coachId": "uuid",
            "studentId": "uuid",
            "scheduledAt": "2024-12-01T10:30:00Z",
            "duration": 60,
            "status": "scheduled",
            "student": {
                "profile": {
                    "firstName": "學生",
                    "lastName": "姓名"
                }
            }
        }
    ],
    "hasConflicts": true,
    "conflictCount": 1
}
```

#### 解決排課衝突
```http
POST /api/v1/intelligent-scheduling/resolve-conflict
Authorization: Bearer {token}
Content-Type: application/json

{
    "conflictingLessonId": "uuid",
    "newScheduledAt": "2024-12-01T11:00:00Z"
}
```

**響應:**
```json
{
    "message": "排課衝突已解決"
}
```

### 匹配因子分析

#### 獲取教練推薦因子
```http
POST /api/v1/intelligent-scheduling/coaches/{coachId}/factors
Authorization: Bearer {token}
Content-Type: application/json

{
    "ntrpLevel": 3.5,
    "preferredTimes": ["09:00-12:00"],
    "preferredDays": [1, 2, 3],
    "maxDistance": 10,
    "maxPrice": 2000,
    "preferredLessonType": "individual",
    "dateRange": ["2024-12-01", "2024-12-02", "2024-12-03"]
}
```

**響應:**
```json
{
    "factors": {
        "coachId": "uuid",
        "coachName": "張 教練",
        "experience": 5,
        "averageRating": 4.8,
        "specialties": ["intermediate", "advanced"],
        "hourlyRate": 1500,
        "isVerified": true,
        "totalLessons": 150,
        "totalReviews": 25,
        "studentLevel": "intermediate",
        "levelMatch": true
    }
}
```

## 匹配演算法

### 匹配分數計算

智能排課系統使用多維度評分機制來計算教練與學生的匹配度：

```
總匹配分數 = 技術等級匹配度 × 0.25 +
           時間相容性 × 0.20 +
           位置評分 × 0.15 +
           價格評分 × 0.15 +
           經驗評分 × 0.15 +
           評分評分 × 0.10
```

### 各項因子說明

#### 1. 技術等級匹配度 (25%)
- **完美匹配 (1.0)**: 教練專長完全符合學生等級
- **較好匹配 (0.7)**: 教練專長與學生等級相鄰
- **基本匹配 (0.3)**: 教練可以教導該等級但非專長

#### 2. 時間相容性 (20%)
- 基於教練可用時間與學生偏好時間的重疊度
- 考慮星期偏好和時間段偏好

#### 3. 位置評分 (15%)
- 基於地理距離計算
- 距離越近評分越高
- 超出最大距離限制的教練被排除

#### 4. 價格評分 (15%)
- 在預算範圍內的價格獲得較高評分
- 超出預算的價格會降低評分
- 價格過低可能影響品質評分

#### 5. 經驗評分 (15%)
- 基於教練的教學年數
- 10年以上經驗獲得滿分
- 線性計算: 經驗年數 / 10

#### 6. 評分評分 (10%)
- 基於教練的平均評分
- 5星制評分系統
- 計算: 平均評分 / 5

## 技術等級分類

### NTRP等級對應
- **初學者 (beginner)**: 1.0 - 2.5
- **中級 (intermediate)**: 2.5 - 4.0  
- **高級 (advanced)**: 4.0 - 7.0

## 錯誤處理

### 常見錯誤碼
- `400 Bad Request`: 請求參數錯誤
- `401 Unauthorized`: 未提供有效認證令牌
- `403 Forbidden`: 無權限執行操作
- `404 Not Found`: 教練或課程不存在

### 錯誤響應格式
```json
{
    "error": "錯誤描述",
    "details": "詳細錯誤信息（可選）"
}
```

## 使用場景

### 1. 學生尋找教練
1. 學生設定偏好條件（技術等級、時間、位置、價格）
2. 系統推薦匹配的教練和時間
3. 學生選擇合適的教練和時間
4. 預訂課程

### 2. 教練管理時間表
1. 教練設定可用時間表
2. 系統檢測新課程的時間衝突
3. 提供衝突解決方案
4. 自動調整課程安排

### 3. 最佳時間推薦
1. 學生選定特定教練
2. 系統分析雙方時間偏好
3. 推薦最佳的課程時間
4. 確保無時間衝突

## 性能優化

### 緩存策略
- 教練基本信息緩存
- 時間表信息緩存
- 推薦結果短期緩存

### 查詢優化
- 使用複合索引提升查詢性能
- 分頁處理大量推薦結果
- 預載入關聯數據

### 擴展性考慮
- 支援水平擴展
- 異步處理複雜推薦計算
- 分散式緩存支援

## 安全考慮

### 權限控制
- 學生只能查看自己的推薦
- 教練只能管理自己的時間表
- 管理員可以查看所有數據

### 數據保護
- 位置信息模糊化處理
- 個人偏好數據加密存儲
- API速率限制防止濫用

## 測試

使用提供的測試腳本 `test_intelligent_scheduling_api.sh` 可以快速驗證所有智能排課API端點的功能。

```bash
cd backend
chmod +x test_intelligent_scheduling_api.sh
./test_intelligent_scheduling_api.sh
```