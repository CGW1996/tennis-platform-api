#!/bin/bash

# 智能排課 API 測試腳本

BASE_URL="http://localhost:8080/api/v1"
CONTENT_TYPE="Content-Type: application/json"

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 測試結果統計
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 測試函數
test_api() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"
    local auth_header="$6"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}測試: $test_name${NC}"
    echo "請求: $method $endpoint"
    
    if [ -n "$data" ]; then
        echo "數據: $data"
    fi
    
    # 構建 curl 命令
    curl_cmd="curl -s -w \"HTTP_STATUS:%{http_code}\" -X $method \"$BASE_URL$endpoint\""
    
    if [ -n "$auth_header" ]; then
        curl_cmd="$curl_cmd -H \"Authorization: $auth_header\""
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H \"$CONTENT_TYPE\" -d '$data'"
    fi
    
    # 執行請求
    response=$(eval $curl_cmd)
    
    # 提取狀態碼
    http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
    response_body=$(echo "$response" | sed 's/HTTP_STATUS:[0-9]*$//')
    
    echo "狀態碼: $http_status"
    
    # 格式化 JSON 響應
    if [ -n "$response_body" ] && echo "$response_body" | jq . >/dev/null 2>&1; then
        echo "響應:"
        echo "$response_body" | jq .
    else
        echo "響應: $response_body"
    fi
    
    # 檢查狀態碼
    if [ "$http_status" = "$expected_status" ]; then
        echo -e "${GREEN}✓ 測試通過${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ 測試失敗 (期望: $expected_status, 實際: $http_status)${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    echo "----------------------------------------"
}

# 檢查服務器是否運行
echo -e "${YELLOW}檢查服務器狀態...${NC}"
if ! curl -s "$BASE_URL/../health" > /dev/null; then
    echo -e "${RED}錯誤: 服務器未運行，請先啟動服務器${NC}"
    exit 1
fi

echo -e "${GREEN}服務器運行正常${NC}"
echo "========================================"

# 全局變量存儲認證令牌
ACCESS_TOKEN=""
COACH_ID=""
STUDENT_ID=""

# 1. 用戶註冊和登入（獲取認證令牌）
echo -e "${YELLOW}步驟 1: 用戶認證${NC}"

# 註冊學生用戶
test_api "註冊學生用戶" "POST" "/auth/register" '{
    "email": "student@example.com",
    "password": "password123",
    "confirmPassword": "password123"
}' "201"

# 學生登入
response=$(curl -s -w "HTTP_STATUS:%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "$CONTENT_TYPE" \
    -d '{
        "email": "student@example.com",
        "password": "password123"
    }')

http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
response_body=$(echo "$response" | sed 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_status" = "200" ]; then
    ACCESS_TOKEN=$(echo "$response_body" | jq -r '.accessToken')
    STUDENT_ID=$(echo "$response_body" | jq -r '.user.id')
    echo -e "${GREEN}學生登入成功，獲取到令牌${NC}"
else
    echo -e "${RED}學生登入失敗${NC}"
    exit 1
fi

# 註冊教練用戶
test_api "註冊教練用戶" "POST" "/auth/register" '{
    "email": "coach@example.com",
    "password": "password123",
    "confirmPassword": "password123"
}' "201"

# 教練登入
response=$(curl -s -w "HTTP_STATUS:%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "$CONTENT_TYPE" \
    -d '{
        "email": "coach@example.com",
        "password": "password123"
    }')

http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
response_body=$(echo "$response" | sed 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_status" = "200" ]; then
    COACH_ACCESS_TOKEN=$(echo "$response_body" | jq -r '.accessToken')
    COACH_USER_ID=$(echo "$response_body" | jq -r '.user.id')
    echo -e "${GREEN}教練登入成功，獲取到令牌${NC}"
else
    echo -e "${RED}教練登入失敗${NC}"
    exit 1
fi

echo "========================================"

# 2. 創建用戶檔案
echo -e "${YELLOW}步驟 2: 創建用戶檔案${NC}"

# 創建學生檔案
test_api "創建學生檔案" "POST" "/users/profile" '{
    "firstName": "學生",
    "lastName": "測試",
    "ntrpLevel": 3.5,
    "playingStyle": "aggressive",
    "preferredHand": "right"
}' "201" "Bearer $ACCESS_TOKEN"

# 創建教練檔案
test_api "創建教練檔案" "POST" "/coaches" '{
    "experience": 5,
    "specialties": ["intermediate", "advanced"],
    "biography": "專業網球教練",
    "hourlyRate": 1500,
    "languages": ["zh-TW", "en"],
    "availableHours": {
        "monday": ["09:00-12:00", "14:00-18:00"],
        "tuesday": ["09:00-12:00", "14:00-18:00"],
        "wednesday": ["09:00-12:00", "14:00-18:00"],
        "thursday": ["09:00-12:00", "14:00-18:00"],
        "friday": ["09:00-12:00", "14:00-18:00"]
    }
}' "201" "Bearer $COACH_ACCESS_TOKEN"

# 獲取教練ID
response=$(curl -s -w "HTTP_STATUS:%{http_code}" -X GET "$BASE_URL/coaches/my-profile" \
    -H "Authorization: Bearer $COACH_ACCESS_TOKEN")

http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
response_body=$(echo "$response" | sed 's/HTTP_STATUS:[0-9]*$//')

if [ "$http_status" = "200" ]; then
    COACH_ID=$(echo "$response_body" | jq -r '.id')
    echo -e "${GREEN}獲取到教練ID: $COACH_ID${NC}"
else
    echo -e "${RED}獲取教練ID失敗${NC}"
fi

echo "========================================"

# 3. 創建課程類型
echo -e "${YELLOW}步驟 3: 創建課程類型${NC}"

test_api "創建個人課程類型" "POST" "/coaches/lesson-types" '{
    "name": "中級個人課程",
    "description": "適合中級球員的一對一課程",
    "type": "individual",
    "level": "intermediate",
    "duration": 60,
    "price": 1500,
    "currency": "TWD"
}' "201" "Bearer $COACH_ACCESS_TOKEN"

test_api "創建團體課程類型" "POST" "/coaches/lesson-types" '{
    "name": "高級團體課程",
    "description": "高級球員團體訓練",
    "type": "group",
    "level": "advanced",
    "duration": 90,
    "price": 1200,
    "currency": "TWD",
    "maxParticipants": 4,
    "minParticipants": 2
}' "201" "Bearer $COACH_ACCESS_TOKEN"

echo "========================================"

# 4. 設定教練時間表
echo -e "${YELLOW}步驟 4: 設定教練時間表${NC}"

test_api "更新教練時間表" "PUT" "/coaches/schedule" '{
    "schedules": [
        {
            "dayOfWeek": 1,
            "startTime": "09:00",
            "endTime": "17:00",
            "isActive": true
        },
        {
            "dayOfWeek": 2,
            "startTime": "09:00",
            "endTime": "17:00",
            "isActive": true
        },
        {
            "dayOfWeek": 3,
            "startTime": "09:00",
            "endTime": "17:00",
            "isActive": true
        },
        {
            "dayOfWeek": 4,
            "startTime": "09:00",
            "endTime": "17:00",
            "isActive": true
        },
        {
            "dayOfWeek": 5,
            "startTime": "09:00",
            "endTime": "17:00",
            "isActive": true
        }
    ]
}' "200" "Bearer $COACH_ACCESS_TOKEN"

echo "========================================"

# 5. 測試智能排課功能
echo -e "${YELLOW}步驟 5: 測試智能排課功能${NC}"

# 獲取智能排課選項
test_api "獲取智能排課選項" "GET" "/intelligent-scheduling/options" "" "200"

# 獲取智能推薦
test_api "獲取智能教練推薦" "POST" "/intelligent-scheduling/recommendations" '{
    "ntrpLevel": 3.5,
    "preferredTimes": ["09:00-12:00", "14:00-18:00"],
    "preferredDays": [1, 2, 3, 4, 5],
    "maxDistance": 10,
    "maxPrice": 2000,
    "preferredLessonType": "individual",
    "dateRange": ["2024-12-16", "2024-12-17", "2024-12-18", "2024-12-19", "2024-12-20"],
    "location": {
        "latitude": 25.0330,
        "longitude": 121.5654,
        "address": "台北市信義區"
    }
}' "200" "Bearer $ACCESS_TOKEN"

# 尋找最佳課程時間
test_api "尋找最佳課程時間" "POST" "/intelligent-scheduling/optimal-time" '{
    "coachId": "'$COACH_ID'",
    "   ": 3.5,
    "preferredTimes": ["09:00-12:00"],
    "preferredDays": [1, 2, 3],
    "maxDistance": 10,
    "maxPrice": 2000,
    "preferredLessonType": "individual",
    "dateRange": ["2024-12-16", "2024-12-17", "2024-12-18"],
    "location": {
        "latitude": 25.0330,
        "longitude": 121.5654,
        "address": "台北市信義區"
    }
}' "200" "Bearer $ACCESS_TOKEN"

# 檢測排課衝突
test_api "檢測排課衝突" "POST" "/intelligent-scheduling/detect-conflicts" '{
    "coachId": "'$COACH_ID'",
    "scheduledAt": "2024-12-16T10:00:00Z",
    "duration": 60
}' "200" "Bearer $COACH_ACCESS_TOKEN"

# 獲取教練推薦因子
test_api "獲取教練推薦因子" "POST" "/intelligent-scheduling/coaches/$COACH_ID/factors" '{
    "   ": 3.5,
    "preferredTimes": ["09:00-12:00"],
    "preferredDays": [1, 2, 3],
    "maxDistance": 10,
    "maxPrice": 2000,
    "preferredLessonType": "individual",
    "dateRange": ["2024-12-16", "2024-12-17", "2024-12-18"]
}' "200" "Bearer $ACCESS_TOKEN"

echo "========================================"

# 6. 測試錯誤情況
echo -e "${YELLOW}步驟 6: 測試錯誤情況${NC}"

# 無認證令牌的請求
test_api "無認證令牌的推薦請求" "POST" "/intelligent-scheduling/recommendations" '{
    "   ": 3.5,
    "dateRange": ["2024-12-16"]
}' "401"

# 無效的教練ID
test_api "無效教練ID的最佳時間查詢" "POST" "/intelligent-scheduling/optimal-time" '{
    "coachId": "invalid-coach-id",
    "   ": 3.5,
    "dateRange": ["2024-12-16"]
}' "400" "Bearer $ACCESS_TOKEN"

# 無效的日期格式
test_api "無效日期格式的推薦請求" "POST" "/intelligent-scheduling/recommendations" '{
    "   ": 3.5,
    "dateRange": ["invalid-date"]
}' "400" "Bearer $ACCESS_TOKEN"

echo "========================================"

# 測試結果統計
echo -e "${YELLOW}測試結果統計:${NC}"
echo -e "總測試數: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "通過測試: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失敗測試: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 所有測試都通過了！${NC}"
    exit 0
else
    echo -e "${RED}❌ 有 $FAILED_TESTS 個測試失敗${NC}"
    exit 1
fi