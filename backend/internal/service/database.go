package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

type DatabaseService struct {
	connectionDAO   *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewDatabaseService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *DatabaseService {
	return &DatabaseService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

// GetDatabases 获取数据库列表
func (s *DatabaseService) GetDatabases(connectionID, userID int) ([]string, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, fmt.Errorf("获取连接: %w", err)
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

	// 获取执行器（PostgreSQL 需连接到一个库，用连接配置的默认库）
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		dbConfig.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, fmt.Errorf("获取执行器: %w", err)
	}
	// 不要关闭连接，连接池会管理连接的生命周期

	// 构建 SQL 语句用于记录
	var sqlQuery string
	if conn.DBType == "postgresql" {
		sqlQuery = "SELECT datname FROM pg_database WHERE datistemplate = false AND datname != 'postgres'"
	} else {
		sqlQuery = "SHOW DATABASES"
	}

	// 记录开始时间
	startTime := time.Now()

	// 执行查询
	databases, err := executor.GetDatabases()

	// 计算执行时间
	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         sqlQuery,
		OperationType:    "GET_DATABASES",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     len(databases),
		Success:          err == nil,
		ErrorMessage:     "",
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)

	if err != nil {
		return nil, err
	}

	return databases, nil
}
