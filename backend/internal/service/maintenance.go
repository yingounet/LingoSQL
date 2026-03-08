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

// BackupDatabase 创建备份任务并异步执行，立即返回 task
func (s *MaintenanceService) BackupDatabase(connectionID, userID int, req *models.BackupRequest, taskService *TaskService) (*models.BackupResponse, error) {
	if req.Database == "" {
		conn, err := s.connectionDAO.GetByID(connectionID)
		if err != nil {
			return nil, err
		}
		var dbConfig models.DbConfig
		if json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig) == nil && dbConfig.Database != "" {
			req.Database = dbConfig.Database
		}
	}
	if req.Database == "" {
		return nil, errors.New("请指定要备份的数据库")
	}

	task, err := taskService.Create(userID, "BACKUP_DATABASE", req)
	if err != nil {
		return nil, err
	}
	if err := taskService.Start(task.ID); err != nil {
		return nil, err
	}

	go func() {
		result, err := s.RunBackup(connectionID, userID, req, task.ID, taskService)
		if err != nil {
			_ = taskService.CompleteFailure(task.ID, err)
			return
		}
		_ = taskService.CompleteSuccess(task.ID, result)
	}()

	return &models.BackupResponse{
		TaskID: &task.ID,
		// backup_id、download_url、file_size 在任务完成后通过 result 返回
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
