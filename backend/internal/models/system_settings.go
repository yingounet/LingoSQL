package models

// SystemSetting 系统配置项
type SystemSetting struct {
	Key       string `json:"key" db:"key"`
	Value     string `json:"value" db:"value"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

// InstallSettings 安装时提交的系统配置
type InstallSettings struct {
	SiteName           string   `json:"site_name"`
	AllowRegistration  bool     `json:"allow_registration"`
	RateLimitEnabled   bool     `json:"rate_limit_enabled"`
	RateLimitDefaultRPM   int   `json:"rate_limit_default_rpm"`
	RateLimitPollingRPM int    `json:"rate_limit_polling_rpm"`
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
}

// InstallSetupRequest 安装完成请求
type InstallSetupRequest struct {
	Admin    UserCreateRequest `json:"admin" binding:"required"`
	Settings InstallSettings   `json:"settings"`
}
