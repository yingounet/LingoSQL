package models

import "time"

// Connection 数据库连接模型
type Connection struct {
	ID             int        `json:"id" db:"id"`
	UserID         int        `json:"user_id" db:"user_id"`
	Name           string     `json:"name" db:"name"`
	DBType         string     `json:"db_type" db:"db_type"`
	ConnectionType string     `json:"connection_type" db:"connection_type"`
	DbConfigJSON   string     `json:"-" db:"db_config"`  // JSON 存储
	SshConfigJSON  *string    `json:"-" db:"ssh_config"` // JSON 存储，可为 NULL
	IsDefault      bool       `json:"is_default" db:"is_default"`
	LastUsedAt     *time.Time `json:"-" db:"last_used_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// DbConfig 数据库配置（JSON 结构）
type DbConfig struct {
	Host              string     `json:"host"`
	Port              int        `json:"port"`
	Database          string     `json:"database,omitempty"`
	Username          string     `json:"username"`
	PasswordEncrypted string     `json:"password_encrypted,omitempty"`
	Password          string     `json:"-"`
	Options           *DbOptions `json:"options,omitempty"`
}

// DbOptions 数据库扩展选项
type DbOptions struct {
	SslMode string `json:"ssl_mode,omitempty"`
	Charset string `json:"charset,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// SshConfig SSH 隧道配置（JSON 结构）
type SshConfig struct {
	Host                string `json:"host"`
	Port                int    `json:"port"`
	Username            string `json:"username"`
	AuthType            string `json:"auth_type"` // "password" 或 "private_key"
	PasswordEncrypted   string `json:"password_encrypted,omitempty"`
	PrivateKeyEncrypted string `json:"private_key_encrypted,omitempty"`
	PassphraseEncrypted string `json:"passphrase_encrypted,omitempty"`
}

// ConnectionListParams 连接列表查询参数
type ConnectionListParams struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	DbType   string `form:"db_type"`
	Search   string `form:"search"`
}

// ConnectionCreateRequest 创建连接请求
type ConnectionCreateRequest struct {
	Name           string                   `json:"name" binding:"required"`
	DBType         string                   `json:"db_type" binding:"required,oneof=mysql postgresql mariadb"`
	ConnectionType string                   `json:"connection_type" binding:"required,oneof=direct ssh_tunnel"`
	DbConfig       ConnectionDbConfigInput  `json:"db_config" binding:"required"`
	SshConfig      *ConnectionSshConfigInput `json:"ssh_config"`
}

// ConnectionDbConfigInput 数据库配置输入
type ConnectionDbConfigInput struct {
	Host     string                  `json:"host" binding:"required"`
	Port     int                     `json:"port" binding:"required,min=1,max=65535"`
	Database string                  `json:"database"`
	Username string                  `json:"username" binding:"required"`
	Password string                  `json:"password" binding:"required"`
	Options  *ConnectionOptionsInput `json:"options"`
}

// ConnectionOptionsInput 数据库选项输入
type ConnectionOptionsInput struct {
	SslMode string `json:"ssl_mode"`
	Charset string `json:"charset"`
	Timeout int    `json:"timeout"`
}

// ConnectionSshConfigInput SSH 配置输入
type ConnectionSshConfigInput struct {
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port" binding:"required,min=1,max=65535"`
	Username   string `json:"username" binding:"required"`
	AuthType   string `json:"auth_type" binding:"required,oneof=password private_key"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	Passphrase string `json:"passphrase"`
}

// ConnectionUpdateRequest 更新连接请求
type ConnectionUpdateRequest struct {
	Name           string                    `json:"name"`
	DBType         string                    `json:"db_type" binding:"omitempty,oneof=mysql postgresql mariadb"`
	ConnectionType string                    `json:"connection_type" binding:"omitempty,oneof=direct ssh_tunnel"`
	DbConfig       *ConnectionDbConfigInput  `json:"db_config"`
	SshConfig      *ConnectionSshConfigInput `json:"ssh_config"`
}

// ConnectionTestRequest 测试连接请求（未保存的连接）
type ConnectionTestRequest struct {
	DBType         string                    `json:"db_type" binding:"required,oneof=mysql postgresql mariadb"`
	ConnectionType string                    `json:"connection_type" binding:"required,oneof=direct ssh_tunnel"`
	DbConfig       ConnectionDbConfigInput   `json:"db_config" binding:"required"`
	SshConfig      *ConnectionSshConfigInput `json:"ssh_config"`
}

// ConnectionResponse 连接响应（列表项）
type ConnectionResponse struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	DBType         string    `json:"db_type"`
	ConnectionType string    `json:"connection_type"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	Database       string    `json:"database,omitempty"`
	IsDefault      bool      `json:"is_default"`
	LastUsedAt     *string   `json:"last_used_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ConnectionDetailResponse 连接详情响应
type ConnectionDetailResponse struct {
	ID             int                       `json:"id"`
	Name           string                    `json:"name"`
	DBType         string                    `json:"db_type"`
	ConnectionType string                    `json:"connection_type"`
	DbConfig       *DbConfigResponse         `json:"db_config"`
	SshConfig      *SshConfigResponse        `json:"ssh_config,omitempty"`
	IsDefault      bool                      `json:"is_default"`
	LastUsedAt     *string                   `json:"last_used_at,omitempty"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

// DbConfigResponse 数据库配置响应（不含密码）
type DbConfigResponse struct {
	Host     string             `json:"host"`
	Port     int                `json:"port"`
	Database string             `json:"database,omitempty"`
	Username string             `json:"username"`
	Options  *DbOptionsResponse `json:"options,omitempty"`
}

// DbOptionsResponse 数据库选项响应
type DbOptionsResponse struct {
	SslMode string `json:"ssl_mode,omitempty"`
	Charset string `json:"charset,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// SshConfigResponse SSH 配置响应（不含密码和私钥）
type SshConfigResponse struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	AuthType string `json:"auth_type"`
}

// ConnectionListResponse 连接列表响应
type ConnectionListResponse struct {
	List     []ConnectionResponse `json:"list"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// TestConnectionResponse 测试连接响应
type TestConnectionResponse struct {
	Connected bool   `json:"connected"`
	Version   string `json:"version,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}
