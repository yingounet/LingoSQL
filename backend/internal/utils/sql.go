package utils

import "strings"

// RemoveSQLComments 移除 SQL 中的注释
// 支持 -- 单行注释和 /* */ 多行注释
// 会正确处理字符串中的注释符号（不会误删）
func RemoveSQLComments(sql string) string {
	var result strings.Builder
	i := 0
	inString := false
	stringChar := byte(0)

	for i < len(sql) {
		// 处理字符串（单引号或双引号）
		if !inString && (sql[i] == '\'' || sql[i] == '"') {
			inString = true
			stringChar = sql[i]
			result.WriteByte(sql[i])
			i++
			continue
		}

		if inString {
			if sql[i] == stringChar {
				// 检查转义（连续两个相同引号表示转义）
				if i+1 < len(sql) && sql[i+1] == stringChar {
					result.WriteByte(sql[i])
					result.WriteByte(sql[i+1])
					i += 2
					continue
				}
				inString = false
			}
			result.WriteByte(sql[i])
			i++
			continue
		}

		// 处理 -- 单行注释
		if i+1 < len(sql) && sql[i] == '-' && sql[i+1] == '-' {
			// 跳过到行尾
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			// 保留换行符，避免语句连在一起
			if i < len(sql) {
				result.WriteByte('\n')
				i++
			}
			continue
		}

		// 处理 # 单行注释（MySQL 风格）
		if sql[i] == '#' {
			// 跳过到行尾
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			// 保留换行符
			if i < len(sql) {
				result.WriteByte('\n')
				i++
			}
			continue
		}

		// 处理 /* */ 多行注释
		if i+1 < len(sql) && sql[i] == '/' && sql[i+1] == '*' {
			i += 2
			for i+1 < len(sql) && !(sql[i] == '*' && sql[i+1] == '/') {
				i++
			}
			if i+1 < len(sql) {
				i += 2 // 跳过 */
			}
			// 添加空格避免语句连在一起
			result.WriteByte(' ')
			continue
		}

		result.WriteByte(sql[i])
		i++
	}

	return result.String()
}

// SplitSQLStatements 将 SQL 按分号分割成多条语句
// 会正确处理字符串中的分号（不会误分割）
// 返回去除前后空白的非空语句列表
func SplitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		// 处理字符串
		if !inString && (ch == '\'' || ch == '"') {
			inString = true
			stringChar = ch
			current.WriteByte(ch)
			continue
		}

		if inString {
			if ch == stringChar {
				// 检查转义
				if i+1 < len(sql) && sql[i+1] == stringChar {
					current.WriteByte(ch)
					current.WriteByte(sql[i+1])
					i++
					continue
				}
				inString = false
			}
			current.WriteByte(ch)
			continue
		}

		// 非字符串中的分号作为分隔符
		if ch == ';' {
			stmt := strings.TrimSpace(current.String())
			if len(stmt) > 0 {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		current.WriteByte(ch)
	}

	// 处理最后一条语句（可能没有分号结尾）
	stmt := strings.TrimSpace(current.String())
	if len(stmt) > 0 {
		statements = append(statements, stmt)
	}

	return statements
}

// IsSQLQuery 判断 SQL 是否为查询语句（返回结果集）
func IsSQLQuery(sql string) bool {
	sqlUpper := strings.ToUpper(strings.TrimSpace(sql))
	return strings.HasPrefix(sqlUpper, "SELECT") ||
		strings.HasPrefix(sqlUpper, "SHOW") ||
		strings.HasPrefix(sqlUpper, "DESCRIBE") ||
		strings.HasPrefix(sqlUpper, "DESC") ||
		strings.HasPrefix(sqlUpper, "EXPLAIN")
}
