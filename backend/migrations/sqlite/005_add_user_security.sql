-- 增加登录安全字段
ALTER TABLE users ADD COLUMN failed_login_count INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN last_failed_login_at DATETIME;
ALTER TABLE users ADD COLUMN locked_until DATETIME;
