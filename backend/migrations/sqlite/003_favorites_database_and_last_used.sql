-- 收藏表增加 database、last_used_at 字段
ALTER TABLE favorites ADD COLUMN database TEXT;
ALTER TABLE favorites ADD COLUMN last_used_at DATETIME;

-- 用于「最近使用」排序的索引
CREATE INDEX IF NOT EXISTS idx_favorites_last_used_at ON favorites(user_id, last_used_at DESC);
