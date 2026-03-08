package handler

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type TaskHandler struct {
	taskService        *service.TaskService
	importService      *service.ImportService
	exportService      *service.ExportService
	maintenanceService *service.MaintenanceService
}

func NewTaskHandler(taskService *service.TaskService, importService *service.ImportService, exportService *service.ExportService, maintenanceService *service.MaintenanceService) *TaskHandler {
	return &TaskHandler{
		taskService:        taskService,
		importService:      importService,
		exportService:      exportService,
		maintenanceService: maintenanceService,
	}
}

// GetTasks 获取任务列表
func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	taskType := c.Query("type")
	var response *models.TaskListResponse
	var err error
	if taskType != "" {
		response, err = h.taskService.ListByUserAndType(userID, taskType, page, pageSize)
	} else {
		response, err = h.taskService.ListByUser(userID, page, pageSize)
	}
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.Success(c, response)
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
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
	utils.Success(c, task)
}

// RetryTask 重新执行任务
func (h *TaskHandler) RetryTask(c *gin.Context) {
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
	if task.Status != "failed" {
		utils.BadRequest(c, "仅支持重试失败的任务")
		return
	}

	switch task.Type {
	case "IMPORT_DATA":
		var payload models.ImportDataRequest
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			utils.BadRequest(c, "任务参数解析失败")
			return
		}
		newTask, err := h.taskService.Create(userID, "IMPORT_DATA", payload)
		if err != nil {
			utils.InternalServerError(c, err.Error())
			return
		}
		go h.importService.RunImportDataTask(newTask.ID, userID, &payload, h.taskService)
		utils.Success(c, newTask)
	case "EXECUTE_SQL_FILE":
		var payload models.ExecuteSQLFileRequest
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			utils.BadRequest(c, "任务参数解析失败")
			return
		}
		newTask, err := h.taskService.Create(userID, "EXECUTE_SQL_FILE", payload)
		if err != nil {
			utils.InternalServerError(c, err.Error())
			return
		}
		go h.importService.RunExecuteSQLFileTask(newTask.ID, userID, &payload, h.taskService)
		utils.Success(c, newTask)
	case "EXPORT_DATA":
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
		utils.Success(c, newTask)
	case "BACKUP_DATABASE":
		var payload models.BackupRequest
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			utils.BadRequest(c, "任务参数解析失败")
			return
		}
		newTask, err := h.taskService.Create(userID, "BACKUP_DATABASE", payload)
		if err != nil {
			utils.InternalServerError(c, err.Error())
			return
		}
		go func() {
			if err := h.taskService.Start(newTask.ID); err != nil {
				return
			}
			result, err := h.maintenanceService.RunBackup(payload.ConnectionID, userID, &payload, newTask.ID, h.taskService)
			if err != nil {
				_ = h.taskService.CompleteFailure(newTask.ID, err)
				return
			}
			_ = h.taskService.CompleteSuccess(newTask.ID, result)
		}()
		utils.Success(c, newTask)
	default:
		utils.BadRequest(c, "不支持重试的任务类型")
	}
}

// DownloadTaskResult 下载导出文件
func (h *TaskHandler) DownloadTaskResult(c *gin.Context) {
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
	if task.Status != "success" {
		utils.BadRequest(c, "任务尚未完成")
		return
	}

	if task.Type == "BACKUP_DATABASE" {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(task.Result), &result); err != nil {
			utils.InternalServerError(c, "备份结果解析失败")
			return
		}
		backupID, _ := result["backup_id"].(string)
		if backupID == "" {
			utils.InternalServerError(c, "备份文件不存在")
			return
		}
		path, name, err := h.maintenanceService.ServeBackupDownload(backupID, userID, "")
		if err != nil {
			utils.InternalServerError(c, err.Error())
			return
		}
		c.FileAttachment(path, name)
		return
	}

	if task.Type != "EXPORT_DATA" {
		utils.BadRequest(c, "该任务类型不支持下载")
		return
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(task.Result), &result); err != nil {
		utils.InternalServerError(c, "导出结果解析失败")
		return
	}
	filePath, _ := result["file_path"].(string)
	fileName, _ := result["file_name"].(string)
	if filePath == "" {
		utils.InternalServerError(c, "导出文件不存在")
		return
	}
	if _, err := os.Stat(filePath); err != nil {
		utils.InternalServerError(c, "导出文件不存在")
		return
	}
	if fileName == "" {
		c.File(filePath)
		return
	}
	c.FileAttachment(filePath, fileName)
}

// CancelTask 取消任务（仅 pending）
func (h *TaskHandler) CancelTask(c *gin.Context) {
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
