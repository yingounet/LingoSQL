package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
)

type InstallService struct {
	db       *sql.DB
	userDAO  *sqlite.UserDAO
	settingsDAO *sqlite.SystemSettingsDAO
}

func NewInstallService(db *sql.DB, userDAO *sqlite.UserDAO, settingsDAO *sqlite.SystemSettingsDAO) *InstallService {
	return &InstallService{
		db:          db,
		userDAO:     userDAO,
		settingsDAO: settingsDAO,
	}
}

// IsInstalled 是否已完成安装（存在用户）
func (s *InstallService) IsInstalled() (bool, error) {
	count, err := s.userDAO.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Setup 执行安装：创建管理员并保存配置
func (s *InstallService) Setup(req *models.InstallSetupRequest) (*models.User, string, string, error) {
	installed, err := s.IsInstalled()
	if err != nil {
		return nil, "", "", err
	}
	if installed {
		return nil, "", "", errors.New("系统已完成安装")
	}

	// 校验管理员信息
	if err := utils.ValidatePasswordStrength(req.Admin.Password); err != nil {
		return nil, "", "", err
	}

	// 加密密码
	passwordHash, err := utils.HashPassword(req.Admin.Password)
	if err != nil {
		return nil, "", "", err
	}

	user := &models.User{
		Username:     req.Admin.Username,
		Email:        req.Admin.Email,
		PasswordHash: passwordHash,
	}

	// 默认配置
	settings := &req.Settings
	if settings.SiteName == "" {
		settings.SiteName = "LingoSQL"
	}
	if settings.RateLimitDefaultRPM <= 0 {
		settings.RateLimitDefaultRPM = 120
	}
	if settings.RateLimitPollingRPM <= 0 {
		settings.RateLimitPollingRPM = 30
	}
	if len(settings.CORSAllowedOrigins) == 0 {
		settings.CORSAllowedOrigins = []string{"http://localhost:5173"}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, "", "", err
	}
	defer tx.Rollback()

	if err := s.userDAO.CreateWithTx(tx, user); err != nil {
		return nil, "", "", err
	}

	// 保存系统配置
	setStr := func(key, value string) error {
		return s.settingsDAO.SetWithTx(tx, key, value)
	}
	if err := setStr("site_name", settings.SiteName); err != nil {
		return nil, "", "", err
	}
	if err := setStr("allow_registration", strconv.FormatBool(settings.AllowRegistration)); err != nil {
		return nil, "", "", err
	}
	if err := setStr("rate_limit_enabled", strconv.FormatBool(settings.RateLimitEnabled)); err != nil {
		return nil, "", "", err
	}
	if err := setStr("rate_limit_default_rpm", strconv.Itoa(settings.RateLimitDefaultRPM)); err != nil {
		return nil, "", "", err
	}
	if err := setStr("rate_limit_polling_rpm", strconv.Itoa(settings.RateLimitPollingRPM)); err != nil {
		return nil, "", "", err
	}
	originsJSON, _ := json.Marshal(settings.CORSAllowedOrigins)
	if err := setStr("cors_allowed_origins", string(originsJSON)); err != nil {
		return nil, "", "", err
	}

	if err := tx.Commit(); err != nil {
		return nil, "", "", err
	}

	// 生成 token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", err
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}
