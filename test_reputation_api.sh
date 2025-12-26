#!/bin/bash

# 信譽評分系統 API 測試腳本

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

# 測試用戶和比賽數據
TEST_USER_ID="test-user-123"
TEST_REVIEWER_ID="test-reviewer-456"
TEST_MATCH_ID="test-match-789"
JWT_TOKEN=""

# 打印測試標題
print_test_title() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# 打印成功消息
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

# 打印失敗消息
print_error() {
    echo -e "${RED}✗ $1${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

# 打印警告消息
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 執行 HTTP 請求並檢查響應
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local auth_header=""
    
    if [ ! -z "$JWT_TOKEN" ]; then
        auth_header="-H \"Authorization: Bearer $JWT_TOKEN\""
    fi
    
    if [ ! -z "$data" ]; then
        response=$(eval curl -s -w "HTTPSTATUS:%{http_code}" -X $method "$BASE_URL$url" -H "$CONTENT_TYPE" $auth_header -d '$data')
    else
        response=$(eval curl -s -w "HTTPSTATUS:%{http_code}" -X $method "$BASE_URL$url" $auth_header)
    fi
    
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo $response | sed -e 's/HTTPSTATUS:.*//g')
    
    echo "Response: $body"
    echo "HTTP Status: $http_code"
    
    if [ "$http_code" -eq "$expected_status" ]; then
        return 0
    else
        return 1
    fi
}

# 測試用戶註冊和登入（獲取 JWT Token）
test_auth_setup() {
    print_test_title "設置測試用戶認證"
    
    # 註冊測試用戶
    register_data='{
        "email": "test-reputation@example.com",
        "password": "TestPassword123!",
        "firstName": "Test",
        "lastName": "User"
    }'
    
    echo "註冊測試用戶..."
    if make_request "POST" "/auth/register" "$register_data" 201; then
        print_success "用戶註冊成功"
    else
        print_warning "用戶可能已存在，嘗試登入"
    fi
    
    # 登入獲取 Token
    login_data='{
        "email": "test-reputation@example.com",
        "password": "TestPassword123!"
    }'
    
    echo "用戶登入..."
    response=$(curl -s -X POST "$BASE_URL/auth/login" -H "$CONTENT_TYPE" -d "$login_data")
    JWT_TOKEN=$(echo $response | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
    
    if [ ! -z "$JWT_TOKEN" ]; then
        print_success "獲取 JWT Token 成功"
        echo "Token: ${JWT_TOKEN:0:20}..."
    else
        print_error "獲取 JWT Token 失敗"
        echo "Response: $response"
        exit 1
    fi
}

# 測試獲取用戶信譽分數（公開API）
test_get_reputation_score() {
    print_test_title "測試獲取用戶信譽分數"
    
    if make_request "GET" "/reputation/users/$TEST_USER_ID/score" "" 200; then
        print_success "獲取信譽分數成功"
    else
        print_error "獲取信譽分數失敗"
    fi
}

# 測試獲取信譽排行榜
test_get_leaderboard() {
    print_test_title "測試獲取信譽排行榜"
    
    if make_request "GET" "/reputation/leaderboard?limit=10" "" 200; then
        print_success "獲取排行榜成功"
    else
        print_error "獲取排行榜失敗"
    fi
}

# 測試獲取信譽統計信息
test_get_stats() {
    print_test_title "測試獲取信譽統計信息"
    
    if make_request "GET" "/reputation/stats" "" 200; then
        print_success "獲取統計信息成功"
    else
        print_error "獲取統計信息失敗"
    fi
}

# 測試獲取用戶信譽歷史記錄（需要認證）
test_get_reputation_history() {
    print_test_title "測試獲取用戶信譽歷史記錄"
    
    if make_request "GET" "/reputation/users/$TEST_USER_ID/history" "" 200; then
        print_success "獲取信譽歷史成功"
    else
        print_error "獲取信譽歷史失敗"
    fi
}

# 測試記錄比賽出席情況
test_record_attendance() {
    print_test_title "測試記錄比賽出席情況"
    
    attendance_data='{
        "matchId": "'$TEST_MATCH_ID'",
        "status": "completed"
    }'
    
    if make_request "POST" "/reputation/users/$TEST_USER_ID/attendance" "$attendance_data" 200; then
        print_success "記錄出席情況成功"
    else
        print_error "記錄出席情況失敗"
    fi
}

# 測試記錄比賽準時情況
test_record_punctuality() {
    print_test_title "測試記錄比賽準時情況"
    
    current_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    punctuality_data='{
        "matchId": "'$TEST_MATCH_ID'",
        "arrivalTime": "'$current_time'"
    }'
    
    if make_request "POST" "/reputation/users/$TEST_USER_ID/punctuality" "$punctuality_data" 200; then
        print_success "記錄準時情況成功"
    else
        print_error "記錄準時情況失敗"
    fi
}

# 測試記錄技術等級準確度
test_record_skill_accuracy() {
    print_test_title "測試記錄技術等級準確度"
    
    skill_data='{
        "matchId": "'$TEST_MATCH_ID'",
        "reportedLevel": 4.0,
        "observedLevel": 3.8
    }'
    
    if make_request "POST" "/reputation/users/$TEST_USER_ID/skill-accuracy" "$skill_data" 200; then
        print_success "記錄技術準確度成功"
    else
        print_error "記錄技術準確度失敗"
    fi
}

# 測試提交行為評價
test_submit_behavior_review() {
    print_test_title "測試提交行為評價"
    
    review_data='{
        "matchId": "'$TEST_MATCH_ID'",
        "rating": 4.5,
        "comment": "很好的球友，技術不錯且態度友善",
        "tags": ["friendly", "skilled", "punctual"]
    }'
    
    if make_request "POST" "/reputation/users/$TEST_USER_ID/behavior-review" "$review_data" 200; then
        print_success "提交行為評價成功"
    else
        print_error "提交行為評價失敗"
    fi
}

# 測試更新用戶NTRP等級
test_update_ntrp_level() {
    print_test_title "測試更新用戶NTRP等級"
    
    if make_request "POST" "/reputation/users/$TEST_USER_ID/update-ntrp" "" 200; then
        print_success "更新NTRP等級成功"
    else
        print_error "更新NTRP等級失敗"
    fi
}

# 測試無效請求
test_invalid_requests() {
    print_test_title "測試無效請求處理"
    
    # 測試無效的評分範圍
    invalid_review_data='{
        "matchId": "'$TEST_MATCH_ID'",
        "rating": 6.0,
        "comment": "無效評分測試"
    }'
    
    if make_request "POST" "/reputation/users/$TEST_USER_ID/behavior-review" "$invalid_review_data" 400; then
        print_success "無效評分請求正確被拒絕"
    else
        print_error "無效評分請求處理失敗"
    fi
    
    # 測試無效的NTRP等級
    invalid_skill_data='{
        "matchId": "'$TEST_MATCH_ID'",
        "reportedLevel": 8.0,
        "observedLevel": 3.8
    }'
    
    if make_request "POST" "/reputation/users/$TEST_USER_ID/skill-accuracy" "$invalid_skill_data" 400; then
        print_success "無效NTRP等級請求正確被拒絕"
    else
        print_error "無效NTRP等級請求處理失敗"
    fi
}

# 測試未認證請求
test_unauthorized_requests() {
    print_test_title "測試未認證請求處理"
    
    # 暫時清空 Token
    temp_token=$JWT_TOKEN
    JWT_TOKEN=""
    
    if make_request "GET" "/reputation/users/$TEST_USER_ID/history" "" 401; then
        print_success "未認證請求正確被拒絕"
    else
        print_error "未認證請求處理失敗"
    fi
    
    # 恢復 Token
    JWT_TOKEN=$temp_token
}

# 打印測試結果摘要
print_test_summary() {
    echo -e "\n${BLUE}=== 測試結果摘要 ===${NC}"
    echo -e "總測試數: $TOTAL_TESTS"
    echo -e "${GREEN}通過: $PASSED_TESTS${NC}"
    echo -e "${RED}失敗: $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 所有測試通過！${NC}"
        exit 0
    else
        echo -e "\n${RED}❌ 有測試失敗，請檢查上述錯誤信息${NC}"
        exit 1
    fi
}

# 主測試流程
main() {
    echo -e "${BLUE}開始信譽評分系統 API 測試${NC}"
    echo -e "測試服務器: $BASE_URL"
    
    # 檢查服務器是否運行
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "無法連接到服務器，請確保服務器正在運行"
        exit 1
    fi
    
    # 執行測試
    test_auth_setup
    test_get_reputation_score
    test_get_leaderboard
    test_get_stats
    test_get_reputation_history
    test_record_attendance
    test_record_punctuality
    test_record_skill_accuracy
    test_submit_behavior_review
    test_update_ntrp_level
    test_invalid_requests
    test_unauthorized_requests
    
    # 打印測試摘要
    print_test_summary
}

# 檢查是否提供了自定義參數
while [[ $# -gt 0 ]]; do
    case $1 in
        --base-url)
            BASE_URL="$2"
            shift 2
            ;;
        --user-id)
            TEST_USER_ID="$2"
            shift 2
            ;;
        --match-id)
            TEST_MATCH_ID="$2"
            shift 2
            ;;
        --help)
            echo "信譽評分系統 API 測試腳本"
            echo ""
            echo "用法: $0 [選項]"
            echo ""
            echo "選項:"
            echo "  --base-url URL    設置 API 基礎 URL (默認: http://localhost:8080/api/v1)"
            echo "  --user-id ID      設置測試用戶 ID (默認: test-user-123)"
            echo "  --match-id ID     設置測試比賽 ID (默認: test-match-789)"
            echo "  --help           顯示此幫助信息"
            echo ""
            echo "示例:"
            echo "  $0"
            echo "  $0 --base-url http://localhost:3000/api/v1"
            echo "  $0 --user-id real-user-id --match-id real-match-id"
            exit 0
            ;;
        *)
            echo "未知選項: $1"
            echo "使用 --help 查看可用選項"
            exit 1
            ;;
    esac
done

# 執行主測試流程
main