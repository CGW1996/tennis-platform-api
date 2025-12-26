#!/bin/bash

# 網球平台聊天 API 測試腳本

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

# 日誌函數
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 測試函數
test_api() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"
    local auth_header="$6"
    
    ((TOTAL_TESTS++))
    
    log_info "測試: $test_name"
    
    # 構建 curl 命令
    local curl_cmd="curl -s -w '%{http_code}' -X $method"
    
    if [ ! -z "$auth_header" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $auth_header'"
    fi
    
    curl_cmd="$curl_cmd -H '$CONTENT_TYPE'"
    
    if [ ! -z "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    curl_cmd="$curl_cmd $BASE_URL$endpoint"
    
    # 執行請求
    local response=$(eval $curl_cmd)
    local status_code="${response: -3}"
    local body="${response%???}"
    
    # 檢查狀態碼
    if [ "$status_code" = "$expected_status" ]; then
        log_success "$test_name - 狀態碼: $status_code"
        if [ ! -z "$body" ] && [ "$body" != "null" ]; then
            echo "響應: $body" | jq . 2>/dev/null || echo "響應: $body"
        fi
    else
        log_error "$test_name - 期望狀態碼: $expected_status, 實際: $status_code"
        if [ ! -z "$body" ]; then
            echo "響應: $body"
        fi
    fi
    
    echo "----------------------------------------"
}

# 檢查服務器是否運行
check_server() {
    log_info "檢查服務器狀態..."
    local response=$(curl -s -w '%{http_code}' $BASE_URL/../health)
    local status_code="${response: -3}"
    
    if [ "$status_code" = "200" ]; then
        log_success "服務器運行正常"
    else
        log_error "服務器未運行或無法訪問 (狀態碼: $status_code)"
        exit 1
    fi
    echo "----------------------------------------"
}

# 用戶註冊和登入
setup_test_users() {
    log_info "設置測試用戶..."
    
    # 註冊測試用戶1
    local user1_data='{
        "email": "chatuser1@example.com",
        "password": "password123",
        "firstName": "Chat",
        "lastName": "User1"
    }'
    
    test_api "註冊測試用戶1" "POST" "/auth/register" "$user1_data" "201"
    
    # 註冊測試用戶2
    local user2_data='{
        "email": "chatuser2@example.com",
        "password": "password123",
        "firstName": "Chat",
        "lastName": "User2"
    }'
    
    test_api "註冊測試用戶2" "POST" "/auth/register" "$user2_data" "201"
    
    # 登入用戶1
    local login1_data='{
        "email": "chatuser1@example.com",
        "password": "password123"
    }'
    
    local login1_response=$(curl -s -X POST -H "$CONTENT_TYPE" -d "$login1_data" $BASE_URL/auth/login)
    USER1_TOKEN=$(echo $login1_response | jq -r '.accessToken' 2>/dev/null)
    USER1_ID=$(echo $login1_response | jq -r '.user.id' 2>/dev/null)
    
    if [ "$USER1_TOKEN" != "null" ] && [ ! -z "$USER1_TOKEN" ]; then
        log_success "用戶1登入成功"
    else
        log_error "用戶1登入失敗"
        echo "響應: $login1_response"
    fi
    
    # 登入用戶2
    local login2_data='{
        "email": "chatuser2@example.com",
        "password": "password123"
    }'
    
    local login2_response=$(curl -s -X POST -H "$CONTENT_TYPE" -d "$login2_data" $BASE_URL/auth/login)
    USER2_TOKEN=$(echo $login2_response | jq -r '.accessToken' 2>/dev/null)
    USER2_ID=$(echo $login2_response | jq -r '.user.id' 2>/dev/null)
    
    if [ "$USER2_TOKEN" != "null" ] && [ ! -z "$USER2_TOKEN" ]; then
        log_success "用戶2登入成功"
    else
        log_error "用戶2登入失敗"
        echo "響應: $login2_response"
    fi
    
    echo "----------------------------------------"
}

# 測試聊天室功能
test_chat_rooms() {
    log_info "開始測試聊天室功能..."
    
    # 測試創建聊天室（未授權）
    local room_data='{
        "type": "direct",
        "participantIds": ["'$USER2_ID'"]
    }'
    
    test_api "創建聊天室（未授權）" "POST" "/chat/rooms" "$room_data" "401"
    
    # 測試創建直接聊天室
    test_api "創建直接聊天室" "POST" "/chat/rooms" "$room_data" "201" "$USER1_TOKEN"
    
    # 獲取聊天室ID
    local rooms_response=$(curl -s -H "Authorization: Bearer $USER1_TOKEN" $BASE_URL/chat/rooms)
    ROOM_ID=$(echo $rooms_response | jq -r '.[0].id' 2>/dev/null)
    
    if [ "$ROOM_ID" != "null" ] && [ ! -z "$ROOM_ID" ]; then
        log_success "獲取聊天室ID: $ROOM_ID"
    else
        log_error "無法獲取聊天室ID"
        echo "響應: $rooms_response"
    fi
    
    # 測試獲取聊天室列表
    test_api "獲取聊天室列表" "GET" "/chat/rooms" "" "200" "$USER1_TOKEN"
    
    # 測試獲取聊天室詳情
    if [ ! -z "$ROOM_ID" ] && [ "$ROOM_ID" != "null" ]; then
        test_api "獲取聊天室詳情" "GET" "/chat/rooms/$ROOM_ID" "" "200" "$USER1_TOKEN"
    fi
    
    # 測試創建群組聊天室
    local group_data='{
        "type": "group",
        "name": "測試群組",
        "participantIds": ["'$USER2_ID'"]
    }'
    
    test_api "創建群組聊天室" "POST" "/chat/rooms" "$group_data" "201" "$USER1_TOKEN"
}

# 測試訊息功能
test_messages() {
    log_info "開始測試訊息功能..."
    
    if [ -z "$ROOM_ID" ] || [ "$ROOM_ID" = "null" ]; then
        log_warning "跳過訊息測試 - 沒有可用的聊天室ID"
        return
    fi
    
    # 測試發送訊息
    local message_data='{
        "chatRoomId": "'$ROOM_ID'",
        "content": "Hello, this is a test message!",
        "messageType": "text"
    }'
    
    test_api "發送訊息" "POST" "/chat/messages" "$message_data" "201" "$USER1_TOKEN"
    
    # 測試獲取訊息列表
    test_api "獲取訊息列表" "GET" "/chat/rooms/$ROOM_ID/messages" "" "200" "$USER1_TOKEN"
    
    # 測試分頁獲取訊息
    test_api "分頁獲取訊息" "GET" "/chat/rooms/$ROOM_ID/messages?page=1&limit=10" "" "200" "$USER1_TOKEN"
    
    # 測試標記訊息為已讀
    test_api "標記訊息為已讀" "POST" "/chat/rooms/$ROOM_ID/read" "" "200" "$USER1_TOKEN"
    
    # 測試用戶2發送訊息
    local message2_data='{
        "chatRoomId": "'$ROOM_ID'",
        "content": "Hello back from user 2!",
        "messageType": "text"
    }'
    
    test_api "用戶2發送訊息" "POST" "/chat/messages" "$message2_data" "201" "$USER2_TOKEN"
}

# 測試聊天室操作
test_room_operations() {
    log_info "開始測試聊天室操作..."
    
    if [ -z "$ROOM_ID" ] || [ "$ROOM_ID" = "null" ]; then
        log_warning "跳過聊天室操作測試 - 沒有可用的聊天室ID"
        return
    fi
    
    # 測試加入聊天室
    test_api "加入聊天室" "POST" "/chat/rooms/$ROOM_ID/join" "" "200" "$USER1_TOKEN"
    
    # 測試獲取在線用戶
    test_api "獲取在線用戶" "GET" "/chat/online-users" "" "200" "$USER1_TOKEN"
    
    # 測試獲取聊天室在線用戶
    test_api "獲取聊天室在線用戶" "GET" "/chat/rooms/$ROOM_ID/online-users" "" "200" "$USER1_TOKEN"
    
    # 測試離開聊天室
    test_api "離開聊天室" "POST" "/chat/rooms/$ROOM_ID/leave" "" "200" "$USER2_TOKEN"
}

# 測試錯誤情況
test_error_cases() {
    log_info "開始測試錯誤情況..."
    
    # 測試訪問不存在的聊天室
    test_api "訪問不存在的聊天室" "GET" "/chat/rooms/nonexistent-id" "" "404" "$USER1_TOKEN"
    
    # 測試發送空訊息
    local empty_message='{
        "chatRoomId": "'$ROOM_ID'",
        "content": "",
        "messageType": "text"
    }'
    
    test_api "發送空訊息" "POST" "/chat/messages" "$empty_message" "400" "$USER1_TOKEN"
    
    # 測試無效的聊天室ID
    local invalid_message='{
        "chatRoomId": "invalid-room-id",
        "content": "Test message",
        "messageType": "text"
    }'
    
    test_api "發送到無效聊天室" "POST" "/chat/messages" "$invalid_message" "403" "$USER1_TOKEN"
}

# 清理測試數據
cleanup() {
    log_info "清理測試數據..."
    # 這裡可以添加清理邏輯，比如刪除測試用戶等
    # 目前暫時跳過，因為需要實現相應的API
    log_info "清理完成"
}

# 顯示測試結果
show_results() {
    echo "========================================"
    echo "測試結果統計:"
    echo "總測試數: $TOTAL_TESTS"
    echo -e "通過: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失敗: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}所有測試通過！${NC}"
        exit 0
    else
        echo -e "${RED}有 $FAILED_TESTS 個測試失敗${NC}"
        exit 1
    fi
}

# 主函數
main() {
    echo "========================================"
    echo "網球平台聊天 API 測試"
    echo "========================================"
    
    # 檢查依賴
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安裝"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_warning "jq 未安裝，JSON 格式化將不可用"
    fi
    
    # 執行測試
    check_server
    setup_test_users
    test_chat_rooms
    test_messages
    test_room_operations
    test_error_cases
    cleanup
    show_results
}

# 執行主函數
main "$@"