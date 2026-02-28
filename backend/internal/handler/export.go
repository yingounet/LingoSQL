package handler

import (
	"encoding/json"
	"strconv"
	"time"

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

// ExportDataSync 同步导出（文档 /api/data/export）
func (h *ExportHandler) ExportDataSync(c *gin.Context) {
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
	if err := h.taskService.Start(task.ID); err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	result, err := h.exportService.ExportData(userID, &req, task.ID)
	if err != nil {
		_ = h.taskService.CompleteFailure(task.ID, err)
		utils.Error(c, 400, "导出失败: "+err.Error())
		return
	}
	_ = h.taskService.CompleteSuccess(task.ID, result)

	utils.Success(c, gin.H{
		"download_url": "/api/data/export/" + strconv.Itoa(task.ID),
		"expires_in":   3600,
	})
}

// ExportDataAsyncDoc 文档版异步导出（返回 job_id）
func (h *ExportHandler) ExportDataAsyncDoc(c *gin.Context) {
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

	utils.Success(c, gin.H{"job_id": task.ID})
}

// GetExportTask 获取导出任务状态
func (h *ExportHandler) GetExportTask(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务 ID")
		return
	}
	task, err := h.taskService.GetByID(id, userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	if task.Type != "EXPORT_DATA" {
		utils.NotFound(c, "任务不存在")
		return
	}

	downloadURL := ""
	var fileName string
	var rows int
	if task.Status == "success" {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(task.Result), &result); err == nil {
			if v, ok := result["file_name"].(string); ok {
				fileName = v
			}
			if v, ok := result["rows"].(float64); ok {
				rows = int(v)
			}
			if _, ok := result["file_path"].(string); ok {
				downloadURL = "/api/data/export/" + strconv.Itoa(task.ID)
			}
		}
	}

	utils.Success(c, gin.H{
		"id":           task.ID,
		"status":       task.Status,
		"progress":     task.Progress,
		"file_name":    fileName,
		"rows":         rows,
		"download_url": downloadURL,
		"created_at":   task.CreatedAt,
		"finished_at":  task.FinishedAt,
	})
}

// ListExportTasks 获取导出任务列表
func (h *ExportHandler) ListExportTasks(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	response, err := h.taskService.ListByUserAndType(userID, "EXPORT_DATA", page, pageSize)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	list := make([]gin.H, 0, len(response.List))
	for _, task := range response.List {
		var fileName string
		var rows int
		if task.Status == "success" {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(task.Result), &result); err == nil {
				if v, ok := result["file_name"].(string); ok {
					fileName = v
				}
				if v, ok := result["rows"].(float64); ok {
					rows = int(v)
				}
			}
		}
		list = append(list, gin.H{
			"id":         task.ID,
			"status":     task.Status,
			"progress":   task.Progress,
			"file_name":  fileName,
			"rows":       rows,
			"created_at": task.CreatedAt,
		})
	}

	utils.Success(c, gin.H{
		"list":      list,
		"total":     response.Total,
		"page":      response.Page,
		"page_size": response.PageSize,
	})
}

// CancelExportTask 取消导出任务
func (h *ExportHandler) CancelExportTask(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务 ID")
		return
	}
	task, err := h.taskService.GetByID(id, userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	if task.Type != "EXPORT_DATA" {
		utils.NotFound(c, "任务不存在")
		return
	}
	if task.Status != "pending" {
		utils.BadRequest(c, "仅支持取消等待中的任务")
		return
	}
	now := time.Now()
	if err := h.taskService.Cancel(task.ID, &now); err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.SuccessWithMessage(c, "任务已取消", nil)
}

// RetryExportTask 重试导出任务
func (h *ExportHandler) RetryExportTask(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务 ID")
		return
	}
	task, err := h.taskService.GetByID(id, userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	if task.Type != "EXPORT_DATA" {
		utils.NotFound(c, "任务不存在")
		return
	}
	if task.Status != "failed" && task.Status != "canceled" {
		utils.BadRequest(c, "仅支持重试失败或已取消的任务")
		return
	}

	var payload models.ExportDataRequest
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		utils.BadRequest(c, "任务参数解析失败")
		return
	}
	newTask, err := h.taskService.Create(userID, "EXPORT_DATA", payload)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	go func() {
		if err := h.taskService.Start(newTask.ID); err != nil {
			return
		}
		result, err := h.exportService.ExportData(userID, &payload, newTask.ID)
		if err != nil {
			_ = h.taskService.CompleteFailure(newTask.ID, err)
			return
		}
		_ = h.taskService.CompleteSuccess(newTask.ID, result)
	}()
	utils.Success(c, gin.H{"job_id": newTask.ID})
}

// GetExportErrors 获取导出错误信息
func (h *ExportHandler) GetExportErrors(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务 ID")
		return
	}
	task, err := h.taskService.GetByID(id, userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	if task.Type != "EXPORT_DATA" {
		utils.NotFound(c, "任务不存在")
		return
	}
	errors := []string{}
	if task.ErrorMessage != "" {
		errors = append(errors, task.ErrorMessage)
	}
	utils.Success(c, gin.H{"errors": errors})
}
