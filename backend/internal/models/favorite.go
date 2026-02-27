package models

import "time"

// Favorite 收藏模型
type Favorite struct {
	ID           int        `json:"id" db:"id"`
	UserID       int        `json:"user_id" db:"user_id"`
	ConnectionID int        `json:"connection_id" db:"connection_id"`
	Database     string     `json:"database" db:"database"`
	Name         string     `json:"name" db:"name"`
	SQLQuery     string     `json:"sql_query" db:"sql_query"`
	Description  string     `json:"description" db:"description"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}

// FavoriteCreateRequest 创建收藏请求
type FavoriteCreateRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database"`
	Name         string `json:"name" binding:"required"`
	SQLQuery     string `json:"sql_query" binding:"required"`
	Description  string `json:"description"`
}

// FavoriteUpdateRequest 更新收藏请求
type FavoriteUpdateRequest struct {
	Name        string `json:"name"`
	SQLQuery    string `json:"sql_query"`
	Description string `json:"description"`
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
	ID             int        `json:"id"`
	ConnectionID   int        `json:"connection_id"`
	ConnectionName string     `json:"connection_name,omitempty"`
	Database       string     `json:"database,omitempty"`
	Name           string     `json:"name"`
	SQLQuery       string     `json:"sql_query"`
	Description    string     `json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
}
