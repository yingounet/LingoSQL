package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type DbObjectsHandler struct {
	dbObjectsService *service.DbObjectsService
}

func NewDbObjectsHandler(dbObjectsService *service.DbObjectsService) *DbObjectsHandler {
	return &DbObjectsHandler{dbObjectsService: dbObjectsService}
}

// GetViews 获取视图列表
func (h *DbObjectsHandler) GetViews(c *gin.Context) {
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
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	views, err := h.dbObjectsService.GetViews(connectionID, userID, database)
	if err != nil {
		utils.Error(c, 400, "获取视图列表失败: "+err.Error())
		return
	}
	utils.Success(c, views)
}

// CreateView 创建视图
func (h *DbObjectsHandler) CreateView(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.CreateViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.CreateView(connectionID, userID, &req); err != nil {
		utils.Error(c, 400, "创建视图失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// DropView 删除视图
func (h *DbObjectsHandler) DropView(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	name := c.Param("name")
	if database == "" || name == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.DropView(connectionID, userID, database, name); err != nil {
		utils.Error(c, 400, "删除视图失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetProcedures 获取存储过程列表
func (h *DbObjectsHandler) GetProcedures(c *gin.Context) {
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
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	procedures, err := h.dbObjectsService.GetProcedures(connectionID, userID, database)
	if err != nil {
		utils.Error(c, 400, "获取存储过程列表失败: "+err.Error())
		return
	}
	utils.Success(c, procedures)
}

// CreateProcedure 创建存储过程
func (h *DbObjectsHandler) CreateProcedure(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.CreateProcedureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.CreateProcedure(connectionID, userID, &req); err != nil {
		utils.Error(c, 400, "创建存储过程失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// DropProcedure 删除存储过程
func (h *DbObjectsHandler) DropProcedure(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	name := c.Param("name")
	if database == "" || name == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.DropProcedure(connectionID, userID, database, name); err != nil {
		utils.Error(c, 400, "删除存储过程失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// ExecuteProcedure 执行存储过程
func (h *DbObjectsHandler) ExecuteProcedure(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	name := c.Param("name")
	if database == "" || name == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	var req models.ExecuteProcedureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	result, err := h.dbObjectsService.ExecuteProcedure(connectionID, userID, database, name, req.Parameters)
	if err != nil {
		utils.Error(c, 400, "执行存储过程失败: "+err.Error())
		return
	}
	utils.Success(c, result)
}

// GetFunctions 获取函数列表
func (h *DbObjectsHandler) GetFunctions(c *gin.Context) {
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
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	functions, err := h.dbObjectsService.GetFunctions(connectionID, userID, database)
	if err != nil {
		utils.Error(c, 400, "获取函数列表失败: "+err.Error())
		return
	}
	utils.Success(c, functions)
}

// CreateFunction 创建函数
func (h *DbObjectsHandler) CreateFunction(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.CreateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.CreateFunction(connectionID, userID, &req); err != nil {
		utils.Error(c, 400, "创建函数失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// DropFunction 删除函数
func (h *DbObjectsHandler) DropFunction(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	name := c.Param("name")
	if database == "" || name == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.DropFunction(connectionID, userID, database, name); err != nil {
		utils.Error(c, 400, "删除函数失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetTriggers 获取触发器列表
func (h *DbObjectsHandler) GetTriggers(c *gin.Context) {
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
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	triggers, err := h.dbObjectsService.GetTriggers(connectionID, userID, database)
	if err != nil {
		utils.Error(c, 400, "获取触发器列表失败: "+err.Error())
		return
	}
	utils.Success(c, triggers)
}

// CreateTrigger 创建触发器
func (h *DbObjectsHandler) CreateTrigger(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.CreateTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.CreateTrigger(connectionID, userID, &req); err != nil {
		utils.Error(c, 400, "创建触发器失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// DropTrigger 删除触发器
func (h *DbObjectsHandler) DropTrigger(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	name := c.Param("name")
	if database == "" || name == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.DropTrigger(connectionID, userID, database, name); err != nil {
		utils.Error(c, 400, "删除触发器失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetEvents 获取事件列表
func (h *DbObjectsHandler) GetEvents(c *gin.Context) {
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
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	events, err := h.dbObjectsService.GetEvents(connectionID, userID, database)
	if err != nil {
		utils.Error(c, 400, "获取事件列表失败: "+err.Error())
		return
	}
	utils.Success(c, events)
}

// CreateEvent 创建事件
func (h *DbObjectsHandler) CreateEvent(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.CreateEvent(connectionID, userID, &req); err != nil {
		utils.Error(c, 400, "创建事件失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}

// DropEvent 删除事件
func (h *DbObjectsHandler) DropEvent(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}
	database := c.Query("database")
	name := c.Param("name")
	if database == "" || name == "" {
		utils.BadRequest(c, "参数不完整")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.dbObjectsService.DropEvent(connectionID, userID, database, name); err != nil {
		utils.Error(c, 400, "删除事件失败: "+err.Error())
		return
	}
	utils.Success(c, nil)
}
