package handler

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

type RowDataHandler struct {
	rowDataService *service.RowDataService
	tableService   *service.TableService
}

func NewRowDataHandler(rowDataService *service.RowDataService, tableService *service.TableService) *RowDataHandler {
	return &RowDataHandler{
		rowDataService: rowDataService,
		tableService:   tableService,
	}
}

// GetTableData 获取表数据（POST方式）
func (h *RowDataHandler) GetTableData(c *gin.Context) {
	var req models.GetTableDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	// 转换filters格式
	filters := make([]db.RowFilter, 0, len(req.Filters))
	for _, f := range req.Filters {
		filter := db.RowFilter{}
		if field, ok := f["field"].(string); ok {
			filter.Field = field
		}
		if operator, ok := f["operator"].(string); ok {
			filter.Operator = operator
		}
		if value, ok := f["value"].(string); ok {
			filter.Value = value
		}
		if filter.Field != "" {
			filters = append(filters, filter)
		}
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	result, err := h.tableService.GetTableRows(req.ConnectionID, userID, req.Database, req.Table, filters, page, pageSize)
	if err != nil {
		utils.Error(c, 400, "获取表数据失败: "+err.Error())
		return
	}
	utils.Success(c, result)
}

// UpdateTableRowData 更新表行数据（POST方式）
func (h *RowDataHandler) UpdateTableRowData(c *gin.Context) {
	var req models.UpdateTableRowDataRequest
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
		req.PrimaryKey, req.UpdateData,
	)
	if err != nil {
		utils.Error(c, 400, "更新表行失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{"affected_rows": affectedRows})
}

// BatchInsertData 批量插入数据
func (h *RowDataHandler) BatchInsertData(c *gin.Context) {
	var req models.BatchInsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	insertedRows, err := h.rowDataService.BatchInsertData(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "批量插入失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{"inserted_rows": insertedRows})
}

// BatchUpdateData 批量更新数据
func (h *RowDataHandler) BatchUpdateData(c *gin.Context) {
	var req models.BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	affectedRows, err := h.rowDataService.BatchUpdateData(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "批量更新失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{"affected_rows": affectedRows})
}

// BatchDeleteData 批量删除数据
func (h *RowDataHandler) BatchDeleteData(c *gin.Context) {
	var req models.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	affectedRows, err := h.rowDataService.BatchDeleteData(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "批量删除失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{"affected_rows": affectedRows})
}

// CompareData 数据对比
func (h *RowDataHandler) CompareData(c *gin.Context) {
	var req models.CompareDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	response, err := h.rowDataService.CompareData(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "数据对比失败: "+err.Error())
		return
	}
	utils.Success(c, response)
}

// FindReplaceData 查找替换数据
func (h *RowDataHandler) FindReplaceData(c *gin.Context) {
	var req models.FindReplaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	response, err := h.rowDataService.FindReplaceData(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "查找替换失败: "+err.Error())
		return
	}
	utils.Success(c, response)
}
