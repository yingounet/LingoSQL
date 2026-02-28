package models

// QueryExecuteRequest SQL 查询执行请求
type QueryExecuteRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database"`
	SQL          string `json:"sql" binding:"required"`
	ConfirmDangerous bool `json:"confirm_dangerous,omitempty"`
}

// QueryExecuteResponse SQL 查询执行响应
type QueryExecuteResponse struct {
	Columns            []string        `json:"columns"`
	Rows               [][]interface{} `json:"rows"`
	ExecutionTimeMs    int             `json:"execution_time_ms"`
	RowsAffected       int             `json:"rows_affected"`
	StatementsExecuted int             `json:"statements_executed"`
	QueryID            int             `json:"query_id,omitempty"`
}

// QueryErrorResponse SQL 查询错误响应
type QueryErrorResponse struct {
	Error   string `json:"error"`
	QueryID int    `json:"query_id,omitempty"`
}

// ExplainRequest SQL 执行计划请求
type ExplainRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database"`
	SQL          string `json:"sql" binding:"required"`
}

// ExplainResponse SQL 执行计划响应
type ExplainResponse struct {
	Plan            []map[string]interface{} `json:"plan"`
	ExecutionTimeMs int                      `json:"execution_time_ms"`
}

// TransactionRequest 事务操作请求
type TransactionRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database"`
}
