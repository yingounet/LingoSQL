package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

type QueryService struct {
	connectionDAO *sqlite.ConnectionDAO
	historyDAO    *sqlite.HistoryDAO
}

func NewQueryService(connectionDAO *sqlite.ConnectionDAO, historyDAO *sqlite.HistoryDAO) *QueryService {
	return &QueryService{
		connectionDAO: connectionDAO,
		historyDAO:    historyDAO,
	}
}

// Execute 执行 SQL 查询
// 支持多条语句（用分号分隔），自动忽略注释
func (s *QueryService) Execute(userID int, req *models.QueryExecuteRequest) (*models.QueryExecuteResponse, int, error) {
	// 获取连接
	conn, err := s.connectionDAO.GetByID(req.ConnectionID)
	if err != nil {
		return nil, 0, fmt.Errorf("获取连接: %w", err)
	}

	if conn.UserID != userID {
		return nil, 0, errors.New("无权访问此连接")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return nil, 0, errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, 0, err
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		req.ConnectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		req.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("获取执行器: %w", err)
	}

	// 预处理 SQL：移除注释
	cleanSQL := utils.RemoveSQLComments(req.SQL)

	// 分割多条语句
	statements := utils.SplitSQLStatements(cleanSQL)

	if len(statements) == 0 {
		return nil, 0, errors.New("没有可执行的 SQL 语句")
	}

	for _, stmt := range statements {
		if dangerous, reason := utils.IsDangerousSQL(stmt); dangerous && !req.ConfirmDangerous {
			return nil, 0, errors.New("危险 SQL 需确认: " + reason)
		}
	}

	var response *models.QueryExecuteResponse
	var queryID int
	var totalRowsAffected int
	var totalExecTime int
	var lastQueryColumns []string
	var lastQueryRows [][]interface{}
	var hasQueryResult bool

	// 依次执行每条语句
	for i, stmt := range statements {
		isQuery := utils.IsSQLQuery(stmt)

		if isQuery {
			// 查询操作
			columns, rows, execTime, err := executor.Execute(stmt)
			if err != nil {
				// 记录失败历史
				history := &models.QueryHistory{
					UserID:          userID,
					ConnectionID:    req.ConnectionID,
					SQLQuery:        req.SQL,
					ExecutionTimeMs: totalExecTime + execTime,
					Success:         false,
					ErrorMessage:    formatErrorWithStatement(err, i+1, len(statements), stmt),
				}
				s.historyDAO.Create(history)
				queryID = history.ID

				return nil, queryID, errors.New(formatErrorWithStatement(err, i+1, len(statements), stmt))
			}

			totalExecTime += execTime
			totalRowsAffected += len(rows)
			// 保存最后一条查询的结果
			lastQueryColumns = columns
			lastQueryRows = rows
			hasQueryResult = true
		} else {
			// 更新操作
			rowsAffected, execTime, err := executor.ExecuteUpdate(stmt)
			if err != nil {
				// 记录失败历史
				history := &models.QueryHistory{
					UserID:          userID,
					ConnectionID:    req.ConnectionID,
					SQLQuery:        req.SQL,
					ExecutionTimeMs: totalExecTime + execTime,
					Success:         false,
					ErrorMessage:    formatErrorWithStatement(err, i+1, len(statements), stmt),
				}
				s.historyDAO.Create(history)
				queryID = history.ID

				return nil, queryID, errors.New(formatErrorWithStatement(err, i+1, len(statements), stmt))
			}

			totalExecTime += execTime
			totalRowsAffected += rowsAffected
		}
	}

	// 构建响应
	if hasQueryResult {
		// 有查询结果，返回最后一条查询的结果
		response = &models.QueryExecuteResponse{
			Columns:            lastQueryColumns,
			Rows:               lastQueryRows,
			ExecutionTimeMs:    totalExecTime,
			RowsAffected:       totalRowsAffected,
			StatementsExecuted: len(statements),
		}
	} else {
		// 全是更新操作，返回影响行数
		response = &models.QueryExecuteResponse{
			Columns:            []string{},
			Rows:               [][]interface{}{},
			ExecutionTimeMs:    totalExecTime,
			RowsAffected:       totalRowsAffected,
			StatementsExecuted: len(statements),
		}
	}

	// 记录成功历史
	history := &models.QueryHistory{
		UserID:          userID,
		ConnectionID:    req.ConnectionID,
		SQLQuery:        req.SQL,
		ExecutionTimeMs: response.ExecutionTimeMs,
		RowsAffected:    response.RowsAffected,
		Success:         true,
	}
	if err := s.historyDAO.Create(history); err == nil {
		queryID = history.ID
		response.QueryID = queryID
	}

	return response, queryID, nil
}

// Explain 执行 SQL 执行计划分析
func (s *QueryService) Explain(userID int, req *models.ExplainRequest) (*models.ExplainResponse, error) {
	// 获取连接
	conn, err := s.connectionDAO.GetByID(req.ConnectionID)
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

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		req.ConnectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		req.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, fmt.Errorf("获取执行器: %w", err)
	}

	// 预处理 SQL：移除注释
	cleanSQL := utils.RemoveSQLComments(req.SQL)
	cleanSQL = strings.TrimSpace(cleanSQL)

	if cleanSQL == "" {
		return nil, errors.New("SQL 语句为空")
	}

	// 执行 EXPLAIN
	plan, execTime, err := executor.Explain(cleanSQL)
	if err != nil {
		return nil, fmt.Errorf("执行计划分析失败: %w", err)
	}

	return &models.ExplainResponse{
		Plan:            plan,
		ExecutionTimeMs: execTime,
	}, nil
}

// BeginTransaction 开始事务
func (s *QueryService) BeginTransaction(userID int, req *models.TransactionRequest) error {
	// 获取连接
	conn, err := s.connectionDAO.GetByID(req.ConnectionID)
	if err != nil {
		return fmt.Errorf("获取连接: %w", err)
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
		req.ConnectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		req.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return fmt.Errorf("获取执行器: %w", err)
	}

	return executor.BeginTransaction()
}

// CommitTransaction 提交事务
func (s *QueryService) CommitTransaction(userID int, req *models.TransactionRequest) error {
	// 获取连接
	conn, err := s.connectionDAO.GetByID(req.ConnectionID)
	if err != nil {
		return fmt.Errorf("获取连接: %w", err)
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
		req.ConnectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		req.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return fmt.Errorf("获取执行器: %w", err)
	}

	return executor.CommitTransaction()
}

// RollbackTransaction 回滚事务
func (s *QueryService) RollbackTransaction(userID int, req *models.TransactionRequest) error {
	// 获取连接
	conn, err := s.connectionDAO.GetByID(req.ConnectionID)
	if err != nil {
		return fmt.Errorf("获取连接: %w", err)
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
		req.ConnectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		req.Database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return fmt.Errorf("获取执行器: %w", err)
	}

	return executor.RollbackTransaction()
}

// formatErrorWithStatement 格式化错误信息，包含语句位置信息
func formatErrorWithStatement(err error, stmtIndex, totalStmts int, stmt string) string {
	if totalStmts == 1 {
		return err.Error()
	}

	// 截取语句的前50个字符作为预览
	preview := stmt
	if len(preview) > 50 {
		preview = preview[:50] + "..."
	}

	return err.Error() + " (第 " + strconv.Itoa(stmtIndex) + "/" + strconv.Itoa(totalStmts) + " 条语句: " + preview + ")"
}
