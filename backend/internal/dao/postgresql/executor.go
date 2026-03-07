package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"lingosql/internal/utils"
	"lingosql/pkg/types"
)

type PostgreSQLExecutor struct {
	db *sql.DB
}

// NewPostgreSQLExecutor 创建 PostgreSQL 执行器
// sslMode: 可选，如 "disable", "require", "verify-ca", "verify-full"，空时默认 "disable"
func NewPostgreSQLExecutor(host string, port int, database, username, password string, sslMode string) (*PostgreSQLExecutor, error) {
	if sslMode == "" {
		sslMode = "disable"
	}
	// PostgreSQL 必须连接到一个数据库，空时使用 postgres（系统默认库）
	if database == "" {
		database = "postgres"
	}
	// 使用 url.URL 构建 DSN，自动处理密码中的特殊字符
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(username, password),
		Host:     fmt.Sprintf("%s:%d", host, port),
		Path:     "/" + database,
		RawQuery: "sslmode=" + url.QueryEscape(sslMode),
	}
	dsn := u.String()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 PostgreSQL 连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("PostgreSQL 连接测试失败: %w", err)
	}

	return &PostgreSQLExecutor{db: db}, nil
}

// Execute 执行 SQL 查询
func (e *PostgreSQLExecutor) Execute(sql string) ([]string, [][]interface{}, int, error) {
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
func (e *PostgreSQLExecutor) ExecuteUpdate(sql string) (int, int, error) {
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
func (e *PostgreSQLExecutor) GetDatabases() ([]string, error) {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false AND datname != 'postgres'`
	rows, err := e.db.Query(query)
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
		databases = append(databases, dbName)
	}

	return databases, nil
}

// GetTables 获取表列表
func (e *PostgreSQLExecutor) GetTables(database string) ([]map[string]interface{}, error) {
	// PostgreSQL 需要先切换到目标数据库
	query := `SELECT table_name, 
	          (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = t.table_name) as column_count
	          FROM information_schema.tables t
	          WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	          ORDER BY table_name`
	
	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []map[string]interface{}
	for rows.Next() {
		var name string
		var columnCount int

		if err := rows.Scan(&name, &columnCount); err != nil {
			return nil, err
		}

		table := map[string]interface{}{
			"name":         name,
			"column_count": columnCount,
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetTableInfo 获取表详细信息
func (e *PostgreSQLExecutor) GetTableInfo(database, table string) (map[string]interface{}, error) {
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	query := `
		SELECT 
			c.relname as name,
			pg_total_relation_size(c.oid) as total_size,
			pg_table_size(c.oid) as data_length,
			pg_indexes_size(c.oid) as index_length,
			c.reltuples::bigint as rows,
			obj_description(c.oid, 'pg_class') as comment
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname = $1 AND n.nspname = 'public'
	`

	row := e.db.QueryRow(query, table)

	var name string
	var totalSize, dataLength, indexLength sql.NullInt64
	var rowsCount sql.NullInt64
	var comment sql.NullString

	if err := row.Scan(&name, &totalSize, &dataLength, &indexLength, &rowsCount, &comment); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("表 %s 不存在", table)
		}
		return nil, err
	}

	result := map[string]interface{}{
		"name":           name,
		"engine":         "PostgreSQL",
		"rows":           rowsCount.Int64,
		"data_length":    dataLength.Int64,
		"index_length":   indexLength.Int64,
		"collation":      "",
		"auto_increment": nil, // PostgreSQL 不支持 AUTO_INCREMENT
		"comment":        comment.String,
		"create_time":    "",
		"update_time":    "",
	}

	// 尝试获取表的编码信息
	encodingQuery := `SELECT pg_encoding_to_char(encoding) FROM pg_database WHERE datname = current_database()`
	var encoding string
	if err := e.db.QueryRow(encodingQuery).Scan(&encoding); err == nil {
		result["collation"] = encoding
	}

	return result, nil
}

// GetTableColumns 获取表字段列表
func (e *PostgreSQLExecutor) GetTableColumns(database, table string) ([]map[string]interface{}, error) {
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	query := `
		SELECT 
			a.attname AS column_name,
			format_type(a.atttypid, a.atttypmod) AS data_type,
			a.attnotnull AS not_null,
			pg_get_expr(d.adbin, d.adrelid) AS default_value,
			col_description(a.attrelid, a.attnum) AS comment,
			CASE WHEN pk.contype = 'p' THEN true ELSE false END AS is_primary,
			a.attidentity AS identity,
			a.attndims AS array_dims,
			a.atttypmod AS type_mod
		FROM pg_attribute a
		LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		LEFT JOIN pg_constraint pk ON pk.conrelid = a.attrelid 
			AND a.attnum = ANY(pk.conkey) AND pk.contype = 'p'
		WHERE a.attrelid = $1::regclass
			AND a.attnum > 0 
			AND NOT a.attisdropped
		ORDER BY a.attnum
	`

	// 构建带 schema 的表名
	fullTableName := fmt.Sprintf("public.%s", table)

	rows, err := e.db.Query(query, fullTableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []map[string]interface{}
	for rows.Next() {
		var columnName, dataType string
		var notNull, isPrimary bool
		var defaultValue, comment sql.NullString
		var identity sql.NullString
		var arrayDims int
		var typeMod int

		if err := rows.Scan(
			&columnName, &dataType, &notNull, &defaultValue,
			&comment, &isPrimary, &identity, &arrayDims, &typeMod,
		); err != nil {
			return nil, err
		}

		// 解析数据类型和长度
		baseType, length, precision, scale := parsePostgreSQLType(dataType, typeMod)

		// 判断是否为数组
		isArray := arrayDims > 0

		// 判断标识列类型
		var identityType string
		if identity.Valid && identity.String != "" {
			switch identity.String {
			case "a":
				identityType = "ALWAYS"
			case "d":
				identityType = "BY DEFAULT"
			default:
				identityType = ""
			}
		}

		column := map[string]interface{}{
			"name":           columnName,
			"type":           toUpper(baseType),
			"nullable":       !notNull,
			"is_primary":     isPrimary,
			"auto_increment": false, // PostgreSQL 使用 identity 或 serial
			"is_array":       isArray,
			"dimension":      arrayDims,
			"identity":       identityType,
			"extra":          "",
		}

		// 处理长度
		if length != nil {
			column["length"] = *length
		} else {
			column["length"] = nil
		}

		// 处理精度和小数位
		if precision != nil {
			column["precision"] = *precision
		} else {
			column["precision"] = nil
		}
		if scale != nil {
			column["scale"] = *scale
		} else {
			column["scale"] = nil
		}

		// 处理默认值
		if defaultValue.Valid {
			column["default_value"] = defaultValue.String
		} else {
			column["default_value"] = nil
		}

		// 处理注释
		if comment.Valid {
			column["comment"] = comment.String
		} else {
			column["comment"] = ""
		}

		columns = append(columns, column)
	}

	return columns, nil
}

// GetTableIndexes 获取表索引列表
func (e *PostgreSQLExecutor) GetTableIndexes(database, table string) ([]map[string]interface{}, error) {
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	// 仅取 attnum > 0 的列（跳过表达式索引中的 0），使用 pq.Array 正确扫描 text[]
	query := `
		SELECT 
			i.relname AS index_name,
			ix.indisunique AS is_unique,
			ix.indisprimary AS is_primary,
			am.amname AS method,
			pg_get_indexdef(ix.indexrelid) AS index_def,
			obj_description(ix.indexrelid, 'pg_class') AS comment,
			COALESCE(
				(SELECT array_agg(a.attname ORDER BY k.ord)
				 FROM unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord)
				 JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = k.attnum
				 WHERE a.attnum > 0 AND NOT a.attisdropped),
				ARRAY[]::text[]
			) AS columns
		FROM pg_index ix
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_am am ON am.oid = i.relam
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE t.relname = $1 AND n.nspname = 'public'
		ORDER BY i.relname
	`

	rows, err := e.db.Query(query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []map[string]interface{}
	for rows.Next() {
		var indexName, method, indexDef string
		var isUnique, isPrimary bool
		var comment sql.NullString
		var columns pq.StringArray

		if err := rows.Scan(
			&indexName, &isUnique, &isPrimary, &method, &indexDef, &comment, &columns,
		); err != nil {
			return nil, err
		}
		columnsStr := []string(columns)

		// 确定索引类型
		var indexType string
		if isPrimary {
			indexType = "PRIMARY"
		} else if isUnique {
			indexType = "UNIQUE"
		} else {
			indexType = "INDEX"
		}

		// 尝试提取 WHERE 子句（部分索引）
		whereClause := extractWhereClause(indexDef)

		index := map[string]interface{}{
			"name":         indexName,
			"type":         indexType,
			"method":       method,
			"columns":      columnsStr,
			"cardinality":  nil, // PostgreSQL 需要从 pg_stats 获取，这里简化处理
			"where_clause": whereClause,
		}

		if comment.Valid {
			index["comment"] = comment.String
		} else {
			index["comment"] = ""
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

// parsePostgreSQLType 解析 PostgreSQL 类型字符串
func parsePostgreSQLType(dataType string, typeMod int) (baseType string, length *int64, precision *int64, scale *int64) {
	baseType = dataType

	// 处理 character varying(n), varchar(n)
	if len(dataType) > 0 {
		// 检查是否包含长度信息
		if idx := indexOf(dataType, '('); idx != -1 {
			endIdx := indexOf(dataType, ')')
			if endIdx > idx {
				baseType = dataType[:idx]
				params := dataType[idx+1 : endIdx]
				
				// 检查是否有逗号（precision, scale）
				if commaIdx := indexOf(params, ','); commaIdx != -1 {
					p := parseInt64(params[:commaIdx])
					s := parseInt64(params[commaIdx+1:])
					precision = &p
					scale = &s
				} else {
					l := parseInt64(params)
					length = &l
				}
			}
		}
	}

	// 标准化类型名
	baseType = normalizeTypeName(baseType)

	return
}

// normalizeTypeName 标准化 PostgreSQL 类型名
func normalizeTypeName(typeName string) string {
	switch typeName {
	case "character varying":
		return "VARCHAR"
	case "character":
		return "CHAR"
	case "integer":
		return "INTEGER"
	case "bigint":
		return "BIGINT"
	case "smallint":
		return "SMALLINT"
	case "double precision":
		return "DOUBLE PRECISION"
	case "real":
		return "REAL"
	case "boolean":
		return "BOOLEAN"
	case "timestamp without time zone":
		return "TIMESTAMP"
	case "timestamp with time zone":
		return "TIMESTAMPTZ"
	case "date":
		return "DATE"
	case "time without time zone":
		return "TIME"
	case "time with time zone":
		return "TIMETZ"
	case "text":
		return "TEXT"
	case "bytea":
		return "BYTEA"
	case "json":
		return "JSON"
	case "jsonb":
		return "JSONB"
	case "uuid":
		return "UUID"
	case "numeric":
		return "NUMERIC"
	default:
		return toUpper(typeName)
	}
}

// extractWhereClause 从索引定义中提取 WHERE 子句
func extractWhereClause(indexDef string) string {
	whereIdx := indexOfStr(toLowerPg(indexDef), " where ")
	if whereIdx == -1 {
		return ""
	}
	return trim(indexDef[whereIdx+7:])
}

// 辅助函数
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func indexOfStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func parseInt64(s string) int64 {
	s = trim(s)
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

func toUpper(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

func toLowerPg(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

// pgValueToJSON 将 PostgreSQL 驱动返回值转为 JSON 可序列化类型
func pgValueToJSON(val interface{}) interface{} {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []byte:
		return string(v)
	case [16]byte:
		return fmt.Sprintf("%x-%x-%x-%x-%x", v[0:4], v[4:6], v[6:8], v[8:10], v[10:16])
	default:
		return val
	}
}

// GetTableRows 获取表数据
func (e *PostgreSQLExecutor) GetTableRows(database, table string, filters []types.RowFilter, page, pageSize int) (*types.TableRowsResult, error) {
	if err := utils.ValidateTableName(table); err != nil {
		return nil, err
	}
	for _, f := range filters {
		if err := utils.ValidateColumnName(f.Field); err != nil {
			return nil, err
		}
	}
	// 1. 获取列信息
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
		columnNames = append(columnNames, fmt.Sprintf("\"%s\"", name))
	}

	// 3. 构建 WHERE 条件
	whereClause, whereArgs := buildPgWhereClause(filters)

	// 4. 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM \"public\".\"%s\"%s", table, whereClause)
	var total int64
	if err := e.db.QueryRow(countSQL, whereArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 5. 查询数据
	offset := (page - 1) * pageSize
	dataSQL := fmt.Sprintf(
		"SELECT %s FROM \"public\".\"%s\"%s LIMIT %d OFFSET %d",
		joinPgStrings(columnNames, ", "),
		table, whereClause, pageSize, offset,
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
			row[col.Name] = pgValueToJSON(val)
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
func (e *PostgreSQLExecutor) UpdateTableRow(database, table string, primaryKey map[string]interface{}, data map[string]interface{}) (int, error) {
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

	// 构建 SET 子句（PostgreSQL 使用 $1, $2 占位符）
	var setClauses []string
	var args []interface{}
	argIndex := 1

	for field, value := range data {
		setClauses = append(setClauses, fmt.Sprintf("\"%s\" = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// 构建 WHERE 子句
	var whereClauses []string
	for field, value := range primaryKey {
		whereClauses = append(whereClauses, fmt.Sprintf("\"%s\" = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// 构建 SQL
	updateSQL := fmt.Sprintf(
		"UPDATE \"public\".\"%s\" SET %s WHERE %s",
		table,
		joinPgStrings(setClauses, ", "),
		joinPgStrings(whereClauses, " AND "),
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

// buildPgWhereClause 构建 PostgreSQL WHERE 子句
func buildPgWhereClause(filters []types.RowFilter) (string, []interface{}) {
	if len(filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	for _, f := range filters {
		switch f.Operator {
		case "=", "!=", "<", "<=", ">", ">=":
			conditions = append(conditions, fmt.Sprintf("\"%s\" %s $%d", f.Field, f.Operator, argIndex))
			args = append(args, f.Value)
			argIndex++
		case "LIKE", "NOT LIKE":
			// PostgreSQL 使用 ILIKE 进行不区分大小写的匹配
			conditions = append(conditions, fmt.Sprintf("\"%s\" %s $%d", f.Field, f.Operator, argIndex))
			value := f.Value
			if len(value) > 0 && value[0] != '%' && value[len(value)-1] != '%' {
				value = "%" + value + "%"
			}
			args = append(args, value)
			argIndex++
		case "IN", "NOT IN":
			values := splitPgAndTrim(f.Value, ",")
			if len(values) > 0 {
				placeholders := make([]string, len(values))
				for i, v := range values {
					placeholders[i] = fmt.Sprintf("$%d", argIndex)
					args = append(args, v)
					argIndex++
				}
				conditions = append(conditions, fmt.Sprintf("\"%s\" %s (%s)", f.Field, f.Operator, joinPgStrings(placeholders, ", ")))
			}
		case "IS NULL":
			conditions = append(conditions, fmt.Sprintf("\"%s\" IS NULL", f.Field))
		case "IS NOT NULL":
			conditions = append(conditions, fmt.Sprintf("\"%s\" IS NOT NULL", f.Field))
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return " WHERE " + joinPgStrings(conditions, " AND "), args
}

// joinPgStrings 连接字符串切片
func joinPgStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// splitPgAndTrim 分割字符串并去除空白
func splitPgAndTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || (i+len(sep) <= len(s) && s[i:i+len(sep)] == sep) {
			part := trimPg(s[start:i])
			if len(part) > 0 {
				result = append(result, part)
			}
			start = i + len(sep)
		}
	}
	return result
}

// trimPg 去除字符串两端的空白
func trimPg(s string) string {
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

// GetVersion 获取数据库版本
func (e *PostgreSQLExecutor) GetVersion() (string, error) {
	var version string
	err := e.db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return "", err
	}
	// 简化版本信息，例如 "PostgreSQL 15.2"
	return version, nil
}

// Explain 执行 EXPLAIN 查询
func (e *PostgreSQLExecutor) Explain(sql string) ([]map[string]interface{}, int, error) {
	start := time.Now()
	
	// PostgreSQL 使用 EXPLAIN (FORMAT JSON) 获取更详细的信息
	explainSQL := "EXPLAIN (FORMAT JSON) " + sql
	
	rows, err := e.db.Query(explainSQL)
	if err != nil {
		// 如果 JSON 格式失败，尝试使用 TEXT 格式
		return e.explainText(sql, start)
	}
	defer rows.Close()

	// PostgreSQL 的 EXPLAIN (FORMAT JSON) 返回 JSON 格式
	var result []map[string]interface{}
	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			// 如果扫描失败，尝试 TEXT 格式
			return e.explainText(sql, start)
		}

		// 解析 JSON - PostgreSQL 返回的是包含 Plan 字段的对象数组
		var planArray []map[string]interface{}
		if err := json.Unmarshal(jsonData, &planArray); err != nil {
			// 如果解析失败，尝试作为单个对象
			var singlePlan map[string]interface{}
			if err2 := json.Unmarshal(jsonData, &singlePlan); err2 != nil {
				// JSON 解析失败，使用 TEXT 格式
				return e.explainText(sql, start)
			}
			// 提取 Plan 字段
			if plan, ok := singlePlan["Plan"].(map[string]interface{}); ok {
				result = append(result, plan)
			} else {
				result = append(result, singlePlan)
			}
		} else {
			// 从数组中提取 Plan 字段
			for _, item := range planArray {
				if plan, ok := item["Plan"].(map[string]interface{}); ok {
					result = append(result, plan)
				} else {
					result = append(result, item)
				}
			}
		}
	}

	executionTime := int(time.Since(start).Milliseconds())
	return result, executionTime, nil
}

// explainText 使用 TEXT 格式执行 EXPLAIN
func (e *PostgreSQLExecutor) explainText(sql string, start time.Time) ([]map[string]interface{}, int, error) {
	explainSQL := "EXPLAIN " + sql
	rows, err := e.db.Query(explainSQL)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}

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
func (e *PostgreSQLExecutor) BeginTransaction() error {
	_, err := e.db.Exec("BEGIN")
	return err
}

// CommitTransaction 提交事务
func (e *PostgreSQLExecutor) CommitTransaction() error {
	_, err := e.db.Exec("COMMIT")
	return err
}

// RollbackTransaction 回滚事务
func (e *PostgreSQLExecutor) RollbackTransaction() error {
	_, err := e.db.Exec("ROLLBACK")
	return err
}

// Close 关闭连接
func (e *PostgreSQLExecutor) Close() error {
	return e.db.Close()
}

// 以下方法在 admin.go 中实现
// CheckAdminPermissions, CreateDatabase, DropDatabase, GetDatabaseInfo, RenameDatabase
// GetUsers, CreateUser, DropUser, AlterUserPassword, GetUserGrants
// GrantDatabasePrivileges, RevokeDatabasePrivileges, GrantTablePrivileges, RevokeTablePrivileges
// GrantColumnPrivileges, RevokeColumnPrivileges, GetPermissionTree
