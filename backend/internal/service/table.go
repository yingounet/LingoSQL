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

type TableService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewTableService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *TableService {
	return &TableService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

// GetTables 获取表列表
func (s *TableService) GetTables(connectionID, userID int, database string) ([]map[string]interface{}, error) {
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

	// 构建 SQL 语句用于记录
	var sqlQuery string
	if conn.DBType == "postgresql" {
		sqlQuery = `SELECT table_name, (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = t.table_name) as column_count FROM information_schema.tables t WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name`
	} else {
		sqlQuery = fmt.Sprintf("SELECT TABLE_NAME, ENGINE, TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH FROM information_schema.TABLES WHERE TABLE_SCHEMA = '%s'", database)
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}
	// 不要关闭连接，连接池会管理连接的生命周期

	// 记录开始时间
	startTime := time.Now()

	// 执行查询
	tables, err := executor.GetTables(database)

	// 计算执行时间
	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         sqlQuery,
		OperationType:    "GET_TABLES",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     len(tables),
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

	return tables, nil
}

// GetTableInfo 获取表详细信息
func (s *TableService) GetTableInfo(connectionID, userID int, database, table string) (map[string]interface{}, error) {
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
		database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}
	// 不要关闭连接，连接池会管理连接的生命周期

	return executor.GetTableInfo(database, table)
}

// GetTableColumns 获取表字段列表
func (s *TableService) GetTableColumns(connectionID, userID int, database, table string) ([]map[string]interface{}, error) {
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

	// 构建 SQL 语句用于记录
	var sqlQuery string
	if conn.DBType == "postgresql" {
		sqlQuery = fmt.Sprintf(`SELECT a.attname AS column_name, format_type(a.atttypid, a.atttypmod) AS data_type FROM pg_attribute a WHERE a.attrelid = 'public.%s'::regclass AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum`, table)
	} else {
		sqlQuery = fmt.Sprintf("SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT, COLUMN_KEY FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s' ORDER BY ORDINAL_POSITION", database, table)
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}
	// 不要关闭连接，连接池会管理连接的生命周期

	// 记录开始时间
	startTime := time.Now()

	// 执行查询
	columns, err := executor.GetTableColumns(database, table)

	// 计算执行时间
	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         sqlQuery,
		OperationType:    "GET_TABLE_COLUMNS",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     len(columns),
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

	return columns, nil
}

// GetTableIndexes 获取表索引列表
func (s *TableService) GetTableIndexes(connectionID, userID int, database, table string) ([]map[string]interface{}, error) {
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

	// 构建 SQL 语句用于记录
	var sqlQuery string
	if conn.DBType == "postgresql" {
		sqlQuery = fmt.Sprintf(`SELECT indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' AND tablename = '%s'`, table)
	} else {
		sqlQuery = fmt.Sprintf("SHOW INDEXES FROM `%s`.`%s`", database, table)
	}

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}
	// 不要关闭连接，连接池会管理连接的生命周期

	// 记录开始时间
	startTime := time.Now()

	// 执行查询
	indexes, err := executor.GetTableIndexes(database, table)

	// 计算执行时间
	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         sqlQuery,
		OperationType:    "GET_TABLE_INDEXES",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     len(indexes),
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

	return indexes, nil
}

// GetTableRows 获取表数据
func (s *TableService) GetTableRows(connectionID, userID int, database, table string, filters []db.RowFilter, page, pageSize int) (*db.TableRowsResult, error) {
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

	// 构建 SQL 语句用于记录（简化版本，实际 SQL 由 executor 内部构建）
	sqlQuery := s.buildSelectSQL(conn.DBType, database, table, filters, page, pageSize)

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return nil, err
	}
	// 不要关闭连接，连接池会管理连接的生命周期

	// 记录开始时间
	startTime := time.Now()

	// 执行查询
	result, err := executor.GetTableRows(database, table, filters, page, pageSize)

	// 计算执行时间
	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	rowsAffected := 0
	if result != nil {
		rowsAffected = len(result.Rows)
	}
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         sqlQuery,
		OperationType:    "GET_TABLE_ROWS",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     rowsAffected,
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

	return result, nil
}

// UpdateTableRow 更新表数据
func (s *TableService) UpdateTableRow(connectionID, userID int, database, table string, primaryKey map[string]interface{}, data map[string]interface{}) (int, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return 0, err
	}

	if conn.UserID != userID {
		return 0, errors.New("无权访问此连接")
	}

	// 解析数据库配置
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return 0, errors.New("配置解析失败")
	}

	// 解密密码
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return 0, err
	}

	// 构建 SQL 语句用于记录
	sqlQuery := s.buildUpdateSQL(conn.DBType, database, table, primaryKey, data)

	// 获取执行器
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		database, dbConfig.Username, password, dbConfig.Options,
	)
	if err != nil {
		return 0, err
	}
	// 不要关闭连接，连接池会管理连接的生命周期

	// 记录开始时间
	startTime := time.Now()

	// 执行更新
	rowsAffected, err := executor.UpdateTableRow(database, table, primaryKey, data)

	// 计算执行时间
	executionTime := int(time.Since(startTime).Milliseconds())

	// 记录系统执行历史
	history := &models.SystemQueryHistory{
		ConnectionID:     connectionID,
		UserID:           userID,
		SQLQuery:         sqlQuery,
		OperationType:    "UPDATE_TABLE_ROW",
		ExecutionTimeMs:  executionTime,
		RowsAffected:     rowsAffected,
		Success:          err == nil,
		ErrorMessage:     "",
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// buildSelectSQL 构建 SELECT SQL 语句（用于记录）
func (s *TableService) buildSelectSQL(dbType, database, table string, filters []db.RowFilter, page, pageSize int) string {
	var tableName string
	if dbType == "postgresql" {
		tableName = fmt.Sprintf(`"public"."%s"`, table)
	} else {
		tableName = fmt.Sprintf("`%s`.`%s`", database, table)
	}

	whereClause := ""
	if len(filters) > 0 {
		var conditions []string
		for _, f := range filters {
			var fieldName string
			if dbType == "postgresql" {
				fieldName = fmt.Sprintf(`"%s"`, f.Field)
			} else {
				fieldName = fmt.Sprintf("`%s`", f.Field)
			}

			switch f.Operator {
			case "=", "!=", "<", "<=", ">", ">=":
				conditions = append(conditions, fmt.Sprintf("%s %s ?", fieldName, f.Operator))
			case "LIKE", "NOT LIKE":
				conditions = append(conditions, fmt.Sprintf("%s %s ?", fieldName, f.Operator))
			case "IS NULL":
				conditions = append(conditions, fmt.Sprintf("%s IS NULL", fieldName))
			case "IS NOT NULL":
				conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", fieldName))
			}
		}
		if len(conditions) > 0 {
			whereClause = " WHERE " + strings.Join(conditions, " AND ")
		}
	}

	offset := (page - 1) * pageSize
	return fmt.Sprintf("SELECT * FROM %s%s LIMIT %d OFFSET %d", tableName, whereClause, pageSize, offset)
}

// buildUpdateSQL 构建 UPDATE SQL 语句（用于记录）
func (s *TableService) buildUpdateSQL(dbType, database, table string, primaryKey, data map[string]interface{}) string {
	var tableName string
	if dbType == "postgresql" {
		tableName = fmt.Sprintf(`"public"."%s"`, table)
	} else {
		tableName = fmt.Sprintf("`%s`.`%s`", database, table)
	}

	var setClauses []string
	for field := range data {
		if dbType == "postgresql" {
			setClauses = append(setClauses, fmt.Sprintf(`"%s" = ?`, field))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", field))
		}
	}

	var whereClauses []string
	for field := range primaryKey {
		if dbType == "postgresql" {
			whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = ?`, field))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("`%s` = ?", field))
		}
	}

	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", tableName, strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))
}
