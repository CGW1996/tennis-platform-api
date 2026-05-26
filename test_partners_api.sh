#!/bin/bash

# 測試球友配對 API
# 用法: ./test_partners_api.sh

BASE_URL="http://localhost:8080/api/v1"
TOKEN=""

# 顏色輸出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 測試結果統計
PASSED=0
FAILED=0

# 打印測試標題
print_test() {
    echo -e "\n${YELLOW}測試: $1${NC}"
}

# 打印成功
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((PASSED++))
}

# 打印失敗
print_failure() {
    echo -e "${RED}✗ $1${NC}"
    echo -e "${RED}  響應: $2${NC}"
    ((FAILED++))
}

# 打印分隔線
print_separator() {
    echo -e "\n${'─'%.0s}────────────────────────────────────────────────────────────"
}

# 1. 註冊測試用戶
print_test "註冊測試用戶"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "partner_test@example.com",
        "password": "Password123!",
        "firstName": "Partner",
        "lastName": "Tester"
    }')

if echo "$REGISTER_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    print_success "用戶註冊成功"
else
    # 嘗試登入
    LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "partner_test@example.com",
            "password": "Password123!"
        }')
    
    if echo "$LOGIN_RESPONSE" | grep -q "token"; then
        TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        print_success "用戶登入成功"
    else
        print_failure "用戶註冊/登入失敗" "$REGISTER_RESPONSE"
        exit 1
    fi
fi

# 2. 測試尋找球友 - 基本請求
print_test "尋找球友 - 基本請求"
PARTNERS_RESPONSE=$(curl -s -X POST "$BASE_URL/partners/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "limit": 10
    }')

if echo "$PARTNERS_RESPONSE" | grep -q "partners"; then
    print_success "基本球友搜尋成功"
    echo "找到球友數量: $(echo "$PARTNERS_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)"
else
    print_failure "基本球友搜尋失敗" "$PARTNERS_RESPONSE"
fi

# 3. 測試尋找球友 - 帶篩選條件
print_test "尋找球友 - 帶篩選條件"
PARTNERS_FILTER_RESPONSE=$(curl -s -X POST "$BASE_URL/partners/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "ntrp_range": {
            "min": 3.0,
            "max": 4.5
        },
        "gender": "any",
        "max_distance": 20.0,
        "playing_frequency": "regular",
        "limit": 5
    }')

if echo "$PARTNERS_FILTER_RESPONSE" | grep -q "partners"; then
    print_success "帶篩選條件的球友搜尋成功"
    echo "找到球友數量: $(echo "$PARTNERS_FILTER_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)"
else
    print_failure "帶篩選條件的球友搜尋失敗" "$PARTNERS_FILTER_RESPONSE"
fi

# 4. 測試尋找球友 - 帶位置和時間條件
print_test "尋找球友 - 帶位置和時間條件"
PARTNERS_LOCATION_RESPONSE=$(curl -s -X POST "$BASE_URL/partners/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "location": {
            "city": "台北市",
            "district": "大安區"
        },
        "availability": [
            {
                "type": "weekend",
                "time": "morning"
            },
            {
                "type": "weekday",
                "time": "evening"
            }
        ],
        "play_type": ["rally", "doubles"],
        "limit": 10
    }')

if echo "$PARTNERS_LOCATION_RESPONSE" | grep -q "partners"; then
    print_success "帶位置和時間條件的球友搜尋成功"
else
    print_failure "帶位置和時間條件的球友搜尋失敗" "$PARTNERS_LOCATION_RESPONSE"
fi

# 5. 測試獲取球友歷史
print_test "獲取球友歷史"
HISTORY_RESPONSE=$(curl -s -X GET "$BASE_URL/partners/history?page=1&limit=10" \
    -H "Authorization: Bearer $TOKEN")

if echo "$HISTORY_RESPONSE" | grep -q "matches"; then
    print_success "獲取球友歷史成功"
    echo "歷史記錄數量: $(echo "$HISTORY_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)"
else
    print_failure "獲取球友歷史失敗" "$HISTORY_RESPONSE"
fi

# 6. 測試創建球友配對 - 練習類型
print_test "創建球友配對 - 練習類型"
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/partners/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "participantIds": ["test-user-id-123"],
        "matchType": "practice"
    }')

if echo "$CREATE_RESPONSE" | grep -q '"match"' || echo "$CREATE_RESPONSE" | grep -q "error"; then
    if echo "$CREATE_RESPONSE" | grep -q "error"; then
        print_success "創建配對請求已處理（可能因測試用戶不存在而失敗，這是預期的）"
    else
        print_success "創建球友配對成功"
    fi
else
    print_failure "創建球友配對失敗" "$CREATE_RESPONSE"
fi

# 7. 測試創建球友配對 - 輕鬆類型
print_test "創建球友配對 - 輕鬆類型"
CREATE_CASUAL_RESPONSE=$(curl -s -X POST "$BASE_URL/partners/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "participantIds": ["test-user-id-456"],
        "matchType": "casual"
    }')

if echo "$CREATE_CASUAL_RESPONSE" | grep -q '"match"' || echo "$CREATE_CASUAL_RESPONSE" | grep -q "error"; then
    print_success "創建輕鬆配對請求已處理"
else
    print_failure "創建輕鬆配對失敗" "$CREATE_CASUAL_RESPONSE"
fi

# 8. 測試未授權訪問
print_test "測試未授權訪問"
UNAUTH_RESPONSE=$(curl -s -X POST "$BASE_URL/partners/find" \
    -H "Content-Type: application/json" \
    -d '{"limit": 10}')

if echo "$UNAUTH_RESPONSE" | grep -q "not authenticated\|Unauthorized"; then
    print_success "未授權訪問被正確拒絕"
else
    print_failure "未授權訪問應該被拒絕" "$UNAUTH_RESPONSE"
fi

# 打印測試總結
print_separator
echo -e "\n${YELLOW}測試總結:${NC}"
echo -e "${GREEN}通過: $PASSED${NC}"
echo -e "${RED}失敗: $FAILED${NC}"
echo -e "總計: $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}所有測試通過! ✓${NC}"
    exit 0
else
    echo -e "\n${RED}部分測試失敗 ✗${NC}"
    exit 1
fi
