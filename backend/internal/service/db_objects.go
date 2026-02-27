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

type DbObjectsService struct {
	connectionDAO    *sqlite.ConnectionDAO
	systemHistoryDAO *sqlite.SystemHistoryDAO
}

func NewDbObjectsService(connectionDAO *sqlite.ConnectionDAO, systemHistoryDAO *sqlite.SystemHistoryDAO) *DbObjectsService {
	return &DbObjectsService{
		connectionDAO:    connectionDAO,
		systemHistoryDAO: systemHistoryDAO,
	}
}

func (s *DbObjectsService) getExecutor(connectionID, userID int, database string) (db.Executor, *models.Connection, error) {
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

// GetViews 获取视图列表
func (s *DbObjectsService) GetViews(connectionID, userID int, database string) ([]models.ViewInfo, error) {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return nil, err
	}

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("SELECT table_name, view_definition FROM information_schema.views WHERE table_schema = '%s'", database)
	} else {
		sql = fmt.Sprintf("SELECT TABLE_NAME, VIEW_DEFINITION FROM information_schema.VIEWS WHERE TABLE_SCHEMA = '%s'", database)
	}

	cols, rows, _, err := executor.Execute(sql)
	if err != nil {
		return nil, err
	}

	views := make([]models.ViewInfo, 0, len(rows))
	for _, row := range rows {
		view := models.ViewInfo{}
		for i := range cols {
			if i == 0 {
				view.Name = fmt.Sprintf("%v", row[i])
			} else if i == 1 {
				view.Definition = fmt.Sprintf("%v", row[i])
			}
		}
		views = append(views, view)
	}
	return views, nil
}

// CreateView 创建视图
func (s *DbObjectsService) CreateView(connectionID, userID int, req *models.CreateViewRequest) error {
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("CREATE VIEW \"%s\".\"%s\" AS %s", req.Database, req.Name, req.Definition)
	} else {
		sql = fmt.Sprintf("CREATE VIEW `%s`.`%s` AS %s", req.Database, req.Name, req.Definition)
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
		OperationType:  "CREATE_VIEW",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// DropView 删除视图
func (s *DbObjectsService) DropView(connectionID, userID int, database, name string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("DROP VIEW \"%s\".\"%s\"", database, name)
	} else {
		sql = fmt.Sprintf("DROP VIEW `%s`.`%s`", database, name)
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
		OperationType:  "DROP_VIEW",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// GetProcedures 获取存储过程列表
func (s *DbObjectsService) GetProcedures(connectionID, userID int, database string) ([]models.ProcedureInfo, error) {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return nil, err
	}

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("SELECT routine_name, routine_definition FROM information_schema.routines WHERE routine_schema = '%s' AND routine_type = 'PROCEDURE'", database)
	} else {
		sql = fmt.Sprintf("SELECT ROUTINE_NAME, ROUTINE_DEFINITION FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = '%s' AND ROUTINE_TYPE = 'PROCEDURE'", database)
	}

	cols, rows, _, err := executor.Execute(sql)
	if err != nil {
		return nil, err
	}

	procedures := make([]models.ProcedureInfo, 0, len(rows))
	for _, row := range rows {
		proc := models.ProcedureInfo{}
		for i := range cols {
			if i == 0 {
				proc.Name = fmt.Sprintf("%v", row[i])
			} else if i == 1 {
				proc.Definition = fmt.Sprintf("%v", row[i])
			}
		}
		procedures = append(procedures, proc)
	}
	return procedures, nil
}

// CreateProcedure 创建存储过程
func (s *DbObjectsService) CreateProcedure(connectionID, userID int, req *models.CreateProcedureRequest) error {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	sql := req.Definition // 存储过程的定义应该包含完整的CREATE PROCEDURE语句

	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "CREATE_PROCEDURE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// DropProcedure 删除存储过程
func (s *DbObjectsService) DropProcedure(connectionID, userID int, database, name string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("DROP PROCEDURE \"%s\".\"%s\"", database, name)
	} else {
		sql = fmt.Sprintf("DROP PROCEDURE `%s`.`%s`", database, name)
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
		OperationType:  "DROP_PROCEDURE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// ExecuteProcedure 执行存储过程
func (s *DbObjectsService) ExecuteProcedure(connectionID, userID int, database, name string, parameters []interface{}) (interface{}, error) {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()

	// 构建CALL语句
	paramStr := ""
	if len(parameters) > 0 {
		paramParts := make([]string, len(parameters))
		for i, p := range parameters {
			paramParts[i] = fmt.Sprintf("'%v'", p)
		}
		paramStr = strings.Join(paramParts, ", ")
	}

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("CALL \"%s\".\"%s\"(%s)", database, name, paramStr)
	} else {
		sql = fmt.Sprintf("CALL `%s`.`%s`(%s)", database, name, paramStr)
	}

	cols, rows, _, err := executor.Execute(sql)
	executionTime := int(time.Since(startTime).Milliseconds())

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "EXECUTE_PROCEDURE",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)

	if err != nil {
		return nil, err
	}

	// 返回结果
	if len(rows) > 0 {
		result := make(map[string]interface{})
		for i, col := range cols {
			if len(rows[0]) > i {
				result[col] = rows[0][i]
			}
		}
		return result, nil
	}
	return nil, nil
}

// GetFunctions 获取函数列表
func (s *DbObjectsService) GetFunctions(connectionID, userID int, database string) ([]models.FunctionInfo, error) {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return nil, err
	}

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("SELECT routine_name, routine_definition, data_type FROM information_schema.routines WHERE routine_schema = '%s' AND routine_type = 'FUNCTION'", database)
	} else {
		sql = fmt.Sprintf("SELECT ROUTINE_NAME, ROUTINE_DEFINITION, DATA_TYPE FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = '%s' AND ROUTINE_TYPE = 'FUNCTION'", database)
	}

	cols, rows, _, err := executor.Execute(sql)
	if err != nil {
		return nil, err
	}

	functions := make([]models.FunctionInfo, 0, len(rows))
	for _, row := range rows {
		fn := models.FunctionInfo{}
		for i, col := range cols {
			switch col {
			case "routine_name", "ROUTINE_NAME":
				fn.Name = fmt.Sprintf("%v", row[i])
			case "routine_definition", "ROUTINE_DEFINITION":
				fn.Definition = fmt.Sprintf("%v", row[i])
			case "data_type", "DATA_TYPE":
				fn.ReturnType = fmt.Sprintf("%v", row[i])
			}
		}
		functions = append(functions, fn)
	}
	return functions, nil
}

// CreateFunction 创建函数
func (s *DbObjectsService) CreateFunction(connectionID, userID int, req *models.CreateFunctionRequest) error {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	sql := req.Definition // 函数的定义应该包含完整的CREATE FUNCTION语句

	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "CREATE_FUNCTION",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// DropFunction 删除函数
func (s *DbObjectsService) DropFunction(connectionID, userID int, database, name string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("DROP FUNCTION \"%s\".\"%s\"", database, name)
	} else {
		sql = fmt.Sprintf("DROP FUNCTION `%s`.`%s`", database, name)
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
		OperationType:  "DROP_FUNCTION",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// GetTriggers 获取触发器列表
func (s *DbObjectsService) GetTriggers(connectionID, userID int, database string) ([]models.TriggerInfo, error) {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return nil, err
	}

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("SELECT trigger_name, event_manipulation, event_object_table, action_timing, action_statement FROM information_schema.triggers WHERE trigger_schema = '%s'", database)
	} else {
		sql = fmt.Sprintf("SELECT TRIGGER_NAME, EVENT_MANIPULATION, EVENT_OBJECT_TABLE, ACTION_TIMING, ACTION_STATEMENT FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA = '%s'", database)
	}

	cols, rows, _, err := executor.Execute(sql)
	if err != nil {
		return nil, err
	}

	triggers := make([]models.TriggerInfo, 0, len(rows))
	for _, row := range rows {
		trigger := models.TriggerInfo{}
		for i, col := range cols {
			switch col {
			case "trigger_name", "TRIGGER_NAME":
				trigger.Name = fmt.Sprintf("%v", row[i])
			case "event_manipulation", "EVENT_MANIPULATION":
				trigger.Event = fmt.Sprintf("%v", row[i])
			case "event_object_table", "EVENT_OBJECT_TABLE":
				trigger.Table = fmt.Sprintf("%v", row[i])
			case "action_timing", "ACTION_TIMING":
				trigger.Timing = fmt.Sprintf("%v", row[i])
			case "action_statement", "ACTION_STATEMENT":
				trigger.Definition = fmt.Sprintf("%v", row[i])
			}
		}
		triggers = append(triggers, trigger)
	}
	return triggers, nil
}

// CreateTrigger 创建触发器
func (s *DbObjectsService) CreateTrigger(connectionID, userID int, req *models.CreateTriggerRequest) error {
	executor, _, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	sql := req.Definition // 触发器的定义应该包含完整的CREATE TRIGGER语句

	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "CREATE_TRIGGER",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// DropTrigger 删除触发器
func (s *DbObjectsService) DropTrigger(connectionID, userID int, database, name string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	startTime := time.Now()

	var sql string
	if conn.DBType == "postgresql" {
		sql = fmt.Sprintf("DROP TRIGGER \"%s\".\"%s\"", database, name)
	} else {
		sql = fmt.Sprintf("DROP TRIGGER `%s`.`%s`", database, name)
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
		OperationType:  "DROP_TRIGGER",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// GetEvents 获取事件列表（MySQL）
func (s *DbObjectsService) GetEvents(connectionID, userID int, database string) ([]models.EventInfo, error) {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return nil, err
	}

	if conn.DBType != "mysql" {
		return []models.EventInfo{}, nil // PostgreSQL不支持事件
	}

	sql := fmt.Sprintf("SELECT EVENT_NAME, EVENT_DEFINITION, STATUS, ON_COMPLETION FROM information_schema.EVENTS WHERE EVENT_SCHEMA = '%s'", database)
	cols, rows, _, err := executor.Execute(sql)
	if err != nil {
		return nil, err
	}

	events := make([]models.EventInfo, 0, len(rows))
	for _, row := range rows {
		event := models.EventInfo{}
		for i, col := range cols {
			switch col {
			case "EVENT_NAME":
				event.Name = fmt.Sprintf("%v", row[i])
			case "EVENT_DEFINITION":
				event.Definition = fmt.Sprintf("%v", row[i])
			case "STATUS":
				event.Status = fmt.Sprintf("%v", row[i])
			case "ON_COMPLETION":
				event.OnCompletion = fmt.Sprintf("%v", row[i])
			}
		}
		events = append(events, event)
	}
	return events, nil
}

// CreateEvent 创建事件（MySQL）
func (s *DbObjectsService) CreateEvent(connectionID, userID int, req *models.CreateEventRequest) error {
	executor, conn, err := s.getExecutor(connectionID, userID, req.Database)
	if err != nil {
		return err
	}
	if conn.DBType != "mysql" {
		return errors.New("事件功能仅支持MySQL")
	}
	startTime := time.Now()

	sql := req.Definition // 事件的定义应该包含完整的CREATE EVENT语句

	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "CREATE_EVENT",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}

// DropEvent 删除事件（MySQL）
func (s *DbObjectsService) DropEvent(connectionID, userID int, database, name string) error {
	executor, conn, err := s.getExecutor(connectionID, userID, database)
	if err != nil {
		return err
	}
	if conn.DBType != "mysql" {
		return errors.New("事件功能仅支持MySQL")
	}
	startTime := time.Now()

	sql := fmt.Sprintf("DROP EVENT `%s`.`%s`", database, name)

	_, execTime, err := executor.ExecuteUpdate(sql)
	executionTime := int(time.Since(startTime).Milliseconds())
	if execTime > 0 {
		executionTime = execTime
	}

	history := &models.SystemQueryHistory{
		ConnectionID:    connectionID,
		UserID:         userID,
		SQLQuery:       sql,
		OperationType:  "DROP_EVENT",
		ExecutionTimeMs: executionTime,
		Success:        err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}
	s.systemHistoryDAO.Create(history)
	return err
}
