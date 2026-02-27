package types

// RowFilter 行数据筛选条件
type RowFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// ColumnDef 列定义
type ColumnDef struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsPrimary bool   `json:"is_primary"`
	IsIndex   bool   `json:"is_index"`
}

// TableRowsResult 表数据查询结果
type TableRowsResult struct {
	Columns  []ColumnDef              `json:"columns"`
	Rows     []map[string]interface{} `json:"rows"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}
