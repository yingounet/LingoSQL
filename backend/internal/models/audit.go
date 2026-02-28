package models

import "time"

// AuditLog 审计日志模型
type AuditLog struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Action      string    `json:"action" db:"action"`
	ResourceType string   `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID  *int      `json:"resource_id,omitempty" db:"resource_id"`
	Success     bool      `json:"success" db:"success"`
	ErrorMessage string   `json:"error_message,omitempty" db:"error_message"`
	Details     string    `json:"details,omitempty" db:"details"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AuditLogListResponse 审计日志列表响应
type AuditLogListResponse struct {
	List     []AuditLog `json:"list"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
