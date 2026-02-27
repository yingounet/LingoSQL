package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type DatabaseHandler struct {
	databaseService *service.DatabaseService
}

func NewDatabaseHandler(databaseService *service.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{databaseService: databaseService}
}

// GetDatabases 获取数据库列表
func (h *DatabaseHandler) GetDatabases(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	databases, err := h.databaseService.GetDatabases(connectionID, userID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, databases)
}
