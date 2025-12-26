#!/bin/bash

# Elasticsearch 管理工具腳本

ELASTICSEARCH_URL="http://localhost:9200"
API_URL="http://localhost:8080/api/v1"

echo "=== Elasticsearch 管理工具 ==="

# 檢查 Elasticsearch 狀態
check_elasticsearch() {
    echo "檢查 Elasticsearch 狀態..."
    curl -s "$ELASTICSEARCH_URL/_cluster/health" | jq '.'
}

# 檢查索引狀態
check_index() {
    echo "檢查 courts 索引狀態..."
    curl -s "$ELASTICSEARCH_URL/courts/_stats" | jq '.indices.courts.total'
}

# 查看索引映射
show_mapping() {
    echo "顯示 courts 索引映射..."
    curl -s "$ELASTICSEARCH_URL/courts/_mapping" | jq '.'
}

# 搜尋所有文檔
search_all() {
    echo "搜尋所有場地文檔..."
    curl -s "$ELASTICSEARCH_URL/courts/_search?size=5" | jq '.hits'
}

# 刪除索引
delete_index() {
    echo "刪除 courts 索引..."
    curl -X DELETE "$ELASTICSEARCH_URL/courts"
    echo "索引已刪除"
}

# 重建索引
rebuild_index() {
    echo "重建索引..."
    delete_index
    echo "等待 3 秒..."
    sleep 3
    
    echo "觸發批量索引..."
    curl -X POST "$API_URL/courts/bulk-index" \
        -H "Content-Type: application/json"
    echo "重建完成"
}

# 測試地理搜尋
test_geo_search() {
    echo "測試地理搜尋（台北市中心 5km 範圍）..."
    curl -s "$ELASTICSEARCH_URL/courts/_search" \
        -H "Content-Type: application/json" \
        -d '{
            "query": {
                "bool": {
                    "must": [
                        {"term": {"is_active": true}}
                    ],
                    "filter": [
                        {
                            "geo_distance": {
                                "distance": "5km",
                                "location": {
                                    "lat": 25.0330,
                                    "lon": 121.5654
                                }
                            }
                        }
                    ]
                }
            },
            "sort": [
                {
                    "_geo_distance": {
                        "location": {
                            "lat": 25.0330,
                            "lon": 121.5654
                        },
                        "order": "asc",
                        "unit": "km"
                    }
                }
            ],
            "size": 5
        }' | jq '.hits'
}

# 測試文字搜尋
test_text_search() {
    echo "測試文字搜尋（搜尋'網球'）..."
    curl -s "$ELASTICSEARCH_URL/courts/_search" \
        -H "Content-Type: application/json" \
        -d '{
            "query": {
                "bool": {
                    "must": [
                        {"term": {"is_active": true}},
                        {
                            "multi_match": {
                                "query": "網球",
                                "fields": ["name^2", "description", "address"],
                                "type": "best_fields"
                            }
                        }
                    ]
                }
            },
            "size": 5
        }' | jq '.hits'
}

# 顯示幫助
show_help() {
    echo "使用方法: $0 [命令]"
    echo ""
    echo "可用命令:"
    echo "  status      - 檢查 Elasticsearch 狀態"
    echo "  index       - 檢查索引狀態"
    echo "  mapping     - 顯示索引映射"
    echo "  search      - 搜尋所有文檔"
    echo "  delete      - 刪除索引"
    echo "  rebuild     - 重建索引"
    echo "  test-geo    - 測試地理搜尋"
    echo "  test-text   - 測試文字搜尋"
    echo "  help        - 顯示此幫助"
}

# 主程序
case "$1" in
    "status")
        check_elasticsearch
        ;;
    "index")
        check_index
        ;;
    "mapping")
        show_mapping
        ;;
    "search")
        search_all
        ;;
    "delete")
        delete_index
        ;;
    "rebuild")
        rebuild_index
        ;;
    "test-geo")
        test_geo_search
        ;;
    "test-text")
        test_text_search
        ;;
    "help"|"")
        show_help
        ;;
    *)
        echo "未知命令: $1"
        show_help
        exit 1
        ;;
esac