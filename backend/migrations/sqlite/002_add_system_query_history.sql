-- 创建系统查询历史表
CREATE TABLE IF NOT EXISTS system_query_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    sql_query TEXT NOT NULL,
    operation_type VARCHAR(50), -- 'GET_DATABASES', 'GET_TABLES', 'GET_TABLE_ROWS', 'USE_DATABASE' 等
    execution_time_ms INTEGER,
    rows_affected INTEGER,
    success BOOLEAN DEFAULT 1,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (connection_id) REFERENCES connections(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_system_query_history_connection_id ON system_query_history(connection_id);
CREATE INDEX IF NOT EXISTS idx_system_query_history_user_id ON system_query_history(user_id);
CREATE INDEX IF NOT EXISTS idx_system_query_history_created_at ON system_query_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_query_history_user_connection ON system_query_history(user_id, connection_id, created_at DESC);
