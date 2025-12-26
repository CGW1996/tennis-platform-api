#!/bin/bash

# 網球平台用戶檔案管理 API 測試腳本

BASE_URL="http://localhost:8080/api/v1"
EMAIL="testuser@example.com"
PASSWORD="testpassword123"

echo "=== 網球平台用戶檔案管理 API 測試 ==="

# 1. 用戶註冊
echo "1. 註冊新用戶..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "註冊響應: $REGISTER_RESPONSE"

# 2. 用戶登入
echo -e "\n2. 用戶登入..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "登入響應: $LOGIN_RESPONSE"

# 提取 access token
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  echo "錯誤: 無法獲取 access token"
  exit 1
fi

echo "Access Token: $ACCESS_TOKEN"

# 3. 獲取 NTRP 等級列表
echo -e "\n3. 獲取 NTRP 等級列表..."
curl -s -X GET "$BASE_URL/users/ntrp-levels" \
  -H "Content-Type: application/json" | jq '.'

# 4. 創建用戶檔案
echo -e "\n4. 創建用戶檔案..."
CREATE_PROFILE_RESPONSE=$(curl -s -X POST "$BASE_URL/users/profile" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "firstName": "測試",
    "lastName": "用戶",
    "ntrpLevel": 3.5,
    "playingStyle": "aggressive",
    "preferredHand": "right",
    "latitude": 25.0330,
    "longitude": 121.5654,
    "bio": "我是一個熱愛網球的玩家",
    "gender": "male",
    "playingFrequency": "regular",
    "preferredTimes": ["morning", "evening"],
    "maxTravelDistance": 10.0,
    "profilePrivacy": "public"
  }')

echo "創建檔案響應: $CREATE_PROFILE_RESPONSE"

# 5. 獲取用戶檔案
echo -e "\n5. 獲取用戶檔案..."
curl -s -X GET "$BASE_URL/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'

# 6. 更新用戶檔案
echo -e "\n6. 更新用戶檔案..."
UPDATE_PROFILE_RESPONSE=$(curl -s -X PUT "$BASE_URL/users/profile" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "firstName": "更新的",
    "ntrpLevel": 4.0,
    "bio": "我是一個更新後的網球愛好者"
  }')

echo "更新檔案響應: $UPDATE_PROFILE_RESPONSE"

# 7. 更新用戶偏好設定
echo -e "\n7. 更新用戶偏好設定..."
UPDATE_PREFERENCES_RESPONSE=$(curl -s -X PUT "$BASE_URL/users/preferences" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "playingStyle": "defensive",
    "playingFrequency": "competitive",
    "preferredTimes": ["afternoon", "evening"],
    "maxTravelDistance": 15.0,
    "profilePrivacy": "friends"
  }')

echo "更新偏好響應: $UPDATE_PREFERENCES_RESPONSE"

# 8. 更新用戶位置
echo -e "\n8. 更新用戶位置..."
UPDATE_LOCATION_RESPONSE=$(curl -s -X PUT "$BASE_URL/users/location" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "latitude": 25.0478,
    "longitude": 121.5319,
    "locationPrivacy": true
  }')

echo "更新位置響應: $UPDATE_LOCATION_RESPONSE"

# 9. 最終獲取用戶檔案確認更新
echo -e "\n9. 最終獲取用戶檔案..."
curl -s -X GET "$BASE_URL/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'

echo -e "\n=== 測試完成 ==="