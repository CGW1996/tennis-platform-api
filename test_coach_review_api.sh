#!/bin/bash

# 網球平台教練評價 API 測試腳本

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
COACH_TOKEN=""
COACH_ID=""
LESSON_ID=""
REVIEW_ID=""

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

increment_test() {
    ((TOTAL_TESTS++))
}

# 測試 API 響應
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5
    local token=$6

    increment_test
    
    echo -e "\n${YELLOW}測試: $description${NC}"
    echo "請求: $method $endpoint"
    
    if [ -n "$data" ]; then
        echo "數據: $data"
    fi
    
    # 構建 curl 命令
    local curl_cmd="curl -s -w \"HTTPSTATUS:%{http_code}\" -X $method \"$BASE_URL$endpoint\""
    
    if [ -n "$token" ]; then
        curl_cmd="$curl_cmd -H \"Authorization: Bearer $token\""
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H \"$CONTENT_TYPE\" -d '$data'"
    fi
    
    # 執行請求
    local response=$(eval $curl_cmd)
    local http_status=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    local body=$(echo $response | sed -e 's/HTTPSTATUS:.*//g')
    
    echo "響應狀態: $http_status"
    echo "響應內容: $body"
    
    # 檢查狀態碼
    if [ "$http_status" -eq "$expected_status" ]; then
        print_success "$description - 狀態碼正確 ($http_status)"
        
        # 提取重要信息
        case $endpoint in
            "/auth/login")
                if [ "$http_status" -eq 200 ]; then
                    USER_TOKEN=$(echo $body | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
                    echo "提取的用戶 Token: ${USER_TOKEN:0:20}..."
                fi
                ;;
            "/coaches")
                if [ "$http_status" -eq 201 ]; then
                    COACH_ID=$(echo $body | grep -o '"id":"[^"]*' | cut -d'"' -f4)
                    echo "提取的教練 ID: $COACH_ID"
                fi
                ;;
            "/lessons")
                if [ "$http_status" -eq 201 ]; then
                    LESSON_ID=$(echo $body | grep -o '"id":"[^"]*' | cut -d'"' -f4)
                    echo "提取的課程 ID: $LESSON_ID"
                fi
                ;;
            "/coach-reviews")
                if [ "$http_status" -eq 201 ]; then
                    REVIEW_ID=$(echo $body | grep -o '"id":"[^"]*' | cut -d'"' -f4)
                    echo "提取的評價 ID: $REVIEW_ID"
                fi
                ;;
        esac
        
        return 0
    else
        print_error "$description - 狀態碼錯誤 (期望: $expected_status, 實際: $http_status)"
        return 1
    fi
}

# 主測試流程
main() {
    echo -e "${BLUE}網球平台教練評價 API 測試${NC}"
    echo "測試開始時間: $(date)"
    
    # 1. 用戶註冊和登入
    print_test_header "用戶認證測試"
    
    # 註冊測試用戶
    test_api "POST" "/auth/register" '{
        "email": "student@example.com",
        "password": "password123",
        "firstName": "Test",
        "lastName": "Student"
    }' 201 "註冊學生用戶"
    
    # 登入獲取 token
    test_api "POST" "/auth/login" '{
        "email": "student@example.com",
        "password": "password123"
    }' 200 "學生用戶登入"
    
    # 註冊教練用戶
    test_api "POST" "/auth/register" '{
        "email": "coach@example.com",
        "password": "password123",
        "firstName": "Test",
        "lastName": "Coach"
    }' 201 "註冊教練用戶"
    
    # 教練登入
    test_api "POST" "/auth/login" '{
        "email": "coach@example.com",
        "password": "password123"
    }' 200 "教練用戶登入"
    
    COACH_TOKEN=$USER_TOKEN
    
    # 重新登入學生用戶
    test_api "POST" "/auth/login" '{
        "email": "student@example.com",
        "password": "password123"
    }' 200 "重新登入學生用戶"
    
    # 2. 創建教練檔案
    print_test_header "教練檔案創建測試"
    
    test_api "POST" "/coaches" '{
        "experience": 5,
        "specialties": ["beginner", "intermediate"],
        "hourlyRate": 1500,
        "currency": "TWD",
        "languages": ["zh-TW", "en"],
        "biography": "專業網球教練，擁有5年教學經驗"
    }' 201 "創建教練檔案" "$COACH_TOKEN"
    
    # 3. 創建課程（模擬已完成的課程）
    print_test_header "課程創建測試"
    
    if [ -n "$COACH_ID" ]; then
        test_api "POST" "/lessons" '{
            "coachId": "'$COACH_ID'",
            "type": "individual",
            "level": "beginner",
            "duration": 60,
            "price": 1500,
            "scheduledAt": "2024-01-15T10:00:00Z",
            "notes": "初學者網球課程"
        }' 201 "創建課程" "$USER_TOKEN"
        
        # 模擬課程完成（直接更新數據庫狀態）
        if [ -n "$LESSON_ID" ]; then
            echo "注意: 需要手動將課程狀態更新為 'completed' 才能進行評價測試"
        fi
    fi
    
    # 4. 教練評價系統測試
    print_test_header "教練評價系統測試"
    
    # 獲取可用評價標籤
    test_api "GET" "/coach-reviews/available-tags" "" 200 "獲取可用評價標籤"
    
    # 檢查是否可以評價教練
    if [ -n "$COACH_ID" ]; then
        test_api "GET" "/coach-reviews/can-review?coachId=$COACH_ID" "" 200 "檢查是否可以評價教練" "$USER_TOKEN"
        
        if [ -n "$LESSON_ID" ]; then
            test_api "GET" "/coach-reviews/can-review?coachId=$COACH_ID&lessonId=$LESSON_ID" "" 200 "檢查是否可以評價特定課程" "$USER_TOKEN"
        fi
    fi
    
    # 創建教練評價
    if [ -n "$COACH_ID" ]; then
        test_api "POST" "/coach-reviews" '{
            "coachId": "'$COACH_ID'",
            "rating": 5,
            "comment": "非常棒的教練！教學方式很清晰，很有耐心。",
            "tags": ["patient", "professional", "knowledgeable"]
        }' 201 "創建教練評價" "$USER_TOKEN"
    fi
    
    # 獲取教練評價列表
    if [ -n "$COACH_ID" ]; then
        test_api "GET" "/coach-reviews?coachId=$COACH_ID" "" 200 "獲取教練評價列表"
        test_api "GET" "/coach-reviews?coachId=$COACH_ID&rating=5" "" 200 "按評分篩選評價"
        test_api "GET" "/coach-reviews?coachId=$COACH_ID&hasComment=true" "" 200 "篩選有評論的評價"
        test_api "GET" "/coach-reviews?coachId=$COACH_ID&tags=patient,professional" "" 200 "按標籤篩選評價"
    fi
    
    # 獲取評價詳情
    if [ -n "$REVIEW_ID" ]; then
        test_api "GET" "/coach-reviews/$REVIEW_ID" "" 200 "獲取評價詳情"
    fi
    
    # 標記評價有用（需要另一個用戶）
    if [ -n "$REVIEW_ID" ]; then
        # 註冊另一個用戶來標記評價有用
        test_api "POST" "/auth/register" '{
            "email": "user2@example.com",
            "password": "password123",
            "firstName": "Test",
            "lastName": "User2"
        }' 201 "註冊第二個用戶"
        
        test_api "POST" "/auth/login" '{
            "email": "user2@example.com",
            "password": "password123"
        }' 200 "第二個用戶登入"
        
        USER2_TOKEN=$USER_TOKEN
        
        test_api "POST" "/coach-reviews/mark-helpful" '{
            "reviewId": "'$REVIEW_ID'",
            "isHelpful": true
        }' 200 "標記評價有用" "$USER2_TOKEN"
    fi
    
    # 更新評價（切回原用戶）
    test_api "POST" "/auth/login" '{
        "email": "student@example.com",
        "password": "password123"
    }' 200 "切回學生用戶"
    
    if [ -n "$REVIEW_ID" ]; then
        test_api "PUT" "/coach-reviews/$REVIEW_ID" '{
            "rating": 4,
            "comment": "更新評價：教練很好，但時間安排可以更靈活一些。",
            "tags": ["professional", "knowledgeable", "punctual"]
        }' 200 "更新教練評價" "$USER_TOKEN"
    fi
    
    # 獲取教練評價統計
    if [ -n "$COACH_ID" ]; then
        test_api "GET" "/coaches/$COACH_ID/review-statistics" "" 200 "獲取教練評價統計"
    fi
    
    # 5. 錯誤情況測試
    print_test_header "錯誤情況測試"
    
    # 未認證用戶嘗試創建評價
    test_api "POST" "/coach-reviews" '{
        "coachId": "'$COACH_ID'",
        "rating": 5,
        "comment": "測試評價"
    }' 401 "未認證用戶創建評價"
    
    # 無效的評分
    test_api "POST" "/coach-reviews" '{
        "coachId": "'$COACH_ID'",
        "rating": 6,
        "comment": "無效評分測試"
    }' 400 "無效評分測試" "$USER_TOKEN"
    
    # 評價不存在的教練
    test_api "POST" "/coach-reviews" '{
        "coachId": "non-existent-coach-id",
        "rating": 5,
        "comment": "評價不存在的教練"
    }' 400 "評價不存在的教練" "$USER_TOKEN"
    
    # 嘗試更新他人的評價
    if [ -n "$REVIEW_ID" ] && [ -n "$USER2_TOKEN" ]; then
        test_api "PUT" "/coach-reviews/$REVIEW_ID" '{
            "rating": 1,
            "comment": "嘗試惡意修改他人評價"
        }' 400 "嘗試更新他人評價" "$USER2_TOKEN"
    fi
    
    # 刪除評價測試（在24小時內）
    if [ -n "$REVIEW_ID" ]; then
        test_api "DELETE" "/coach-reviews/$REVIEW_ID" "" 204 "刪除評價" "$USER_TOKEN"
    fi
    
    # 6. 測試結果統計
    print_test_header "測試結果統計"
    
    echo -e "\n${BLUE}測試完成統計:${NC}"
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

# 檢查服務器是否運行
check_server() {
    echo "檢查服務器狀態..."
    if curl -s "$BASE_URL/../health" > /dev/null; then
        echo -e "${GREEN}✓ 服務器正在運行${NC}"
        return 0
    else
        echo -e "${RED}✗ 服務器未運行，請先啟動服務器${NC}"
        echo "啟動命令: cd backend && go run cmd/server/main.go"
        exit 1
    fi
}

# 腳本入口
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "網球平台教練評價 API 測試腳本"
    echo ""
    echo "用法: $0 [選項]"
    echo ""
    echo "選項:"
    echo "  -h, --help     顯示此幫助信息"
    echo "  --no-check     跳過服務器狀態檢查"
    echo ""
    echo "環境變量:"
    echo "  BASE_URL       API 基礎 URL (默認: http://localhost:8080/api/v1)"
    echo ""
    exit 0
fi

# 檢查依賴
if ! command -v curl &> /dev/null; then
    echo -e "${RED}錯誤: 需要安裝 curl${NC}"
    exit 1
fi

# 檢查服務器狀態（除非跳過）
if [ "$1" != "--no-check" ]; then
    check_server
fi

# 執行主測試
main