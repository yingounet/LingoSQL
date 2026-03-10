package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/pkg/db"
)

func formatJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func getIntArg(req mcp.CallToolRequest, name string, defaultVal int) int {
	v := req.GetFloat(name, float64(defaultVal))
	iv := int(v)
	if iv > 0 {
		return iv
	}
	return defaultVal
}

func requireIntArg(req mcp.CallToolRequest, name string) (int, error) {
	v, err := req.RequireFloat(name)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func requireUserID(ctx context.Context) (int, error) {
	uid, ok := GetUserID(ctx)
	if !ok {
		return 0, fmt.Errorf("未认证: 缺少 user_id")
	}
	return uid, nil
}

// NewHandlers 创建 MCP tool handlers，依赖注入 service
func NewHandlers(
	connSvc *service.ConnectionService,
	dbSvc *service.DatabaseService,
	tableSvc *service.TableService,
	querySvc *service.QueryService,
) map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return map[string]func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error){
		"lingosql_list_connections":    listConnections(connSvc),
		"lingosql_list_databases":      listDatabases(dbSvc),
		"lingosql_list_tables":         listTables(tableSvc),
		"lingosql_get_table_structure": getTableStructure(tableSvc),
		"lingosql_execute_query":       executeQuery(querySvc),
		"lingosql_get_table_data":      getTableData(tableSvc),
	}
}

func listConnections(s *service.ConnectionService) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := requireUserID(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		page := getIntArg(req, "page", 1)
		pageSize := getIntArg(req, "page_size", 20)
		params := &models.ConnectionListParams{Page: page, PageSize: pageSize}
		resp, err := s.GetList(userID, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(formatJSON(resp)), nil
	}
}

func listDatabases(s *service.DatabaseService) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := requireUserID(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		connID, err := requireIntArg(req, "connection_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		names, err := s.GetDatabases(connID, userID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		list := make([]map[string]interface{}, len(names))
		for i, n := range names {
			list[i] = map[string]interface{}{"name": n}
		}
		return mcp.NewToolResultText(formatJSON(list)), nil
	}
}

func listTables(s *service.TableService) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := requireUserID(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		connID, err := requireIntArg(req, "connection_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dbName, err := req.RequireString("database")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		list, err := s.GetTables(connID, userID, dbName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(formatJSON(list)), nil
	}
}

func getTableStructure(s *service.TableService) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := requireUserID(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		connID, err := requireIntArg(req, "connection_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dbName, err := req.RequireString("database")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		tableName, err := req.RequireString("table_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		columns, err := s.GetTableColumns(connID, userID, dbName, tableName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		indexes, err := s.GetTableIndexes(connID, userID, dbName, tableName)
		if err != nil {
			indexes = []map[string]interface{}{}
		}
		out := map[string]interface{}{"columns": columns, "indexes": indexes}
		return mcp.NewToolResultText(formatJSON(out)), nil
	}
}

func executeQuery(s *service.QueryService) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := requireUserID(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		connID, err := requireIntArg(req, "connection_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dbName, err := req.RequireString("database")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sqlStr, err := req.RequireString("sql")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		execReq := &models.QueryExecuteRequest{
			ConnectionID: connID,
			Database:     dbName,
			SQL:          sqlStr,
		}
		resp, _, err := s.Execute(userID, execReq)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text := formatQueryResult(resp)
		return mcp.NewToolResultText(text), nil
	}
}

func getTableData(s *service.TableService) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := requireUserID(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		connID, err := requireIntArg(req, "connection_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dbName, err := req.RequireString("database")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		tableName, err := req.RequireString("table_name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		page := getIntArg(req, "page", 1)
		pageSize := getIntArg(req, "page_size", 100)
		if pageSize > 1000 {
			pageSize = 1000
		}
		result, err := s.GetTableRows(connID, userID, dbName, tableName, nil, page, pageSize)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text := formatTableRows(result)
		return mcp.NewToolResultText(text), nil
	}
}

func formatQueryResult(r *models.QueryExecuteResponse) string {
	var buf string
	buf += fmt.Sprintf("Columns: %v\n", r.Columns)
	buf += fmt.Sprintf("Rows affected: %d | Execution time: %d ms\n\n", r.RowsAffected, r.ExecutionTimeMs)
	if len(r.Rows) == 0 {
		return buf + "(no rows)"
	}
	cols := r.Columns
	colWidth := 20
	for _, col := range cols {
		buf += fmt.Sprintf("%-*s ", colWidth, truncate(col, colWidth))
	}
	buf += "\n"
	for _, row := range r.Rows {
		for i := 0; i < len(cols); i++ {
			var cell interface{}
			if i < len(row) {
				cell = row[i]
			}
			buf += fmt.Sprintf("%-*s ", colWidth, truncate(strval(cell), colWidth))
		}
		buf += "\n"
	}
	return buf
}

func formatTableRows(r *db.TableRowsResult) string {
	var buf string
	colNames := make([]string, len(r.Columns))
	for i, c := range r.Columns {
		colNames[i] = c.Name
	}
	buf += fmt.Sprintf("Total: %d | Page: %d/%d\n\n", r.Total, r.Page, r.PageSize)
	buf += fmt.Sprintf("Columns: %v\n\n", colNames)
	if len(r.Rows) == 0 {
		return buf + "(no rows)"
	}
	colWidth := 20
	for _, col := range colNames {
		if len(col) > colWidth {
			colWidth = len(col) + 2
		}
	}
	for _, col := range colNames {
		buf += fmt.Sprintf("%-*s ", colWidth, truncate(col, colWidth))
	}
	buf += "\n"
	for _, row := range r.Rows {
		for _, col := range colNames {
			val := row[col]
			buf += fmt.Sprintf("%-*s ", colWidth, truncate(strval(val), colWidth))
		}
		buf += "\n"
	}
	return buf
}

func strval(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
