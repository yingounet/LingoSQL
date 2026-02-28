package handler

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type ExportHandler struct {
	taskService   *service.TaskService
	exportService *service.ExportService
	auditService  *service.AuditService
}

func NewExportHandler(taskService *service.TaskService, exportService *service.ExportService, auditService *service.AuditService) *ExportHandler {
	return &ExportHandler{
		taskService:   taskService,
		exportService: exportService,
		auditService:  auditService,
	}
}

// ExportDataAsync 创建导出任务
func (h *ExportHandler) ExportDataAsync(c *gin.Context) {
	var req models.ExportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	task, err := h.taskService.Create(userID, "EXPORT_DATA", req)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	h.auditService.Record(userID, "export.create_task", "task", &task.ID, true, "", nil)

	go func() {
		if err := h.taskService.Start(task.ID); err != nil {
			return
		}
		result, err := h.exportService.ExportData(userID, &req, task.ID)
		if err != nil {
			_ = h.taskService.CompleteFailure(task.ID, err)
			return
		}
		_ = h.taskService.CompleteSuccess(task.ID, result)
	}()

	utils.Success(c, task)
}
