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

type UserAdminService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewUserAdminService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *UserAdminService {
	return &UserAdminService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

// GetUsers 获取用户列表
func (s *UserAdminService) GetUsers(connectionID, userID int) ([]map[string]interface{}, error) {
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

	return executor.GetUsers()
}

// CreateUser 创建用户
func (s *UserAdminService) CreateUser(connectionID, userID int, req *models.CreateUserRequest) error {
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

	// 创建用户
	var err2 error
	if conn.DBType == "postgresql" {
		// PostgreSQL需要特殊处理
		// sql := "CREATE USER"
		// if req.IsSuperuser {
		// 	sql = "CREATE ROLE"
		// }
		// 这里简化处理，实际应该调用executor的方法
		err2 = executor.CreateUser(req.Username, "", req.Password)
	} else {
		err2 = executor.CreateUser(req.Username, req.Host, req.Password)
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         "CREATE USER",
		OperationType:    "CREATE_USER",
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

// DeleteUser 删除用户
func (s *UserAdminService) DeleteUser(connectionID, userID int, req *models.DeleteUserRequest) error {
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

	// 删除用户
	err2 := executor.DropUser(req.Username, req.Host)

	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         "DROP USER",
		OperationType:    "DROP_USER",
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

// ChangeUserPassword 修改用户密码
func (s *UserAdminService) ChangeUserPassword(connectionID, userID int, req *models.ChangePasswordRequest) error {
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

	return executor.AlterUserPassword(req.Username, req.Host, req.NewPassword)
}

// GetUserGrants 获取用户权限
func (s *UserAdminService) GetUserGrants(connectionID, userID int, username, host string) ([]string, error) {
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

	return executor.GetUserGrants(username, host)
}
