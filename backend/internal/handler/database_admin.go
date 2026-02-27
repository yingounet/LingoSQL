package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type DatabaseAdminHandler struct {
	databaseAdminService *service.DatabaseAdminService
}

func NewDatabaseAdminHandler(databaseAdminService *service.DatabaseAdminService) *DatabaseAdminHandler {
	return &DatabaseAdminHandler{
		databaseAdminService: databaseAdminService,
	}
}

// CheckAdminPermissions 检查管理权限
func (h *DatabaseAdminHandler) CheckAdminPermissions(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	permissions, err := h.databaseAdminService.CheckAdminPermissions(connectionID, userID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, permissions)
}

// GetDatabaseList 获取数据库列表（带详细信息）
func (h *DatabaseAdminHandler) GetDatabaseList(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	databases, err := h.databaseAdminService.GetDatabaseList(connectionID, userID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, databases)
}

// GetDatabaseInfo 获取数据库详情
func (h *DatabaseAdminHandler) GetDatabaseInfo(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	databaseName := c.Param("name")
	if databaseName == "" {
		utils.BadRequest(c, "数据库名称不能为空")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	info, err := h.databaseAdminService.GetDatabaseInfo(connectionID, userID, databaseName)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, info)
}

// CreateDatabase 创建数据库
func (h *DatabaseAdminHandler) CreateDatabase(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	var req models.CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.databaseAdminService.CreateDatabase(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// DropDatabase 删除数据库
func (h *DatabaseAdminHandler) DropDatabase(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	databaseName := c.Param("name")
	if databaseName == "" {
		utils.BadRequest(c, "数据库名称不能为空")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.databaseAdminService.DropDatabase(connectionID, userID, databaseName); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// RenameDatabase 重命名数据库
func (h *DatabaseAdminHandler) RenameDatabase(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	databaseName := c.Param("name")
	if databaseName == "" {
		utils.BadRequest(c, "数据库名称不能为空")
		return
	}

	var req models.RenameDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.databaseAdminService.RenameDatabase(connectionID, userID, databaseName, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}
