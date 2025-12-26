#!/bin/bash

# 網球平台評價系統 API 測試腳本

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
USER_TOKEN=""
COURT_ID=""
REVIEW_ID=""

# 輔助函數
print_header() {
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

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# 測試 API 請求
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5
    local auth_header=""
    
    ((TOTAL_TESTS++))
    
    if [ ! -z "$USER_TOKEN" ]; then
        auth_header="Authorization: Bearer $USER_TOKEN"
    fi
    
    print_info "測試: $description"
    
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
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    echo "請求: $method $BASE_URL$endpoint"
    if [ ! -z "$data" ]; then
        echo "數據: $data"
    fi
    echo "響應狀態: $status_code"
    echo "響應內容: $body"
    
    if [ "$status_code" -eq "$expected_status" ]; then
        print_success "$description - 狀態碼正確 ($status_code)"
        echo "$body"
        return 0
    else
        print_error "$description - 狀態碼錯誤 (期望: $expected_status, 實際: $status_code)"
        echo "$body"
        return 1
    fi
}

# 用戶註冊和登入
setup_user() {
    print_header "設置測試用戶"
    
    # 註冊測試用戶
    register_data='{
        "email": "reviewer@example.com",
        "password": "password123",
        "firstName": "Test",
        "lastName": "Reviewer"
    }'
    
    test_api "POST" "/auth/register" "$register_data" 201 "註冊測試用戶"
    
    # 登入獲取 token
    login_data='{
        "email": "reviewer@example.com",
        "password": "password123"
    }'
    
    if test_api "POST" "/auth/login" "$login_data" 200 "用戶登入"; then
        USER_TOKEN=$(echo "$body" | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
        if [ ! -z "$USER_TOKEN" ]; then
            print_success "獲取用戶 token: ${USER_TOKEN:0:20}..."
        else
            print_error "無法獲取用戶 token"
            exit 1
        fi
    else
        print_error "用戶登入失敗"
        exit 1
    fi
}

# 創建測試場地
setup_court() {
    print_header "創建測試場地"
    
    court_data='{
        "name": "測試網球場",
        "description": "用於評價系統測試的場地",
        "address": "台北市信義區測試路123號",
        "latitude": 25.0330,
        "longitude": 121.5654,
        "facilities": ["parking", "restroom", "lighting"],
        "courtType": "hard",
        "pricePerHour": 800,
        "currency": "TWD",
        "operatingHours": {
            "monday": "06:00-22:00",
            "tuesday": "06:00-22:00",
            "wednesday": "06:00-22:00",
            "thursday": "06:00-22:00",
            "friday": "06:00-22:00",
            "saturday": "06:00-22:00",
            "sunday": "06:00-22:00"
        },
        "contactPhone": "02-1234-5678",
        "contactEmail": "test@court.com"
    }'
    
    if test_api "POST" "/courts" "$court_data" 201 "創建測試場地"; then
        COURT_ID=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        if [ ! -z "$COURT_ID" ]; then
            print_success "獲取場地 ID: $COURT_ID"
        else
            print_error "無法獲取場地 ID"
            exit 1
        fi
    else
        print_error "創建測試場地失敗"
        exit 1
    fi
}

# 測試評價 CRUD 操作
test_review_crud() {
    print_header "測試評價 CRUD 操作"
    
    # 創建評價
    review_data='{
        "courtId": "'$COURT_ID'",
        "rating": 4,
        "comment": "場地很不錯，設施齊全，但價格稍高。",
        "images": ["https://example.com/image1.jpg", "https://example.com/image2.jpg"]
    }'
    
    if test_api "POST" "/reviews" "$review_data" 201 "創建場地評價"; then
        REVIEW_ID=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        if [ ! -z "$REVIEW_ID" ]; then
            print_success "獲取評價 ID: $REVIEW_ID"
        else
            print_error "無法獲取評價 ID"
            return 1
        fi
    else
        print_error "創建評價失敗"
        return 1
    fi
    
    # 獲取評價詳情
    test_api "GET" "/reviews/$REVIEW_ID" "" 200 "獲取評價詳情"
    
    # 更新評價
    update_data='{
        "rating": 5,
        "comment": "重新評估後覺得這個場地非常棒！",
        "images": ["https://example.com/updated_image.jpg"]
    }'
    
    test_api "PUT" "/reviews/$REVIEW_ID" "$update_data" 200 "更新評價"
    
    # 測試重複評價（應該失敗）
    test_api "POST" "/reviews" "$review_data" 400 "測試重複評價（應該失敗）"
}

# 測試評價列表和篩選
test_review_list() {
    print_header "測試評價列表和篩選"
    
    # 獲取所有評價
    test_api "GET" "/reviews" "" 200 "獲取所有評價"
    
    # 根據場地篩選評價
    test_api "GET" "/reviews?courtId=$COURT_ID" "" 200 "根據場地篩選評價"
    
    # 根據評分篩選評價
    test_api "GET" "/reviews?rating=5" "" 200 "根據評分篩選評價"
    
    # 測試排序
    test_api "GET" "/reviews?sortBy=rating&sortOrder=desc" "" 200 "按評分降序排序"
    
    # 測試分頁
    test_api "GET" "/reviews?page=1&pageSize=10" "" 200 "測試分頁"
}

# 測試評價統計
test_review_statistics() {
    print_header "測試評價統計"
    
    # 獲取場地評價統計
    test_api "GET" "/courts/$COURT_ID/reviews/statistics" "" 200 "獲取場地評價統計"
    
    # 驗證場地評分是否更新
    test_api "GET" "/courts/$COURT_ID" "" 200 "驗證場地評分更新"
}

# 測試評價舉報功能
test_review_reporting() {
    print_header "測試評價舉報功能"
    
    # 舉報評價
    report_data='{
        "reason": "inappropriate",
        "comment": "評價內容不當"
    }'
    
    test_api "POST" "/reviews/$REVIEW_ID/report" "$report_data" 200 "舉報評價"
    
    # 測試重複舉報（應該失敗）
    test_api "POST" "/reviews/$REVIEW_ID/report" "$report_data" 400 "測試重複舉報（應該失敗）"
}

# 測試評價有用性標記
test_review_helpful() {
    print_header "測試評價有用性標記"
    
    # 標記評價為有用
    test_api "POST" "/reviews/$REVIEW_ID/helpful?helpful=true" "" 200 "標記評價為有用"
    
    # 取消有用標記
    test_api "POST" "/reviews/$REVIEW_ID/helpful?helpful=false" "" 200 "取消有用標記"
}

# 測試圖片上傳
test_image_upload() {
    print_header "測試評價圖片上傳"
    
    # 創建測試圖片文件
    echo "fake image content" > /tmp/test_image.jpg
    
    # 上傳圖片（使用 multipart/form-data）
    if [ ! -z "$USER_TOKEN" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/reviews/images" \
            -H "Authorization: Bearer $USER_TOKEN" \
            -F "images=@/tmp/test_image.jpg")
        
        body=$(echo "$response" | head -n -1)
        status_code=$(echo "$response" | tail -n 1)
        
        echo "上傳圖片響應狀態: $status_code"
        echo "上傳圖片響應內容: $body"
        
        if [ "$status_code" -eq 200 ]; then
            print_success "評價圖片上傳成功"
        else
            print_warning "評價圖片上傳失敗 (可能是文件服務未配置)"
        fi
    fi
    
    # 清理測試文件
    rm -f /tmp/test_image.jpg
}

# 測試錯誤處理
test_error_handling() {
    print_header "測試錯誤處理"
    
    # 測試無效的場地ID
    invalid_review_data='{
        "courtId": "invalid-uuid",
        "rating": 4,
        "comment": "測試評價"
    }'
    
    test_api "POST" "/reviews" "$invalid_review_data" 400 "測試無效場地ID"
    
    # 測試無效的評分
    invalid_rating_data='{
        "courtId": "'$COURT_ID'",
        "rating": 6,
        "comment": "測試評價"
    }'
    
    test_api "POST" "/reviews" "$invalid_rating_data" 400 "測試無效評分"
    
    # 測試不存在的評價ID
    test_api "GET" "/reviews/non-existent-id" "" 404 "測試不存在的評價ID"
    
    # 測試未認證的操作
    USER_TOKEN_BACKUP=$USER_TOKEN
    USER_TOKEN=""
    
    test_api "POST" "/reviews" "$invalid_review_data" 401 "測試未認證的創建操作"
    
    USER_TOKEN=$USER_TOKEN_BACKUP
}

# 清理測試數據
cleanup() {
    print_header "清理測試數據"
    
    # 刪除評價
    if [ ! -z "$REVIEW_ID" ]; then
        test_api "DELETE" "/reviews/$REVIEW_ID" "" 200 "刪除測試評價"
    fi
    
    # 刪除場地
    if [ ! -z "$COURT_ID" ]; then
        test_api "DELETE" "/courts/$COURT_ID" "" 200 "刪除測試場地"
    fi
}

# 顯示測試結果摘要
show_summary() {
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
    print_header "網球平台評價系統 API 測試"
    
    # 檢查服務器是否運行
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "無法連接到服務器 $BASE_URL"
        print_info "請確保服務器正在運行"
        exit 1
    fi
    
    print_success "服務器連接正常"
    
    # 執行測試
    setup_user
    setup_court
    test_review_crud
    test_review_list
    test_review_statistics
    test_review_reporting
    test_review_helpful
    test_image_upload
    test_error_handling
    cleanup
    
    # 顯示結果
    show_summary
}

# 執行主函數
main "$@"