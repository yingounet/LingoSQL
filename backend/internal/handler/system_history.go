package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type SystemHistoryHandler struct {
	systemHistoryService *service.SystemHistoryService
}

func NewSystemHistoryHandler(systemHistoryService *service.SystemHistoryService) *SystemHistoryHandler {
	return &SystemHistoryHandler{systemHistoryService: systemHistoryService}
}

// GetSystemHistory 获取系统执行记录列表
func (h *SystemHistoryHandler) GetSystemHistory(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	// connection_id 是必填参数
	connectionIDStr := c.Query("connection_id")
	if connectionIDStr == "" {
		utils.BadRequest(c, "connection_id 参数必填")
		return
	}

	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的 connection_id")
		return
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 50
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	response, err := h.systemHistoryService.GetByConnectionID(connectionID, userID, page, pageSize)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, response)
}

// DeleteSystemHistory 删除系统执行记录
func (h *SystemHistoryHandler) DeleteSystemHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的历史记录 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.systemHistoryService.Delete(id, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "历史记录删除成功", nil)
}
