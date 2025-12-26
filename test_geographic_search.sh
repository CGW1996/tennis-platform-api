#!/bin/bash

# 測試地理搜尋功能
BASE_URL="http://localhost:8080/api/v1"

echo "=== 網球場地地理搜尋功能測試 ==="

# 1. 測試基本搜尋（無地理位置）
echo "1. 測試基本搜尋..."
curl -s -X GET "$BASE_URL/courts?page=1&pageSize=5" | jq '.'

echo -e "\n"

# 2. 測試地理位置搜尋（台北市中心）
echo "2. 測試地理位置搜尋（台北市中心，半徑5公里）..."
curl -s -X GET "$BASE_URL/courts?latitude=25.0330&longitude=121.5654&radius=5&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 3. 測試文字搜尋
echo "3. 測試文字搜尋（搜尋'網球'）..."
curl -s -X GET "$BASE_URL/courts?query=網球&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 4. 測試價格篩選
echo "4. 測試價格篩選（500-1000元）..."
curl -s -X GET "$BASE_URL/courts?minPrice=500&maxPrice=1000&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 5. 測試場地類型篩選
echo "5. 測試場地類型篩選（硬地球場）..."
curl -s -X GET "$BASE_URL/courts?courtType=hard&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 6. 測試設施篩選
echo "6. 測試設施篩選（停車場+照明）..."
curl -s -X GET "$BASE_URL/courts?facilities=parking&facilities=lighting&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 7. 測試評分篩選
echo "7. 測試評分篩選（4星以上）..."
curl -s -X GET "$BASE_URL/courts?minRating=4.0&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 8. 測試距離排序
echo "8. 測試距離排序（台北市中心，按距離排序）..."
curl -s -X GET "$BASE_URL/courts?latitude=25.0330&longitude=121.5654&radius=10&sortBy=distance&sortOrder=asc&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 9. 測試價格排序
echo "9. 測試價格排序（按價格升序）..."
curl -s -X GET "$BASE_URL/courts?sortBy=price&sortOrder=asc&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 10. 測試評分排序
echo "10. 測試評分排序（按評分降序）..."
curl -s -X GET "$BASE_URL/courts?sortBy=rating&sortOrder=desc&page=1&pageSize=5" | jq '.'

echo -e "\n"

# 11. 測試綜合搜尋（地理位置 + 文字 + 篩選）
echo "11. 測試綜合搜尋（台北 + '網球' + 硬地 + 停車場）..."
curl -s -X GET "$BASE_URL/courts?query=網球&latitude=25.0330&longitude=121.5654&radius=10&courtType=hard&facilities=parking&sortBy=distance&sortOrder=asc&page=1&pageSize=5" | jq '.'

echo -e "\n=== 測試完成 ==="