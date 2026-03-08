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
	taskService       *service.TaskService
}

func NewMaintenanceHandler(maintenanceService *service.MaintenanceService, taskService *service.TaskService) *MaintenanceHandler {
	return &MaintenanceHandler{
		maintenanceService: maintenanceService,
		taskService:       taskService,
	}
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
	response, err := h.maintenanceService.BackupDatabase(req.ConnectionID, userID, &req, h.taskService)
	if err != nil {
		utils.Error(c, 400, "备份数据库失败: "+err.Error())
		return
	}
	utils.Success(c, response)
}

// ListBackups 备份列表
func (h *MaintenanceHandler) ListBackups(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var connID *int
	if v := c.Query("connection_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			connID = &id
		}
	}
	list, total, err := h.maintenanceService.ListBackups(userID, connID, page, pageSize)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	items := make([]map[string]interface{}, 0, len(list))
	for _, m := range list {
		items = append(items, map[string]interface{}{
			"id":            m.BackupID,
			"name":          m.BackupID,
			"database":      m.Database,
			"size":          m.TotalSize,
			"file_count":    len(m.Files),
			"created_at":    m.CreatedAt.Format("2006-01-02 15:04:05"),
			"connection_id": m.ConnectionID,
		})
	}
	utils.Success(c, gin.H{"list": items, "total": total})
}

// DownloadBackup 下载备份
func (h *MaintenanceHandler) DownloadBackup(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	backupID := c.Param("id")
	singleFile := c.Query("file")
	path, name, err := h.maintenanceService.ServeBackupDownload(backupID, userID, singleFile)
	if err != nil {
		utils.Error(c, 404, err.Error())
		return
	}
	c.FileAttachment(path, name)
}

// DeleteBackup 删除备份
func (h *MaintenanceHandler) DeleteBackup(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	backupID := c.Param("id")
	if err := h.maintenanceService.DeleteBackup(backupID, userID); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, nil)
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
