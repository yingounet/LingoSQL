package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type ConnectionHandler struct {
	connectionService *service.ConnectionService
}

func NewConnectionHandler(connectionService *service.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{connectionService: connectionService}
}

// GetConnections 获取连接列表
// GET /api/connections?page=1&page_size=10&db_type=mysql&search=prod
func (h *ConnectionHandler) GetConnections(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}

	// 解析查询参数
	var params models.ConnectionListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}

	response, err := h.connectionService.GetList(userID, &params)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, response)
}

// GetConnection 获取连接详情
// GET /api/connections/:id
func (h *ConnectionHandler) GetConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	response, err := h.connectionService.GetDetail(id, userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, response)
}

// CreateConnection 创建连接
// POST /api/connections
func (h *ConnectionHandler) CreateConnection(c *gin.Context) {
	var req models.ConnectionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证 SSH 配置
	if req.ConnectionType == "ssh_tunnel" {
		if req.SshConfig == nil {
			utils.BadRequest(c, "SSH 隧道模式需要提供 SSH 配置")
			return
		}
		if req.SshConfig.AuthType == "password" && req.SshConfig.Password == "" {
			utils.BadRequest(c, "密码认证模式需要提供 SSH 密码")
			return
		}
		if req.SshConfig.AuthType == "private_key" && req.SshConfig.PrivateKey == "" {
			utils.BadRequest(c, "私钥认证模式需要提供 SSH 私钥")
			return
		}
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	conn, err := h.connectionService.Create(userID, &req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 获取详情响应
	response, err := h.connectionService.GetDetail(conn.ID, userID)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "连接创建成功", response)
}

// UpdateConnection 更新连接
// PUT /api/connections/:id
func (h *ConnectionHandler) UpdateConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	var req models.ConnectionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证 SSH 配置
	if req.ConnectionType == "ssh_tunnel" && req.SshConfig != nil {
		if req.SshConfig.AuthType == "password" && req.SshConfig.Password == "" {
			utils.BadRequest(c, "密码认证模式需要提供 SSH 密码")
			return
		}
		if req.SshConfig.AuthType == "private_key" && req.SshConfig.PrivateKey == "" {
			utils.BadRequest(c, "私钥认证模式需要提供 SSH 私钥")
			return
		}
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	conn, err := h.connectionService.Update(id, userID, &req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 获取详情响应
	response, err := h.connectionService.GetDetail(conn.ID, userID)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "连接更新成功", response)
}

// DeleteConnection 删除连接
// DELETE /api/connections/:id
func (h *ConnectionHandler) DeleteConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.connectionService.Delete(id, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "连接删除成功", nil)
}

// TestConnection 测试已保存的连接
// POST /api/connections/:id/test
func (h *ConnectionHandler) TestConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	result, err := h.connectionService.Test(id, userID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if result.Connected {
		utils.SuccessWithMessage(c, "连接测试成功", result)
	} else {
		utils.BadRequest(c, result.Error)
	}
}

// TestConnectionConfig 测试连接配置（未保存的连接）
// POST /api/connections/test
func (h *ConnectionHandler) TestConnectionConfig(c *gin.Context) {
	var req models.ConnectionTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证 SSH 配置
	if req.ConnectionType == "ssh_tunnel" {
		if req.SshConfig == nil {
			utils.BadRequest(c, "SSH 隧道模式需要提供 SSH 配置")
			return
		}
	}

	result, err := h.connectionService.TestConfig(&req)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	if result.Connected {
		utils.SuccessWithMessage(c, "连接测试成功", result)
	} else {
		utils.BadRequest(c, result.Error)
	}
}

// UpdateLastUsed 更新连接的最后使用时间
// PUT /api/connections/:id/last-used
func (h *ConnectionHandler) UpdateLastUsed(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.connectionService.UpdateLastUsed(id, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

// SetDefaultConnection 设置默认连接
// PUT /api/connections/:id/default
func (h *ConnectionHandler) SetDefaultConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.connectionService.SetDefault(id, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "默认连接设置成功", nil)
}
