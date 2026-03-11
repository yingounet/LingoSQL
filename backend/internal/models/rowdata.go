package models

// GetTableDataRequest 获取表数据请求
type GetTableDataRequest struct {
	ConnectionID int                    `json:"connection_id" binding:"required"`
	Database     string                 `json:"database" binding:"required"`
	Table        string                 `json:"table" binding:"required"`
	Filters      []map[string]interface{} `json:"filters,omitempty"`
	Page         int                    `json:"page,omitempty"`
	PageSize     int                    `json:"pageSize,omitempty"` // 前端使用 pageSize
}

// UpdateTableRowDataRequest 更新表行数据请求
type UpdateTableRowDataRequest struct {
	ConnectionID int                    `json:"connection_id" binding:"required"`
	Database     string                 `json:"database" binding:"required"`
	Table        string                 `json:"table" binding:"required"`
	PrimaryKey   map[string]interface{} `json:"primary_key" binding:"required"`
	UpdateData   map[string]interface{} `json:"update_data" binding:"required"`
}

// BatchInsertRequest 批量插入请求
type BatchInsertRequest struct {
	ConnectionID int           `json:"connection_id" binding:"required"`
	Database     string        `json:"database" binding:"required"`
	Table        string        `json:"table" binding:"required"`
	Data         [][]interface{} `json:"data" binding:"required"`
	Columns      []string      `json:"columns" binding:"required"`
}

// BatchUpdateRequest 批量更新请求
type BatchUpdateRequest struct {
	ConnectionID int                    `json:"connection_id" binding:"required"`
	Database     string                 `json:"database" binding:"required"`
	Table        string                 `json:"table" binding:"required"`
	Filters      []map[string]interface{} `json:"filters" binding:"required"` // 筛选条件
	UpdateData   map[string]interface{} `json:"update_data" binding:"required"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	ConnectionID int                    `json:"connection_id" binding:"required"`
	Database     string                 `json:"database" binding:"required"`
	Table        string                 `json:"table" binding:"required"`
	Filters      []map[string]interface{} `json:"filters" binding:"required"` // 筛选条件
}

// DeleteByPrimaryKeysRequest 按主键删除请求
type DeleteByPrimaryKeysRequest struct {
	ConnectionID  int                       `json:"connection_id" binding:"required"`
	Database      string                    `json:"database" binding:"required"`
	Table         string                    `json:"table" binding:"required"`
	PrimaryKeys   []map[string]interface{}  `json:"primary_keys" binding:"required"` // 主键值列表，每个元素为 {col1: val1, col2: val2}
}

// CompareDataRequest 数据对比请求
type CompareDataRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database1    string `json:"database1" binding:"required"`
	Table1       string `json:"table1" binding:"required"`
	Database2    string `json:"database2,omitempty"` // 为空则使用 database1
	Table2       string `json:"table2" binding:"required"`
	KeyColumns   []string `json:"key_columns" binding:"required"`
}

// CompareDataResponse 数据对比响应
type CompareDataResponse struct {
	OnlyInTable1 []map[string]interface{} `json:"only_in_table1"`
	OnlyInTable2 []map[string]interface{} `json:"only_in_table2"`
	Different    []struct {
		Key        map[string]interface{} `json:"key"`
		Table1Data map[string]interface{} `json:"table1_data"`
		Table2Data map[string]interface{} `json:"table2_data"`
		Differences []string              `json:"differences"`
	} `json:"different"`
	SameCount int `json:"same_count"`
}

// FindReplaceRequest 查找替换请求
type FindReplaceRequest struct {
	ConnectionID int                    `json:"connection_id" binding:"required"`
	Database     string                 `json:"database" binding:"required"`
	Table        string                 `json:"table" binding:"required"`
	Column       string                 `json:"column" binding:"required"`
	FindValue    string                 `json:"find_value" binding:"required"`
	ReplaceValue string                 `json:"replace_value" binding:"required"`
	Filters      []map[string]interface{} `json:"filters,omitempty"` // 可选的筛选条件
}

// FindReplaceResponse 查找替换响应
type FindReplaceResponse struct {
	AffectedRows int `json:"affected_rows"`
	MatchedRows  int `json:"matched_rows"`
}
