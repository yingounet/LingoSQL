package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

// safeNumericRegex 用于校验 IN 子句中的数值（防止 SQL 注入）
var safeNumericRegex = regexp.MustCompile(`^-?[0-9]+\.?[0-9]*$`)

type RowDataService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewRowDataService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *RowDataService {
	return &RowDataService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

func (s *RowDataService) getExecutor(connectionID, userID int, database string) (db.Executor, *models.Connection, error) {
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

// escapeSQLString 转义 SQL 字符串字面量中的单引号和反斜杠
func escapeSQLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return strings.ReplaceAll(s, "'", "''")
}

// escapeINValue 安全处理 IN 子句中的值，防止 SQL 注入
func escapeINValue(part string, dbType string) string {
	part = strings.TrimSpace(part)
	if part == "" {
		return ""
	}
	if safeNumericRegex.MatchString(part) {
		return part
	}
	escaped := escapeSQLString(part)
	return "'" + escaped + "'"
}

func (s *RowDataService) buildWhereClause(filters []map[string]interface{}, dbType string) (string, error) {
	if len(filters) == 0 {
		return "", nil
	}
	var conditions []string
	for _, filter := range filters {
		field, _ := filter["field"].(string)
		operator, _ := filter["operator"].(string)
		value, _ := filter["value"].(string)

		if field == "" {
			continue
		}
		if err := utils.ValidateColumnName(field); err != nil {
			return "", err
		}

		var cond string
		if dbType == "postgresql" {
			field = fmt.Sprintf("\"%s\"", field)
		} else {
			field = fmt.Sprintf("`%s`", field)
		}

		switch operator {
		case "=", "!=", "<", ">", "<=", ">=":
			cond = fmt.Sprintf("%s %s '%s'", field, operator, escapeSQLString(value))
		case "LIKE":
			cond = fmt.Sprintf("%s LIKE '%%%s%%'", field, escapeSQLString(value))
		case "IN":
			parts := strings.Split(value, ",")
			var safeParts []string
			for _, p := range parts {
				if escaped := escapeINValue(p, dbType); escaped != "" {
					safeParts = append(safeParts, escaped)
				}
			}
			if len(safeParts) == 0 {
				continue
			}
			cond = fmt.Sprintf("%s IN (%s)", field, strings.Join(safeParts, ", "))
		case "IS NULL":
			cond = fmt.Sprintf("%s IS NULL", field)
		case "IS NOT NULL":
			cond = fmt.Sprintf("%s IS NOT NULL", field)
		default:
			cond = fmt.Sprintf("%s = '%s'", field, escapeSQLString(value))
		}
		conditions = append(conditions, cond)
	}
	if len(conditions) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(conditions, " AND "), nil
}

// BatchInsertData 批量插入数据
func (s *RowDataService) BatchInsertData(connectionID, userID int, req *models.BatchInsertRequest) (int, error) {
	if err := utils.ValidateTableName(req.Table); err != nil {
		return 0, err
	}
	for _, col := range req.Columns {
		if err := utils.ValidateColumnName(col); err != nil {
			return 0, err
		}
	}
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return 0, err
	}
	startTime := time.Now()

	// 构建批量INSERT语句
	batchSize := 100
	insertedRows := 0
	for i := 0; i < len(req.Data); i += batchSize {
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
					valStr := fmt.Sprintf("'%s'", escapeSQLString(fmt.Sprintf("%v", val)))
					valueStrs[j] = valStr
				}
			}
			values = append(values, fmt.Sprintf("(%s)", strings.Join(valueStrs, ", ")))
		}
		var columnsStr, tableQuoted string
		if conn.DBType == "postgresql" {
			tableQuoted = fmt.Sprintf("\"%s\"", req.Table)
			cols := make([]string, len(req.Columns))
			for j, c := range req.Columns {
				cols[j] = fmt.Sprintf("\"%s\"", c)
			}
			columnsStr = strings.Join(cols, ", ")
		} else {
			tableQuoted = fmt.Sprintf("`%s`", req.Table)
			cols := make([]string, len(req.Columns))
			for j, c := range req.Columns {
				cols[j] = fmt.Sprintf("`%s`", c)
			}
			columnsStr = strings.Join(cols, ", ")
		}
		valuesStr := strings.Join(values, ", ")
		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableQuoted, columnsStr, valuesStr)

		rowsAffected, _, err := executor.ExecuteUpdate(sql)
		if err != nil {
			return insertedRows, err
		}
		insertedRows += rowsAffected
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       "BATCH INSERT",
		OperationType:  "BATCH_INSERT",
		ExecutionTimeMs: executionTime,
		RowsAffected:   insertedRows,
		Success:        true,
	}
	s.systemHistoryDAO.Create(history)

	return insertedRows, nil
}

// BatchUpdateData 批量更新数据
func (s *RowDataService) BatchUpdateData(connectionID, userID int, req *models.BatchUpdateRequest) (int, error) {
	if err := utils.ValidateTableName(req.Table); err != nil {
		return 0, err
	}
	for k := range req.UpdateData {
		if err := utils.ValidateColumnName(k); err != nil {
			return 0, err
		}
	}
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return 0, err
	}
	startTime := time.Now()

	// 构建UPDATE语句
	var setParts []string
	for k, v := range req.UpdateData {
		if conn.DBType == "postgresql" {
			setParts = append(setParts, fmt.Sprintf("\"%s\" = '%s'", k, escapeSQLString(fmt.Sprintf("%v", v))))
		} else {
			setParts = append(setParts, fmt.Sprintf("`%s` = '%s'", k, escapeSQLString(fmt.Sprintf("%v", v))))
		}
	}
	setClause := strings.Join(setParts, ", ")
	whereClause, err := s.buildWhereClause(req.Filters, conn.DBType)
	if err != nil {
		return 0, err
	}
	if whereClause == "" {
		return 0, errors.New("批量更新必须提供筛选条件")
	}

	var updateSQL string
	if conn.DBType == "postgresql" {
		updateSQL = fmt.Sprintf("UPDATE \"%s\" SET %s %s", req.Table, setClause, whereClause)
	} else {
		updateSQL = fmt.Sprintf("UPDATE `%s` SET %s %s", req.Table, setClause, whereClause)
	}
	rowsAffected, _, err := executor.ExecuteUpdate(updateSQL)
	if err != nil {
		return 0, err
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       updateSQL,
		OperationType:  "BATCH_UPDATE",
		ExecutionTimeMs: executionTime,
		RowsAffected:   rowsAffected,
		Success:        true,
	}
	s.systemHistoryDAO.Create(history)

	return rowsAffected, nil
}

// BatchDeleteData 批量删除数据
func (s *RowDataService) BatchDeleteData(connectionID, userID int, req *models.BatchDeleteRequest) (int, error) {
	if err := utils.ValidateTableName(req.Table); err != nil {
		return 0, err
	}
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return 0, err
	}
	startTime := time.Now()

	whereClause, err := s.buildWhereClause(req.Filters, conn.DBType)
	if err != nil {
		return 0, err
	}
	if whereClause == "" {
		return 0, errors.New("批量删除必须提供筛选条件")
	}

	var deleteSQL string
	if conn.DBType == "postgresql" {
		deleteSQL = fmt.Sprintf("DELETE FROM \"%s\" %s", req.Table, whereClause)
	} else {
		deleteSQL = fmt.Sprintf("DELETE FROM `%s` %s", req.Table, whereClause)
	}
	rowsAffected, _, err := executor.ExecuteUpdate(deleteSQL)
	if err != nil {
		return 0, err
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       deleteSQL,
		OperationType:  "BATCH_DELETE",
		ExecutionTimeMs: executionTime,
		RowsAffected:   rowsAffected,
		Success:        true,
	}
	s.systemHistoryDAO.Create(history)

	return rowsAffected, nil
}

// CompareData 数据对比
func (s *RowDataService) CompareData(connectionID, userID int, req *models.CompareDataRequest) (*models.CompareDataResponse, error) {
	if err := utils.ValidateDatabaseName(req.Database1); err != nil {
		return nil, err
	}
	if err := utils.ValidateTableName(req.Table1); err != nil {
		return nil, err
	}
	if err := utils.ValidateTableName(req.Table2); err != nil {
		return nil, err
	}
	for _, col := range req.KeyColumns {
		if err := utils.ValidateColumnName(col); err != nil {
			return nil, err
		}
	}
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database1)
	if err != nil {
		return nil, err
	}

	db2 := req.Database2
	if db2 == "" {
		db2 = req.Database1
	} else if err := utils.ValidateDatabaseName(db2); err != nil {
		return nil, err
	}

	// 安全构建 ORDER BY 子句
	var keyCols []string
	for _, col := range req.KeyColumns {
		if conn.DBType == "postgresql" {
			keyCols = append(keyCols, fmt.Sprintf("\"%s\"", col))
		} else {
			keyCols = append(keyCols, fmt.Sprintf("`%s`", col))
		}
	}
	orderBy := strings.Join(keyCols, ", ")
	var sql1, sql2 string
	if conn.DBType == "postgresql" {
		sql1 = fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" ORDER BY %s", req.Database1, req.Table1, orderBy)
		sql2 = fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" ORDER BY %s", db2, req.Table2, orderBy)
	} else {
		sql1 = fmt.Sprintf("SELECT * FROM `%s`.`%s` ORDER BY %s", req.Database1, req.Table1, orderBy)
		sql2 = fmt.Sprintf("SELECT * FROM `%s`.`%s` ORDER BY %s", db2, req.Table2, orderBy)
	}

	cols1, rows1, _, err := executor.Execute(sql1)
	if err != nil {
		return nil, err
	}
	cols2, rows2, _, err := executor.Execute(sql2)
	if err != nil {
		return nil, err
	}

	response := &models.CompareDataResponse{
		OnlyInTable1: make([]map[string]interface{}, 0),
		OnlyInTable2: make([]map[string]interface{}, 0),
		Different:    make([]struct {
			Key        map[string]interface{} `json:"key"`
			Table1Data map[string]interface{} `json:"table1_data"`
			Table2Data map[string]interface{} `json:"table2_data"`
			Differences []string              `json:"differences"`
		}, 0),
		SameCount: 0,
	}

	// 简化的对比逻辑（实际需要更完善的实现）
	_ = cols1
	_ = cols2
	_ = rows1
	_ = rows2

	return response, nil
}

// FindReplaceData 查找替换数据
func (s *RowDataService) FindReplaceData(connectionID, userID int, req *models.FindReplaceRequest) (*models.FindReplaceResponse, error) {
	if err := utils.ValidateTableName(req.Table); err != nil {
		return nil, err
	}
	if err := utils.ValidateColumnName(req.Column); err != nil {
		return nil, err
	}
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()

	whereClause, err := s.buildWhereClause(req.Filters, conn.DBType)
	if err != nil {
		return nil, err
	}
	additionalWhere := ""
	if whereClause != "" {
		additionalWhere = " AND " + strings.TrimPrefix(whereClause, "WHERE ")
	}

	findEscaped := escapeSQLString(req.FindValue)
	replaceEscaped := escapeSQLString(req.ReplaceValue)

	// 先查询匹配的行数
	var countSQL string
	if conn.DBType == "postgresql" {
		countSQL = fmt.Sprintf("SELECT COUNT(*) FROM \"%s\" WHERE \"%s\" LIKE '%%%s%%' %s",
			req.Table, req.Column, findEscaped, additionalWhere)
	} else {
		countSQL = fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s` LIKE '%%%s%%' %s",
			req.Table, req.Column, findEscaped, additionalWhere)
	}
	_, countRows, _, err := executor.Execute(countSQL)
	var matchedRows int
	if err == nil && len(countRows) > 0 && len(countRows[0]) > 0 {
		if val, ok := countRows[0][0].(int64); ok {
			matchedRows = int(val)
		}
	}

	// 执行替换
	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("UPDATE \"%s\" SET \"%s\" = REPLACE(\"%s\", '%s', '%s') WHERE \"%s\" LIKE '%%%s%%' %s",
			req.Table, req.Column, req.Column,
			findEscaped, replaceEscaped,
			req.Column, findEscaped, additionalWhere)
	} else {
		sql = fmt.Sprintf("UPDATE `%s` SET `%s` = REPLACE(`%s`, '%s', '%s') WHERE `%s` LIKE '%%%s%%' %s",
			req.Table, req.Column, req.Column,
			findEscaped, replaceEscaped,
			req.Column, findEscaped, additionalWhere)
	}

	rowsAffected, _, err := executor.ExecuteUpdate(sql)
	if err != nil {
		return nil, err
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "FIND_REPLACE",
		ExecutionTimeMs: executionTime,
		RowsAffected:   rowsAffected,
		Success:        true,
	}
	s.systemHistoryDAO.Create(history)

	return &models.FindReplaceResponse{
		AffectedRows: rowsAffected,
		MatchedRows:  matchedRows,
	}, nil
}
