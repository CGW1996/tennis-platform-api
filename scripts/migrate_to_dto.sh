#!/bin/bash

# 此腳本用於將 usecases 包中的 request/response 結構體引用替換為 dto 包的引用

echo "開始更新 controllers 和 usecases 中的引用..."

# 更新 controllers 中的所有文件
for file in /Users/stevenchen/tennis/backend/internal/controllers/*.go; do
    if [[ ! "$file" =~ _test\.go$ ]]; then
        echo "正在處理: $file"
        
        # 添加 dto import (如果文件中使用了 usecases 包)
        if grep -q "tennis-platform/backend/internal/usecases" "$file"; then
            # 在 usecases import 之後添加 dto import
            sed -i.bak '/tennis-platform\/backend\/internal\/usecases/a\
	"tennis-platform/backend/internal/dto"
' "$file"
        fi
        
        # 替換所有的 usecases.XXXRequest 為 dto.XXXRequest
        # 替換所有的 usecases.XXXResponse 為 dto.XXXResponse
        sed -i.bak 's/usecases\.\([A-Z][a-zA-Z]*Request\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\([A-Z][a-zA-Z]*Response\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(CourtWithDistance\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(TimeSlot\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(ReviewStatistics\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(Location\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(ScheduleItem\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(UserInfo\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(PlayingStyleStat\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(UsageDurationStat\)/dto.\1/g' "$file"
        sed -i.bak 's/usecases\.\(RacketReviewStatistics\)/dto.\1/g' "$file"
        
        # 刪除備份文件
        rm -f "$file.bak"
    fi
done

# 更新 usecases 中的所有文件
for file in /Users/stevenchen/tennis/backend/internal/usecases/*.go; do
    if [[ ! "$file" =~ _test\.go$ ]]; then
        echo "正在處理: $file"
        
        # 添加 dto import
        if ! grep -q "tennis-platform/backend/internal/dto" "$file"; then
            # 在 package usecases 之後，import 之前添加 dto import
            sed -i.bak '/^import (/a\
	"tennis-platform/backend/internal/dto"
' "$file"
        fi
        
        # 替換方法簽名中的類型
        sed -i.bak 's/func (.*) \([A-Z][a-zA-Z]*\)(.*\*\([A-Z][a-zA-Z]*Request\))/func (.*) \1(.*\*dto.\2)/g' "$file"
        sed -i.bak 's/) (\*\([A-Z][a-zA-Z]*Response\),/) (*dto.\1,/g' "$file"
        
        # 刪除舊的 type 定義（保留註釋）
        # 這部分需要手動檢查，因為sed不適合處理多行刪除
        
        # 刪除備份文件
        rm -f "$file.bak"
    fi
done

echo "更新完成！請檢查文件並進行必要的手動調整。"
echo "特別注意："
echo "1. 檢查 import 語句是否正確"
echo "2. 檢查是否有重複的 import"
echo "3. 從 usecases 文件中刪除已移到 dto 的類型定義"
