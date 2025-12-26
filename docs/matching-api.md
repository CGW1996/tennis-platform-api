# 配對系統 API 文檔

## 概述

配對系統 API 提供球友配對、信譽管理和配對歷史等功能。所有 API 端點都需要用戶認證（除非另有說明）。

## 基礎資訊

- **基礎 URL**: `/api/v1/matching`
- **認證方式**: Bearer Token (JWT)
- **內容類型**: `application/json`

## API 端點

### 1. 尋找配對

根據指定條件尋找合適的球友配對。

**端點**: `POST /find`

**請求體**:
```json
{
  "ntrpLevel": 3.5,
  "maxDistance": 20.0,
  "preferredTimes": ["morning", "evening"],
  "playingFrequency": "regular",
  "ageRange": {
    "min": 25,
    "max": 40
  },
  "gender": "any",
  "minReputationScore": 80.0,
  "limit": 10
}
```

**請求參數說明**:
- `ntrpLevel` (float, 可選): NTRP 技術等級 (1.0-7.0)
- `maxDistance` (float, 可選): 最大距離（公里）
- `preferredTimes` (array, 可選): 偏好時間 ["morning", "afternoon", "evening"]
- `playingFrequency` (string, 可選): 打球頻率 ("casual", "regular", "competitive")
- `ageRange` (object, 可選): 年齡範圍
  - `min` (int): 最小年齡
  - `max` (int): 最大年齡
- `gender` (string, 可選): 性別偏好 ("male", "female", "any")
- `minReputationScore` (float, 可選): 最低信譽分數 (0-100)
- `limit` (int, 可選): 結果數量限制，預設 10

**成功響應** (200):
```json
{
  "matches": [
    {
      "userId": "user-uuid",
      "score": 0.85,
      "factors": {
        "skillLevel": 0.9,
        "distance": 0.8,
        "timeCompatibility": 0.7,
        "playingStyle": 0.9,
        "age": 0.8,
        "reputation": 0.9
      },
      "user": {
        "id": "user-uuid",
        "profile": {
          "firstName": "John",
          "lastName": "Doe",
          "ntrpLevel": 3.5,
          "playingStyle": "aggressive",
          "avatarUrl": "https://example.com/avatar.jpg"
        }
      }
    }
  ],
  "total": 1
}
```

### 2. 隨機配對（抽卡功能）

提供抽卡式的隨機配對功能。

**端點**: `GET /random`

**查詢參數**:
- `count` (int, 可選): 配對數量，預設 5，最大 20

**成功響應** (200):
```json
{
  "matches": [
    {
      "userId": "user-uuid",
      "score": 0.75,
      "factors": {
        "skillLevel": 0.8,
        "distance": 0.7,
        "timeCompatibility": 0.6,
        "playingStyle": 0.8,
        "age": 0.9,
        "reputation": 0.8
      },
      "user": {
        "id": "user-uuid",
        "profile": {
          "firstName": "Jane",
          "lastName": "Smith",
          "ntrpLevel": 3.0,
          "playingStyle": "defensive"
        }
      }
    }
  ],
  "total": 1
}
```

### 3. 獲取信譽分數

獲取當前用戶的信譽分數詳情。

**端點**: `GET /reputation`

**成功響應** (200):
```json
{
  "reputation": {
    "id": "reputation-uuid",
    "userId": "user-uuid",
    "attendanceRate": 95.5,
    "punctualityScore": 88.2,
    "skillAccuracy": 92.0,
    "behaviorRating": 4.5,
    "totalMatches": 25,
    "completedMatches": 23,
    "cancelledMatches": 2,
    "overallScore": 91.3,
    "updatedAt": "2024-01-15T10:30:00Z"
  }
}
```

### 4. 更新信譽分數

更新指定用戶的信譽分數（通常在比賽結束後調用）。

**端點**: `PUT /reputation/{userID}`

**路徑參數**:
- `userID` (string): 用戶 ID

**請求體**:
```json
{
  "matchCompleted": true,
  "wasOnTime": true,
  "behaviorRating": 4.5
}
```

**請求參數說明**:
- `matchCompleted` (bool): 比賽是否完成
- `wasOnTime` (bool): 是否準時
- `behaviorRating` (float): 行為評分 (1.0-5.0)

**成功響應** (200):
```json
{
  "message": "Reputation updated successfully"
}
```

### 5. 獲取配對歷史

獲取用戶的配對歷史記錄。

**端點**: `GET /history`

**查詢參數**:
- `page` (int, 可選): 頁碼，預設 1
- `limit` (int, 可選): 每頁數量，預設 10，最大 50

**成功響應** (200):
```json
{
  "matches": [
    {
      "id": "match-uuid",
      "type": "casual",
      "status": "completed",
      "scheduledAt": "2024-01-15T14:00:00Z",
      "completedAt": "2024-01-15T16:00:00Z",
      "duration": 120,
      "participants": [
        {
          "id": "user-uuid",
          "profile": {
            "firstName": "John",
            "lastName": "Doe"
          }
        }
      ],
      "court": {
        "id": "court-uuid",
        "name": "中央網球場",
        "address": "台北市信義區"
      },
      "results": [
        {
          "score": "6-4, 6-2",
          "winnerId": "user-uuid",
          "isConfirmed": true
        }
      ]
    }
  ],
  "page": 1,
  "limit": 10,
  "total": 1
}
```

### 6. 創建配對

創建新的球友配對。

**端點**: `POST /create`

**請求體**:
```json
{
  "participantIds": ["user-uuid-1", "user-uuid-2"],
  "matchType": "casual",
  "courtId": "court-uuid",
  "scheduledAt": "2024-01-20T14:00:00Z"
}
```

**請求參數說明**:
- `participantIds` (array): 參與者 ID 列表
- `matchType` (string): 配對類型 ("casual", "practice", "tournament")
- `courtId` (string, 可選): 場地 ID
- `scheduledAt` (string, 可選): 預定時間 (ISO 8601 格式)

**成功響應** (201):
```json
{
  "match": {
    "id": "match-uuid",
    "type": "casual",
    "status": "pending",
    "courtId": "court-uuid",
    "scheduledAt": "2024-01-20T14:00:00Z",
    "participants": [
      {
        "id": "user-uuid-1",
        "profile": {
          "firstName": "John",
          "lastName": "Doe"
        }
      },
      {
        "id": "user-uuid-2",
        "profile": {
          "firstName": "Jane",
          "lastName": "Smith"
        }
      }
    ],
    "chatRoom": {
      "id": "chatroom-uuid",
      "type": "match",
      "isActive": true
    },
    "createdAt": "2024-01-15T10:30:00Z"
  }
}
```

### 7. 獲取配對統計

獲取用戶的配對統計資訊。

**端點**: `GET /statistics`

**成功響應** (200):
```json
{
  "statistics": {
    "totalMatches": 25,
    "completedMatches": 23,
    "cancelledMatches": 2,
    "reputationScore": 91.3,
    "successRate": 92.0
  }
}
```

### 8. 處理抽卡動作

處理用戶對抽卡配對的動作（喜歡、不喜歡、跳過）。

**端點**: `POST /card-action`

**請求體**:
```json
{
  "targetUserId": "user-uuid",
  "action": "like"
}
```

**請求參數說明**:
- `targetUserId` (string): 目標用戶 ID
- `action` (string): 動作類型 ("like", "dislike", "skip")

**成功響應** (200):
```json
{
  "result": {
    "isMatch": true,
    "matchId": "match-uuid",
    "chatRoomId": "chatroom-uuid",
    "message": "配對成功！你們可以開始聊天了"
  }
}
```

**配對失敗響應** (200):
```json
{
  "result": {
    "isMatch": false,
    "message": "已表達興趣，等待對方回應"
  }
}
```

### 9. 獲取抽卡互動歷史

獲取用戶的抽卡互動歷史記錄。

**端點**: `GET /card-history`

**查詢參數**:
- `page` (int, 可選): 頁碼，預設 1
- `limit` (int, 可選): 每頁數量，預設 20，最大 100
- `action` (string, 可選): 動作類型篩選 ("like", "dislike", "skip")

**成功響應** (200):
```json
{
  "interactions": [
    {
      "id": "interaction-uuid",
      "userId": "user-uuid",
      "targetUserId": "target-user-uuid",
      "action": "like",
      "isMatch": true,
      "matchId": "match-uuid",
      "createdAt": "2024-01-15T10:30:00Z",
      "targetUser": {
        "id": "target-user-uuid",
        "profile": {
          "firstName": "Jane",
          "lastName": "Smith",
          "avatarUrl": "https://example.com/avatar.jpg"
        }
      }
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 1
}
```

### 10. 獲取配對通知

獲取用戶的配對相關通知。

**端點**: `GET /notifications`

**查詢參數**:
- `page` (int, 可選): 頁碼，預設 1
- `limit` (int, 可選): 每頁數量，預設 20，最大 50
- `unread_only` (bool, 可選): 只顯示未讀，預設 false

**成功響應** (200):
```json
{
  "notifications": [
    {
      "id": "notification-uuid",
      "userId": "user-uuid",
      "type": "match_success",
      "title": "配對成功！",
      "message": "你們互相喜歡，現在可以開始聊天了",
      "data": "{\"matchId\":\"match-uuid\",\"targetUserId\":\"target-user-uuid\"}",
      "isRead": false,
      "readAt": null,
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 1
}
```

### 11. 標記通知為已讀

標記指定通知為已讀狀態。

**端點**: `PUT /notifications/{notificationID}/read`

**路徑參數**:
- `notificationID` (string): 通知 ID

**成功響應** (200):
```json
{
  "message": "Notification marked as read"
}
```

## 配對演算法說明

### 配對評分因子

配對系統使用多個因子來計算配對分數：

1. **技術等級匹配** (權重: 35%)
   - NTRP 等級差異越小，分數越高
   - 0.0-0.5 差異: 1.0 分
   - 0.5-1.0 差異: 0.8 分
   - 1.0-1.5 差異: 0.6 分
   - 1.5-2.0 差異: 0.4 分
   - 2.0+ 差異: 0.2 分

2. **地理距離匹配** (權重: 25%)
   - 距離越近，分數越高
   - 0-5km: 1.0 分
   - 5-10km: 0.8 分
   - 10-15km: 0.6 分
   - 15-20km: 0.4 分
   - 20km+: 動態計算

3. **時間相容性** (權重: 20%)
   - 共同時間偏好越多，分數越高
   - 基於共同時間數量與總時間偏好的比例

4. **打球風格匹配** (權重: 10%)
   - 相同風格: 高分
   - all-court 與任何風格: 中等分
   - 不同風格: 較低分

5. **年齡匹配** (權重: 5%)
   - 年齡差異越小，分數越高
   - 0-3 歲差異: 1.0 分
   - 3-5 歲差異: 0.8 分
   - 5-10 歲差異: 0.6 分

6. **信譽匹配** (權重: 5%)
   - 基於用戶信譽分數的正規化值

### 信譽分數計算

信譽分數由以下因子組成：

- **出席率** (權重: 30%): 完成比賽數 / 總比賽數
- **準時度** (權重: 20%): 基於歷史準時記錄的移動平均
- **技術準確度** (權重: 20%): NTRP 等級的準確性評估
- **行為評分** (權重: 30%): 其他用戶給予的行為評分平均

## 錯誤響應

### 400 Bad Request
```json
{
  "error": "Invalid request format"
}
```

### 401 Unauthorized
```json
{
  "error": "User not authenticated"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to find matches"
}
```

## 使用範例

### 尋找技術水平相近的球友

```bash
curl -X POST "http://localhost:8080/api/v1/matching/find" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "ntrpLevel": 3.5,
    "maxDistance": 15.0,
    "playingFrequency": "regular",
    "limit": 5
  }'
```

### 抽卡式隨機配對

```bash
curl -X GET "http://localhost:8080/api/v1/matching/random?count=3" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 創建配對

```bash
curl -X POST "http://localhost:8080/api/v1/matching/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "participantIds": ["user-uuid-1"],
    "matchType": "casual"
  }'
```

## 注意事項

1. **認證要求**: 所有 API 端點都需要有效的 JWT token
2. **速率限制**: API 可能有速率限制，請避免過於頻繁的請求
3. **數據隱私**: 用戶位置資訊會根據隱私設定進行模糊化處理
4. **配對品質**: 系統會優先推薦高品質的配對，可能不會返回所有可能的候選人
5. **實時性**: 配對結果基於當前數據庫狀態，可能不包含最新的用戶狀態變更

## 測試

使用提供的測試腳本來測試 API：

```bash
./backend/test_matching_api.sh
```

測試腳本會自動創建測試用戶並執行所有 API 端點的測試。