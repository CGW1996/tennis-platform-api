#!/bin/bash

# 網球平台配對 API 測試腳本

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

# 測試用戶憑證
ACCESS_TOKEN=""
USER_ID=""

# 輔助函數
print_test_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((PASSED_TESTS++))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    ((FAILED_TESTS++))
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 測試 API 響應
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local test_name=$5
    local auth_header=""
    
    ((TOTAL_TESTS++))
    
    if [ ! -z "$ACCESS_TOKEN" ]; then
        auth_header="Authorization: Bearer $ACCESS_TOKEN"
    fi
    
    echo -e "\n${YELLOW}測試: $test_name${NC}"
    echo "請求: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        if [ ! -z "$auth_header" ]; then
            response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$endpoint" \
                -H "$CONTENT_TYPE" \
                -H "$auth_header")
        else
            response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL$endpoint" \
                -H "$CONTENT_TYPE")
        fi
    else
        if [ ! -z "$auth_header" ]; then
            response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
                -H "$CONTENT_TYPE" \
                -H "$auth_header" \
                -d "$data")
        else
            response=$(curl -s -w "\n%{http_code}" -X $method "$BASE_URL$endpoint" \
                -H "$CONTENT_TYPE" \
                -d "$data")
        fi
    fi
    
    # 分離響應體和狀態碼
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    echo "狀態碼: $status_code"
    echo "響應: $response_body" | jq . 2>/dev/null || echo "響應: $response_body"
    
    if [ "$status_code" = "$expected_status" ]; then
        print_success "$test_name - 狀態碼正確 ($status_code)"
        return 0
    else
        print_error "$test_name - 狀態碼錯誤 (期望: $expected_status, 實際: $status_code)"
        return 1
    fi
}

# 用戶註冊和登入
setup_test_user() {
    print_test_header "設置測試用戶"
    
    # 生成隨機郵箱
    RANDOM_EMAIL="test_matching_$(date +%s)@example.com"
    
    # 註冊用戶
    register_data='{
        "email": "'$RANDOM_EMAIL'",
        "password": "TestPassword123!",
        "firstName": "Test",
        "lastName": "User"
    }'
    
    test_api "POST" "/auth/register" "$register_data" "201" "用戶註冊"
    
    # 登入用戶
    login_data='{
        "email": "'$RANDOM_EMAIL'",
        "password": "TestPassword123!"
    }'
    
    response=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "$CONTENT_TYPE" \
        -d "$login_data")
    
    ACCESS_TOKEN=$(echo "$response" | jq -r '.accessToken // empty')
    USER_ID=$(echo "$response" | jq -r '.user.id // empty')
    
    if [ ! -z "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
        print_success "用戶登入成功，獲取到 Access Token"
        echo "User ID: $USER_ID"
    else
        print_error "用戶登入失敗，無法獲取 Access Token"
        exit 1
    fi
    
    # 創建用戶檔案
    profile_data='{
        "firstName": "Test",
        "lastName": "User",
        "ntrpLevel": 3.5,
        "playingStyle": "all-court",
        "preferredHand": "right",
        "latitude": 25.0330,
        "longitude": 121.5654,
        "playingFrequency": "regular",
        "preferredTimes": ["morning", "evening"],
        "maxTravelDistance": 15.0,
        "gender": "male"
    }'
    
    test_api "POST" "/users/profile" "$profile_data" "201" "創建用戶檔案"
}

# 測試配對搜尋功能
test_matching_search() {
    print_test_header "測試配對搜尋功能"
    
    # 測試基本配對搜尋
    search_data='{
        "ntrpLevel": 3.5,
        "maxDistance": 20.0,
        "preferredTimes": ["morning", "evening"],
        "playingFrequency": "regular",
        "gender": "any",
        "limit": 10
    }'
    
    test_api "POST" "/matching/find" "$search_data" "200" "基本配對搜尋"
    
    # 測試帶年齡範圍的配對搜尋
    search_with_age_data='{
        "ntrpLevel": 3.0,
        "maxDistance": 15.0,
        "ageRange": {
            "min": 25,
            "max": 40
        },
        "gender": "male",
        "minReputationScore": 80.0,
        "limit": 5
    }'
    
    test_api "POST" "/matching/find" "$search_with_age_data" "200" "帶年齡範圍的配對搜尋"
    
    # 測試無效的配對搜尋請求
    invalid_search_data='{
        "ntrpLevel": 10.0,
        "maxDistance": -5.0
    }'
    
    test_api "POST" "/matching/find" "$invalid_search_data" "200" "無效參數的配對搜尋"
}

# 測試隨機配對功能
test_random_matching() {
    print_test_header "測試隨機配對功能"
    
    # 測試基本隨機配對
    test_api "GET" "/matching/random" "" "200" "基本隨機配對"
    
    # 測試指定數量的隨機配對
    test_api "GET" "/matching/random?count=3" "" "200" "指定數量的隨機配對"
    
    # 測試最大數量限制
    test_api "GET" "/matching/random?count=25" "" "200" "超過最大數量的隨機配對"
    
    # 測試無效數量參數
    test_api "GET" "/matching/random?count=invalid" "" "200" "無效數量參數的隨機配對"
}

# 測試信譽系統
test_reputation_system() {
    print_test_header "測試信譽系統"
    
    # 測試獲取信譽分數
    test_api "GET" "/matching/reputation" "" "200" "獲取信譽分數"
    
    # 測試更新信譽分數
    update_reputation_data='{
        "matchCompleted": true,
        "wasOnTime": true,
        "behaviorRating": 4.5
    }'
    
    test_api "PUT" "/matching/reputation/$USER_ID" "$update_reputation_data" "200" "更新信譽分數"
    
    # 測試無效的行為評分
    invalid_reputation_data='{
        "matchCompleted": true,
        "wasOnTime": false,
        "behaviorRating": 6.0
    }'
    
    test_api "PUT" "/matching/reputation/$USER_ID" "$invalid_reputation_data" "400" "無效行為評分的信譽更新"
}

# 測試配對歷史
test_matching_history() {
    print_test_header "測試配對歷史"
    
    # 測試獲取配對歷史
    test_api "GET" "/matching/history" "" "200" "獲取配對歷史"
    
    # 測試分頁配對歷史
    test_api "GET" "/matching/history?page=1&limit=5" "" "200" "分頁配對歷史"
    
    # 測試無效分頁參數
    test_api "GET" "/matching/history?page=0&limit=-1" "" "200" "無效分頁參數的配對歷史"
}

# 測試創建配對
test_create_match() {
    print_test_header "測試創建配對"
    
    # 測試創建基本配對
    create_match_data='{
        "participantIds": ["'$USER_ID'"],
        "matchType": "casual"
    }'
    
    test_api "POST" "/matching/create" "$create_match_data" "201" "創建基本配對"
    
    # 測試創建帶場地的配對
    create_match_with_court_data='{
        "participantIds": ["'$USER_ID'"],
        "matchType": "practice",
        "courtId": "550e8400-e29b-41d4-a716-446655440000",
        "scheduledAt": "2024-12-01T10:00:00Z"
    }'
    
    test_api "POST" "/matching/create" "$create_match_with_court_data" "201" "創建帶場地的配對"
    
    # 測試無效的配對類型
    invalid_match_data='{
        "participantIds": ["'$USER_ID'"],
        "matchType": "invalid_type"
    }'
    
    test_api "POST" "/matching/create" "$invalid_match_data" "400" "無效配對類型的創建"
    
    # 測試缺少必要參數
    incomplete_match_data='{
        "matchType": "casual"
    }'
    
    test_api "POST" "/matching/create" "$incomplete_match_data" "400" "缺少參數的配對創建"
}

# 測試配對統計
test_matching_statistics() {
    print_test_header "測試配對統計"
    
    # 測試獲取配對統計
    test_api "GET" "/matching/statistics" "" "200" "獲取配對統計"
}

# 測試抽卡配對功能
test_card_matching() {
    print_test_header "測試抽卡配對功能"
    
    # 測試處理抽卡動作 - 喜歡
    card_action_like_data='{
        "targetUserId": "550e8400-e29b-41d4-a716-446655440000",
        "action": "like"
    }'
    
    test_api "POST" "/matching/card-action" "$card_action_like_data" "200" "處理抽卡動作 - 喜歡"
    
    # 測試處理抽卡動作 - 不喜歡
    card_action_dislike_data='{
        "targetUserId": "550e8400-e29b-41d4-a716-446655440001",
        "action": "dislike"
    }'
    
    test_api "POST" "/matching/card-action" "$card_action_dislike_data" "200" "處理抽卡動作 - 不喜歡"
    
    # 測試處理抽卡動作 - 跳過
    card_action_skip_data='{
        "targetUserId": "550e8400-e29b-41d4-a716-446655440002",
        "action": "skip"
    }'
    
    test_api "POST" "/matching/card-action" "$card_action_skip_data" "200" "處理抽卡動作 - 跳過"
    
    # 測試無效的動作類型
    invalid_action_data='{
        "targetUserId": "550e8400-e29b-41d4-a716-446655440000",
        "action": "invalid"
    }'
    
    test_api "POST" "/matching/card-action" "$invalid_action_data" "400" "無效動作類型的抽卡動作"
    
    # 測試對自己執行動作
    self_action_data='{
        "targetUserId": "'$USER_ID'",
        "action": "like"
    }'
    
    test_api "POST" "/matching/card-action" "$self_action_data" "400" "對自己執行抽卡動作"
}

# 測試抽卡互動歷史
test_card_history() {
    print_test_header "測試抽卡互動歷史"
    
    # 測試獲取抽卡互動歷史
    test_api "GET" "/matching/card-history" "" "200" "獲取抽卡互動歷史"
    
    # 測試分頁抽卡互動歷史
    test_api "GET" "/matching/card-history?page=1&limit=10" "" "200" "分頁抽卡互動歷史"
    
    # 測試按動作類型篩選
    test_api "GET" "/matching/card-history?action=like" "" "200" "按動作類型篩選互動歷史"
    
    # 測試無效分頁參數
    test_api "GET" "/matching/card-history?page=0&limit=-1" "" "200" "無效分頁參數的互動歷史"
}

# 測試配對通知
test_match_notifications() {
    print_test_header "測試配對通知"
    
    # 測試獲取配對通知
    test_api "GET" "/matching/notifications" "" "200" "獲取配對通知"
    
    # 測試只獲取未讀通知
    test_api "GET" "/matching/notifications?unread_only=true" "" "200" "獲取未讀通知"
    
    # 測試分頁通知
    test_api "GET" "/matching/notifications?page=1&limit=10" "" "200" "分頁配對通知"
    
    # 測試標記通知為已讀（使用假的通知ID）
    test_api "PUT" "/matching/notifications/550e8400-e29b-41d4-a716-446655440000/read" "" "404" "標記不存在的通知為已讀"
}

# 測試未授權訪問
test_unauthorized_access() {
    print_test_header "測試未授權訪問"
    
    # 暫時清除 token
    local temp_token=$ACCESS_TOKEN
    ACCESS_TOKEN=""
    
    # 測試未授權的配對搜尋
    search_data='{"ntrpLevel": 3.5}'
    test_api "POST" "/matching/find" "$search_data" "401" "未授權的配對搜尋"
    
    # 測試未授權的隨機配對
    test_api "GET" "/matching/random" "" "401" "未授權的隨機配對"
    
    # 測試未授權的信譽獲取
    test_api "GET" "/matching/reputation" "" "401" "未授權的信譽獲取"
    
    # 恢復 token
    ACCESS_TOKEN=$temp_token
}

# 主測試流程
main() {
    echo -e "${BLUE}開始網球平台配對 API 測試${NC}"
    echo "測試目標: $BASE_URL"
    
    # 檢查服務器是否運行
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "無法連接到服務器，請確保服務器正在運行"
        exit 1
    fi
    
    print_success "服務器連接正常"
    
    # 執行測試
    setup_test_user
    test_matching_search
    test_random_matching
    test_reputation_system
    test_matching_history
    test_create_match
    test_matching_statistics
    test_card_matching
    test_card_history
    test_match_notifications
    test_unauthorized_access
    
    # 測試結果統計
    echo -e "\n${BLUE}=== 測試結果統計 ===${NC}"
    echo -e "總測試數: $TOTAL_TESTS"
    echo -e "${GREEN}通過: $PASSED_TESTS${NC}"
    echo -e "${RED}失敗: $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 所有測試通過！${NC}"
        exit 0
    else
        echo -e "\n${RED}❌ 有 $FAILED_TESTS 個測試失敗${NC}"
        exit 1
    fi
}

# 檢查依賴
if ! command -v curl &> /dev/null; then
    print_error "curl 未安裝，請先安裝 curl"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    print_warning "jq 未安裝，JSON 響應將不會格式化"
fi

# 執行主函數
main "$@"