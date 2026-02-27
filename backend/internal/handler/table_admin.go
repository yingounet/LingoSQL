package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type TableAdminHandler struct {
	tableAdminService *service.TableAdminService
}

func NewTableAdminHandler(tableAdminService *service.TableAdminService) *TableAdminHandler {
	return &TableAdminHandler{tableAdminService: tableAdminService}
}

// CreateTable 建表
func (h *TableAdminHandler) CreateTable(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.tableAdminService.CreateTable(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DropTable 删表
func (h *TableAdminHandler) DropTable(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	if database == "" {
		utils.BadRequest(c, "数据库名不能为空")
		return
	}
	tableName := c.Param("name")
	if tableName == "" {
		utils.BadRequest(c, "表名不能为空")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.tableAdminService.DropTable(connectionID, userID, database, tableName); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

// AlterTable 修改表结构
func (h *TableAdminHandler) AlterTable(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.AlterTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.tableAdminService.AlterTable(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

// CreateIndex 创建索引
func (h *TableAdminHandler) CreateIndex(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.CreateIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.tableAdminService.CreateIndex(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

// DropIndex 删除索引
func (h *TableAdminHandler) DropIndex(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	table := c.Query("table")
	indexName := c.Param("index_name")
	if database == "" || table == "" || indexName == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	req := models.DropIndexRequest{
		Database:  database,
		Table:     table,
		IndexName: indexName,
	}
	if err := h.tableAdminService.DropIndex(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}

// RenameTable 重命名表
func (h *TableAdminHandler) RenameTable(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.RenameTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.tableAdminService.RenameTable(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.Success(c, nil)
}
