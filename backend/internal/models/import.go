package models

// ImportDataRequest 导入数据请求
type ImportDataRequest struct {
	ConnectionID int                    `json:"connection_id" binding:"required"`
	Database     string                 `json:"database" binding:"required"`
	Table        string                 `json:"table" binding:"required"`
	Data         [][]interface{}        `json:"data" binding:"required"`
	Headers      []string               `json:"headers" binding:"required"`
	FieldMapping map[string]string     `json:"field_mapping,omitempty"` // 字段映射：文件列名 -> 表字段名
	SkipFirstRow bool                   `json:"skip_first_row,omitempty"` // 是否跳过第一行（表头）
	OnDuplicate  string                 `json:"on_duplicate,omitempty"`   // 重复数据处理方式: ignore, update, error
}

// ImportDataResponse 导入数据响应
type ImportDataResponse struct {
	TotalRows    int `json:"total_rows"`
	InsertedRows int `json:"inserted_rows"`
	UpdatedRows  int `json:"updated_rows"`
	ErrorRows    int `json:"error_rows"`
	Errors       []struct {
		Row   int    `json:"row"`
		Error string `json:"error"`
	} `json:"errors,omitempty"`
}

// ExecuteSQLFileRequest 执行 SQL 文件请求
type ExecuteSQLFileRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database"`
	SQL          string `json:"sql" binding:"required"`
	Transaction  bool   `json:"transaction,omitempty"` // 是否在事务中执行
	ConfirmDangerous bool `json:"confirm_dangerous,omitempty"`
}

// ExecuteSQLFileResponse 执行 SQL 文件响应
type ExecuteSQLFileResponse struct {
	ExecutedStatements int `json:"executed_statements"`
	SuccessCount       int `json:"success_count"`
	ErrorCount         int `json:"error_count"`
	Errors             []struct {
		Statement int    `json:"statement"`
		Error     string `json:"error"`
	} `json:"errors,omitempty"`
}
