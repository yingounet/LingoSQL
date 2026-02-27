package handler

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type ImportHandler struct {
	importService *service.ImportService
}

func NewImportHandler(importService *service.ImportService) *ImportHandler {
	return &ImportHandler{importService: importService}
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
