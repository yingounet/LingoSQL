package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
	CORS       CORSConfig       `mapstructure:"cors"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Log        LogConfig        `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type JWTConfig struct {
	Secret           string `mapstructure:"secret"`
	ExpiresIn        string `mapstructure:"expires_in"`
	RefreshExpiresIn string `mapstructure:"refresh_expires_in"`
}

type EncryptionConfig struct {
	Key string `mapstructure:"key"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type RateLimitConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	DefaultRPM int  `mapstructure:"default_rpm"`
	PollingRPM int  `mapstructure:"polling_rpm"`
}

var AppConfig *Config

// Load 加载配置文件
func Load(configPath string) error {
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./backend")
	}

	// 设置环境变量
	viper.AutomaticEnv()
	_ = viper.BindEnv("jwt.secret", "JWT_SECRET")
	_ = viper.BindEnv("encryption.key", "ENCRYPTION_KEY")
	_ = viper.BindEnv("cors.allowed_origins", "ALLOWED_ORIGINS")
	_ = viper.BindEnv("rate_limit.default_rpm", "RATE_LIMIT_RPM")
	_ = viper.BindEnv("rate_limit.polling_rpm", "POLLING_RATE_LIMIT_RPM")
	_ = viper.BindEnv("rate_limit.enabled", "RATE_LIMIT_ENABLED")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := normalizeAndValidateConfig(AppConfig); err != nil {
		return err
	}

	// 确保数据目录存在
	if AppConfig.Database.Path != "" {
		dataDir := filepath.Dir(AppConfig.Database.Path)
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("创建数据目录失败: %w", err)
		}
	}

	return nil
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	return AppConfig
}

func normalizeAndValidateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("配置为空")
	}

	if envOrigins := viper.GetString("cors.allowed_origins"); envOrigins != "" {
		cfg.CORS.AllowedOrigins = splitCSV(envOrigins)
	}

	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = []string{"http://localhost:5173"}
		fmt.Println("警告: CORS 白名单未设置，已使用默认值 http://localhost:5173")
	}

	if strings.TrimSpace(cfg.JWT.Secret) == "" || strings.Contains(cfg.JWT.Secret, "change-in-production") {
		cfg.JWT.Secret = "dev-jwt-secret-change-this"
		fmt.Println("警告: JWT_SECRET 未设置，已使用默认值，请尽快修改")
	}

	if strings.TrimSpace(cfg.Encryption.Key) == "" {
		cfg.Encryption.Key = "1234567890abcdef1234567890abcdef"
		fmt.Println("警告: ENCRYPTION_KEY 未设置，已使用默认值，请尽快修改")
	} else if cfg.Encryption.Key == "1234567890abcdef1234567890abcdef" || cfg.Encryption.Key == "abcdefghijklmnopqrstuvwxyz123456" {
		fmt.Println("警告: ENCRYPTION_KEY 仍为默认值，请尽快修改")
	}
	if len(cfg.Encryption.Key) != 32 {
		return errors.New("ENCRYPTION_KEY 长度必须为 32 字节")
	}

	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.DefaultRPM <= 0 {
			cfg.RateLimit.DefaultRPM = 120
		}
		if cfg.RateLimit.PollingRPM <= 0 {
			cfg.RateLimit.PollingRPM = 30
		}
	}

	return nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}
