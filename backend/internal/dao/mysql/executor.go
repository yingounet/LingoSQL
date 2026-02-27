package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"lingosql/internal/utils"
	"lingosql/pkg/types"
)

type MySQLExecutor struct {
	db *sql.DB
}

// NewMySQLExecutor 创建 MySQL 执行器
func NewMySQLExecutor(host string, port int, database, username, password string) (*MySQLExecutor, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 MySQL 连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL 连接测试失败: %w", err)
	}

	return &MySQLExecutor{db: db}, nil
}

// Execute 执行 SQL 查询
func (e *MySQLExecutor) Execute(sql string) ([]string, [][]interface{}, int, error) {
	start := time.Now()
	rows, err := e.db.Query(sql)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, 0, err
	}

	// 读取数据
	var result [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, 0, err
		}

		// 转换值类型
		row := make([]interface{}, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = nil
			} else if b, ok := val.([]byte); ok {
				// 将 []byte 转换为 string，避免 JSON 序列化时 Base64 编码
				row[i] = string(b)
			} else {
				row[i] = val
			}
		}
		result = append(result, row)
	}

	executionTime := int(time.Since(start).Milliseconds())
	return columns, result, executionTime, nil
}

// ExecuteUpdate 执行更新操作（INSERT, UPDATE, DELETE）
func (e *MySQLExecutor) ExecuteUpdate(sql string) (int, int, error) {
	start := time.Now()
	result, err := e.db.Exec(sql)
	if err != nil {
		return 0, 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	executionTime := int(time.Since(start).Milliseconds())
	return int(rowsAffected), executionTime, nil
}

// GetDatabases 获取数据库列表
func (e *MySQLExecutor) GetDatabases() ([]string, error) {
	rows, err := e.db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		// 排除系统数据库
		if dbName != "information_schema" && dbName != "performance_schema" && 
		   dbName != "mysql" && dbName != "sys" {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// GetTables 获取表列表
func (e *MySQLExecutor) GetTables(database string) ([]map[string]interface{}, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT TABLE_NAME, ENGINE, TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH FROM information_schema.TABLES WHERE TABLE_SCHEMA = '%s'", database)
	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []map[string]interface{}
	for rows.Next() {
		var name, engine string
		var rowsCount sql.NullInt64
		var dataLength, indexLength sql.NullInt64

		if err := rows.Scan(&name, &engine, &rowsCount, &dataLength, &indexLength); err != nil {
			return nil, err
		}

		table := map[string]interface{}{
			"name":        name,
			"engine":      engine,
			"rows":        rowsCount.Int64,
			"data_length": dataLength.Int64,
			"index_length": indexLength.Int64,
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetTableInfo 获取表详细信息
func (e *MySQLExecutor) GetTableInfo(database, table string) (map[string]interface{}, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return nil, err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT 
			TABLE_NAME, ENGINE, TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH,
			TABLE_COLLATION, AUTO_INCREMENT, TABLE_COMMENT, 
			CREATE_TIME, UPDATE_TIME
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'
	`, database, table)

	row := e.db.QueryRow(query)

	var name string
	var engine sql.NullString
	var rowsCount sql.NullInt64
	var dataLength, indexLength sql.NullInt64
	var collation sql.NullString
	var autoIncrement sql.NullInt64
	var comment sql.NullString
	var createTime, updateTime sql.NullTime

	if err := row.Scan(&name, &engine, &rowsCount, &dataLength, &indexLength,
		&collation, &autoIncrement, &comment, &createTime, &updateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("表 %s 不存在", table)
		}
		return nil, err
	}

	result := map[string]interface{}{
		"name":         name,
		"engine":       engine.String,
		"rows":         rowsCount.Int64,
		"data_length":  dataLength.Int64,
		"index_length": indexLength.Int64,
		"collation":    collation.String,
		"comment":      comment.String,
	}

	// 处理可能为 NULL 的自增ID
	if autoIncrement.Valid {
		result["auto_increment"] = autoIncrement.Int64
	} else {
		result["auto_increment"] = nil
	}

	// 处理时间字段
	if createTime.Valid {
		result["create_time"] = createTime.Time.Format("2006-01-02 15:04:05")
	} else {
		result["create_time"] = ""
	}

	if updateTime.Valid {
		result["update_time"] = updateTime.Time.Format("2006-01-02 15:04:05")
	} else {
		result["update_time"] = ""
	}

	return result, nil
}

// GetTableColumns 获取表字段列表
func (e *MySQLExecutor) GetTableColumns(database, table string) ([]map[string]interface{}, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return nil, err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_COMMENT,
			COLUMN_KEY,
			EXTRA,
			COLUMN_TYPE
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'
		ORDER BY ORDINAL_POSITION
	`, database, table)

	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []map[string]interface{}
	for rows.Next() {
		var columnName, dataType string
		var charMaxLen, numPrecision, numScale sql.NullInt64
		var isNullable, columnKey, extra, columnType string
		var columnDefault, columnComment sql.NullString

		if err := rows.Scan(
			&columnName, &dataType, &charMaxLen, &numPrecision, &numScale,
			&isNullable, &columnDefault, &columnComment, &columnKey, &extra, &columnType,
		); err != nil {
			return nil, err
		}

		// 判断是否为无符号
		unsigned := false
		if len(columnType) > 0 {
			unsigned = strings.Contains(strings.ToLower(columnType), "unsigned")
		}

		// 判断是否为自增
		autoIncrement := strings.Contains(strings.ToLower(extra), "auto_increment")

		// 判断是否为主键
		isPrimary := columnKey == "PRI"

		// 构建字段信息
		column := map[string]interface{}{
			"name":           columnName,
			"type":           strings.ToUpper(dataType),
			"nullable":       isNullable == "YES",
			"is_primary":     isPrimary,
			"auto_increment": autoIncrement,
			"unsigned":       unsigned,
			"extra":          extra,
			"comment":        columnComment.String,
		}

		// 处理长度
		if charMaxLen.Valid {
			column["length"] = charMaxLen.Int64
		} else {
			column["length"] = nil
		}

		// 处理精度和小数位
		if numPrecision.Valid {
			column["precision"] = numPrecision.Int64
		} else {
			column["precision"] = nil
		}
		if numScale.Valid {
			column["scale"] = numScale.Int64
		} else {
			column["scale"] = nil
		}

		// 处理默认值
		if columnDefault.Valid {
			column["default_value"] = columnDefault.String
		} else {
			column["default_value"] = nil
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// GetTableIndexes 获取表索引列表
func (e *MySQLExecutor) GetTableIndexes(database, table string) ([]map[string]interface{}, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return nil, err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT 
			INDEX_NAME,
			NON_UNIQUE,
			COLUMN_NAME,
			SEQ_IN_INDEX,
			INDEX_TYPE,
			CARDINALITY,
			INDEX_COMMENT
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s'
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`, database, table)

	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 使用 map 来聚合相同索引的多个列
	indexMap := make(map[string]*indexData)
	var indexOrder []string

	for rows.Next() {
		var indexName, columnName, indexType string
		var nonUnique int
		var seqInIndex int
		var cardinality sql.NullInt64
		var indexComment sql.NullString

		if err := rows.Scan(
			&indexName, &nonUnique, &columnName, &seqInIndex,
			&indexType, &cardinality, &indexComment,
		); err != nil {
			return nil, err
		}

		if _, exists := indexMap[indexName]; !exists {
			// 确定索引类型
			var idxType string
			if indexName == "PRIMARY" {
				idxType = "PRIMARY"
			} else if nonUnique == 0 {
				idxType = "UNIQUE"
			} else {
				idxType = "INDEX"
			}

			indexMap[indexName] = &indexData{
				name:        indexName,
				indexType:   idxType,
				method:      indexType,
				columns:     []string{},
				cardinality: cardinality,
				comment:     indexComment.String,
			}
			indexOrder = append(indexOrder, indexName)
		}

		// 添加列名
		indexMap[indexName].columns = append(indexMap[indexName].columns, columnName)
	}

	// 转换为返回格式
	var indexes []map[string]interface{}
	for _, name := range indexOrder {
		idx := indexMap[name]
		index := map[string]interface{}{
			"name":    idx.name,
			"type":    idx.indexType,
			"method":  idx.method,
			"columns": idx.columns,
			"comment": idx.comment,
		}

		if idx.cardinality.Valid {
			index["cardinality"] = idx.cardinality.Int64
		} else {
			index["cardinality"] = nil
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

// indexData 用于聚合索引信息的辅助结构
type indexData struct {
	name        string
	indexType   string
	method      string
	columns     []string
	cardinality sql.NullInt64
	comment     string
}

// GetVersion 获取数据库版本
func (e *MySQLExecutor) GetVersion() (string, error) {
	var version string
	err := e.db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", err
	}
	return "MySQL " + version, nil
}

// GetTableRows 获取表数据
func (e *MySQLExecutor) GetTableRows(database, table string, filters []types.RowFilter, page, pageSize int) (*types.TableRowsResult, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return nil, err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	for _, f := range filters {
		if err := utils.ValidateColumnName(f.Field); err != nil {
			return nil, err
		}
	}
	// 1. 获取列信息（包含主键和索引信息）
	columnsInfo, err := e.GetTableColumns(database, table)
	if err != nil {
		return nil, fmt.Errorf("获取表结构失败: %w", err)
	}

	// 2. 获取索引信息
	indexesInfo, err := e.GetTableIndexes(database, table)
	if err != nil {
		return nil, fmt.Errorf("获取索引信息失败: %w", err)
	}

	// 构建索引字段集合
	indexFields := make(map[string]bool)
	for _, idx := range indexesInfo {
		if cols, ok := idx["columns"].([]string); ok {
			for _, col := range cols {
				indexFields[col] = true
			}
		}
	}

	// 构建列定义
	var columns []types.ColumnDef
	var columnNames []string
	for _, col := range columnsInfo {
		name := col["name"].(string)
		colType := col["type"].(string)
		isPrimary := false
		if v, ok := col["is_primary"].(bool); ok {
			isPrimary = v
		}
		isIndex := indexFields[name]

		columns = append(columns, types.ColumnDef{
			Name:      name,
			Type:      colType,
			IsPrimary: isPrimary,
			IsIndex:   isIndex,
		})
		columnNames = append(columnNames, fmt.Sprintf("`%s`", name))
	}

	// 3. 构建 WHERE 条件
	whereClause, whereArgs := buildWhereClause(filters)

	// 4. 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`%s", database, table, whereClause)
	var total int64
	if err := e.db.QueryRow(countSQL, whereArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 5. 查询数据
	offset := (page - 1) * pageSize
	dataSQL := fmt.Sprintf(
		"SELECT %s FROM `%s`.`%s`%s LIMIT %d OFFSET %d",
		joinStrings(columnNames, ", "),
		database, table, whereClause, pageSize, offset,
	)

	rows, err := e.db.Query(dataSQL, whereArgs...)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %w", err)
	}
	defer rows.Close()

	// 6. 读取数据
	var resultRows []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("读取数据失败: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// 处理 []byte 类型
			if b, ok := val.([]byte); ok {
				row[col.Name] = string(b)
			} else {
				row[col.Name] = val
			}
		}
		resultRows = append(resultRows, row)
	}

	return &types.TableRowsResult{
		Columns:  columns,
		Rows:     resultRows,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateTableRow 更新表数据
func (e *MySQLExecutor) UpdateTableRow(database, table string, primaryKey map[string]interface{}, data map[string]interface{}) (int, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return 0, err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return 0, err
	}
	if len(primaryKey) == 0 {
		return 0, fmt.Errorf("主键不能为空")
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("更新数据不能为空")
	}
	for field := range primaryKey {
		if err := utils.ValidateColumnName(field); err != nil {
			return 0, err
		}
	}
	for field := range data {
		if err := utils.ValidateColumnName(field); err != nil {
			return 0, err
		}
	}

	// 构建 SET 子句
	var setClauses []string
	var setArgs []interface{}
	for field, value := range data {
		setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", field))
		setArgs = append(setArgs, value)
	}

	// 构建 WHERE 子句
	var whereClauses []string
	var whereArgs []interface{}
	for field, value := range primaryKey {
		whereClauses = append(whereClauses, fmt.Sprintf("`%s` = ?", field))
		whereArgs = append(whereArgs, value)
	}

	// 合并参数
	args := append(setArgs, whereArgs...)

	// 构建 SQL
	updateSQL := fmt.Sprintf(
		"UPDATE `%s`.`%s` SET %s WHERE %s",
		database, table,
		joinStrings(setClauses, ", "),
		joinStrings(whereClauses, " AND "),
	)

	result, err := e.db.Exec(updateSQL, args...)
	if err != nil {
		return 0, fmt.Errorf("更新数据失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %w", err)
	}

	return int(affected), nil
}

// buildWhereClause 构建 WHERE 子句
func buildWhereClause(filters []types.RowFilter) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	for _, f := range filters {
		switch f.Operator {
		case "=", "!=", "<", "<=", ">", ">=":
			conditions = append(conditions, fmt.Sprintf("`%s` %s ?", f.Field, f.Operator))
			args = append(args, f.Value)
		case "LIKE", "NOT LIKE":
			conditions = append(conditions, fmt.Sprintf("`%s` %s ?", f.Field, f.Operator))
			// 如果值不包含 %，自动添加
			value := f.Value
			if len(value) > 0 && value[0] != '%' && value[len(value)-1] != '%' {
				value = "%" + value + "%"
			}
			args = append(args, value)
		case "IN", "NOT IN":
			// IN 条件需要特殊处理，值用逗号分隔
			values := splitAndTrim(f.Value, ",")
			if len(values) > 0 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = "?"
					args = append(args, v)
				}
				conditions = append(conditions, fmt.Sprintf("`%s` %s (%s)", f.Field, f.Operator, joinStrings(placeholders, ", ")))
			}
		case "IS NULL":
			conditions = append(conditions, fmt.Sprintf("`%s` IS NULL", f.Field))
		case "IS NOT NULL":
			conditions = append(conditions, fmt.Sprintf("`%s` IS NOT NULL", f.Field))
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return " WHERE " + joinStrings(conditions, " AND "), args
}

// joinStrings 连接字符串切片
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// splitAndTrim 分割字符串并去除空白
func splitAndTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || (i+len(sep) <= len(s) && s[i:i+len(sep)] == sep) {
			part := trimSpace(s[start:i])
			if len(part) > 0 {
				result = append(result, part)
			}
			start = i + len(sep)
		}
	}
	return result
}

// trimSpace 去除字符串两端的空白
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// Explain 执行 EXPLAIN 查询
func (e *MySQLExecutor) Explain(sql string) ([]map[string]interface{}, int, error) {
	start := time.Now()
	
	// 构建 EXPLAIN 语句
	explainSQL := "EXPLAIN " + sql
	
	rows, err := e.db.Query(explainSQL)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}

	// 读取数据
	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, 0, err
		}

		// 转换为 map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				row[col] = nil
			} else if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result = append(result, row)
	}

	executionTime := int(time.Since(start).Milliseconds())
	return result, executionTime, nil
}

// BeginTransaction 开始事务
func (e *MySQLExecutor) BeginTransaction() error {
	_, err := e.db.Exec("START TRANSACTION")
	return err
}

// CommitTransaction 提交事务
func (e *MySQLExecutor) CommitTransaction() error {
	_, err := e.db.Exec("COMMIT")
	return err
}

// RollbackTransaction 回滚事务
func (e *MySQLExecutor) RollbackTransaction() error {
	_, err := e.db.Exec("ROLLBACK")
	return err
}

// Close 关闭连接
func (e *MySQLExecutor) Close() error {
	return e.db.Close()
}

// 以下方法在 admin.go 中实现
// CheckAdminPermissions, CreateDatabase, DropDatabase, GetDatabaseInfo, RenameDatabase
// GetUsers, CreateUser, DropUser, AlterUserPassword, GetUserGrants
// GrantDatabasePrivileges, RevokeDatabasePrivileges, GrantTablePrivileges, RevokeTablePrivileges
// GrantColumnPrivileges, RevokeColumnPrivileges, GetPermissionTree
