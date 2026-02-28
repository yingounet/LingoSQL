package handler

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type ImportHandler struct {
	importService *service.ImportService
	taskService   *service.TaskService
	auditService  *service.AuditService
}

func NewImportHandler(importService *service.ImportService, taskService *service.TaskService, auditService *service.AuditService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
		taskService:   taskService,
		auditService:  auditService,
	}
}

// ImportData 导入数据
func (h *ImportHandler) ImportData(c *gin.Context) {
	var req models.ImportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	response, err := h.importService.ImportData(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "导入数据失败: "+err.Error())
		return
	}

	utils.Success(c, response)
}

// ExecuteSQLFile 执行SQL文件
func (h *ImportHandler) ExecuteSQLFile(c *gin.Context) {
	var req models.ExecuteSQLFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	response, err := h.importService.ExecuteSQLFile(req.ConnectionID, userID, &req)
	if err != nil {
		utils.Error(c, 400, "执行SQL文件失败: "+err.Error())
		return
	}

	utils.Success(c, response)
}

// ImportDataAsync 异步导入数据
func (h *ImportHandler) ImportDataAsync(c *gin.Context) {
	var req models.ImportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	task, err := h.taskService.Create(userID, "IMPORT_DATA", req)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	h.auditService.Record(userID, "import.create_task", "task", &task.ID, true, "", gin.H{
		"type": "IMPORT_DATA",
	})

	go h.importService.RunImportDataTask(task.ID, userID, &req, h.taskService)
	utils.Success(c, task)
}

// ExecuteSQLFileAsync 异步执行 SQL 文件
func (h *ImportHandler) ExecuteSQLFileAsync(c *gin.Context) {
	var req models.ExecuteSQLFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	task, err := h.taskService.Create(userID, "EXECUTE_SQL_FILE", req)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	h.auditService.Record(userID, "import.create_task", "task", &task.ID, true, "", gin.H{
		"type": "EXECUTE_SQL_FILE",
	})

	go h.importService.RunExecuteSQLFileTask(task.ID, userID, &req, h.taskService)
	utils.Success(c, task)
}
