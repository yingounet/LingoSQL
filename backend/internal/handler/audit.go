package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// GetAuditLogs 获取审计日志
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	pageSize := 20
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}
	action := c.Query("action")

	response, err := h.auditService.List(userID, action, page, pageSize)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.Success(c, response)
}
