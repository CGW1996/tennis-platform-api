#!/bin/bash

# 配對統計 API 測試腳本

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
TEST_EMAIL="test@example.com"
TEST_PASSWORD="password123"
TEST_USER_ID=""
JWT_TOKEN=""

# 測試比賽ID和結果ID（需要在實際測試中替換）
TEST_MATCH_ID="test-match-123"
TEST_RESULT_ID="test-result-123"
TARGET_USER_ID="target-user-456"

# 輔助函數
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_test() {
    echo -e "${YELLOW}測試: $1${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

print_response() {
    echo -e "${BLUE}響應:${NC} $1"
}

# 執行 HTTP 請求並檢查響應
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local auth_header=""
    
    if [ ! -z "$JWT_TOKEN" ]; then
        auth_header="-H \"Authorization: Bearer $JWT_TOKEN\""
    fi
    
    if [ ! -z "$data" ]; then
        response=$(eval curl -s -w "HTTPSTATUS:%{http_code}" -X $method "$BASE_URL$endpoint" -H "$CONTENT_TYPE" $auth_header -d '$data')
    else
        response=$(eval curl -s -w "HTTPSTATUS:%{http_code}" -X $method "$BASE_URL$endpoint" $auth_header)
    fi
    
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    response_body=$(echo $response | sed -e 's/HTTPSTATUS:.*//g')
    
    print_response "$response_body"
    
    if [ "$http_code" -eq "$expected_status" ]; then
        print_success "HTTP 狀態碼: $http_code (預期: $expected_status)"
        return 0
    else
        print_error "HTTP 狀態碼: $http_code (預期: $expected_status)"
        return 1
    fi
}

# 用戶註冊和登入
setup_test_user() {
    print_header "設置測試用戶"
    
    # 註冊測試用戶
    print_test "註冊測試用戶"
    register_data="{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\",\"firstName\":\"Test\",\"lastName\":\"User\"}"
    make_request "POST" "/auth/register" "$register_data" 201
    
    # 登入獲取 JWT Token
    print_test "用戶登入"
    login_data="{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}"
    response=$(curl -s -X POST "$BASE_URL/auth/login" -H "$CONTENT_TYPE" -d "$login_data")
    JWT_TOKEN=$(echo $response | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
    TEST_USER_ID=$(echo $response | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ ! -z "$JWT_TOKEN" ]; then
        print_success "登入成功，獲得 JWT Token"
        print_success "用戶ID: $TEST_USER_ID"
    else
        print_error "登入失敗"
        exit 1
    fi
}

# 測試獲取用戶配對統計資訊
test_get_user_statistics() {
    print_header "測試獲取用戶配對統計資訊"
    
    print_test "獲取自己的統計資訊"
    make_request "GET" "/match-statistics/users/$TEST_USER_ID" "" 200
    
    print_test "獲取其他用戶的統計資訊（可能受隱私限制）"
    make_request "GET" "/match-statistics/users/$TARGET_USER_ID" "" 200
}

# 測試獲取用戶配對歷史
test_get_match_history() {
    print_header "測試獲取用戶配對歷史"
    
    print_test "獲取配對歷史（默認參數）"
    make_request "GET" "/match-statistics/users/$TEST_USER_ID/history" "" 200
    
    print_test "獲取配對歷史（指定限制和偏移）"
    make_request "GET" "/match-statistics/users/$TEST_USER_ID/history?limit=10&offset=0" "" 200
    
    print_test "獲取配對歷史（無效用戶ID）"
    make_request "GET" "/match-statistics/users/invalid-user/history" "" 400
}

# 測試記錄比賽結果
test_record_match_result() {
    print_header "測試記錄比賽結果"
    
    print_test "記錄比賽結果"
    result_data="{\"winnerId\":\"$TEST_USER_ID\",\"loserId\":\"$TARGET_USER_ID\",\"score\":\"6-4, 6-2\"}"
    make_request "POST" "/match-statistics/matches/$TEST_MATCH_ID/result" "$result_data" 200
    
    print_test "記錄比賽結果（無效比賽ID）"
    make_request "POST" "/match-statistics/matches/invalid-match/result" "$result_data" 400
    
    print_test "記錄比賽結果（缺少必要參數）"
    invalid_data="{\"winnerId\":\"$TEST_USER_ID\"}"
    make_request "POST" "/match-statistics/matches/$TEST_MATCH_ID/result" "$invalid_data" 400
}

# 測試確認比賽結果
test_confirm_match_result() {
    print_header "測試確認比賽結果"
    
    print_test "確認比賽結果"
    make_request "POST" "/match-statistics/results/$TEST_RESULT_ID/confirm" "" 200
    
    print_test "確認比賽結果（無效結果ID）"
    make_request "POST" "/match-statistics/results/invalid-result/confirm" "" 400
}

# 測試獲取待確認的比賽結果
test_get_pending_confirmations() {
    print_header "測試獲取待確認的比賽結果"
    
    print_test "獲取待確認的比賽結果"
    make_request "GET" "/match-statistics/pending-confirmations" "" 200
}

# 測試獲取技術等級進展
test_get_skill_progression() {
    print_header "測試獲取技術等級進展"
    
    print_test "獲取技術等級進展"
    make_request "GET" "/match-statistics/users/$TEST_USER_ID/skill-progression" "" 200
    
    print_test "獲取其他用戶的技術等級進展（可能受隱私限制）"
    make_request "GET" "/match-statistics/users/$TARGET_USER_ID/skill-progression" "" 200
}

# 測試手動調整技術等級
test_adjust_skill_level() {
    print_header "測試手動調整技術等級"
    
    print_test "手動調整技術等級"
    adjust_data="{\"newLevel\":4.0,\"reason\":\"測試調整\"}"
    make_request "POST" "/match-statistics/users/$TEST_USER_ID/adjust-skill-level" "$adjust_data" 200
    
    print_test "手動調整技術等級（無效等級）"
    invalid_adjust_data="{\"newLevel\":8.0,\"reason\":\"無效等級\"}"
    make_request "POST" "/match-statistics/users/$TEST_USER_ID/adjust-skill-level" "$invalid_adjust_data" 400
    
    print_test "手動調整技術等級（缺少原因）"
    missing_reason_data="{\"newLevel\":3.5}"
    make_request "POST" "/match-statistics/users/$TEST_USER_ID/adjust-skill-level" "$missing_reason_data" 400
}

# 測試隱私設定
test_privacy_settings() {
    print_header "測試隱私設定"
    
    print_test "獲取隱私設定"
    make_request "GET" "/match-statistics/privacy-settings" "" 200
    
    print_test "更新隱私設定"
    privacy_data="{\"showReputationScore\":true,\"showMatchHistory\":false,\"showWinLossRecord\":true,\"showSkillProgression\":true,\"showBehaviorReviews\":false,\"showDetailedStats\":false,\"allowStatisticsSharing\":false}"
    make_request "PUT" "/match-statistics/privacy-settings" "$privacy_data" 200
    
    print_test "驗證隱私設定已更新"
    make_request "GET" "/match-statistics/privacy-settings" "" 200
}

# 測試根據隱私設定獲取信譽分數
test_get_reputation_with_privacy() {
    print_header "測試根據隱私設定獲取信譽分數"
    
    print_test "獲取自己的信譽分數"
    make_request "GET" "/match-statistics/users/$TEST_USER_ID/reputation" "" 200
    
    print_test "獲取其他用戶的信譽分數（可能受隱私限制）"
    make_request "GET" "/match-statistics/users/$TARGET_USER_ID/reputation" "" 200
}

# 測試獲取統計摘要
test_get_statistics_summary() {
    print_header "測試獲取統計摘要"
    
    print_test "獲取統計摘要"
    make_request "GET" "/match-statistics/summary" "" 200
}

# 測試未認證請求
test_unauthorized_requests() {
    print_header "測試未認證請求"
    
    # 暫時清除 JWT Token
    local temp_token=$JWT_TOKEN
    JWT_TOKEN=""
    
    print_test "未認證請求 - 記錄比賽結果"
    result_data="{\"winnerId\":\"$TEST_USER_ID\",\"loserId\":\"$TARGET_USER_ID\",\"score\":\"6-4, 6-2\"}"
    make_request "POST" "/match-statistics/matches/$TEST_MATCH_ID/result" "$result_data" 401
    
    print_test "未認證請求 - 獲取隱私設定"
    make_request "GET" "/match-statistics/privacy-settings" "" 401
    
    print_test "未認證請求 - 獲取統計摘要"
    make_request "GET" "/match-statistics/summary" "" 401
    
    # 恢復 JWT Token
    JWT_TOKEN=$temp_token
}

# 測試無效參數
test_invalid_parameters() {
    print_header "測試無效參數"
    
    print_test "無效用戶ID"
    make_request "GET" "/match-statistics/users//history" "" 404
    
    print_test "無效查詢參數"
    make_request "GET" "/match-statistics/users/$TEST_USER_ID/history?limit=abc&offset=xyz" "" 200
    
    print_test "超大限制參數"
    make_request "GET" "/match-statistics/users/$TEST_USER_ID/history?limit=1000" "" 200
}

# 清理測試數據
cleanup_test_data() {
    print_header "清理測試數據"
    
    print_test "登出用戶"
    make_request "POST" "/auth/logout" "" 200
    
    print_success "測試數據清理完成"
}

# 顯示測試結果摘要
show_test_summary() {
    print_header "測試結果摘要"
    
    echo -e "總測試數: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "通過測試: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失敗測試: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 所有測試通過！${NC}"
        exit 0
    else
        echo -e "\n${RED}❌ 有 $FAILED_TESTS 個測試失敗${NC}"
        exit 1
    fi
}

# 主測試流程
main() {
    echo -e "${BLUE}配對統計 API 測試開始${NC}"
    echo -e "基礎 URL: $BASE_URL"
    echo -e "測試時間: $(date)"
    
    # 檢查服務器是否運行
    print_header "檢查服務器狀態"
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "無法連接到服務器，請確保服務器正在運行"
        exit 1
    fi
    print_success "服務器連接正常"
    
    # 執行測試
    setup_test_user
    test_get_user_statistics
    test_get_match_history
    test_record_match_result
    test_confirm_match_result
    test_get_pending_confirmations
    test_get_skill_progression
    test_adjust_skill_level
    test_privacy_settings
    test_get_reputation_with_privacy
    test_get_statistics_summary
    test_unauthorized_requests
    test_invalid_parameters
    cleanup_test_data
    
    # 顯示測試結果
    show_test_summary
}

# 檢查是否提供了命令行參數
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "配對統計 API 測試腳本"
    echo ""
    echo "用法: $0 [選項]"
    echo ""
    echo "選項:"
    echo "  -h, --help     顯示此幫助信息"
    echo "  --base-url     指定基礎 URL (默認: http://localhost:8080/api/v1)"
    echo ""
    echo "示例:"
    echo "  $0"
    echo "  $0 --base-url http://localhost:3000/api/v1"
    exit 0
fi

# 處理命令行參數
while [[ $# -gt 0 ]]; do
    case $1 in
        --base-url)
            BASE_URL="$2"
            shift 2
            ;;
        *)
            echo "未知參數: $1"
            echo "使用 --help 查看可用選項"
            exit 1
            ;;
    esac
done

# 執行主測試流程
main