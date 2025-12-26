#!/bin/bash

# 網球平台教練 API 測試腳本

BASE_URL="http://localhost:8080/api/v1"
CONTENT_TYPE="Content-Type: application/json"

echo "=== 網球平台教練 API 測試 ==="
echo

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 測試函數
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4
    local description=$5
    
    echo -e "${BLUE}測試: $description${NC}"
    echo "請求: $method $endpoint"
    
    if [ -n "$data" ]; then
        echo "數據: $data"
    fi
    
    if [ -n "$headers" ]; then
        response=$(curl -s -X $method "$BASE_URL$endpoint" -H "$CONTENT_TYPE" -H "$headers" -d "$data")
    else
        response=$(curl -s -X $method "$BASE_URL$endpoint" -H "$CONTENT_TYPE" -d "$data")
    fi
    
    echo "響應: $response"
    echo
    
    # 檢查響應是否包含錯誤
    if echo "$response" | grep -q '"error"'; then
        echo -e "${RED}❌ 測試失敗${NC}"
    else
        echo -e "${GREEN}✅ 測試成功${NC}"
    fi
    echo "----------------------------------------"
}

# 1. 測試獲取教練專長選項
test_endpoint "GET" "/coaches/specialties" "" "" "獲取教練專長選項"

# 2. 測試獲取教練認證選項
test_endpoint "GET" "/coaches/certifications" "" "" "獲取教練認證選項"

# 3. 測試獲取可用語言選項
test_endpoint "GET" "/coaches/languages" "" "" "獲取可用語言選項"

# 4. 測試獲取可用貨幣選項
test_endpoint "GET" "/coaches/currencies" "" "" "獲取可用貨幣選項"

# 5. 測試搜尋教練（無參數）
test_endpoint "GET" "/coaches" "" "" "搜尋教練（無參數）"

# 6. 測試搜尋教練（帶參數）
test_endpoint "GET" "/coaches?specialties=beginner&minExperience=1&maxHourlyRate=2000&isVerified=true&page=1&limit=10" "" "" "搜尋教練（帶參數）"

# 7. 測試獲取特定教練（使用假ID）
test_endpoint "GET" "/coaches/123e4567-e89b-12d3-a456-426614174000" "" "" "獲取特定教練"

echo -e "${YELLOW}注意: 以下測試需要有效的認證令牌${NC}"
echo

# 模擬認證令牌（實際使用時需要真實的JWT令牌）
AUTH_TOKEN="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

# 8. 測試創建教練檔案（需要認證）
COACH_DATA='{
    "experience": 5,
    "specialties": ["beginner", "intermediate"],
    "biography": "專業網球教練，擁有5年教學經驗",
    "hourlyRate": 1500,
    "currency": "TWD",
    "languages": ["zh-TW", "en"],
    "certifications": ["PTR", "CTCA"],
    "availableHours": {
        "monday": ["09:00-12:00", "14:00-18:00"],
        "tuesday": ["09:00-12:00", "14:00-18:00"],
        "wednesday": ["09:00-12:00"],
        "friday": ["14:00-18:00"],
        "saturday": ["09:00-17:00"]
    }
}'

test_endpoint "POST" "/coaches" "$COACH_DATA" "Authorization: $AUTH_TOKEN" "創建教練檔案（需要認證）"

# 9. 測試獲取我的教練檔案（需要認證）
test_endpoint "GET" "/coaches/my-profile" "" "Authorization: $AUTH_TOKEN" "獲取我的教練檔案（需要認證）"

# 10. 測試更新教練檔案（需要認證）
UPDATE_DATA='{
    "biography": "更新後的教練簡介",
    "hourlyRate": 1800,
    "languages": ["zh-TW", "en", "ja"]
}'

test_endpoint "PUT" "/coaches/123e4567-e89b-12d3-a456-426614174000" "$UPDATE_DATA" "Authorization: $AUTH_TOKEN" "更新教練檔案（需要認證）"

# 11. 測試教練認證（需要管理員權限）
VERIFY_DATA='{
    "coachId": "123e4567-e89b-12d3-a456-426614174000",
    "isVerified": true,
    "verificationNotes": "已驗證教練資格證書"
}'

test_endpoint "POST" "/coaches/verify" "$VERIFY_DATA" "Authorization: $AUTH_TOKEN" "教練認證（需要管理員權限）"

echo -e "${GREEN}=== 教練 API 測試完成 ===${NC}"
echo
echo -e "${YELLOW}注意事項:${NC}"
echo "1. 需要認證的 API 需要有效的 JWT 令牌"
echo "2. 某些測試可能因為數據不存在而失敗，這是正常的"
echo "3. 在實際使用前，請確保數據庫中有相應的測試數據"
echo "4. 教練認證功能需要管理員權限"