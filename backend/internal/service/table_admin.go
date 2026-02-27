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

type TableAdminService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewTableAdminService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *TableAdminService {
	return &TableAdminService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

func (s *TableAdminService) getExecutor(connectionID, userID int, database string) (db.Executor, *models.Connection, error) {
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

// CreateTable 在指定数据库中建表
func (s *TableAdminService) CreateTable(connectionID, userID int, req *models.CreateTableRequest) error {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()
	err = executor.CreateTable(req.Database, req.TableName, req.CreateDDL)
	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       "CREATE TABLE",
		OperationType:  "CREATE_TABLE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// DropTable 在指定数据库中删表
func (s *TableAdminService) DropTable(connectionID, userID int, database, tableName string) error {
	executor, _, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	startTime := time.Now()
	err = executor.DropTable(database, tableName)
	executionTime := int(time.Since(startTime).Milliseconds())
	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       "DROP TABLE",
		OperationType:  "DROP_TABLE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// AlterTable 修改表结构
func (s *TableAdminService) AlterTable(connectionID, userID int, req *models.AlterTableRequest) error {
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	// 构建ALTER TABLE语句
	var sqlParts []string
	for _, op := range req.Operations {
		switch op.Type {
		case "add_column":
			if col := op.Column; col != nil {
				colDef := s.buildColumnDefinition(col, conn.DBType)
				sqlParts = append(sqlParts, fmt.Sprintf("ADD COLUMN %s", colDef))
			}
		case "drop_column":
			if op.OldColumnName != "" {
				sqlParts = append(sqlParts, fmt.Sprintf("DROP COLUMN `%s`", op.OldColumnName))
			}
		case "modify_column":
			if col := op.Column; col != nil {
				colDef := s.buildColumnDefinition(col, conn.DBType)
				sqlParts = append(sqlParts, fmt.Sprintf("MODIFY COLUMN %s", colDef))
			}
		case "rename_column":
			if op.OldColumnName != "" && op.NewColumnName != "" {
				if conn.DBType == "postgresql" {
					sqlParts = append(sqlParts, fmt.Sprintf("RENAME COLUMN \"%s\" TO \"%s\"", op.OldColumnName, op.NewColumnName))
				} else {
					sqlParts = append(sqlParts, fmt.Sprintf("CHANGE COLUMN `%s` `%s`", op.OldColumnName, op.NewColumnName))
				}
			}
		}
	}

	if len(sqlParts) == 0 {
		return errors.New("没有有效的修改操作")
	}

	sql := fmt.Sprintf("ALTER TABLE `%s` %s", req.Table, strings.Join(sqlParts, ", "))
	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "ALTER_TABLE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// CreateIndex 创建索引
func (s *TableAdminService) CreateIndex(connectionID, userID int, req *models.CreateIndexRequest) error {
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	idx := req.Index
	name, _ := idx["name"].(string)
	idxType, _ := idx["type"].(string)
	columns, _ := idx["columns"].([]interface{})

	if name == "" || len(columns) == 0 {
		return errors.New("索引名称和列不能为空")
	}

	colStrs := make([]string, len(columns))
	for i, col := range columns {
		colStrs[i] = fmt.Sprintf("`%v`", col)
	}

	var sql string
	if conn.DBType == "postgresql" {
		unique := ""
		if idxType == "UNIQUE" {
			unique = "UNIQUE "
		}
		sql = fmt.Sprintf("CREATE %sINDEX \"%s\" ON \"%s\" (%s)", unique, name, req.Table, strings.Join(colStrs, ", "))
	} else {
		unique := ""
		if idxType == "UNIQUE" {
			unique = "UNIQUE "
		}
		sql = fmt.Sprintf("CREATE %sINDEX `%s` ON `%s` (%s)", unique, name, req.Table, strings.Join(colStrs, ", "))
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
		OperationType:  "CREATE_INDEX",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// DropIndex 删除索引
func (s *TableAdminService) DropIndex(connectionID, userID int, req *models.DropIndexRequest) error {
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("DROP INDEX \"%s\"", req.IndexName)
	} else {
		sql = fmt.Sprintf("DROP INDEX `%s` ON `%s`", req.IndexName, req.Table)
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
		OperationType:  "DROP_INDEX",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// RenameTable 重命名表
func (s *TableAdminService) RenameTable(connectionID, userID int, req *models.RenameTableRequest) error {
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\"", req.OldName, req.NewName)
	} else {
		sql = fmt.Sprintf("RENAME TABLE `%s` TO `%s`", req.OldName, req.NewName)
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
		OperationType:  "RENAME_TABLE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

func (s *TableAdminService) buildColumnDefinition(col map[string]interface{}, dbType string) string {
	name, _ := col["name"].(string)
	colType, _ := col["type"].(string)
	nullable, _ := col["nullable"].(bool)
	defaultVal, _ := col["default_value"].(string)
	isPrimary, _ := col["is_primary"].(bool)
	autoIncrement, _ := col["auto_increment"].(bool)

	var parts []string
	if dbType == "postgresql" {
		parts = append(parts, fmt.Sprintf("\"%s\" %s", name, colType))
		if !nullable {
			parts = append(parts, "NOT NULL")
		}
		if defaultVal != "" {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", defaultVal))
		}
		if isPrimary {
			parts = append(parts, "PRIMARY KEY")
		}
	} else {
		parts = append(parts, fmt.Sprintf("`%s` %s", name, colType))
		if !nullable {
			parts = append(parts, "NOT NULL")
		}
		if defaultVal != "" {
			parts = append(parts, fmt.Sprintf("DEFAULT '%s'", defaultVal))
		}
		if autoIncrement {
			parts = append(parts, "AUTO_INCREMENT")
		}
		if isPrimary {
			parts = append(parts, "PRIMARY KEY")
		}
	}
	return strings.Join(parts, " ")
}
