# 聊天系統 API 文檔

## 概述

聊天系統提供即時通訊功能，支援直接聊天、群組聊天和配對聊天。系統使用 WebSocket 進行即時通訊，同時提供 REST API 進行聊天室和訊息管理。

## 認證

所有聊天 API 都需要 JWT 認證。在請求頭中包含：
```
Authorization: Bearer <access_token>
```

## WebSocket 連接

### 建立 WebSocket 連接

**端點:** `GET /api/v1/chat/ws`

**描述:** 建立 WebSocket 連接用於即時通訊

**認證:** 需要 JWT Token

**WebSocket 訊息格式:**

```json
{
  "type": "message_type",
  "chatRoomId": "room_id",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**支援的訊息類型:**

- `join_room`: 加入聊天室
- `leave_room`: 離開聊天室
- `new_message`: 新訊息
- `messages_read`: 訊息已讀
- `user_joined`: 用戶加入
- `user_left`: 用戶離開
- `ping/pong`: 心跳檢測

## REST API

### 聊天室管理

#### 創建聊天室

**端點:** `POST /api/v1/chat/rooms`

**描述:** 創建新的聊天室

**請求體:**
```json
{
  "matchId": "match_id",          // 可選，配對ID
  "type": "direct",               // 必需：direct, group, match
  "name": "聊天室名稱",            // 可選，群組聊天室名稱
  "participantIds": ["user_id"]   // 必需，參與者ID列表
}
```

**響應:** `201 Created`
```json
{
  "id": "room_id",
  "matchId": "match_id",
  "type": "direct",
  "name": "聊天室名稱",
  "isActive": true,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "participants": [
    {
      "id": "participant_id",
      "userId": "user_id",
      "joinedAt": "2024-01-01T00:00:00Z",
      "lastReadAt": "2024-01-01T00:00:00Z",
      "isActive": true,
      "user": {
        "id": "user_id",
        "firstName": "John",
        "lastName": "Doe",
        "avatarUrl": "https://example.com/avatar.jpg"
      }
    }
  ],
  "lastMessage": {
    "id": "message_id",
    "chatRoomId": "room_id",
    "senderId": "user_id",
    "content": "Hello!",
    "messageType": "text",
    "isRead": false,
    "createdAt": "2024-01-01T00:00:00Z",
    "sender": {
      "id": "user_id",
      "firstName": "John",
      "lastName": "Doe",
      "avatarUrl": "https://example.com/avatar.jpg"
    }
  },
  "unreadCount": 5
}
```

#### 獲取聊天室列表

**端點:** `GET /api/v1/chat/rooms`

**描述:** 獲取用戶參與的所有聊天室

**響應:** `200 OK`
```json
[
  {
    "id": "room_id",
    "type": "direct",
    "name": "聊天室名稱",
    "isActive": true,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z",
    "participants": [...],
    "lastMessage": {...},
    "unreadCount": 5
  }
]
```

#### 獲取聊天室詳情

**端點:** `GET /api/v1/chat/rooms/{roomId}`

**描述:** 獲取指定聊天室的詳細信息

**路徑參數:**
- `roomId`: 聊天室ID

**響應:** `200 OK` - 同創建聊天室響應格式

#### 加入聊天室 (WebSocket)

**端點:** `POST /api/v1/chat/rooms/{roomId}/join`

**描述:** 加入指定的聊天室 WebSocket 頻道

**路徑參數:**
- `roomId`: 聊天室ID

**響應:** `200 OK`
```json
{
  "message": "加入聊天室成功"
}
```

#### 離開聊天室

**端點:** `POST /api/v1/chat/rooms/{roomId}/leave`

**描述:** 離開指定的聊天室

**路徑參數:**
- `roomId`: 聊天室ID

**響應:** `200 OK`
```json
{
  "message": "離開聊天室成功"
}
```

### 訊息管理

#### 發送訊息

**端點:** `POST /api/v1/chat/messages`

**描述:** 向聊天室發送訊息

**請求體:**
```json
{
  "chatRoomId": "room_id",        // 必需
  "content": "Hello, world!",     // 必需
  "messageType": "text"           // 可選：text, image, file
}
```

**響應:** `201 Created`
```json
{
  "id": "message_id",
  "chatRoomId": "room_id",
  "senderId": "user_id",
  "content": "Hello, world!",
  "messageType": "text",
  "isRead": false,
  "createdAt": "2024-01-01T00:00:00Z",
  "sender": {
    "id": "user_id",
    "firstName": "John",
    "lastName": "Doe",
    "avatarUrl": "https://example.com/avatar.jpg"
  }
}
```

#### 獲取聊天室訊息

**端點:** `GET /api/v1/chat/rooms/{roomId}/messages`

**描述:** 獲取指定聊天室的訊息列表

**路徑參數:**
- `roomId`: 聊天室ID

**查詢參數:**
- `page`: 頁碼 (默認: 1)
- `limit`: 每頁數量 (默認: 50, 最大: 100)
- `before`: 獲取此時間之前的訊息 (RFC3339格式)

**響應:** `200 OK`
```json
[
  {
    "id": "message_id",
    "chatRoomId": "room_id",
    "senderId": "user_id",
    "content": "Hello, world!",
    "messageType": "text",
    "isRead": false,
    "createdAt": "2024-01-01T00:00:00Z",
    "sender": {
      "id": "user_id",
      "firstName": "John",
      "lastName": "Doe",
      "avatarUrl": "https://example.com/avatar.jpg"
    }
  }
]
```

#### 標記訊息為已讀

**端點:** `POST /api/v1/chat/rooms/{roomId}/read`

**描述:** 標記聊天室中的訊息為已讀

**路徑參數:**
- `roomId`: 聊天室ID

**響應:** `200 OK`
```json
{
  "message": "標記已讀成功"
}
```

### 在線狀態

#### 獲取在線用戶

**端點:** `GET /api/v1/chat/online-users`

**描述:** 獲取當前在線的用戶列表

**響應:** `200 OK`
```json
{
  "onlineUsers": ["user_id1", "user_id2"],
  "count": 2
}
```

#### 獲取聊天室在線用戶

**端點:** `GET /api/v1/chat/rooms/{roomId}/online-users`

**描述:** 獲取指定聊天室的在線用戶列表

**路徑參數:**
- `roomId`: 聊天室ID

**響應:** `200 OK`
```json
{
  "roomUsers": ["user_id1", "user_id2"],
  "count": 2
}
```

## 錯誤響應

### 常見錯誤狀態碼

- `400 Bad Request`: 請求格式錯誤
- `401 Unauthorized`: 未授權
- `403 Forbidden`: 無權限訪問
- `404 Not Found`: 資源不存在
- `500 Internal Server Error`: 服務器內部錯誤

### 錯誤響應格式

```json
{
  "error": "錯誤描述",
  "details": "詳細錯誤信息"
}
```

## WebSocket 事件

### 客戶端發送事件

#### 加入聊天室
```json
{
  "type": "join_room",
  "data": "room_id"
}
```

#### 離開聊天室
```json
{
  "type": "leave_room",
  "data": "room_id"
}
```

#### 心跳檢測
```json
{
  "type": "ping"
}
```

### 服務器發送事件

#### 新訊息
```json
{
  "type": "new_message",
  "chatRoomId": "room_id",
  "data": {
    "id": "message_id",
    "chatRoomId": "room_id",
    "senderId": "user_id",
    "content": "Hello!",
    "messageType": "text",
    "createdAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 訊息已讀
```json
{
  "type": "messages_read",
  "chatRoomId": "room_id",
  "data": {
    "userId": "user_id",
    "chatRoomId": "room_id",
    "readAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 用戶加入
```json
{
  "type": "user_joined",
  "chatRoomId": "room_id",
  "data": {
    "userId": "user_id",
    "chatRoomId": "room_id",
    "joinedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 用戶離開
```json
{
  "type": "user_left",
  "chatRoomId": "room_id",
  "data": {
    "userId": "user_id",
    "chatRoomId": "room_id",
    "leftAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 心跳響應
```json
{
  "type": "pong",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 使用示例

### JavaScript WebSocket 客戶端

```javascript
// 建立 WebSocket 連接
const ws = new WebSocket('ws://localhost:8080/api/v1/chat/ws', [], {
  headers: {
    'Authorization': 'Bearer ' + accessToken
  }
});

// 監聽連接開啟
ws.onopen = function(event) {
  console.log('WebSocket 連接已建立');
  
  // 加入聊天室
  ws.send(JSON.stringify({
    type: 'join_room',
    data: 'room_id'
  }));
};

// 監聽訊息
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  switch(message.type) {
    case 'new_message':
      console.log('收到新訊息:', message.data);
      break;
    case 'user_joined':
      console.log('用戶加入:', message.data.userId);
      break;
    case 'user_left':
      console.log('用戶離開:', message.data.userId);
      break;
  }
};

// 監聽錯誤
ws.onerror = function(error) {
  console.error('WebSocket 錯誤:', error);
};

// 監聽連接關閉
ws.onclose = function(event) {
  console.log('WebSocket 連接已關閉');
};
```

### 發送訊息示例

```javascript
// 通過 REST API 發送訊息
fetch('/api/v1/chat/messages', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + accessToken
  },
  body: JSON.stringify({
    chatRoomId: 'room_id',
    content: 'Hello, world!',
    messageType: 'text'
  })
})
.then(response => response.json())
.then(data => {
  console.log('訊息發送成功:', data);
})
.catch(error => {
  console.error('發送訊息失敗:', error);
});
```

## 最佳實踐

1. **連接管理**: 實現自動重連機制處理網絡中斷
2. **心跳檢測**: 定期發送 ping 訊息保持連接活躍
3. **錯誤處理**: 妥善處理各種錯誤情況
4. **訊息去重**: 客戶端應實現訊息去重邏輯
5. **離線處理**: 處理用戶離線時的訊息同步
6. **性能優化**: 合理使用分頁避免一次載入過多訊息

## 限制

- 每個用戶最多同時建立 5 個 WebSocket 連接
- 單條訊息最大長度 1000 字符
- 每分鐘最多發送 60 條訊息
- 聊天室最多 100 個參與者
- 訊息歷史保留 90 天