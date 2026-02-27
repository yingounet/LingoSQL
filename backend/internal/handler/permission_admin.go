package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type PermissionAdminHandler struct {
	permissionAdminService *service.PermissionAdminService
}

func NewPermissionAdminHandler(permissionAdminService *service.PermissionAdminService) *PermissionAdminHandler {
	return &PermissionAdminHandler{
		permissionAdminService: permissionAdminService,
	}
}

// GetPermissionTree 获取权限树
func (h *PermissionAdminHandler) GetPermissionTree(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	username := c.Query("username")
	if username == "" {
		utils.BadRequest(c, "用户名不能为空")
		return
	}

	host := c.Query("host")

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	tree, err := h.permissionAdminService.GetPermissionTree(connectionID, userID, username, host)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, tree)
}

// GrantPermission 授予权限
func (h *PermissionAdminHandler) GrantPermission(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	var req models.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.permissionAdminService.GrantPermission(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// RevokePermission 撤销权限
func (h *PermissionAdminHandler) RevokePermission(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	var req models.RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.permissionAdminService.RevokePermission(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}
