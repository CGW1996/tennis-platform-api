#!/bin/bash

# 網球場地 API 測試腳本

BASE_URL="http://localhost:8080/api/v1"

echo "=== 網球場地 API 測試 ==="
echo

# 測試獲取場地類型
echo "1. 測試獲取場地類型"
curl -s -X GET "$BASE_URL/courts/types" | jq '.'
echo
echo

# 測試獲取可用設施
echo "2. 測試獲取可用設施"
curl -s -X GET "$BASE_URL/courts/facilities" | jq '.'
echo
echo

# 測試搜尋場地（無參數）
echo "3. 測試搜尋場地（無參數）"
curl -s -X GET "$BASE_URL/courts" | jq '.'
echo
echo

# 測試搜尋場地（帶地理位置）
echo "4. 測試搜尋場地（帶地理位置）"
curl -s -X GET "$BASE_URL/courts?latitude=25.0330&longitude=121.5654&radius=10&sortBy=distance" | jq '.'
echo
echo

# 測試搜尋場地（帶價格篩選）
echo "5. 測試搜尋場地（帶價格篩選）"
curl -s -X GET "$BASE_URL/courts?minPrice=300&maxPrice=700&sortBy=price&sortOrder=asc" | jq '.'
echo
echo

# 測試搜尋場地（帶場地類型篩選）
echo "6. 測試搜尋場地（帶場地類型篩選）"
curl -s -X GET "$BASE_URL/courts?courtType=hard&sortBy=rating&sortOrder=desc" | jq '.'
echo
echo

# 獲取第一個場地的ID用於後續測試
COURT_ID=$(curl -s -X GET "$BASE_URL/courts?pageSize=1" | jq -r '.courts[0].id // empty')

if [ -n "$COURT_ID" ]; then
    echo "7. 測試獲取場地詳情（ID: $COURT_ID）"
    curl -s -X GET "$BASE_URL/courts/$COURT_ID" | jq '.'
    echo
    echo
else
    echo "7. 無法獲取場地ID，跳過場地詳情測試"
    echo
fi

# 測試需要認證的功能（需要先登入獲取 token）
echo "8. 測試需要認證的功能"
echo "   註：以下功能需要先登入獲取 JWT token"
echo

# 測試用戶登入
echo "   8.1 測試用戶登入"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@tennis-platform.com",
    "password": "password123"
  }')

echo "$LOGIN_RESPONSE" | jq '.'

# 提取 access token
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.accessToken // empty')

if [ -n "$ACCESS_TOKEN" ]; then
    echo
    echo "   8.2 測試創建場地（需要認證）"
    CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/courts" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -d '{
        "name": "測試網球場",
        "description": "這是一個測試場地",
        "address": "台北市測試區測試路123號",
        "latitude": 25.0500,
        "longitude": 121.5500,
        "facilities": ["parking", "restroom", "lighting"],
        "courtType": "hard",
        "pricePerHour": 500,
        "currency": "TWD",
        "operatingHours": {
          "monday": "08:00-20:00",
          "tuesday": "08:00-20:00",
          "wednesday": "08:00-20:00",
          "thursday": "08:00-20:00",
          "friday": "08:00-20:00",
          "saturday": "08:00-20:00",
          "sunday": "08:00-20:00"
        },
        "contactPhone": "+886-2-1234-5678",
        "contactEmail": "test@example.com"
      }')
    
    echo "$CREATE_RESPONSE" | jq '.'
    
    # 提取新創建場地的ID
    NEW_COURT_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id // empty')
    
    if [ -n "$NEW_COURT_ID" ]; then
        echo
        echo "   8.3 測試更新場地（ID: $NEW_COURT_ID）"
        curl -s -X PUT "$BASE_URL/courts/$NEW_COURT_ID" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $ACCESS_TOKEN" \
          -d '{
            "name": "更新後的測試網球場",
            "pricePerHour": 600,
            "facilities": ["parking", "restroom", "lighting", "wifi"]
          }' | jq '.'
        
        echo
        echo "   8.4 測試刪除場地（ID: $NEW_COURT_ID）"
        curl -s -X DELETE "$BASE_URL/courts/$NEW_COURT_ID" \
          -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
    fi
else
    echo "   無法獲取 access token，跳過需要認證的測試"
fi

echo
echo "=== 測試完成 ==="