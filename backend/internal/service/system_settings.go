package service

import (
	"encoding/json"
	"strconv"

	"lingosql/internal/config"
	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
)

type SystemSettingsService struct {
	dao *sqlite.SystemSettingsDAO
}

func NewSystemSettingsService(dao *sqlite.SystemSettingsDAO) *SystemSettingsService {
	return &SystemSettingsService{dao: dao}
}

// SaveInstallSettings 保存安装时的配置
func (s *SystemSettingsService) SaveInstallSettings(settings *models.InstallSettings) error {
	if err := s.dao.Set("site_name", settings.SiteName); err != nil {
		return err
	}
	if err := s.dao.Set("allow_registration", strconv.FormatBool(settings.AllowRegistration)); err != nil {
		return err
	}
	if err := s.dao.Set("rate_limit_enabled", strconv.FormatBool(settings.RateLimitEnabled)); err != nil {
		return err
	}
	if err := s.dao.Set("rate_limit_default_rpm", strconv.Itoa(settings.RateLimitDefaultRPM)); err != nil {
		return err
	}
	if err := s.dao.Set("rate_limit_polling_rpm", strconv.Itoa(settings.RateLimitPollingRPM)); err != nil {
		return err
	}
	originsJSON, err := json.Marshal(settings.CORSAllowedOrigins)
	if err != nil {
		return err
	}
	return s.dao.Set("cors_allowed_origins", string(originsJSON))
}

// ApplyToConfig 从 DB 读取配置并覆盖 AppConfig（启动时调用）
func (s *SystemSettingsService) ApplyToConfig() error {
	all, err := s.dao.GetAll()
	if err != nil || len(all) == 0 {
		return err
	}

	cfg := config.GetConfig()
	if cfg == nil {
		return nil
	}

	if v, ok := all["site_name"]; ok && v != "" {
		// 可扩展：用于页面标题等，当前 config 无此字段，后续可加
		_ = v
	}
	if v, ok := all["allow_registration"]; ok {
		_ = v
		// 由 handler 读取，不写入 config
	}
	if v, ok := all["rate_limit_enabled"]; ok {
		if enabled, err := strconv.ParseBool(v); err == nil {
			cfg.RateLimit.Enabled = enabled
		}
	}
	if v, ok := all["rate_limit_default_rpm"]; ok {
		if rpm, err := strconv.Atoi(v); err == nil && rpm > 0 {
			cfg.RateLimit.DefaultRPM = rpm
		}
	}
	if v, ok := all["rate_limit_polling_rpm"]; ok {
		if rpm, err := strconv.Atoi(v); err == nil && rpm > 0 {
			cfg.RateLimit.PollingRPM = rpm
		}
	}
	if v, ok := all["cors_allowed_origins"]; ok && v != "" {
		var origins []string
		if err := json.Unmarshal([]byte(v), &origins); err == nil && len(origins) > 0 {
			cfg.CORS.AllowedOrigins = origins
		}
	}

	return nil
}

// AllowRegistration 是否允许公开注册
func (s *SystemSettingsService) AllowRegistration() bool {
	v, err := s.dao.Get("allow_registration")
	if err != nil || v == "" {
		return true // 未设置时默认允许（兼容旧数据）
	}
	allow, _ := strconv.ParseBool(v)
	return allow
}
