package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

type MaintenanceService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewMaintenanceService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *MaintenanceService {
	return &MaintenanceService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

func (s *MaintenanceService) getExecutor(connectionID, userID int, database string) (db.Executor, *models.Connection, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, nil, err
	}
	if conn.UserID != userID {
		return nil, nil, errors.New("无权访问此连接")
	}
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return nil, nil, errors.New("配置解析失败")
	}
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, nil, err
	}
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, nil, err
	}
	return executor, conn, nil
}

// OptimizeTable 优化表
func (s *MaintenanceService) OptimizeTable(connectionID, userID int, database, table string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		// PostgreSQL: 使用 VACUUM ANALYZE 等效于 MySQL 的 OPTIMIZE TABLE
		sql = fmt.Sprintf("VACUUM ANALYZE \"public\".\"%s\"", table)
	} else {
		sql = fmt.Sprintf("OPTIMIZE TABLE `%s`.`%s`", database, table)
	}
	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "OPTIMIZE_TABLE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// RepairTable 修复表
func (s *MaintenanceService) RepairTable(connectionID, userID int, database, table string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	if conn.DBType == "postgresql" {
		return errors.New("PostgreSQL 不支持 REPAIR TABLE，可使用 VACUUM FULL 回收空间")
	}
	startTime := time.Now()

	sql := fmt.Sprintf("REPAIR TABLE `%s`.`%s`", database, table)
	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "REPAIR_TABLE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// AnalyzeTable 分析表
func (s *MaintenanceService) AnalyzeTable(connectionID, userID int, database, table string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		// PostgreSQL: 使用 public schema，连接已绑定到目标 database
		sql = fmt.Sprintf("ANALYZE \"public\".\"%s\"", table)
	} else {
		sql = fmt.Sprintf("ANALYZE TABLE `%s`.`%s`", database, table)
	}

	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "ANALYZE_TABLE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// BackupDatabase 备份数据库（简化实现，返回SQL导出）
func (s *MaintenanceService) BackupDatabase(connectionID, userID int, req *models.BackupRequest) (*models.BackupResponse, error) {
	// 简化实现：返回备份ID和下载URL（实际应该生成文件）
	backupID := fmt.Sprintf("backup_%d_%d", connectionID, time.Now().Unix())
	return &models.BackupResponse{
		BackupID:    backupID,
		DownloadURL: fmt.Sprintf("/api/admin/backup/%s/download", backupID),
		FileSize:    0,
	}, nil
}

// RestoreDatabase 恢复数据库
func (s *MaintenanceService) RestoreDatabase(connectionID, userID int, req *models.RestoreRequest) error {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	// 执行SQL文件内容
	statements := utils.SplitSQLStatements(req.SQLFile)
	for _, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		_, _, err := executor.ExecuteUpdate(stmt)
		if err != nil {
			executionTime := int(time.Since(startTime).Milliseconds())
			history := &models.SystemQueryHistory{
				ConnectionID:    connectionID,
				UserID:         userID,
				SQLQuery:       "RESTORE DATABASE",
				OperationType:  "RESTORE_DATABASE",
				ExecutionTimeMs: executionTime,
				Success:        false,
				ErrorMessage:   err.Error(),
			}
			s.systemHistoryDAO.Create(history)
			return err
		}
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       "RESTORE DATABASE",
		OperationType:  "RESTORE_DATABASE",
		ExecutionTimeMs: executionTime,
		Success:        true,
	}
	s.systemHistoryDAO.Create(history)
	return nil
}
