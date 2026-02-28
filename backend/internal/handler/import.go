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

// ImportDataPreview 导入预检（不写入）
func (h *ImportHandler) ImportDataPreview(c *gin.Context) {
	var req models.ImportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	totalRows := len(req.Data)
	sample := req.Data
	if len(sample) > 5 {
		sample = sample[:5]
	}

	utils.Success(c, gin.H{
		"total_rows": totalRows,
		"headers":    req.Headers,
		"sample":     sample,
	})
}

// ImportDataAsyncDoc 文档版异步导入（返回 job_id）
func (h *ImportHandler) ImportDataAsyncDoc(c *gin.Context) {
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
	utils.Success(c, gin.H{"job_id": task.ID})
}

// GetImportTask 获取导入任务详情
func (h *ImportHandler) GetImportTask(c *gin.Context) {
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
	if task.Type != "IMPORT_DATA" {
		utils.NotFound(c, "任务不存在")
		return
	}

	rowsImported := 0
	errorCount := 0
	if task.Status == "success" {
		var result struct {
			InsertedRows int `json:"inserted_rows"`
			ErrorRows    int `json:"error_rows"`
		}
		if err := json.Unmarshal([]byte(task.Result), &result); err == nil {
			rowsImported = result.InsertedRows
			errorCount = result.ErrorRows
		}
	}

	utils.Success(c, gin.H{
		"id":            task.ID,
		"status":        task.Status,
		"progress":      task.Progress,
		"rows_imported": rowsImported,
		"error_count":   errorCount,
		"created_at":    task.CreatedAt,
		"finished_at":   task.FinishedAt,
	})
}

// ListImportTasks 获取导入任务列表
func (h *ImportHandler) ListImportTasks(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	response, err := h.taskService.ListByUserAndType(userID, "IMPORT_DATA", page, pageSize)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	list := make([]gin.H, 0, len(response.List))
	for _, task := range response.List {
		rowsImported := 0
		errorCount := 0
		if task.Status == "success" {
			var result struct {
				InsertedRows int `json:"inserted_rows"`
				ErrorRows    int `json:"error_rows"`
			}
			if err := json.Unmarshal([]byte(task.Result), &result); err == nil {
				rowsImported = result.InsertedRows
				errorCount = result.ErrorRows
			}
		}
		list = append(list, gin.H{
			"id":            task.ID,
			"status":        task.Status,
			"progress":      task.Progress,
			"rows_imported": rowsImported,
			"error_count":   errorCount,
			"created_at":    task.CreatedAt,
		})
	}

	utils.Success(c, gin.H{
		"list":      list,
		"total":     response.Total,
		"page":      response.Page,
		"page_size": response.PageSize,
	})
}

// CancelImportTask 取消导入任务
func (h *ImportHandler) CancelImportTask(c *gin.Context) {
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
	if task.Type != "IMPORT_DATA" {
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

// RetryImportTask 重试导入任务
func (h *ImportHandler) RetryImportTask(c *gin.Context) {
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
	if task.Type != "IMPORT_DATA" {
		utils.NotFound(c, "任务不存在")
		return
	}
	if task.Status != "failed" && task.Status != "canceled" {
		utils.BadRequest(c, "仅支持重试失败或已取消的任务")
		return
	}

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
	utils.Success(c, gin.H{"job_id": newTask.ID})
}

// GetImportErrors 获取导入错误信息
func (h *ImportHandler) GetImportErrors(c *gin.Context) {
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
	if task.Type != "IMPORT_DATA" {
		utils.NotFound(c, "任务不存在")
		return
	}
	errors := []string{}
	if task.ErrorMessage != "" {
		errors = append(errors, task.ErrorMessage)
	}
	utils.Success(c, gin.H{"errors": errors})
}
