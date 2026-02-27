package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type MaintenanceHandler struct {
	maintenanceService *service.MaintenanceService
}

func NewMaintenanceHandler(maintenanceService *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{maintenanceService: maintenanceService}
}

// OptimizeTable 优化表
func (h *MaintenanceHandler) OptimizeTable(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.TableMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.maintenanceService.OptimizeTable(connectionID, userID, req.Database, req.Table); err != nil {
		utils.Error(c, 400, "优化表失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// RepairTable 修复表
func (h *MaintenanceHandler) RepairTable(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.TableMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.maintenanceService.RepairTable(connectionID, userID, req.Database, req.Table); err != nil {
		utils.Error(c, 400, "修复表失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// AnalyzeTable 分析表
func (h *MaintenanceHandler) AnalyzeTable(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.TableMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.maintenanceService.AnalyzeTable(connectionID, userID, req.Database, req.Table); err != nil {
		utils.Error(c, 400, "分析表失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// BackupDatabase 备份数据库
func (h *MaintenanceHandler) BackupDatabase(c *gin.Context) {
	var req models.BackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	response, err := h.maintenanceService.BackupDatabase(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "备份数据库失败: "+err.Error())
		return
	}
	utils.Success(c, response)
}

// RestoreDatabase 恢复数据库
func (h *MaintenanceHandler) RestoreDatabase(c *gin.Context) {
	var req models.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.maintenanceService.RestoreDatabase(req.ConnectionID, userID, &req); err != nil {
		utils.Error(c, 400, "恢复数据库失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}
