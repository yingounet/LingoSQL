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

type DatabaseAdminService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewDatabaseAdminService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *DatabaseAdminService {
	return &DatabaseAdminService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

// CheckAdminPermissions 检查管理权限
func (s *DatabaseAdminService) CheckAdminPermissions(connectionID, userID int) (*models.AdminPermissionResponse, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, err
	}

	if conn.UserID != userID {
		return nil, errors.New("无权访问此连接")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return nil, errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, err
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		dbConfig.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}

	// 检查权限
	permissions, err := executor.CheckAdminPermissions()
	if err != nil {
		return nil, err
	}

	return &models.AdminPermissionResponse{
		HasDatabaseAdmin:  permissions["has_database_admin"],
		HasUserAdmin:      permissions["has_user_admin"],
		HasPermissionAdmin: permissions["has_permission_admin"],
	}, nil
}

// GetDatabaseList 获取数据库列表（带详细信息）
func (s *DatabaseAdminService) GetDatabaseList(connectionID, userID int) ([]map[string]interface{}, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, err
	}

	if conn.UserID != userID {
		return nil, errors.New("无权访问此连接")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return nil, errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, err
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		dbConfig.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}

	// 获取数据库列表
	databases, err := executor.GetDatabases()
	if err != nil {
		return nil, err
	}

	// 获取每个数据库的详细信息
	var result []map[string]interface{}
	for _, dbName := range databases {
		info, err := executor.GetDatabaseInfo(dbName)
		if err != nil {
			// 如果获取详细信息失败，至少返回名称
			result = append(result, map[string]interface{}{
				"name": dbName,
			})
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// GetDatabaseInfo 获取数据库详情
func (s *DatabaseAdminService) GetDatabaseInfo(connectionID, userID int, databaseName string) (map[string]interface{}, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, err
	}

	if conn.UserID != userID {
		return nil, errors.New("无权访问此连接")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return nil, errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, err
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		dbConfig.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}

	return executor.GetDatabaseInfo(databaseName)
}

// CreateDatabase 创建数据库
func (s *DatabaseAdminService) CreateDatabase(connectionID, userID int, req *models.CreateDatabaseRequest) error {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return err
	}

	if conn.UserID != userID {
		return errors.New("无权访问此连接")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return err
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		dbConfig.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return err
	}

	// 记录开始时间
	startTime := time.Now()

	// 创建数据库
	var err2 error
	if conn.DBType == "postgresql" {
		// PostgreSQL: encoding作为charset参数，lc_collate作为collation参数
		encoding := req.Encoding
		if encoding == "" {
			encoding = "UTF8"
		}
		lcCollate := req.LcCollate
		if lcCollate == "" {
			lcCollate = "en_US.UTF-8"
		}
		err2 = executor.CreateDatabase(req.Name, encoding, lcCollate)
	} else {
		// MySQL
		charset := req.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		collation := req.Collation
		if collation == "" {
			collation = "utf8mb4_unicode_ci"
		}
		err2 = executor.CreateDatabase(req.Name, charset, collation)
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         "CREATE DATABASE",
		OperationType:    "CREATE_DATABASE",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     0,
		Success:          err2 == nil,
		ErrorMessage:     "",
	}
	if err2 != nil {
		history.ErrorMessage = err2.Error()
	}
	s.systemHistoryDAO.Create(history)

	return err2
}

// DropDatabase 删除数据库
func (s *DatabaseAdminService) DropDatabase(connectionID, userID int, databaseName string) error {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return err
	}

	if conn.UserID != userID {
		return errors.New("无权访问此连接")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return err
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		dbConfig.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return err
	}

	// 记录开始时间
	startTime := time.Now()

	// 删除数据库
	err2 := executor.DropDatabase(databaseName)

	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         "DROP DATABASE",
		OperationType:    "DROP_DATABASE",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     0,
		Success:          err2 == nil,
		ErrorMessage:     "",
	}
	if err2 != nil {
		history.ErrorMessage = err2.Error()
	}
	s.systemHistoryDAO.Create(history)

	return err2
}

// RenameDatabase 重命名数据库（仅PostgreSQL）
func (s *DatabaseAdminService) RenameDatabase(connectionID, userID int, databaseName string, req *models.RenameDatabaseRequest) error {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return err
	}

	if conn.UserID != userID {
		return errors.New("无权访问此连接")
	}

	if conn.DBType != "postgresql" {
		return errors.New("只有PostgreSQL支持重命名数据库")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return err
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		dbConfig.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return err
	}

	return executor.RenameDatabase(databaseName, req.NewName)
}
