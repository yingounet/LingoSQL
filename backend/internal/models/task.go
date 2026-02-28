package models

import "time"

// Task 任务模型
type Task struct {
	ID          int        `json:"id" db:"id"`
	UserID      int        `json:"user_id" db:"user_id"`
	Type        string     `json:"type" db:"type"`
	Status      string     `json:"status" db:"status"`
	Progress    int        `json:"progress" db:"progress"`
	Payload     string     `json:"-" db:"payload"`
	Result      string     `json:"-" db:"result"`
	ErrorMessage string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty" db:"finished_at"`
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	List     []Task `json:"list"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
