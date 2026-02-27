package models

// AdminPermissionResponse 管理权限响应
type AdminPermissionResponse struct {
	HasDatabaseAdmin  bool `json:"has_database_admin"`
	HasUserAdmin      bool `json:"has_user_admin"`
	HasPermissionAdmin bool `json:"has_permission_admin"`
}

// CreateDatabaseRequest 创建数据库请求
type CreateDatabaseRequest struct {
	Name      string `json:"name" binding:"required"`
	Charset   string `json:"charset,omitempty"`   // MySQL
	Collation string `json:"collation,omitempty"` // MySQL
	Encoding  string `json:"encoding,omitempty"`  // PostgreSQL
	LcCollate string `json:"lc_collate,omitempty"` // PostgreSQL
	LcCtype   string `json:"lc_ctype,omitempty"`   // PostgreSQL
}

// RenameDatabaseRequest 重命名数据库请求
type RenameDatabaseRequest struct {
	NewName string `json:"new_name" binding:"required"`
}
