package service

import (
	"encoding/json"
	"errors"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

type ConnectionService struct {
	connectionDAO *sqlite.ConnectionDAO
}

func NewConnectionService(connectionDAO *sqlite.ConnectionDAO) *ConnectionService {
	return &ConnectionService{connectionDAO: connectionDAO}
}

// Create 创建连接
func (s *ConnectionService) Create(userID int, req *models.ConnectionCreateRequest) (*models.Connection, error) {
	// 构建数据库配置
	dbConfig := &models.DbConfig{
		Host:     req.DbConfig.Host,
		Port:     req.DbConfig.Port,
		Database: req.DbConfig.Database,
		Username: req.DbConfig.Username,
	}

	// 加密数据库密码
	encryptedPassword, err := utils.Encrypt(req.DbConfig.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	dbConfig.PasswordEncrypted = encryptedPassword

	// 处理选项
	if req.DbConfig.Options != nil {
		dbConfig.Options = &models.DbOptions{
			SslMode: req.DbConfig.Options.SslMode,
			Charset: req.DbConfig.Options.Charset,
			Timeout: req.DbConfig.Options.Timeout,
		}
	}

	// 序列化数据库配置
	dbConfigJSON, err := json.Marshal(dbConfig)
	if err != nil {
		return nil, errors.New("配置序列化失败")
	}

	conn := &models.Connection{
		UserID:         userID,
		Name:           req.Name,
		DBType:         req.DBType,
		ConnectionType: req.ConnectionType,
		DbConfigJSON:   string(dbConfigJSON),
		IsDefault:      false,
	}

	// 处理 SSH 配置
	if req.ConnectionType == "ssh_tunnel" && req.SshConfig != nil {
		sshConfig := &models.SshConfig{
			Host:     req.SshConfig.Host,
			Port:     req.SshConfig.Port,
			Username: req.SshConfig.Username,
			AuthType: req.SshConfig.AuthType,
		}

		// 加密 SSH 密码
		if req.SshConfig.AuthType == "password" && req.SshConfig.Password != "" {
			encryptedSshPassword, err := utils.Encrypt(req.SshConfig.Password)
			if err != nil {
				return nil, errors.New("SSH 密码加密失败")
			}
			sshConfig.PasswordEncrypted = encryptedSshPassword
		}

		// 加密私钥
		if req.SshConfig.AuthType == "private_key" && req.SshConfig.PrivateKey != "" {
			encryptedPrivateKey, err := utils.Encrypt(req.SshConfig.PrivateKey)
			if err != nil {
				return nil, errors.New("私钥加密失败")
			}
			sshConfig.PrivateKeyEncrypted = encryptedPrivateKey

			// 加密私钥密码
			if req.SshConfig.Passphrase != "" {
				encryptedPassphrase, err := utils.Encrypt(req.SshConfig.Passphrase)
				if err != nil {
					return nil, errors.New("私钥密码加密失败")
				}
				sshConfig.PassphraseEncrypted = encryptedPassphrase
			}
		}

		sshConfigJSON, err := json.Marshal(sshConfig)
		if err != nil {
			return nil, errors.New("SSH 配置序列化失败")
		}
		sshConfigStr := string(sshConfigJSON)
		conn.SshConfigJSON = &sshConfigStr
	}

	// 如果是第一个连接，设为默认
	count, err := s.connectionDAO.GetUserConnectionCount(userID)
	if err == nil && count == 0 {
		conn.IsDefault = true
	}

	if err := s.connectionDAO.Create(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// GetList 获取连接列表
func (s *ConnectionService) GetList(userID int, params *models.ConnectionListParams) (*models.ConnectionListResponse, error) {
	connections, total, err := s.connectionDAO.GetByUserID(userID, params)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	list := make([]models.ConnectionResponse, len(connections))
	for i, conn := range connections {
		dbConfig, _ := s.parseDbConfig(conn.DbConfigJSON)
		var lastUsedAt *string
		if conn.LastUsedAt != nil {
			s := conn.LastUsedAt.Format(time.RFC3339)
			lastUsedAt = &s
		}
		list[i] = models.ConnectionResponse{
			ID:             conn.ID,
			Name:           conn.Name,
			DBType:         conn.DBType,
			ConnectionType: conn.ConnectionType,
			Host:           dbConfig.Host,
			Port:           dbConfig.Port,
			Database:       dbConfig.Database,
			IsDefault:      conn.IsDefault,
			LastUsedAt:     lastUsedAt,
			CreatedAt:      conn.CreatedAt,
			UpdatedAt:      conn.UpdatedAt,
		}
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	return &models.ConnectionListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetByID 获取连接详情
func (s *ConnectionService) GetByID(id, userID int) (*models.Connection, error) {
	conn, err := s.connectionDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	if conn.UserID != userID {
		return nil, errors.New("无权访问此连接")
	}

	return conn, nil
}

// GetDetail 获取连接详情响应
func (s *ConnectionService) GetDetail(id, userID int) (*models.ConnectionDetailResponse, error) {
	conn, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	// 解析数据库配置
	dbConfig, err := s.parseDbConfig(conn.DbConfigJSON)
	if err != nil {
		return nil, errors.New("配置解析失败")
	}

	var lastUsedAt *string
	if conn.LastUsedAt != nil {
		s := conn.LastUsedAt.Format(time.RFC3339)
		lastUsedAt = &s
	}
	response := &models.ConnectionDetailResponse{
		ID:             conn.ID,
		Name:           conn.Name,
		DBType:         conn.DBType,
		ConnectionType: conn.ConnectionType,
		DbConfig: &models.DbConfigResponse{
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Database: dbConfig.Database,
			Username: dbConfig.Username,
		},
		IsDefault:  conn.IsDefault,
		LastUsedAt: lastUsedAt,
		CreatedAt:  conn.CreatedAt,
		UpdatedAt:  conn.UpdatedAt,
	}

	// 处理选项
	if dbConfig.Options != nil {
		response.DbConfig.Options = &models.DbOptionsResponse{
			SslMode: dbConfig.Options.SslMode,
			Charset: dbConfig.Options.Charset,
			Timeout: dbConfig.Options.Timeout,
		}
	}

	// 解析 SSH 配置
	if conn.SshConfigJSON != nil && *conn.SshConfigJSON != "" {
		sshConfig, err := s.parseSshConfig(*conn.SshConfigJSON)
		if err == nil {
			response.SshConfig = &models.SshConfigResponse{
				Host:     sshConfig.Host,
				Port:     sshConfig.Port,
				Username: sshConfig.Username,
				AuthType: sshConfig.AuthType,
			}
		}
	}

	return response, nil
}

// Update 更新连接
func (s *ConnectionService) Update(id, userID int, req *models.ConnectionUpdateRequest) (*models.Connection, error) {
	conn, err := s.connectionDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	if conn.UserID != userID {
		return nil, errors.New("无权访问此连接")
	}

	// 解析现有配置
	dbConfig, err := s.parseDbConfig(conn.DbConfigJSON)
	if err != nil {
		return nil, errors.New("配置解析失败")
	}

	// 更新基本字段
	if req.Name != "" {
		conn.Name = req.Name
	}
	if req.DBType != "" {
		conn.DBType = req.DBType
	}
	if req.ConnectionType != "" {
		conn.ConnectionType = req.ConnectionType
	}

	// 更新数据库配置
	if req.DbConfig != nil {
		if req.DbConfig.Host != "" {
			dbConfig.Host = req.DbConfig.Host
		}
		if req.DbConfig.Port > 0 {
			dbConfig.Port = req.DbConfig.Port
		}
		if req.DbConfig.Database != "" {
			dbConfig.Database = req.DbConfig.Database
		}
		if req.DbConfig.Username != "" {
			dbConfig.Username = req.DbConfig.Username
		}
		if req.DbConfig.Password != "" {
			encryptedPassword, err := utils.Encrypt(req.DbConfig.Password)
			if err != nil {
				return nil, errors.New("密码加密失败")
			}
			dbConfig.PasswordEncrypted = encryptedPassword
		}
		if req.DbConfig.Options != nil {
			if dbConfig.Options == nil {
				dbConfig.Options = &models.DbOptions{}
			}
			if req.DbConfig.Options.SslMode != "" {
				dbConfig.Options.SslMode = req.DbConfig.Options.SslMode
			}
			if req.DbConfig.Options.Charset != "" {
				dbConfig.Options.Charset = req.DbConfig.Options.Charset
			}
			if req.DbConfig.Options.Timeout > 0 {
				dbConfig.Options.Timeout = req.DbConfig.Options.Timeout
			}
		}

		dbConfigJSON, err := json.Marshal(dbConfig)
		if err != nil {
			return nil, errors.New("配置序列化失败")
		}
		conn.DbConfigJSON = string(dbConfigJSON)
	}

	// 更新 SSH 配置
	if req.ConnectionType == "ssh_tunnel" && req.SshConfig != nil {
		sshConfig := &models.SshConfig{
			Host:     req.SshConfig.Host,
			Port:     req.SshConfig.Port,
			Username: req.SshConfig.Username,
			AuthType: req.SshConfig.AuthType,
		}

		if req.SshConfig.AuthType == "password" && req.SshConfig.Password != "" {
			encryptedSshPassword, err := utils.Encrypt(req.SshConfig.Password)
			if err != nil {
				return nil, errors.New("SSH 密码加密失败")
			}
			sshConfig.PasswordEncrypted = encryptedSshPassword
		}

		if req.SshConfig.AuthType == "private_key" && req.SshConfig.PrivateKey != "" {
			encryptedPrivateKey, err := utils.Encrypt(req.SshConfig.PrivateKey)
			if err != nil {
				return nil, errors.New("私钥加密失败")
			}
			sshConfig.PrivateKeyEncrypted = encryptedPrivateKey

			if req.SshConfig.Passphrase != "" {
				encryptedPassphrase, err := utils.Encrypt(req.SshConfig.Passphrase)
				if err != nil {
					return nil, errors.New("私钥密码加密失败")
				}
				sshConfig.PassphraseEncrypted = encryptedPassphrase
			}
		}

		sshConfigJSON, err := json.Marshal(sshConfig)
		if err != nil {
			return nil, errors.New("SSH 配置序列化失败")
		}
		sshConfigStr := string(sshConfigJSON)
		conn.SshConfigJSON = &sshConfigStr
	} else if req.ConnectionType == "direct" {
		// 切换到直连模式，清除 SSH 配置
		conn.SshConfigJSON = nil
	}

	if err := s.connectionDAO.Update(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// Delete 删除连接
func (s *ConnectionService) Delete(id, userID int) error {
	conn, err := s.connectionDAO.GetByID(id)
	if err != nil {
		return err
	}

	if conn.UserID != userID {
		return errors.New("无权访问此连接")
	}

	// 关闭连接池中的连接
	db.GetPool().CloseConnection(id)

	return s.connectionDAO.Delete(id, userID)
}

// UpdateLastUsed 更新连接的最后使用时间
func (s *ConnectionService) UpdateLastUsed(id, userID int) error {
	exists, err := s.connectionDAO.Exists(id, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("连接不存在")
	}
	return s.connectionDAO.UpdateLastUsedAt(id, userID, time.Now())
}

// SetDefault 设置默认连接
func (s *ConnectionService) SetDefault(id, userID int) error {
	// 检查连接是否存在
	exists, err := s.connectionDAO.Exists(id, userID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("连接不存在")
	}

	return s.connectionDAO.SetDefault(id, userID)
}

// Test 测试已保存的连接
func (s *ConnectionService) Test(id, userID int) (*models.TestConnectionResponse, error) {
	conn, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	// 解析配置
	dbConfig, err := s.parseDbConfig(conn.DbConfigJSON)
	if err != nil {
		return &models.TestConnectionResponse{
			Connected: false,
			Error:     "配置解析失败",
		}, nil
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return &models.TestConnectionResponse{
			Connected: false,
			Error:     "密码解密失败",
		}, nil
	}

	// 测试连接
	return s.testConnection(conn.DBType, dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.Username, password)
}

// TestConfig 测试未保存的连接配置
func (s *ConnectionService) TestConfig(req *models.ConnectionTestRequest) (*models.TestConnectionResponse, error) {
	// 直接使用请求中的密码测试
	return s.testConnection(req.DBType, req.DbConfig.Host, req.DbConfig.Port,
		req.DbConfig.Database, req.DbConfig.Username, req.DbConfig.Password)
}

// testConnection 测试数据库连接
func (s *ConnectionService) testConnection(dbType, host string, port int, database, username, password string) (*models.TestConnectionResponse, error) {
	startTime := time.Now()

	// 创建执行器测试连接
	// connectionID=0 表示临时连接，不会被连接池缓存，需要手动关闭
	executor, err := db.GetPool().GetExecutor(
		0, dbType, host, port, database, username, password,
	)
	if err != nil {
		return &models.TestConnectionResponse{
			Connected: false,
			Error:     err.Error(),
		}, nil
	}
	defer executor.Close() // 临时连接需要手动关闭

	// 尝试获取数据库版本
	version, err := executor.GetVersion()
	if err != nil {
		// 尝试获取数据库列表作为备选
		_, listErr := executor.GetDatabases()
		if listErr != nil {
			return &models.TestConnectionResponse{
				Connected: false,
				Error:     err.Error(),
			}, nil
		}
	}

	latency := time.Since(startTime).Milliseconds()

	return &models.TestConnectionResponse{
		Connected: true,
		Version:   version,
		LatencyMs: latency,
	}, nil
}

// parseDbConfig 解析数据库配置
func (s *ConnectionService) parseDbConfig(jsonStr string) (*models.DbConfig, error) {
	var config models.DbConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// parseSshConfig 解析 SSH 配置
func (s *ConnectionService) parseSshConfig(jsonStr string) (*models.SshConfig, error) {
	var config models.SshConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// GetDecryptedDbConfig 获取解密后的数据库配置（内部使用）
func (s *ConnectionService) GetDecryptedDbConfig(id, userID int) (*models.DbConfig, error) {
	conn, err := s.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	dbConfig, err := s.parseDbConfig(conn.DbConfigJSON)
	if err != nil {
		return nil, err
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, err
	}
	dbConfig.Password = password

	return dbConfig, nil
}
