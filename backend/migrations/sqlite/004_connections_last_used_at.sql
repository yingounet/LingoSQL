-- 连接表增加 last_used_at 字段，用于记录最后一次连接时间
ALTER TABLE connections ADD COLUMN last_used_at DATETIME;

CREATE INDEX IF NOT EXISTS idx_connections_last_used_at ON connections(user_id, last_used_at DESC);
