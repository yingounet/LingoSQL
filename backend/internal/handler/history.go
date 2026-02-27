package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type HistoryHandler struct {
	historyService *service.HistoryService
}

func NewHistoryHandler(historyService *service.HistoryService) *HistoryHandler {
	return &HistoryHandler{historyService: historyService}
}

// GetHistory 获取历史记录列表
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	var connectionID *int
	if connectionIDStr := c.Query("connection_id"); connectionIDStr != "" {
		if id, err := strconv.Atoi(connectionIDStr); err == nil {
			connectionID = &id
		}
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

	response, err := h.historyService.GetByUserID(userID, connectionID, page, pageSize)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, response)
}

// GetHistoryByID 获取单条历史记录
func (h *HistoryHandler) GetHistoryByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的历史记录 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	history, err := h.historyService.GetByID(id, userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, history)
}

// DeleteHistory 删除历史记录
func (h *HistoryHandler) DeleteHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的历史记录 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.historyService.Delete(id, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "历史记录删除成功", nil)
}
