#!/bin/bash

# OAuth API 測試腳本
BASE_URL="http://localhost:8080/api/v1"

echo "=== 網球平台 OAuth API 測試 ==="
echo

# 測試獲取 Google OAuth 授權 URL
echo "1. 測試獲取 Google OAuth 授權 URL"
curl -X GET "$BASE_URL/auth/oauth/google" \
  -H "Content-Type: application/json" | jq .
echo -e "\n"

# 測試獲取 Facebook OAuth 授權 URL
echo "2. 測試獲取 Facebook OAuth 授權 URL"
curl -X GET "$BASE_URL/auth/oauth/facebook" \
  -H "Content-Type: application/json" | jq .
echo -e "\n"

# 測試獲取 Apple OAuth 授權 URL
echo "3. 測試獲取 Apple OAuth 授權 URL"
curl -X GET "$BASE_URL/auth/oauth/apple" \
  -H "Content-Type: application/json" | jq .
echo -e "\n"

# 註冊測試用戶以測試 OAuth 帳號管理功能
echo "4. 註冊測試用戶"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "oauth-test@example.com",
    "password": "password123",
    "firstName": "OAuth",
    "lastName": "Test"
  }')

echo $REGISTER_RESPONSE | jq .

# 提取訪問令牌
ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.accessToken')

if [ "$ACCESS_TOKEN" != "null" ] && [ "$ACCESS_TOKEN" != "" ]; then
    echo -e "\n5. 測試獲取已關聯的 OAuth 帳號（需要認證）"
    curl -X GET "$BASE_URL/auth/oauth/accounts" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
    echo -e "\n"
    
    echo "6. 測試解除關聯不存在的 OAuth 帳號（應該失敗）"
    curl -X DELETE "$BASE_URL/auth/oauth/google/unlink" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
    echo -e "\n"
else
    echo "無法獲取訪問令牌，跳過需要認證的測試"
fi

echo "=== OAuth API 測試完成 ==="