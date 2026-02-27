package models

// AlterTableRequest 修改表结构请求
type AlterTableRequest struct {
	Database  string                 `json:"database" binding:"required"`
	Table     string                 `json:"table" binding:"required"`
	Operations []AlterTableOperation `json:"operations" binding:"required"`
}

// AlterTableOperation 表结构修改操作
type AlterTableOperation struct {
	Type          string                 `json:"type" binding:"required,oneof=add_column drop_column modify_column rename_column"`
	Column        map[string]interface{} `json:"column,omitempty"`        // 用于 add_column 和 modify_column
	OldColumnName string                 `json:"old_column_name,omitempty"` // 用于 drop_column, modify_column, rename_column
	NewColumnName string                 `json:"new_column_name,omitempty"` // 用于 rename_column
}

// CreateIndexRequest 创建索引请求
type CreateIndexRequest struct {
	Database string                 `json:"database" binding:"required"`
	Table    string                 `json:"table" binding:"required"`
	Index    map[string]interface{} `json:"index" binding:"required"`
}

// DropIndexRequest 删除索引请求
type DropIndexRequest struct {
	Database  string `json:"database" binding:"required"`
	Table     string `json:"table" binding:"required"`
	IndexName string `json:"index_name" binding:"required"`
}

// RenameTableRequest 重命名表请求
type RenameTableRequest struct {
	Database string `json:"database" binding:"required"`
	OldName  string `json:"old_name" binding:"required"`
	NewName  string `json:"new_name" binding:"required"`
}
