# 配對統計 API 文檔

## 概述

配對統計 API 提供用戶配對歷史記錄、比賽結果統計、技術等級進展追蹤和隱私控制功能。

## 基礎 URL

```
/api/v1/match-statistics
```

## 認證

大部分端點需要 JWT 認證。在請求頭中包含：

```
Authorization: Bearer <your-jwt-token>
```

## 端點列表

### 1. 獲取用戶配對統計資訊

獲取指定用戶的詳細配對統計資訊。

**端點：** `GET /users/{userId}`

**參數：**
- `userId` (path, required): 用戶ID

**響應：**
```json
{
  "userId": "user-123",
  "totalMatches": 45,
  "completedMatches": 38,
  "cancelledMatches": 7,
  "wonMatches": 22,
  "lostMatches": 16,
  "winRate": 57.89,
  "attendanceRate": 84.44,
  "averageMatchDuration": 90,
  "favoriteCourtType": "hard",
  "mostPlayedWith": ["user-456", "user-789"],
  "recentMatches": [...],
  "monthlyStats": [...],
  "skillLevelProgression": [...]
}
```

### 2. 獲取用戶配對歷史

獲取指定用戶的配對歷史記錄。

**端點：** `GET /users/{userId}/history`

**參數：**
- `userId` (path, required): 用戶ID
- `limit` (query, optional): 返回數量限制，默認20，最大100
- `offset` (query, optional): 偏移量，默認0

**響應：**
```json
[
  {
    "id": "match-123",
    "type": "casual",
    "status": "completed",
    "scheduledAt": "2024-11-15T14:00:00Z",
    "completedAt": "2024-11-15T15:30:00Z",
    "duration": 90,
    "participants": [...],
    "court": {...},
    "results": [...]
  }
]
```

### 3. 記錄比賽結果

記錄比賽的勝負結果和比分。

**端點：** `POST /matches/{matchId}/result`

**參數：**
- `matchId` (path, required): 比賽ID

**請求體：**
```json
{
  "winnerId": "user-123",
  "loserId": "user-456",
  "score": "6-4, 6-2"
}
```

**響應：**
```json
{
  "message": "比賽結果已記錄"
}
```

### 4. 確認比賽結果

確認比賽結果的準確性。

**端點：** `POST /results/{resultId}/confirm`

**參數：**
- `resultId` (path, required): 比賽結果ID

**響應：**
```json
{
  "message": "比賽結果已確認"
}
```

### 5. 獲取待確認的比賽結果

獲取用戶需要確認的比賽結果列表。

**端點：** `GET /pending-confirmations`

**響應：**
```json
[
  {
    "id": "result-123",
    "matchId": "match-456",
    "winnerId": "user-123",
    "loserId": "user-456",
    "score": "6-4, 6-2",
    "isConfirmed": false,
    "confirmedBy": ["user-123"],
    "match": {...},
    "winner": {...},
    "loser": {...}
  }
]
```

### 6. 獲取技術等級進展

獲取用戶的技術等級變化歷史。

**端點：** `GET /users/{userId}/skill-progression`

**參數：**
- `userId` (path, required): 用戶ID

**響應：**
```json
[
  {
    "id": "record-123",
    "userId": "user-123",
    "oldLevel": 3.0,
    "newLevel": 3.5,
    "reason": "auto_adjustment",
    "matchId": "match-456",
    "createdAt": "2024-11-15T10:00:00Z"
  }
]
```

### 7. 手動調整技術等級

手動調整用戶的NTRP技術等級。

**端點：** `POST /users/{userId}/adjust-skill-level`

**參數：**
- `userId` (path, required): 用戶ID

**請求體：**
```json
{
  "newLevel": 4.0,
  "reason": "教練評估調整"
}
```

**響應：**
```json
{
  "message": "技術等級已調整"
}
```

### 8. 獲取用戶隱私設定

獲取用戶的統計資訊隱私設定。

**端點：** `GET /privacy-settings`

**響應：**
```json
{
  "userId": "user-123",
  "showReputationScore": true,
  "showMatchHistory": true,
  "showWinLossRecord": true,
  "showSkillProgression": true,
  "showBehaviorReviews": false,
  "showDetailedStats": true,
  "allowStatisticsSharing": false
}
```

### 9. 更新用戶隱私設定

更新用戶的統計資訊隱私設定。

**端點：** `PUT /privacy-settings`

**請求體：**
```json
{
  "showReputationScore": true,
  "showMatchHistory": false,
  "showWinLossRecord": true,
  "showSkillProgression": true,
  "showBehaviorReviews": false,
  "showDetailedStats": false,
  "allowStatisticsSharing": false
}
```

**響應：**
```json
{
  "message": "隱私設定已更新"
}
```

### 10. 根據隱私設定獲取信譽分數

根據用戶隱私設定獲取信譽分數。

**端點：** `GET /users/{userId}/reputation`

**參數：**
- `userId` (path, required): 用戶ID

**響應：**
```json
{
  "id": "reputation-123",
  "userId": "user-123",
  "attendanceRate": 85.5,
  "punctualityScore": 92.3,
  "skillAccuracy": 78.9,
  "behaviorRating": 4.2,
  "totalMatches": 45,
  "completedMatches": 38,
  "cancelledMatches": 7,
  "overallScore": 86.7
}
```

### 11. 獲取配對統計摘要

獲取用戶配對統計的簡要摘要。

**端點：** `GET /summary`

**響應：**
```json
{
  "totalMatches": 45,
  "monthlyMatches": 8,
  "winRate": 57.89,
  "reputationScore": 86.7,
  "ntrpLevel": 3.5
}
```

## 錯誤響應

所有端點可能返回以下錯誤：

### 400 Bad Request
```json
{
  "error": "INVALID_REQUEST",
  "message": "請求參數無效",
  "details": "具體錯誤信息"
}
```

### 401 Unauthorized
```json
{
  "error": "UNAUTHORIZED",
  "message": "未授權的請求"
}
```

### 403 Forbidden
```json
{
  "error": "PRIVATE_STATISTICS",
  "message": "用戶統計資訊為私人設定"
}
```

### 404 Not Found
```json
{
  "error": "NOT_FOUND",
  "message": "資源不存在"
}
```

### 500 Internal Server Error
```json
{
  "error": "INTERNAL_ERROR",
  "message": "服務器內部錯誤",
  "details": "具體錯誤信息"
}
```

## 隱私控制

用戶可以通過隱私設定控制以下資訊的可見性：

1. **信譽分數** (`showReputationScore`): 控制其他用戶是否能看到信譽分數
2. **配對歷史** (`showMatchHistory`): 控制配對歷史記錄的可見性
3. **勝負記錄** (`showWinLossRecord`): 控制勝率和勝負統計的可見性
4. **技術進展** (`showSkillProgression`): 控制技術等級變化歷史的可見性
5. **行為評價** (`showBehaviorReviews`): 控制其他用戶評價的可見性
6. **詳細統計** (`showDetailedStats`): 控制詳細統計資訊的可見性
7. **統計分享** (`allowStatisticsSharing`): 控制是否允許統計資訊被分享

## 自動技術等級調整

系統會根據以下因素自動調整用戶的NTRP技術等級：

1. **技術準確度記錄**: 基於其他用戶對技術水平的評估
2. **比賽勝率**: 最近比賽的勝負表現
3. **對手等級**: 與不同等級對手的比賽結果

調整規則：
- 需要至少5次技術評估記錄
- 建議等級與當前等級差距超過0.3時才進行調整
- 每次調整幅度限制在差距的30%以內
- 等級範圍限制在1.0-7.0之間

## 使用示例

### 獲取自己的統計資訊
```bash
curl -X GET "http://localhost:8080/api/v1/match-statistics/users/user-123" \
  -H "Authorization: Bearer your-jwt-token"
```

### 記錄比賽結果
```bash
curl -X POST "http://localhost:8080/api/v1/match-statistics/matches/match-456/result" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "winnerId": "user-123",
    "loserId": "user-456",
    "score": "6-4, 6-2"
  }'
```

### 更新隱私設定
```bash
curl -X PUT "http://localhost:8080/api/v1/match-statistics/privacy-settings" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "showReputationScore": true,
    "showMatchHistory": false,
    "showWinLossRecord": true
  }'
```