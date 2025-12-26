#!/bin/bash

# 球拍管理 API 測試腳本
# 使用方法: ./test_racket_api.sh

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

# 全局變量
ACCESS_TOKEN=""
RACKET_ID=""
PRICE_ID=""
REVIEW_ID=""

# 輔助函數
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_test() {
    echo -e "\n${YELLOW}測試: $1${NC}"
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

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# 檢查 HTTP 狀態碼
check_status() {
    local expected=$1
    local actual=$2
    local description=$3
    
    if [ "$actual" -eq "$expected" ]; then
        print_success "$description (狀態碼: $actual)"
        return 0
    else
        print_error "$description (期望: $expected, 實際: $actual)"
        return 1
    fi
}

# 檢查回應是否包含特定欄位
check_field() {
    local response=$1
    local field=$2
    local description=$3
    
    if echo "$response" | jq -e ".$field" > /dev/null 2>&1; then
        print_success "$description"
        return 0
    else
        print_error "$description"
        return 1
    fi
}

# 用戶註冊和登入
setup_auth() {
    print_header "設置認證"
    
    # 註冊測試用戶
    print_test "註冊測試用戶"
    local register_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
        -H "$CONTENT_TYPE" \
        -d '{
            "email": "racket_test@example.com",
            "password": "TestPassword123!",
            "firstName": "球拍",
            "lastName": "測試員"
        }')
    
    local register_body=$(echo "$register_response" | head -n -1)
    local register_status=$(echo "$register_response" | tail -n 1)
    
    if [ "$register_status" -eq 201 ] || [ "$register_status" -eq 409 ]; then
        print_success "用戶註冊成功或用戶已存在"
    else
        print_error "用戶註冊失敗 (狀態碼: $register_status)"
    fi
    
    # 登入獲取 token
    print_test "用戶登入"
    local login_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
        -H "$CONTENT_TYPE" \
        -d '{
            "email": "racket_test@example.com",
            "password": "TestPassword123!"
        }')
    
    local login_body=$(echo "$login_response" | head -n -1)
    local login_status=$(echo "$login_response" | tail -n 1)
    
    if check_status 200 "$login_status" "用戶登入"; then
        ACCESS_TOKEN=$(echo "$login_body" | jq -r '.accessToken')
        print_info "獲取到 Access Token: ${ACCESS_TOKEN:0:20}..."
    else
        print_error "無法獲取 Access Token，後續需要認證的測試將失敗"
        echo "登入回應: $login_body"
    fi
}

# 測試球拍規格選項
test_racket_specifications() {
    print_header "球拍規格選項測試"
    
    print_test "獲取球拍規格選項"
    local response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets/specifications")
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -n 1)
    
    if check_status 200 "$status" "獲取球拍規格選項"; then
        check_field "$body" "headSizeRanges" "包含拍面大小範圍"
        check_field "$body" "weightRanges" "包含重量範圍"
        check_field "$body" "stringPatterns" "包含線床模式"
        check_field "$body" "currencies" "包含貨幣選項"
        check_field "$body" "levels" "包含等級選項"
    fi
}

# 測試可用品牌
test_available_brands() {
    print_header "可用品牌測試"
    
    print_test "獲取可用品牌列表"
    local response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets/brands")
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -n 1)
    
    if check_status 200 "$status" "獲取可用品牌列表"; then
        check_field "$body" "brands" "包含品牌列表"
        local brand_count=$(echo "$body" | jq '.brands | length')
        print_info "找到 $brand_count 個品牌"
    fi
}

# 測試創建球拍
test_create_racket() {
    print_header "創建球拍測試"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_error "沒有 Access Token，跳過需要認證的測試"
        return
    fi
    
    print_test "創建新球拍"
    local response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rackets" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d '{
            "brand": "Wilson",
            "model": "Test Pro Staff 97",
            "year": 2023,
            "headSize": 97,
            "weight": 315,
            "balance": 315,
            "stringPattern": "16x19",
            "beamWidth": 21.5,
            "length": 27,
            "stiffness": 68,
            "swingWeight": 335,
            "powerLevel": 6,
            "controlLevel": 9,
            "maneuverLevel": 7,
            "stabilityLevel": 8,
            "description": "測試用球拍，專業級網球拍",
            "images": ["https://example.com/image1.jpg"],
            "msrp": 8500.0,
            "currency": "TWD"
        }')
    
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -n 1)
    
    if check_status 201 "$status" "創建球拍"; then
        RACKET_ID=$(echo "$body" | jq -r '.id')
        print_info "創建的球拍 ID: $RACKET_ID"
        check_field "$body" "brand" "包含品牌資訊"
        check_field "$body" "model" "包含型號資訊"
        check_field "$body" "headSize" "包含拍面大小"
        check_field "$body" "weight" "包含重量"
    else
        print_error "創建球拍失敗，回應: $body"
    fi
    
    # 測試重複創建（應該失敗）
    print_test "測試重複創建球拍（應該失敗）"
    local duplicate_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rackets" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d '{
            "brand": "Wilson",
            "model": "Test Pro Staff 97",
            "headSize": 97,
            "weight": 315,
            "stringPattern": "16x19"
        }')
    
    local duplicate_status=$(echo "$duplicate_response" | tail -n 1)
    check_status 409 "$duplicate_status" "重複創建球拍被拒絕"
}

# 測試獲取球拍詳情
test_get_racket() {
    print_header "獲取球拍詳情測試"
    
    if [ -z "$RACKET_ID" ]; then
        print_error "沒有球拍 ID，跳過測試"
        return
    fi
    
    print_test "獲取球拍詳情"
    local response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets/$RACKET_ID")
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -n 1)
    
    if check_status 200 "$status" "獲取球拍詳情"; then
        check_field "$body" "id" "包含球拍 ID"
        check_field "$body" "brand" "包含品牌"
        check_field "$body" "model" "包含型號"
        check_field "$body" "averageRating" "包含平均評分"
        check_field "$body" "totalReviews" "包含評價總數"
    fi
    
    # 測試獲取不存在的球拍
    print_test "獲取不存在的球拍（應該失敗）"
    local not_found_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets/non-existent-id")
    local not_found_status=$(echo "$not_found_response" | tail -n 1)
    check_status 404 "$not_found_status" "不存在的球拍返回 404"
}

# 測試搜尋球拍
test_search_rackets() {
    print_header "搜尋球拍測試"
    
    print_test "基本搜尋"
    local response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets")
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -n 1)
    
    if check_status 200 "$status" "基本搜尋"; then
        check_field "$body" "rackets" "包含球拍列表"
        check_field "$body" "total" "包含總數"
        check_field "$body" "page" "包含頁碼"
        check_field "$body" "pageSize" "包含每頁數量"
        check_field "$body" "totalPages" "包含總頁數"
    fi
    
    print_test "按品牌搜尋"
    local brand_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets?brand=Wilson")
    local brand_status=$(echo "$brand_response" | tail -n 1)
    check_status 200 "$brand_status" "按品牌搜尋"
    
    print_test "按規格篩選"
    local spec_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets?minHeadSize=95&maxHeadSize=100&minWeight=300&maxWeight=320")
    local spec_status=$(echo "$spec_response" | tail -n 1)
    check_status 200 "$spec_status" "按規格篩選"
    
    print_test "排序測試"
    local sort_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets?sortBy=brand&sortOrder=asc")
    local sort_status=$(echo "$sort_response" | tail -n 1)
    check_status 200 "$sort_status" "排序測試"
}

# 測試更新球拍
test_update_racket() {
    print_header "更新球拍測試"
    
    if [ -z "$ACCESS_TOKEN" ] || [ -z "$RACKET_ID" ]; then
        print_error "沒有 Access Token 或球拍 ID，跳過測試"
        return
    fi
    
    print_test "更新球拍資訊"
    local response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/rackets/$RACKET_ID" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d '{
            "description": "更新後的描述 - 專業級網球拍，適合高級球員",
            "year": 2024,
            "powerLevel": 7
        }')
    
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -n 1)
    
    if check_status 200 "$status" "更新球拍資訊"; then
        local updated_description=$(echo "$body" | jq -r '.description')
        local updated_year=$(echo "$body" | jq -r '.year')
        local updated_power=$(echo "$body" | jq -r '.powerLevel')
        
        if [[ "$updated_description" == *"更新後的描述"* ]]; then
            print_success "描述更新成功"
        else
            print_error "描述更新失敗"
        fi
        
        if [ "$updated_year" -eq 2024 ]; then
            print_success "年份更新成功"
        else
            print_error "年份更新失敗"
        fi
        
        if [ "$updated_power" -eq 7 ]; then
            print_success "力量等級更新成功"
        else
            print_error "力量等級更新失敗"
        fi
    fi
}

# 測試球拍價格管理
test_racket_prices() {
    print_header "球拍價格管理測試"
    
    if [ -z "$ACCESS_TOKEN" ] || [ -z "$RACKET_ID" ]; then
        print_error "沒有 Access Token 或球拍 ID，跳過測試"
        return
    fi
    
    # 創建價格
    print_test "創建球拍價格"
    local create_price_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rackets/$RACKET_ID/prices" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d '{
            "retailer": "測試網球專賣店",
            "price": 7500.0,
            "currency": "TWD",
            "url": "https://example.com/product",
            "isAvailable": true
        }')
    
    local create_price_body=$(echo "$create_price_response" | head -n -1)
    local create_price_status=$(echo "$create_price_response" | tail -n 1)
    
    if check_status 201 "$create_price_status" "創建球拍價格"; then
        PRICE_ID=$(echo "$create_price_body" | jq -r '.id')
        print_info "創建的價格 ID: $PRICE_ID"
        check_field "$create_price_body" "retailer" "包含零售商資訊"
        check_field "$create_price_body" "price" "包含價格"
        check_field "$create_price_body" "isAvailable" "包含可用性"
    fi
    
    # 獲取球拍價格
    print_test "獲取球拍價格列表"
    local get_prices_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets/$RACKET_ID/prices")
    local get_prices_body=$(echo "$get_prices_response" | head -n -1)
    local get_prices_status=$(echo "$get_prices_response" | tail -n 1)
    
    if check_status 200 "$get_prices_status" "獲取球拍價格列表"; then
        check_field "$get_prices_body" "prices" "包含價格列表"
        check_field "$get_prices_body" "lowestPrice" "包含最低價格"
    fi
    
    # 更新價格
    if [ -n "$PRICE_ID" ]; then
        print_test "更新球拍價格"
        local update_price_response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/racket-prices/$PRICE_ID" \
            -H "$CONTENT_TYPE" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -d '{
                "price": 7000.0,
                "isAvailable": true
            }')
        
        local update_price_status=$(echo "$update_price_response" | tail -n 1)
        check_status 200 "$update_price_status" "更新球拍價格"
        
        # 更新價格可用性
        print_test "更新價格可用性"
        local availability_response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/racket-prices/$PRICE_ID/availability" \
            -H "$CONTENT_TYPE" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -d '{
                "isAvailable": false
            }')
        
        local availability_status=$(echo "$availability_response" | tail -n 1)
        check_status 200 "$availability_status" "更新價格可用性"
    fi
}

# 測試球拍評價管理
test_racket_reviews() {
    print_header "球拍評價管理測試"
    
    if [ -z "$ACCESS_TOKEN" ] || [ -z "$RACKET_ID" ]; then
        print_error "沒有 Access Token 或球拍 ID，跳過測試"
        return
    fi
    
    # 創建評價
    print_test "創建球拍評價"
    local create_review_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rackets/$RACKET_ID/reviews" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d '{
            "rating": 5,
            "powerRating": 4,
            "controlRating": 5,
            "comfortRating": 4,
            "comment": "非常好的球拍，控制性極佳，適合進階球員使用",
            "playingStyle": "all-court",
            "usageDuration": 6
        }')
    
    local create_review_body=$(echo "$create_review_response" | head -n -1)
    local create_review_status=$(echo "$create_review_response" | tail -n 1)
    
    if check_status 201 "$create_review_status" "創建球拍評價"; then
        REVIEW_ID=$(echo "$create_review_body" | jq -r '.id')
        print_info "創建的評價 ID: $REVIEW_ID"
        check_field "$create_review_body" "rating" "包含評分"
        check_field "$create_review_body" "comment" "包含評論"
        check_field "$create_review_body" "playingStyle" "包含打法風格"
    fi
    
    # 獲取球拍評價
    print_test "獲取球拍評價列表"
    local get_reviews_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets/$RACKET_ID/reviews")
    local get_reviews_body=$(echo "$get_reviews_response" | head -n -1)
    local get_reviews_status=$(echo "$get_reviews_response" | tail -n 1)
    
    if check_status 200 "$get_reviews_status" "獲取球拍評價列表"; then
        check_field "$get_reviews_body" "reviews" "包含評價列表"
        check_field "$get_reviews_body" "total" "包含總數"
        check_field "$get_reviews_body" "page" "包含頁碼"
    fi
    
    # 獲取評價統計
    print_test "獲取球拍評價統計"
    local stats_response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/rackets/$RACKET_ID/reviews/statistics")
    local stats_body=$(echo "$stats_response" | head -n -1)
    local stats_status=$(echo "$stats_response" | tail -n 1)
    
    if check_status 200 "$stats_status" "獲取球拍評價統計"; then
        check_field "$stats_body" "totalReviews" "包含評價總數"
        check_field "$stats_body" "averageRating" "包含平均評分"
        check_field "$stats_body" "ratingDistribution" "包含評分分佈"
        check_field "$stats_body" "playingStyleStats" "包含打法統計"
    fi
    
    # 標記評價有用
    if [ -n "$REVIEW_ID" ]; then
        print_test "標記評價有用"
        local helpful_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/racket-reviews/$REVIEW_ID/helpful" \
            -H "$CONTENT_TYPE" \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -d '{
                "helpful": true
            }')
        
        local helpful_status=$(echo "$helpful_response" | tail -n 1)
        # 注意：用戶不能標記自己的評價為有用，所以這裡應該返回錯誤
        if [ "$helpful_status" -eq 400 ] || [ "$helpful_status" -eq 403 ]; then
            print_success "正確阻止用戶標記自己的評價"
        else
            print_error "應該阻止用戶標記自己的評價"
        fi
    fi
}

# 測試圖片上傳
test_image_upload() {
    print_header "圖片上傳測試"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_error "沒有 Access Token，跳過測試"
        return
    fi
    
    # 創建測試圖片文件
    echo "測試圖片內容" > /tmp/test_racket_image.txt
    
    print_test "上傳球拍圖片"
    local upload_response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rackets/images" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -F "images=@/tmp/test_racket_image.txt")
    
    local upload_body=$(echo "$upload_response" | head -n -1)
    local upload_status=$(echo "$upload_response" | tail -n 1)
    
    if check_status 200 "$upload_status" "上傳球拍圖片"; then
        check_field "$upload_body" "images" "包含圖片 URL 列表"
    fi
    
    # 清理測試文件
    rm -f /tmp/test_racket_image.txt
}

# 清理測試數據
cleanup_test_data() {
    print_header "清理測試數據"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_error "沒有 Access Token，無法清理數據"
        return
    fi
    
    # 刪除價格
    if [ -n "$PRICE_ID" ]; then
        print_test "刪除測試價格"
        local delete_price_response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/racket-prices/$PRICE_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        local delete_price_status=$(echo "$delete_price_response" | tail -n 1)
        check_status 204 "$delete_price_status" "刪除測試價格"
    fi
    
    # 刪除球拍
    if [ -n "$RACKET_ID" ]; then
        print_test "刪除測試球拍"
        local delete_racket_response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/rackets/$RACKET_ID" \
            -H "Authorization: Bearer $ACCESS_TOKEN")
        local delete_racket_status=$(echo "$delete_racket_response" | tail -n 1)
        check_status 204 "$delete_racket_status" "刪除測試球拍"
    fi
}

# 顯示測試結果摘要
show_test_summary() {
    print_header "測試結果摘要"
    
    echo -e "總測試數: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "通過測試: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失敗測試: ${RED}$FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 所有測試都通過了！${NC}"
        exit 0
    else
        echo -e "\n${RED}❌ 有 $FAILED_TESTS 個測試失敗${NC}"
        exit 1
    fi
}

# 主測試流程
main() {
    echo -e "${BLUE}球拍管理 API 測試開始${NC}"
    echo -e "測試服務器: $BASE_URL"
    echo -e "時間: $(date)"
    
    # 檢查服務器是否運行
    print_header "檢查服務器狀態"
    local health_response=$(curl -s -w "\n%{http_code}" -X GET "http://localhost:8080/health")
    local health_status=$(echo "$health_response" | tail -n 1)
    
    if ! check_status 200 "$health_status" "服務器健康檢查"; then
        echo -e "${RED}服務器未運行或無法訪問，請先啟動服務器${NC}"
        exit 1
    fi
    
    # 執行測試
    setup_auth
    test_racket_specifications
    test_available_brands
    test_create_racket
    test_get_racket
    test_search_rackets
    test_update_racket
    test_racket_prices
    test_racket_reviews
    test_image_upload
    cleanup_test_data
    
    # 顯示結果
    show_test_summary
}

# 檢查依賴
if ! command -v curl &> /dev/null; then
    echo -e "${RED}錯誤: 需要安裝 curl${NC}"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "${RED}錯誤: 需要安裝 jq${NC}"
    exit 1
fi

# 執行主函數
main "$@"