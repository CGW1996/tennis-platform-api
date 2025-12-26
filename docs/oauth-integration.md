# OAuth 社交登入整合文件

## 概述

本文件描述了網球平台 OAuth 社交登入功能的實作，支援 Google、Facebook 和 Apple 三種主要的社交登入提供商。

## 支援的提供商

### 1. Google OAuth 2.0
- **授權端點**: `https://accounts.google.com/oauth2/auth`
- **令牌端點**: `https://oauth2.googleapis.com/token`
- **用戶資訊端點**: `https://www.googleapis.com/oauth2/v2/userinfo`
- **所需範圍**: `openid`, `profile`, `email`

### 2. Facebook Login
- **授權端點**: `https://www.facebook.com/v18.0/dialog/oauth`
- **令牌端點**: `https://graph.facebook.com/v18.0/oauth/access_token`
- **用戶資訊端點**: `https://graph.facebook.com/me`
- **所需範圍**: `email`, `public_profile`

### 3. Apple Sign In
- **授權端點**: `https://appleid.apple.com/auth/authorize`
- **令牌端點**: `https://appleid.apple.com/auth/token`
- **所需範圍**: `name`, `email`

## API 端點

### 1. 獲取 OAuth 授權 URL
```http
GET /api/v1/auth/oauth/{provider}
```

**參數:**
- `provider`: OAuth 提供商 (`google`, `facebook`, `apple`)

**響應:**
```json
{
  "authUrl": "https://accounts.google.com/oauth2/auth?client_id=..."
}
```

### 2. OAuth 回調處理
```http
POST /api/v1/auth/oauth/{provider}/callback
```

**請求體:**
```json
{
  "code": "授權碼",
  "state": "狀態參數"
}
```

**響應:**
```json
{
  "user": {
    "id": "用戶ID",
    "email": "用戶郵箱",
    "profile": {
      "firstName": "名字",
      "lastName": "姓氏",
      "avatarUrl": "頭像URL"
    }
  },
  "accessToken": "JWT訪問令牌",
  "refreshToken": "刷新令牌"
}
```

### 3. 關聯 OAuth 帳號 (需要認證)
```http
POST /api/v1/auth/oauth/{provider}/link
```

**請求頭:**
```
Authorization: Bearer {accessToken}
```

**請求體:**
```json
{
  "code": "授權碼",
  "state": "狀態參數"
}
```

### 4. 解除關聯 OAuth 帳號 (需要認證)
```http
DELETE /api/v1/auth/oauth/{provider}/unlink
```

**請求頭:**
```
Authorization: Bearer {accessToken}
```

### 5. 獲取已關聯的 OAuth 帳號 (需要認證)
```http
GET /api/v1/auth/oauth/accounts
```

**請求頭:**
```
Authorization: Bearer {accessToken}
```

**響應:**
```json
{
  "accounts": [
    {
      "id": "關聯ID",
      "provider": "google",
      "email": "oauth@example.com",
      "createdAt": "2023-01-01T00:00:00Z"
    }
  ]
}
```

## 環境配置

在 `.env` 文件中配置 OAuth 提供商的憑證：

```env
# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/google/callback

# Facebook OAuth
FACEBOOK_CLIENT_ID=your-facebook-client-id
FACEBOOK_CLIENT_SECRET=your-facebook-client-secret
FACEBOOK_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/facebook/callback

# Apple OAuth
APPLE_CLIENT_ID=your-apple-client-id
APPLE_CLIENT_SECRET=your-apple-client-secret
APPLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/apple/callback
```

## 數據庫結構

### oauth_accounts 表
```sql
CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(provider, provider_id)
);
```

## 業務邏輯

### 1. OAuth 登入流程

1. **獲取授權 URL**: 客戶端調用 `/auth/oauth/{provider}` 獲取授權 URL
2. **用戶授權**: 用戶在提供商頁面完成授權
3. **回調處理**: 提供商重定向到回調 URL，客戶端獲取授權碼
4. **令牌交換**: 客戶端調用 `/auth/oauth/{provider}/callback` 交換令牌
5. **用戶創建/登入**: 系統根據 OAuth 資訊創建新用戶或登入現有用戶

### 2. 帳號關聯邏輯

- **新用戶**: 如果 OAuth 郵箱不存在，創建新用戶並關聯 OAuth 帳號
- **現有用戶**: 如果郵箱已存在，將 OAuth 帳號關聯到現有用戶
- **已關聯**: 如果 OAuth 帳號已關聯，直接登入

### 3. 帳號合併規則

- 同一郵箱的多個 OAuth 帳號會自動關聯到同一用戶
- 用戶可以手動關聯多個 OAuth 提供商
- 解除關聯時會檢查用戶是否有其他登入方式

### 4. 安全考慮

- **狀態參數驗證**: 防止 CSRF 攻擊
- **令牌安全存儲**: OAuth 令牌加密存儲
- **權限檢查**: 關聯/解除關聯需要用戶認證
- **帳號保護**: 防止惡意解除關聯導致用戶無法登入

## 錯誤處理

### 常見錯誤碼

- `400 Bad Request`: 請求參數錯誤
- `401 Unauthorized`: 未授權或令牌無效
- `404 Not Found`: 提供商不支援
- `409 Conflict`: 帳號已關聯或衝突

### 錯誤響應格式
```json
{
  "error": "錯誤描述",
  "details": "詳細錯誤信息"
}
```

## 測試

### 單元測試
```bash
go test ./internal/services -v -run TestOAuth
```

### 整合測試
```bash
./test_oauth_api.sh
```

### 測試覆蓋範圍
- OAuth 服務功能測試
- 授權 URL 生成測試
- 狀態參數驗證測試
- 錯誤處理測試

## 部署注意事項

1. **回調 URL 配置**: 確保在各提供商控制台配置正確的回調 URL
2. **HTTPS 要求**: 生產環境必須使用 HTTPS
3. **域名驗證**: 確保域名在提供商白名單中
4. **憑證安全**: 妥善保管 OAuth 憑證，不要提交到版本控制

## 未來擴展

1. **更多提供商**: 支援 Twitter、LinkedIn 等
2. **企業登入**: 支援 Microsoft Azure AD、SAML
3. **多因素認證**: 整合 2FA 功能
4. **帳號遷移**: 支援帳號合併和遷移工具

## 故障排除

### 常見問題

1. **授權失敗**: 檢查客戶端 ID 和回調 URL 配置
2. **令牌過期**: 實作令牌刷新機制
3. **用戶資訊獲取失敗**: 檢查 API 權限和範圍設定
4. **帳號關聯失敗**: 檢查數據庫約束和業務邏輯

### 日誌監控

- OAuth 授權流程日誌
- 令牌交換成功/失敗統計
- 用戶關聯/解除關聯操作記錄
- 錯誤率和響應時間監控