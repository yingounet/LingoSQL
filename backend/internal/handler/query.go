package handler

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type QueryHandler struct {
	queryService *service.QueryService
}

func NewQueryHandler(queryService *service.QueryService) *QueryHandler {
	return &QueryHandler{queryService: queryService}
}

// Execute 执行 SQL 查询
func (h *QueryHandler) Execute(c *gin.Context) {
	var req models.QueryExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	response, queryID, err := h.queryService.Execute(userID, &req)
	if err != nil {
		utils.ErrorWithData(c, 400, "SQL 执行失败", gin.H{
			"error":    err.Error(),
			"query_id": queryID,
		})
		return
	}

	utils.Success(c, response)
}

// Explain 执行 SQL 执行计划分析
func (h *QueryHandler) Explain(c *gin.Context) {
	var req models.ExplainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	response, err := h.queryService.Explain(userID, &req)
	if err != nil {
		utils.Error(c, 400, "执行计划分析失败: "+err.Error())
		return
	}

	utils.Success(c, response)
}

// BeginTransaction 开始事务
func (h *QueryHandler) BeginTransaction(c *gin.Context) {
	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	if err := h.queryService.BeginTransaction(userID, &req); err != nil {
		utils.Error(c, 400, "开始事务失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "事务已开始"})
}

// CommitTransaction 提交事务
func (h *QueryHandler) CommitTransaction(c *gin.Context) {
	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	if err := h.queryService.CommitTransaction(userID, &req); err != nil {
		utils.Error(c, 400, "提交事务失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "事务已提交"})
}

// RollbackTransaction 回滚事务
func (h *QueryHandler) RollbackTransaction(c *gin.Context) {
	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	if err := h.queryService.RollbackTransaction(userID, &req); err != nil {
		utils.Error(c, 400, "回滚事务失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "事务已回滚"})
}
