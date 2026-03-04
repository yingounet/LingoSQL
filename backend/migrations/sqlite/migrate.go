package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed *.sql
var migrationFS embed.FS

// Migrate 执行数据库迁移
func Migrate(dbPath string) error {
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()

	// 迁移文件列表（按顺序执行）
	migrationFiles := []string{
		"001_init.sql",
		"002_add_system_query_history.sql",
		"003_favorites_database_and_last_used.sql",
		"004_connections_last_used_at.sql",
		"005_add_user_security.sql",
		"006_audit_and_tasks.sql",
		"007_system_settings.sql",
	}

	// 依次执行每个迁移文件
	for _, filename := range migrationFiles {
		sqlBytes, err := migrationFS.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("读取迁移文件 %s 失败: %w", filename, err)
		}

		sqlStr := strings.TrimSpace(string(sqlBytes))
		statements := splitSQLStatements(sqlStr)
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			_, err = db.Exec(stmt)
			if err != nil {
				errStr := err.Error()
				// 幂等：列已存在或索引已存在时忽略
				if strings.Contains(errStr, "duplicate column name") ||
					strings.Contains(errStr, "already exists") {
					continue
				}
				return fmt.Errorf("执行迁移文件 %s 失败: %w", filename, err)
			}
		}
	}

	return nil
}

// splitSQLStatements 按分号拆分 SQL，保留注释内的分号不拆分（简单实现：按行和分号拆分）
func splitSQLStatements(s string) []string {
	var list []string
	var buf strings.Builder
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		buf.WriteString(line)
		buf.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			list = append(list, strings.TrimSpace(strings.TrimSuffix(buf.String(), "\n")))
			buf.Reset()
		}
	}
	if buf.Len() > 0 {
		list = append(list, strings.TrimSpace(buf.String()))
	}
	return list
}

// InitDB 初始化数据库连接
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 执行迁移
	if err := Migrate(dbPath); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return db, nil
}
