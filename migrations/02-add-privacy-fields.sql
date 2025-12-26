-- 添加隱私控制欄位到用戶檔案表

-- 添加位置隱私控制欄位
ALTER TABLE user_profiles 
ADD COLUMN IF NOT EXISTS location_privacy BOOLEAN DEFAULT FALSE;

-- 添加檔案隱私控制欄位
ALTER TABLE user_profiles 
ADD COLUMN IF NOT EXISTS profile_privacy VARCHAR(20) DEFAULT 'public';

-- 添加檔案隱私控制的約束
ALTER TABLE user_profiles 
ADD CONSTRAINT check_profile_privacy 
CHECK (profile_privacy IN ('public', 'friends', 'private'));

-- 為隱私欄位添加索引以提升查詢性能
CREATE INDEX IF NOT EXISTS idx_user_profiles_location_privacy ON user_profiles(location_privacy);
CREATE INDEX IF NOT EXISTS idx_user_profiles_profile_privacy ON user_profiles(profile_privacy);

-- 添加註釋
COMMENT ON COLUMN user_profiles.location_privacy IS '位置隱私設定：true=隱藏精確位置，false=顯示精確位置';
COMMENT ON COLUMN user_profiles.profile_privacy IS '檔案隱私設定：public=公開，friends=僅朋友，private=私人';