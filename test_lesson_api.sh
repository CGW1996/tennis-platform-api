#!/bin/bash

# 課程管理 API 測試腳本
# 測試課程類型和課程管理功能

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
    echo -e "${YELLOW}請求: $method $endpoint${NC}"
    
    if [ -n "$data" ]; then
        echo -e "${YELLOW}數據: $data${NC}"
    fi
    
    # 構建 curl 命令
    curl_cmd="curl -s -w \"HTTP_STATUS:%{http_code}\" -X $method \"$BASE_URL$endpoint\""
    
    if [ -n "$auth_header" ]; then
        curl_cmd="$curl_cmd -H \"Authorization: $auth_header\""
    fi
    
    curl_cmd="$curl_cmd -H \"$CONTENT_TYPE\""
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    # 執行請求
    response=$(eval $curl_cmd)
    
    # 提取狀態碼
    http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
    response_body=$(echo "$response" | sed 's/HTTP_STATUS:[0-9]*$//')
    
    echo -e "${YELLOW}響應狀態: $http_status${NC}"
    echo -e "${YELLOW}響應內容: $response_body${NC}"
    
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

echo -e "${BLUE}開始課程管理 API 測試${NC}"
echo "========================================"

# 測試變數
COACH_ID="test-coach-id"
LESSON_TYPE_ID="test-lesson-type-id"
LESSON_ID="test-lesson-id"
AUTH_TOKEN="Bearer test-token"

echo -e "${BLUE}1. 課程類型管理測試${NC}"

# 測試創建課程類型
test_api "創建個人課程類型" "POST" "/coaches/lesson-types" '{
    "name": "初級個人課程",
    "description": "適合初學者的一對一網球課程",
    "type": "individual",
    "level": "beginner",
    "duration": 60,
    "price": 1500,
    "currency": "TWD",
    "equipment": ["網球拍", "網球"],
    "prerequisites": "無需經驗"
}' "201" "$AUTH_TOKEN"

# 測試創建團體課程類型
test_api "創建團體課程類型" "POST" "/coaches/lesson-types" '{
    "name": "中級團體課程",
    "description": "適合中級球員的團體訓練",
    "type": "group",
    "level": "intermediate",
    "duration": 90,
    "price": 800,
    "currency": "TWD",
    "maxParticipants": 6,
    "minParticipants": 3,
    "equipment": ["網球拍", "網球", "訓練錐"],
    "prerequisites": "具備基本網球技巧"
}' "201" "$AUTH_TOKEN"

# 測試獲取課程類型列表
test_api "獲取課程類型列表" "GET" "/coaches/$COACH_ID/lesson-types" "" "200" ""

# 測試更新課程類型
test_api "更新課程類型" "PUT" "/lesson-types/$LESSON_TYPE_ID" '{
    "name": "初級個人課程（更新）",
    "price": 1600,
    "description": "更新後的課程描述"
}' "200" "$AUTH_TOKEN"

# 測試刪除課程類型
test_api "刪除課程類型" "DELETE" "/lesson-types/$LESSON_TYPE_ID" "" "204" "$AUTH_TOKEN"

echo -e "${BLUE}2. 課程管理測試${NC}"

# 測試創建課程
test_api "創建課程" "POST" "/lessons" '{
    "coachId": "'$COACH_ID'",
    "lessonTypeId": "'$LESSON_TYPE_ID'",
    "type": "individual",
    "level": "beginner",
    "duration": 60,
    "price": 1500,
    "currency": "TWD",
    "scheduledAt": "2024-12-01T10:00:00Z",
    "notes": "第一次課程"
}' "201" "$AUTH_TOKEN"

# 測試獲取課程詳情
test_api "獲取課程詳情" "GET" "/lessons/$LESSON_ID" "" "200" "$AUTH_TOKEN"

# 測試獲取課程列表
test_api "獲取課程列表" "GET" "/lessons?coachId=$COACH_ID&status=scheduled" "" "200" "$AUTH_TOKEN"

# 測試更新課程
test_api "更新課程" "PUT" "/lessons/$LESSON_ID" '{
    "scheduledAt": "2024-12-01T11:00:00Z",
    "notes": "時間調整"
}' "200" "$AUTH_TOKEN"

# 測試取消課程
test_api "取消課程" "POST" "/lessons/$LESSON_ID/cancel" '{
    "reason": "學生臨時有事"
}' "200" "$AUTH_TOKEN"

echo -e "${BLUE}3. 教練時間表管理測試${NC}"

# 測試獲取教練可用時間
test_api "獲取教練可用時間" "GET" "/coaches/$COACH_ID/availability?date=2024-12-01" "" "200" ""

# 測試更新教練時間表
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
            "startTime": "10:00",
            "endTime": "18:00",
            "isActive": true
        }
    ]
}' "200" "$AUTH_TOKEN"

# 測試獲取教練時間表
test_api "獲取教練時間表" "GET" "/coaches/$COACH_ID/schedule" "" "200" ""

echo -e "${BLUE}4. 錯誤處理測試${NC}"

# 測試無效的課程類型
test_api "創建無效課程類型" "POST" "/coaches/lesson-types" '{
    "name": "",
    "type": "invalid",
    "duration": 0,
    "price": -100
}' "400" "$AUTH_TOKEN"

# 測試時間衝突
test_api "創建時間衝突課程" "POST" "/lessons" '{
    "coachId": "'$COACH_ID'",
    "type": "individual",
    "duration": 60,
    "price": 1500,
    "scheduledAt": "2024-12-01T10:00:00Z"
}' "400" "$AUTH_TOKEN"

# 測試未授權訪問
test_api "未授權創建課程類型" "POST" "/coaches/lesson-types" '{
    "name": "測試課程",
    "type": "individual",
    "duration": 60,
    "price": 1000
}' "401" ""

echo "========================================"
echo -e "${BLUE}測試完成${NC}"
echo -e "${GREEN}通過: $PASSED_TESTS${NC}"
echo -e "${RED}失敗: $FAILED_TESTS${NC}"
echo -e "${YELLOW}總計: $TOTAL_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}所有測試通過！${NC}"
    exit 0
else
    echo -e "${RED}有測試失敗，請檢查 API 實現${NC}"
    exit 1
fi