package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"lingosql/internal/service"
)

// NewServer 创建 MCP HTTP 服务（StreamableHTTP）
func NewServer(
	connSvc *service.ConnectionService,
	dbSvc *service.DatabaseService,
	tableSvc *service.TableService,
	querySvc *service.QueryService,
) *server.StreamableHTTPServer {
	s := server.NewMCPServer(
		"LingoSQL",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	handlers := NewHandlers(connSvc, dbSvc, tableSvc, querySvc)

	// 注册 6 个 tools
	s.AddTool(
		mcp.NewTool("lingosql_list_connections",
			mcp.WithDescription("List all database connections configured in LingoSQL for the current user. Use this first to get connection_id before other operations."),
			mcp.WithNumber("page", mcp.Description("Page number, default 1"), mcp.DefaultNumber(1)),
			mcp.WithNumber("page_size", mcp.Description("Items per page, default 20"), mcp.DefaultNumber(20)),
		),
		handlers["lingosql_list_connections"],
	)
	s.AddTool(
		mcp.NewTool("lingosql_list_databases",
			mcp.WithDescription("List all databases in a connection. Requires connection_id from lingosql_list_connections."),
			mcp.WithNumber("connection_id", mcp.Required(), mcp.Description("Connection ID")),
		),
		handlers["lingosql_list_databases"],
	)
	s.AddTool(
		mcp.NewTool("lingosql_list_tables",
			mcp.WithDescription("List all tables in a database."),
			mcp.WithNumber("connection_id", mcp.Required(), mcp.Description("Connection ID")),
			mcp.WithString("database", mcp.Required(), mcp.Description("Database name")),
		),
		handlers["lingosql_list_tables"],
	)
	s.AddTool(
		mcp.NewTool("lingosql_get_table_structure",
			mcp.WithDescription("Get column definitions, indexes, and foreign keys of a table. Useful for writing correct SQL."),
			mcp.WithNumber("connection_id", mcp.Required(), mcp.Description("Connection ID")),
			mcp.WithString("database", mcp.Required(), mcp.Description("Database name")),
			mcp.WithString("table_name", mcp.Required(), mcp.Description("Table name")),
		),
		handlers["lingosql_get_table_structure"],
	)
	s.AddTool(
		mcp.NewTool("lingosql_execute_query",
			mcp.WithDescription("Execute a SQL query and return results. Supports SELECT and write operations based on connection settings."),
			mcp.WithNumber("connection_id", mcp.Required(), mcp.Description("Connection ID")),
			mcp.WithString("database", mcp.Required(), mcp.Description("Database name")),
			mcp.WithString("sql", mcp.Required(), mcp.Description("SQL statement to execute")),
		),
		handlers["lingosql_execute_query"],
	)
	s.AddTool(
		mcp.NewTool("lingosql_get_table_data",
			mcp.WithDescription("Get paginated rows from a table."),
			mcp.WithNumber("connection_id", mcp.Required(), mcp.Description("Connection ID")),
			mcp.WithString("database", mcp.Required(), mcp.Description("Database name")),
			mcp.WithString("table_name", mcp.Required(), mcp.Description("Table name")),
			mcp.WithNumber("page", mcp.Description("Page number"), mcp.DefaultNumber(1)),
			mcp.WithNumber("page_size", mcp.Description("Rows per page"), mcp.DefaultNumber(100)),
		),
		handlers["lingosql_get_table_data"],
	)

	return server.NewStreamableHTTPServer(s)
}

// GinHandler 返回 Gin 处理器，将 user_id 从 Gin 上下文注入到 MCP 请求上下文中
func GinHandler(mcpHandler http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists || userIDVal == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
			c.Abort()
			return
		}
		userID, ok := userIDVal.(int)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
			c.Abort()
			return
		}
		ctx := WithUserID(c.Request.Context(), userID)
		req := c.Request.WithContext(ctx)
		mcpHandler.ServeHTTP(c.Writer, req)
	}
}
