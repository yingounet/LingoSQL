package handler

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type AuthHandler struct {
	authService         *service.AuthService
	auditService        *service.AuditService
	systemSettingsService *service.SystemSettingsService
}

func NewAuthHandler(authService *service.AuthService, auditService *service.AuditService, systemSettingsService *service.SystemSettingsService) *AuthHandler {
	return &AuthHandler{
		authService:           authService,
		auditService:          auditService,
		systemSettingsService: systemSettingsService,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	if !h.systemSettingsService.AllowRegistration() {
		utils.Forbidden(c, "系统已关闭公开注册")
		return
	}

	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, accessToken, refreshToken, err := h.authService.Register(&req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	h.auditService.Record(user.ID, "auth.register", "user", &user.ID, true, "", gin.H{
		"username": user.Username,
	})

	utils.SuccessWithMessage(c, "注册成功", gin.H{
		"user": models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(&req)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}
	h.auditService.Record(user.ID, "auth.login", "user", &user.ID, true, "", nil)

	utils.SuccessWithMessage(c, "登录成功", gin.H{
		"user": models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := GetUserID(c)
	if ok {
		h.auditService.Record(userID, "auth.logout", "user", &userID, true, "", nil)
	}
	utils.SuccessWithMessage(c, "登出成功", nil)
}

// GetMe 获取当前用户信息
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	user, err := h.authService.GetUser(userID)
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// Refresh 刷新访问令牌
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	accessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "刷新成功", gin.H{
		"token": accessToken,
	})
}
