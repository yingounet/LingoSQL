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

type PermissionAdminService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewPermissionAdminService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *PermissionAdminService {
	return &PermissionAdminService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

// GetPermissionTree 获取权限树
func (s *PermissionAdminService) GetPermissionTree(connectionID, userID int, username, host string) ([]map[string]interface{}, error) {
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
		"", dbConfig.Username, password,
	)
	if err != nil {
		return nil, err
	}

	return executor.GetPermissionTree(username, host)
}

// GrantPermission 授予权限
func (s *PermissionAdminService) GrantPermission(connectionID, userID int, req *models.GrantPermissionRequest) error {
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
		"", dbConfig.Username, password,
	)
	if err != nil {
		return err
	}

	// 记录开始时间
	startTime := time.Now()

	var err2 error
	switch req.TargetType {
	case "database":
		err2 = executor.GrantDatabasePrivileges(req.Username, req.Host, req.TargetName, req.Privileges)
	case "table":
		if req.Database == "" {
			return errors.New("表权限需要指定数据库")
		}
		err2 = executor.GrantTablePrivileges(req.Username, req.Host, req.Database, req.TargetName, req.Privileges)
	case "column":
		if req.Database == "" || req.Table == "" {
			return errors.New("列权限需要指定数据库和表")
		}
		// 将权限列表转换为列权限映射
		columnPrivileges := make(map[string][]string)
		columnPrivileges[req.TargetName] = req.Privileges
		err2 = executor.GrantColumnPrivileges(req.Username, req.Host, req.Database, req.Table, columnPrivileges)
	default:
		return errors.New("不支持的目标类型")
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         "GRANT",
		OperationType:    "GRANT_PERMISSION",
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

// RevokePermission 撤销权限
func (s *PermissionAdminService) RevokePermission(connectionID, userID int, req *models.RevokePermissionRequest) error {
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
		"", dbConfig.Username, password,
	)
	if err != nil {
		return err
	}

	// 记录开始时间
	startTime := time.Now()

	var err2 error
	switch req.TargetType {
	case "database":
		err2 = executor.RevokeDatabasePrivileges(req.Username, req.Host, req.TargetName, req.Privileges)
	case "table":
		if req.Database == "" {
			return errors.New("表权限需要指定数据库")
		}
		err2 = executor.RevokeTablePrivileges(req.Username, req.Host, req.Database, req.TargetName, req.Privileges)
	case "column":
		if req.Database == "" || req.Table == "" {
			return errors.New("列权限需要指定数据库和表")
		}
		// 将权限列表转换为列权限映射
		columnPrivileges := make(map[string][]string)
		columnPrivileges[req.TargetName] = req.Privileges
		err2 = executor.RevokeColumnPrivileges(req.Username, req.Host, req.Database, req.Table, columnPrivileges)
	default:
		return errors.New("不支持的目标类型")
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         "REVOKE",
		OperationType:    "REVOKE_PERMISSION",
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
