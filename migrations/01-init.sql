-- 初始化數據庫擴展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 設置時區
SET timezone = 'Asia/Taipei';

-- 創建自定義函數用於更新 updated_at 欄位
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 創建地理位置搜索函數
CREATE OR REPLACE FUNCTION calculate_distance(lat1 float, lon1 float, lat2 float, lon2 float)
RETURNS float AS $$
BEGIN
    RETURN ST_Distance(
        ST_GeogFromText('POINT(' || lon1 || ' ' || lat1 || ')'),
        ST_GeogFromText('POINT(' || lon2 || ' ' || lat2 || ')')
    ) / 1000; -- 返回公里
END;
$$ LANGUAGE plpgsql;