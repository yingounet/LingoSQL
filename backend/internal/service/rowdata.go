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
		database, dbConfig.Username, password,
	)
	if err != nil {
		return nil, nil, err
	}
	return executor, conn, nil
}

func (s *RowDataService) buildWhereClause(filters []map[string]interface{}, dbType string) string {
	if len(filters) == 0 {
		return ""
	}
	var conditions []string
	for _, filter := range filters {
		field, _ := filter["field"].(string)
		operator, _ := filter["operator"].(string)
		value, _ := filter["value"].(string)

		if field == "" {
			continue
		}

		var cond string
		if dbType == "postgresql" {
			field = fmt.Sprintf("\"%s\"", field)
		} else {
			field = fmt.Sprintf("`%s`", field)
		}

		switch operator {
		case "=", "!=", "<", ">", "<=", ">=":
			cond = fmt.Sprintf("%s %s '%s'", field, operator, strings.ReplaceAll(value, "'", "''"))
		case "LIKE":
			cond = fmt.Sprintf("%s LIKE '%%%s%%'", field, strings.ReplaceAll(value, "'", "''"))
		case "IN":
			cond = fmt.Sprintf("%s IN (%s)", field, value)
		case "IS NULL":
			cond = fmt.Sprintf("%s IS NULL", field)
		case "IS NOT NULL":
			cond = fmt.Sprintf("%s IS NOT NULL", field)
		default:
			cond = fmt.Sprintf("%s = '%s'", field, strings.ReplaceAll(value, "'", "''"))
		}
		conditions = append(conditions, cond)
	}
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

// BatchInsertData 批量插入数据
func (s *RowDataService) BatchInsertData(connectionID, userID int, req *models.BatchInsertRequest) (int, error) {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
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
					valStr := fmt.Sprintf("'%s'", strings.ReplaceAll(fmt.Sprintf("%v", val), "'", "''"))
					valueStrs[j] = valStr
				}
			}
			values = append(values, fmt.Sprintf("(%s)", strings.Join(valueStrs, ", ")))
		}
		columnsStr := strings.Join(req.Columns, ", ")
		valuesStr := strings.Join(values, ", ")
		sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", req.Table, columnsStr, valuesStr)

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
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return 0, err
	}
	startTime := time.Now()

	// 构建UPDATE语句
	var setParts []string
	for k, v := range req.UpdateData {
		if conn.DBType == "postgresql" {
			setParts = append(setParts, fmt.Sprintf("\"%s\" = '%s'", k, strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")))
		} else {
			setParts = append(setParts, fmt.Sprintf("`%s` = '%s'", k, strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")))
		}
	}
	setClause := strings.Join(setParts, ", ")
	whereClause := s.buildWhereClause(req.Filters, conn.DBType)

	sql := fmt.Sprintf("UPDATE `%s` SET %s %s", req.Table, setClause, whereClause)
	rowsAffected, _, err := executor.ExecuteUpdate(sql)
	if err != nil {
		return 0, err
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
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
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return 0, err
	}
	startTime := time.Now()

	whereClause := s.buildWhereClause(req.Filters, conn.DBType)
	if whereClause == "" {
		return 0, errors.New("批量删除必须提供筛选条件")
	}

	sql := fmt.Sprintf("DELETE FROM `%s` %s", req.Table, whereClause)
	rowsAffected, _, err := executor.ExecuteUpdate(sql)
	if err != nil {
		return 0, err
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
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
	// 简化实现：通过SQL查询对比数据
	executor, _, err := s.getExecutor(connectionID, userID, req.Database1)
	if err != nil {
		return nil, err
	}

	db2 := req.Database2
	if db2 == "" {
		db2 = req.Database1
	}

	// 构建对比SQL（简化版，实际需要更复杂的逻辑）
	keyCols := strings.Join(req.KeyColumns, ", ")
	sql1 := fmt.Sprintf("SELECT * FROM `%s`.`%s` ORDER BY %s", req.Database1, req.Table1, keyCols)
	sql2 := fmt.Sprintf("SELECT * FROM `%s`.`%s` ORDER BY %s", db2, req.Table2, keyCols)

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
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()

	whereClause := s.buildWhereClause(req.Filters, conn.DBType)
	additionalWhere := ""
	if whereClause != "" {
		additionalWhere = " AND " + strings.TrimPrefix(whereClause, "WHERE ")
	}

	// 先查询匹配的行数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s` LIKE '%%%s%%' %s",
		req.Table, req.Column, strings.ReplaceAll(req.FindValue, "'", "''"), additionalWhere)
	_, countRows, _, err := executor.Execute(countSQL)
	matchedRows := 0
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
			strings.ReplaceAll(req.FindValue, "'", "''"),
			strings.ReplaceAll(req.ReplaceValue, "'", "''"),
			req.Column, strings.ReplaceAll(req.FindValue, "'", "''"), additionalWhere)
	} else {
		sql = fmt.Sprintf("UPDATE `%s` SET `%s` = REPLACE(`%s`, '%s', '%s') WHERE `%s` LIKE '%%%s%%' %s",
			req.Table, req.Column, req.Column,
			strings.ReplaceAll(req.FindValue, "'", "''"),
			strings.ReplaceAll(req.ReplaceValue, "'", "''"),
			req.Column, strings.ReplaceAll(req.FindValue, "'", "''"), additionalWhere)
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
