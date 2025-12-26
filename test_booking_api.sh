#!/bin/bash

# 場地預訂 API 測試腳本
# 使用方法: ./test_booking_api.sh

set -e

# 配置
BASE_URL="http://localhost:8080/api/v1"
CONTENT_TYPE="Content-Type: application/json"

# 顏色輸出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 輔助函數
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 檢查服務器是否運行
check_server() {
    print_header "檢查服務器狀態"
    
    if curl -s "$BASE_URL/../health" > /dev/null; then
        print_success "服務器運行正常"
    else
        print_error "服務器未運行，請先啟動服務器"
        exit 1
    fi
}

# 用戶註冊和登入
setup_auth() {
    print_header "設置認證"
    
    # 註冊測試用戶
    echo "註冊測試用戶..."
    REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
        -H "$CONTENT_TYPE" \
        -d '{
            "email": "booking_test@example.com",
            "password": "password123",
            "confirmPassword": "password123"
        }')
    
    if echo "$REGISTER_RESPONSE" | grep -q "error"; then
        print_warning "用戶可能已存在，嘗試登入"
    else
        print_success "用戶註冊成功"
    fi
    
    # 登入獲取 token
    echo "用戶登入..."
    LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "$CONTENT_TYPE" \
        -d '{
            "email": "booking_test@example.com",
            "password": "password123"
        }')
    
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
    USER_ID=$(echo "$LOGIN_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$ACCESS_TOKEN" ]; then
        print_error "登入失敗"
        echo "回應: $LOGIN_RESPONSE"
        exit 1
    fi
    
    print_success "登入成功，獲得 access token"
    AUTH_HEADER="Authorization: Bearer $ACCESS_TOKEN"
}

# 創建測試場地
create_test_court() {
    print_header "創建測試場地"
    
    COURT_RESPONSE=$(curl -s -X POST "$BASE_URL/courts" \
        -H "$CONTENT_TYPE" \
        -H "$AUTH_HEADER" \
        -d '{
            "name": "預訂測試場地",
            "description": "用於測試預訂功能的場地",
            "address": "台北市信義區信義路五段7號",
            "latitude": 25.0330,
            "longitude": 121.5654,
            "facilities": ["parking", "restroom", "lighting"],
            "courtType": "hard",
            "pricePerHour": 150.0,
            "currency": "TWD",
            "operatingHours": {
                "Monday": "09:00-21:00",
                "Tuesday": "09:00-21:00",
                "Wednesday": "09:00-21:00",
                "Thursday": "09:00-21:00",
                "Friday": "09:00-21:00",
                "Saturday": "08:00-22:00",
                "Sunday": "08:00-22:00"
            },
            "contactPhone": "02-1234-5678",
            "contactEmail": "test@court.com"
        }')
    
    COURT_ID=$(echo "$COURT_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$COURT_ID" ]; then
        print_error "創建場地失敗"
        echo "回應: $COURT_RESPONSE"
        exit 1
    fi
    
    print_success "測試場地創建成功，ID: $COURT_ID"
}

# 測試查詢可用時間
test_availability() {
    print_header "測試查詢場地可用時間"
    
    # 查詢明天的可用時間
    TOMORROW=$(date -d "+1 day" +%Y-%m-%d)
    
    echo "查詢 $TOMORROW 的可用時間..."
    AVAILABILITY_RESPONSE=$(curl -s -X GET "$BASE_URL/courts/availability?courtId=$COURT_ID&date=$TOMORROW&duration=120")
    
    if echo "$AVAILABILITY_RESPONSE" | grep -q "timeSlots"; then
        print_success "可用時間查詢成功"
        echo "可用時間段數量: $(echo "$AVAILABILITY_RESPONSE" | grep -o '"available":true' | wc -l)"
    else
        print_error "可用時間查詢失敗"
        echo "回應: $AVAILABILITY_RESPONSE"
        return 1
    fi
}

# 測試創建預訂
test_create_booking() {
    print_header "測試創建預訂"
    
    # 計算明天上午10點到12點的時間
    TOMORROW_10AM=$(date -d "+1 day 10:00" -u +%Y-%m-%dT%H:%M:%SZ)
    TOMORROW_12PM=$(date -d "+1 day 12:00" -u +%Y-%m-%dT%H:%M:%SZ)
    
    echo "創建預訂: $TOMORROW_10AM 到 $TOMORROW_12PM"
    BOOKING_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings" \
        -H "$CONTENT_TYPE" \
        -H "$AUTH_HEADER" \
        -d "{
            \"courtId\": \"$COURT_ID\",
            \"startTime\": \"$TOMORROW_10AM\",
            \"endTime\": \"$TOMORROW_12PM\",
            \"notes\": \"API 測試預訂\"
        }")
    
    BOOKING_ID=$(echo "$BOOKING_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$BOOKING_ID" ]; then
        print_error "創建預訂失敗"
        echo "回應: $BOOKING_RESPONSE"
        return 1
    fi
    
    print_success "預訂創建成功，ID: $BOOKING_ID"
    
    # 檢查預訂狀態
    STATUS=$(echo "$BOOKING_RESPONSE" | grep -o '"status":"[^"]*' | cut -d'"' -f4)
    TOTAL_PRICE=$(echo "$BOOKING_RESPONSE" | grep -o '"totalPrice":[0-9.]*' | cut -d':' -f2)
    
    echo "預訂狀態: $STATUS"
    echo "總價格: $TOTAL_PRICE TWD"
}

# 測試獲取預訂詳情
test_get_booking() {
    print_header "測試獲取預訂詳情"
    
    if [ -z "$BOOKING_ID" ]; then
        print_warning "跳過測試：沒有可用的預訂ID"
        return 0
    fi
    
    echo "獲取預訂詳情: $BOOKING_ID"
    BOOKING_DETAIL_RESPONSE=$(curl -s -X GET "$BASE_URL/bookings/$BOOKING_ID" \
        -H "$AUTH_HEADER")
    
    if echo "$BOOKING_DETAIL_RESPONSE" | grep -q "\"id\":\"$BOOKING_ID\""; then
        print_success "獲取預訂詳情成功"
        
        # 顯示預訂信息
        COURT_NAME=$(echo "$BOOKING_DETAIL_RESPONSE" | grep -o '"name":"[^"]*' | cut -d'"' -f4)
        START_TIME=$(echo "$BOOKING_DETAIL_RESPONSE" | grep -o '"startTime":"[^"]*' | cut -d'"' -f4)
        
        echo "場地名稱: $COURT_NAME"
        echo "開始時間: $START_TIME"
    else
        print_error "獲取預訂詳情失敗"
        echo "回應: $BOOKING_DETAIL_RESPONSE"
        return 1
    fi
}

# 測試更新預訂
test_update_booking() {
    print_header "測試更新預訂"
    
    if [ -z "$BOOKING_ID" ]; then
        print_warning "跳過測試：沒有可用的預訂ID"
        return 0
    fi
    
    echo "更新預訂備註和狀態..."
    UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/bookings/$BOOKING_ID" \
        -H "$CONTENT_TYPE" \
        -H "$AUTH_HEADER" \
        -d '{
            "notes": "更新後的備註 - API 測試",
            "status": "confirmed"
        }')
    
    if echo "$UPDATE_RESPONSE" | grep -q "confirmed"; then
        print_success "預訂更新成功"
        echo "新狀態: confirmed"
    else
        print_error "預訂更新失敗"
        echo "回應: $UPDATE_RESPONSE"
        return 1
    fi
}

# 測試獲取預訂列表
test_get_bookings() {
    print_header "測試獲取預訂列表"
    
    echo "獲取用戶的所有預訂..."
    BOOKINGS_RESPONSE=$(curl -s -X GET "$BASE_URL/bookings?userId=$USER_ID&page=1&pageSize=10" \
        -H "$AUTH_HEADER")
    
    if echo "$BOOKINGS_RESPONSE" | grep -q "bookings"; then
        print_success "獲取預訂列表成功"
        
        TOTAL=$(echo "$BOOKINGS_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)
        echo "總預訂數量: $TOTAL"
    else
        print_error "獲取預訂列表失敗"
        echo "回應: $BOOKINGS_RESPONSE"
        return 1
    fi
    
    # 測試按狀態篩選
    echo "測試按狀態篩選預訂..."
    CONFIRMED_BOOKINGS=$(curl -s -X GET "$BASE_URL/bookings?status=confirmed&page=1&pageSize=10" \
        -H "$AUTH_HEADER")
    
    if echo "$CONFIRMED_BOOKINGS" | grep -q "bookings"; then
        print_success "按狀態篩選成功"
    else
        print_warning "按狀態篩選可能失敗"
    fi
}

# 測試時間衝突檢測
test_conflict_detection() {
    print_header "測試時間衝突檢測"
    
    if [ -z "$COURT_ID" ]; then
        print_warning "跳過測試：沒有可用的場地ID"
        return 0
    fi
    
    # 嘗試創建衝突的預訂
    TOMORROW_11AM=$(date -d "+1 day 11:00" -u +%Y-%m-%dT%H:%M:%SZ)
    TOMORROW_1PM=$(date -d "+1 day 13:00" -u +%Y-%m-%dT%H:%M:%SZ)
    
    echo "嘗試創建衝突預訂: $TOMORROW_11AM 到 $TOMORROW_1PM"
    CONFLICT_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings" \
        -H "$CONTENT_TYPE" \
        -H "$AUTH_HEADER" \
        -d "{
            \"courtId\": \"$COURT_ID\",
            \"startTime\": \"$TOMORROW_11AM\",
            \"endTime\": \"$TOMORROW_1PM\",
            \"notes\": \"衝突測試預訂\"
        }")
    
    if echo "$CONFLICT_RESPONSE" | grep -q "已被預訂\|衝突"; then
        print_success "時間衝突檢測正常工作"
    else
        print_warning "時間衝突檢測可能有問題"
        echo "回應: $CONFLICT_RESPONSE"
    fi
}

# 測試取消預訂
test_cancel_booking() {
    print_header "測試取消預訂"
    
    if [ -z "$BOOKING_ID" ]; then
        print_warning "跳過測試：沒有可用的預訂ID"
        return 0
    fi
    
    echo "取消預訂: $BOOKING_ID"
    CANCEL_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings/$BOOKING_ID/cancel" \
        -H "$AUTH_HEADER")
    
    if echo "$CANCEL_RESPONSE" | grep -q "取消成功"; then
        print_success "預訂取消成功"
    else
        print_error "預訂取消失敗"
        echo "回應: $CANCEL_RESPONSE"
        return 1
    fi
    
    # 驗證預訂狀態已更改
    echo "驗證預訂狀態..."
    CANCELLED_BOOKING=$(curl -s -X GET "$BASE_URL/bookings/$BOOKING_ID" \
        -H "$AUTH_HEADER")
    
    if echo "$CANCELLED_BOOKING" | grep -q "cancelled"; then
        print_success "預訂狀態已更新為 cancelled"
    else
        print_warning "預訂狀態更新可能有問題"
    fi
}

# 測試邊界條件
test_edge_cases() {
    print_header "測試邊界條件"
    
    if [ -z "$COURT_ID" ]; then
        print_warning "跳過測試：沒有可用的場地ID"
        return 0
    fi
    
    # 測試過去時間預訂
    echo "測試預訂過去時間（應該失敗）..."
    YESTERDAY=$(date -d "-1 day 10:00" -u +%Y-%m-%dT%H:%M:%SZ)
    YESTERDAY_12=$(date -d "-1 day 12:00" -u +%Y-%m-%dT%H:%M:%SZ)
    
    PAST_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings" \
        -H "$CONTENT_TYPE" \
        -H "$AUTH_HEADER" \
        -d "{
            \"courtId\": \"$COURT_ID\",
            \"startTime\": \"$YESTERDAY\",
            \"endTime\": \"$YESTERDAY_12\",
            \"notes\": \"過去時間測試\"
        }")
    
    if echo "$PAST_RESPONSE" | grep -q "過去的時間\|error"; then
        print_success "過去時間預訂正確被拒絕"
    else
        print_warning "過去時間預訂檢查可能有問題"
    fi
    
    # 測試超長預訂
    echo "測試超長預訂（應該失敗）..."
    TOMORROW_9AM=$(date -d "+1 day 09:00" -u +%Y-%m-%dT%H:%M:%SZ)
    TOMORROW_9PM=$(date -d "+1 day 21:00" -u +%Y-%m-%dT%H:%M:%SZ)
    
    LONG_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings" \
        -H "$CONTENT_TYPE" \
        -H "$AUTH_HEADER" \
        -d "{
            \"courtId\": \"$COURT_ID\",
            \"startTime\": \"$TOMORROW_9AM\",
            \"endTime\": \"$TOMORROW_9PM\",
            \"notes\": \"超長預訂測試\"
        }")
    
    if echo "$LONG_RESPONSE" | grep -q "超過.*小時\|error"; then
        print_success "超長預訂正確被拒絕"
    else
        print_warning "超長預訂檢查可能有問題"
    fi
    
    # 測試無效場地ID
    echo "測試無效場地ID（應該失敗）..."
    INVALID_RESPONSE=$(curl -s -X POST "$BASE_URL/bookings" \
        -H "$CONTENT_TYPE" \
        -H "$AUTH_HEADER" \
        -d '{
            "courtId": "invalid-court-id",
            "startTime": "'$TOMORROW_10AM'",
            "endTime": "'$TOMORROW_12PM'",
            "notes": "無效場地測試"
        }')
    
    if echo "$INVALID_RESPONSE" | grep -q "不存在\|error"; then
        print_success "無效場地ID正確被拒絕"
    else
        print_warning "無效場地ID檢查可能有問題"
    fi
}

# 清理測試數據
cleanup() {
    print_header "清理測試數據"
    
    # 刪除測試場地（會級聯刪除相關預訂）
    if [ -n "$COURT_ID" ]; then
        echo "刪除測試場地..."
        DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/courts/$COURT_ID" \
            -H "$AUTH_HEADER")
        
        if echo "$DELETE_RESPONSE" | grep -q "刪除成功"; then
            print_success "測試場地刪除成功"
        else
            print_warning "測試場地刪除可能失敗"
        fi
    fi
    
    print_success "測試數據清理完成"
}

# 主測試流程
main() {
    echo -e "${BLUE}"
    echo "========================================"
    echo "       場地預訂 API 測試腳本"
    echo "========================================"
    echo -e "${NC}"
    
    # 執行測試
    check_server
    setup_auth
    create_test_court
    
    # 預訂功能測試
    test_availability
    test_create_booking
    test_get_booking
    test_update_booking
    test_get_bookings
    test_conflict_detection
    test_cancel_booking
    
    # 邊界條件測試
    test_edge_cases
    
    # 清理
    cleanup
    
    print_header "測試完成"
    print_success "所有預訂 API 測試已完成！"
    
    echo -e "\n${YELLOW}注意事項:${NC}"
    echo "1. 確保服務器在 localhost:8080 運行"
    echo "2. 確保數據庫連接正常"
    echo "3. 某些測試可能因為業務邏輯而失敗，這是正常的"
    echo "4. 檢查服務器日誌以獲取更多詳細信息"
}

# 執行主函數
main "$@"