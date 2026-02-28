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

type ImportService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewImportService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *ImportService {
	return &ImportService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

func (s *ImportService) getExecutor(connectionID, userID int, database string) (db.Executor, *models.Connection, error) {
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
		database, dbConfig.Username, password,
	)
	if err != nil {
		return nil, nil, err
	}
	return executor, conn, nil
}

// ImportData 导入数据到表
func (s *ImportService) ImportData(connectionID, userID int, req *models.ImportDataRequest) (*models.ImportDataResponse, error) {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	response := &models.ImportDataResponse{
		TotalRows: len(req.Data),
	}

	// 构建INSERT语句
	dataStart := 0
	if req.SkipFirstRow && len(req.Data) > 0 {
		dataStart = 1
	}

	// 映射字段名
	columnNames := req.Headers
	if req.FieldMapping != nil && len(req.FieldMapping) > 0 {
		mappedColumns := make([]string, len(req.Headers))
		for i, h := range req.Headers {
			if mapped, ok := req.FieldMapping[h]; ok {
				mappedColumns[i] = mapped
			} else {
				mappedColumns[i] = h
			}
		}
		columnNames = mappedColumns
	}

	// 批量插入数据
	batchSize := 100
	for i := dataStart; i < len(req.Data); i += batchSize {
		end := i + batchSize
		if end > len(req.Data) {
			end = len(req.Data)
		}

		batch := req.Data[i:end]
		values := make([]string, 0, len(batch))
		for _, row := range batch {
			valueStrs := make([]string, len(row))
			for j, val := range row {
				if val == nil {
					valueStrs[j] = "NULL"
				} else {
					valStr := fmt.Sprintf("'%s'", strings.ReplaceAll(fmt.Sprintf("%v", val), "'", "''"))
					valueStrs[j] = valStr
				}
			}
			values = append(values, fmt.Sprintf("(%s)", strings.Join(valueStrs, ", ")))
		}

		columnsStr := strings.Join(columnNames, ", ")
		valuesStr := strings.Join(values, ", ")
		sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", req.Table, columnsStr, valuesStr)

		if req.OnDuplicate == "ignore" {
			sql = strings.Replace(sql, "INSERT INTO", "INSERT IGNORE INTO", 1)
		} else if req.OnDuplicate == "update" {
			// MySQL: ON DUPLICATE KEY UPDATE
			updateParts := make([]string, len(columnNames))
			for j, col := range columnNames {
				updateParts[j] = fmt.Sprintf("`%s`=VALUES(`%s`)", col, col)
			}
			sql += " ON DUPLICATE KEY UPDATE " + strings.Join(updateParts, ", ")
		}

		_, execTime, err := executor.ExecuteUpdate(sql)
		if err != nil {
			response.ErrorRows += len(batch)
			response.Errors = append(response.Errors, struct {
				Row   int    `json:"row"`
				Error string `json:"error"`
			}{Row: i + 1, Error: err.Error()})
		} else {
			response.InsertedRows += len(batch)
		}
		_ = execTime
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       "IMPORT DATA",
		OperationType:  "IMPORT_DATA",
		ExecutionTimeMs: executionTime,
		RowsAffected:   response.InsertedRows,
		Success:        response.ErrorRows == 0,
	}
	s.systemHistoryDAO.Create(history)

	return response, nil
}

// RunImportDataTask 异步导入任务
func (s *ImportService) RunImportDataTask(taskID int, userID int, req *models.ImportDataRequest, taskService *TaskService) {
	if err := taskService.Start(taskID); err != nil {
		return
	}
	response, err := s.ImportData(req.ConnectionID, userID, req)
	if err != nil {
		_ = taskService.CompleteFailure(taskID, err)
		return
	}
	_ = taskService.CompleteSuccess(taskID, response)
}

// ExecuteSQLFile 执行SQL文件
func (s *ImportService) ExecuteSQLFile(connectionID, userID int, req *models.ExecuteSQLFileRequest) (*models.ExecuteSQLFileResponse, error) {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	response := &models.ExecuteSQLFileResponse{}

	// 分割SQL语句
	statements := utils.SplitSQLStatements(req.SQL)
	response.ExecutedStatements = len(statements)

	for _, stmt := range statements {
		if dangerous, reason := utils.IsDangerousSQL(stmt); dangerous && !req.ConfirmDangerous {
			return nil, errors.New("危险 SQL 需确认: " + reason)
		}
	}

	if req.Transaction {
		// 在事务中执行
		if err := executor.BeginTransaction(); err != nil {
			return nil, fmt.Errorf("开始事务失败: %w", err)
		}
		defer func() {
			if response.ErrorCount > 0 {
				executor.RollbackTransaction()
			} else {
				executor.CommitTransaction()
			}
		}()
	}

	// 执行每条语句
	for i, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}

		_, _, err := executor.ExecuteUpdate(stmt)
		if err != nil {
			response.ErrorCount++
			response.Errors = append(response.Errors, struct {
				Statement int    `json:"statement"`
				Error     string `json:"error"`
			}{Statement: i + 1, Error: err.Error()})
		} else {
			response.SuccessCount++
		}
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       "EXECUTE SQL FILE",
		OperationType:  "EXECUTE_SQL_FILE",
		ExecutionTimeMs: executionTime,
		Success:        response.ErrorCount == 0,
	}
	if response.ErrorCount > 0 {
		history.ErrorMessage = fmt.Sprintf("%d errors", response.ErrorCount)
	}
	s.systemHistoryDAO.Create(history)

	return response, nil
}

// RunExecuteSQLFileTask 异步执行 SQL 文件任务
func (s *ImportService) RunExecuteSQLFileTask(taskID int, userID int, req *models.ExecuteSQLFileRequest, taskService *TaskService) {
	if err := taskService.Start(taskID); err != nil {
		return
	}
	response, err := s.ExecuteSQLFile(req.ConnectionID, userID, req)
	if err != nil {
		_ = taskService.CompleteFailure(taskID, err)
		return
	}
	_ = taskService.CompleteSuccess(taskID, response)
}
