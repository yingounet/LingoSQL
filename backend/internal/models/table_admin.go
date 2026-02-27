package models

// CreateTableRequest 建表请求
type CreateTableRequest struct {
	Database   string `json:"database" binding:"required"`
	TableName  string `json:"table_name" binding:"required"`
	CreateDDL  string `json:"create_ddl,omitempty"` // 可选，完整 DDL；为空则创建默认简单表
}
