package models

import "time"

// SystemQueryHistory 系统查询历史模型
type SystemQueryHistory struct {
	ID              int       `json:"id" db:"id"`
	ConnectionID     int       `json:"connection_id" db:"connection_id"`
	UserID          int       `json:"user_id" db:"user_id"`
	SQLQuery        string    `json:"sql_query" db:"sql_query"`
	OperationType   string    `json:"operation_type" db:"operation_type"`
	ExecutionTimeMs int       `json:"execution_time_ms" db:"execution_time_ms"`
	RowsAffected    int       `json:"rows_affected" db:"rows_affected"`
	Success         bool      `json:"success" db:"success"`
	ErrorMessage    string    `json:"error_message" db:"error_message"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// SystemQueryHistoryResponse 系统查询历史响应
type SystemQueryHistoryResponse struct {
	ID              int       `json:"id"`
	ConnectionID    int       `json:"connection_id"`
	ConnectionName  string    `json:"connection_name,omitempty"`
	SQLQuery        string    `json:"sql_query"`
	OperationType   string    `json:"operation_type"`
	ExecutionTimeMs int       `json:"execution_time_ms"`
	RowsAffected    int       `json:"rows_affected"`
	Success         bool      `json:"success"`
	ErrorMessage    string    `json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
}

// SystemQueryHistoryListResponse 系统查询历史列表响应
type SystemQueryHistoryListResponse struct {
	List     []SystemQueryHistoryResponse `json:"list"`
	Total    int                          `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
}
