-- 添加 PostGIS 擴展支援地理位置查詢
-- 這個擴展用於數據庫回退搜尋時的地理位置計算

-- 創建 PostGIS 擴展（如果不存在）
CREATE EXTENSION IF NOT EXISTS postgis;

-- 為場地表添加地理位置索引以提升查詢性能
CREATE INDEX IF NOT EXISTS idx_courts_location ON courts USING GIST (ST_Point(longitude, latitude));

-- 添加地理位置查詢函數的註釋
COMMENT ON INDEX idx_courts_location IS '場地地理位置索引，用於提升地理搜尋性能';