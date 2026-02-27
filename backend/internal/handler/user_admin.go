package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type UserAdminHandler struct {
	userAdminService *service.UserAdminService
}

func NewUserAdminHandler(userAdminService *service.UserAdminService) *UserAdminHandler {
	return &UserAdminHandler{
		userAdminService: userAdminService,
	}
}

// GetUsers 获取用户列表
func (h *UserAdminHandler) GetUsers(c *gin.Context) {
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
	users, err := h.userAdminService.GetUsers(connectionID, userID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, users)
}

// CreateUser 创建用户
func (h *UserAdminHandler) CreateUser(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.userAdminService.CreateUser(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// DeleteUser 删除用户
func (h *UserAdminHandler) DeleteUser(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	var req models.DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.userAdminService.DeleteUser(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// ChangeUserPassword 修改用户密码
func (h *UserAdminHandler) ChangeUserPassword(c *gin.Context) {
	connectionIDStr := c.Query("connection_id")
	connectionID, err := strconv.Atoi(connectionIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.userAdminService.ChangeUserPassword(connectionID, userID, &req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetUserGrants 获取用户权限
func (h *UserAdminHandler) GetUserGrants(c *gin.Context) {
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
	grants, err := h.userAdminService.GetUserGrants(connectionID, userID, username, host)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, grants)
}
