package models

import "time"

// QueryHistory 查询历史模型
type QueryHistory struct {
	ID              int       `json:"id" db:"id"`
	UserID          int       `json:"user_id" db:"user_id"`
	ConnectionID    int       `json:"connection_id" db:"connection_id"`
	SQLQuery        string    `json:"sql_query" db:"sql_query"`
	ExecutionTimeMs int       `json:"execution_time_ms" db:"execution_time_ms"`
	RowsAffected    int       `json:"rows_affected" db:"rows_affected"`
	Success         bool      `json:"success" db:"success"`
	ErrorMessage    string    `json:"error_message" db:"error_message"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// QueryHistoryResponse 查询历史响应
type QueryHistoryResponse struct {
	ID              int       `json:"id"`
	ConnectionID    int       `json:"connection_id"`
	ConnectionName  string    `json:"connection_name,omitempty"`
	SQLQuery        string    `json:"sql_query"`
	ExecutionTimeMs int       `json:"execution_time_ms"`
	RowsAffected    int       `json:"rows_affected"`
	Success         bool      `json:"success"`
	ErrorMessage    string    `json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
}

// QueryHistoryListResponse 查询历史列表响应
type QueryHistoryListResponse struct {
	List     []QueryHistoryResponse `json:"list"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}
