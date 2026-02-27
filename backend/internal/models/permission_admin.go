package models

// GrantPermissionRequest 授予权限请求
type GrantPermissionRequest struct {
	Username    string   `json:"username" binding:"required"`
	Host        string   `json:"host,omitempty"` // MySQL only
	TargetType  string   `json:"target_type" binding:"required"` // database, table, column
	TargetName  string   `json:"target_name" binding:"required"`
	Database    string   `json:"database,omitempty"` // for table/column
	Table       string   `json:"table,omitempty"`    // for column
	Privileges  []string `json:"privileges" binding:"required"`
}

// RevokePermissionRequest 撤销权限请求
type RevokePermissionRequest struct {
	Username   string   `json:"username" binding:"required"`
	Host       string   `json:"host,omitempty"` // MySQL only
	TargetType string   `json:"target_type" binding:"required"`
	TargetName string   `json:"target_name" binding:"required"`
	Database   string   `json:"database,omitempty"`
	Table      string   `json:"table,omitempty"`
	Privileges []string `json:"privileges" binding:"required"`
}
