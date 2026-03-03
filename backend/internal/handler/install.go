package handler

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type InstallHandler struct {
	installService        *service.InstallService
	auditService          *service.AuditService
	systemSettingsService *service.SystemSettingsService
}

func NewInstallHandler(installService *service.InstallService, auditService *service.AuditService, systemSettingsService *service.SystemSettingsService) *InstallHandler {
	return &InstallHandler{
		installService:        installService,
		auditService:          auditService,
		systemSettingsService: systemSettingsService,
	}
}

// GetStatus 获取安装状态（已安装时返回 allow_registration 供登录页使用）
func (h *InstallHandler) GetStatus(c *gin.Context) {
	installed, err := h.installService.IsInstalled()
	if err != nil {
		utils.InternalServerError(c, "获取安装状态失败")
		return
	}
	resp := gin.H{"installed": installed}
	if installed {
		resp["allow_registration"] = h.systemSettingsService.AllowRegistration()
	}
	utils.Success(c, resp)
}

// Setup 执行安装
func (h *InstallHandler) Setup(c *gin.Context) {
	var req models.InstallSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, accessToken, refreshToken, err := h.installService.Setup(&req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	h.auditService.Record(user.ID, "install.setup", "user", &user.ID, true, "", gin.H{
		"username": user.Username,
	})

	utils.SuccessWithMessage(c, "安装成功", gin.H{
		"user": models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		"token":          accessToken,
		"refresh_token":  refreshToken,
	})
}
