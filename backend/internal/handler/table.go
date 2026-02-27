package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/service"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

type TableHandler struct {
	tableService *service.TableService
}

func NewTableHandler(tableService *service.TableService) *TableHandler {
	return &TableHandler{tableService: tableService}
}

// GetTables 获取表列表
func (h *TableHandler) GetTables(c *gin.Context) {
	connectionID, ok := ParseConnectionID(c)
	if !ok {
		return
	}
	database := c.Query("database")
	if database == "" {
		utils.BadRequest(c, "数据库名不能为空")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	tables, err := h.tableService.GetTables(connectionID, userID, database)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if tables == nil {
		tables = []map[string]interface{}{}
	}
	utils.Success(c, tables)
}

// GetTableInfo 获取表详细信息
func (h *TableHandler) GetTableInfo(c *gin.Context) {
	connectionID, ok := ParseConnectionID(c)
	if !ok {
		return
	}
	database, table, ok := ParseDatabaseTable(c)
	if !ok {
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	tableInfo, err := h.tableService.GetTableInfo(connectionID, userID, database, table)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, tableInfo)
}

// GetTableColumns 获取表字段列表
func (h *TableHandler) GetTableColumns(c *gin.Context) {
	connectionID, ok := ParseConnectionID(c)
	if !ok {
		return
	}
	database, table, ok := ParseDatabaseTable(c)
	if !ok {
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	columns, err := h.tableService.GetTableColumns(connectionID, userID, database, table)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, columns)
}

// GetTableIndexes 获取表索引列表
func (h *TableHandler) GetTableIndexes(c *gin.Context) {
	connectionID, ok := ParseConnectionID(c)
	if !ok {
		return
	}
	database, table, ok := ParseDatabaseTable(c)
	if !ok {
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	indexes, err := h.tableService.GetTableIndexes(connectionID, userID, database, table)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, indexes)
}

// GetTableRows 获取表数据
func (h *TableHandler) GetTableRows(c *gin.Context) {
	connectionID, ok := ParseConnectionID(c)
	if !ok {
		return
	}
	database, table, ok := ParseDatabaseTable(c)
	if !ok {
		return
	}

	// 解析分页参数
	page := 1
	pageSize := 100
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}

	// 解析筛选条件
	var filters []db.RowFilter
	filtersStr := c.Query("filters")
	if filtersStr != "" {
		if err := json.Unmarshal([]byte(filtersStr), &filters); err != nil {
			utils.BadRequest(c, "筛选条件格式错误")
			return
		}
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	result, err := h.tableService.GetTableRows(connectionID, userID, database, table, filters, page, pageSize)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, result)
}

// UpdateTableRowRequest 更新表数据请求
type UpdateTableRowRequest struct {
	ConnectionID int                    `json:"connection_id" binding:"required"`
	Database     string                 `json:"database" binding:"required"`
	Table        string                 `json:"table" binding:"required"`
	PrimaryKey   map[string]interface{} `json:"primary_key" binding:"required"`
	Data         map[string]interface{} `json:"data" binding:"required"`
}

// UpdateTableRow 更新表数据
func (h *TableHandler) UpdateTableRow(c *gin.Context) {
	var req UpdateTableRowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	affectedRows, err := h.tableService.UpdateTableRow(
		req.ConnectionID, userID,
		req.Database, req.Table,
		req.PrimaryKey, req.Data,
	)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, map[string]interface{}{
		"affected_rows": affectedRows,
	})
}
