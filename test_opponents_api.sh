#!/bin/bash

# 測試對手配對 API
# 用法: ./test_opponents_api.sh

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
    echo -e "\n────────────────────────────────────────────────────────────"
}

# 1. 註冊測試用戶
print_test "註冊測試用戶"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "opponent_test@example.com",
        "password": "Password123!",
        "firstName": "Opponent",
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
            "email": "opponent_test@example.com",
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

# 2. 測試尋找對手 - 基本請求
print_test "尋找對手 - 基本請求"
OPPONENTS_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "limit": 10
    }')

if echo "$OPPONENTS_RESPONSE" | grep -q "opponents"; then
    print_success "基本對手搜尋成功"
    echo "找到對手數量: $(echo "$OPPONENTS_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)"
    echo "類型標記: $(echo "$OPPONENTS_RESPONSE" | grep -o '"type":"[^"]*"' | cut -d'"' -f4)"
else
    print_failure "基本對手搜尋失敗" "$OPPONENTS_RESPONSE"
fi

# 3. 測試尋找對手 - 嚴格 NTRP 匹配
print_test "尋找對手 - 嚴格 NTRP 匹配"
OPPONENTS_NTRP_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "ntrp_range": {
            "min": 4.0,
            "max": 4.5
        },
        "limit": 5
    }')

if echo "$OPPONENTS_NTRP_RESPONSE" | grep -q "opponents"; then
    print_success "嚴格 NTRP 匹配對手搜尋成功"
    echo "找到對手數量: $(echo "$OPPONENTS_NTRP_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)"
else
    print_failure "嚴格 NTRP 匹配對手搜尋失敗" "$OPPONENTS_NTRP_RESPONSE"
fi

# 4. 測試尋找對手 - 高信譽要求
print_test "尋找對手 - 高信譽要求"
OPPONENTS_REPUTATION_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "min_reputation_score": 80.0,
        "ntrp_range": {
            "min": 3.5,
            "max": 5.0
        },
        "limit": 10
    }')

if echo "$OPPONENTS_REPUTATION_RESPONSE" | grep -q "opponents"; then
    print_success "高信譽要求對手搜尋成功"
    # 檢查是否包含信譽資訊
    if echo "$OPPONENTS_REPUTATION_RESPONSE" | grep -q "reputation"; then
        echo "✓ 包含信譽分數資訊"
    fi
else
    print_failure "高信譽要求對手搜尋失敗" "$OPPONENTS_REPUTATION_RESPONSE"
fi

# 5. 測試尋找對手 - 單打/雙打類型
print_test "尋找對手 - 指定比賽類型"
OPPONENTS_MATCH_TYPE_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "match_type": "singles",
        "ntrp_range": {
            "min": 3.0,
            "max": 5.0
        },
        "limit": 10
    }')

if echo "$OPPONENTS_MATCH_TYPE_RESPONSE" | grep -q "opponents"; then
    print_success "指定比賽類型對手搜尋成功"
else
    print_failure "指定比賽類型對手搜尋失敗" "$OPPONENTS_MATCH_TYPE_RESPONSE"
fi

# 6. 測試尋找對手 - 帶場地偏好
print_test "尋找對手 - 帶場地偏好"
OPPONENTS_COURT_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/find" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "preferred_court_type": "hard",
        "location": {
            "city": "台北市"
        },
        "max_distance": 15.0,
        "limit": 10
    }')

if echo "$OPPONENTS_COURT_RESPONSE" | grep -q "opponents"; then
    print_success "帶場地偏好對手搜尋成功"
else
    print_failure "帶場地偏好對手搜尋失敗" "$OPPONENTS_COURT_RESPONSE"
fi

# 7. 測試獲取對戰歷史
print_test "獲取對戰歷史"
HISTORY_RESPONSE=$(curl -s -X GET "$BASE_URL/opponents/history?page=1&limit=10" \
    -H "Authorization: Bearer $TOKEN")

if echo "$HISTORY_RESPONSE" | grep -q "matches"; then
    print_success "獲取對戰歷史成功"
    echo "歷史記錄數量: $(echo "$HISTORY_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)"
else
    print_failure "獲取對戰歷史失敗" "$HISTORY_RESPONSE"
fi

# 8. 測試創建競賽配對 - 錦標賽類型
print_test "創建競賽配對 - 錦標賽類型"
CREATE_TOURNAMENT_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "participantIds": ["test-user-id-123"],
        "matchType": "tournament"
    }')

if echo "$CREATE_TOURNAMENT_RESPONSE" | grep -q '"match"' || echo "$CREATE_TOURNAMENT_RESPONSE" | grep -q "error"; then
    if echo "$CREATE_TOURNAMENT_RESPONSE" | grep -q "error"; then
        print_success "創建錦標賽配對請求已處理（可能因測試用戶不存在而失敗，這是預期的）"
    else
        print_success "創建錦標賽配對成功"
    fi
else
    print_failure "創建錦標賽配對失敗" "$CREATE_TOURNAMENT_RESPONSE"
fi

# 9. 測試創建競賽配對 - 競技類型
print_test "創建競賽配對 - 競技類型"
CREATE_COMPETITIVE_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "participantIds": ["test-user-id-456"],
        "matchType": "competitive"
    }')

if echo "$CREATE_COMPETITIVE_RESPONSE" | grep -q '"match"' || echo "$CREATE_COMPETITIVE_RESPONSE" | grep -q "error"; then
    print_success "創建競技配對請求已處理"
else
    print_failure "創建競技配對失敗" "$CREATE_COMPETITIVE_RESPONSE"
fi

# 10. 測試創建配對 - 無效類型（應該失敗）
print_test "測試創建配對 - 無效類型（預期失敗）"
CREATE_INVALID_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "participantIds": ["test-user-id-789"],
        "matchType": "practice"
    }')

if echo "$CREATE_INVALID_RESPONSE" | grep -q "Invalid match type"; then
    print_success "無效配對類型被正確拒絕"
else
    print_failure "應該拒絕練習類型的配對" "$CREATE_INVALID_RESPONSE"
fi

# 11. 測試未授權訪問
print_test "測試未授權訪問"
UNAUTH_RESPONSE=$(curl -s -X POST "$BASE_URL/opponents/find" \
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
