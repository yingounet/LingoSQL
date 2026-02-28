package postgresql

import (
	"database/sql"
	"fmt"
	"strings"

	"lingosql/internal/utils"
)

// CheckAdminPermissions 检查管理员权限
func (e *PostgreSQLExecutor) CheckAdminPermissions() (map[string]bool, error) {
	permissions := make(map[string]bool)

	// 检查是否为超级用户
	var isSuperuser bool
	err := e.db.QueryRow("SELECT usesuper FROM pg_user WHERE usename = current_user").Scan(&isSuperuser)
	if err != nil {
		return permissions, fmt.Errorf("检查权限失败: %w", err)
	}

	if isSuperuser {
		permissions["has_database_admin"] = true
		permissions["has_user_admin"] = true
		permissions["has_permission_admin"] = true
		return permissions, nil
	}

	// 检查是否有CREATEDB权限
	var canCreateDb bool
	err = e.db.QueryRow("SELECT usecreatedb FROM pg_user WHERE usename = current_user").Scan(&canCreateDb)
	if err == nil && canCreateDb {
		permissions["has_database_admin"] = true
	}

	// 检查是否有CREATEROLE权限
	var canCreateRole bool
	err = e.db.QueryRow("SELECT usecreaterole FROM pg_user WHERE usename = current_user").Scan(&canCreateRole)
	if err == nil && canCreateRole {
		permissions["has_user_admin"] = true
	}

	return permissions, nil
}

// CreateDatabase 创建数据库
// 参数说明：name=数据库名, charset=编码(PostgreSQL使用), collation=LC_COLLATE(PostgreSQL使用)
func (e *PostgreSQLExecutor) CreateDatabase(name string, charset string, collation string) error {
	if err := utils.ValidateDatabaseName(name); err != nil {
		return err
	}
	if err := utils.ValidateCharsetOrCollation(charset); err != nil {
		return err
	}
	if err := utils.ValidateCharsetOrCollation(collation); err != nil {
		return err
	}
	var sql strings.Builder
	sql.WriteString(fmt.Sprintf("CREATE DATABASE \"%s\"", name))

	// PostgreSQL使用encoding而不是charset
	if charset != "" {
		sql.WriteString(fmt.Sprintf(" WITH ENCODING '%s'", charset))
	}

	if collation != "" {
		sql.WriteString(fmt.Sprintf(" LC_COLLATE = '%s'", collation))
	}

	_, err := e.db.Exec(sql.String())
	return err
}

// DropDatabase 删除数据库
func (e *PostgreSQLExecutor) DropDatabase(name string) error {
	if err := utils.ValidateDatabaseName(name); err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP DATABASE \"%s\"", name)
	_, err := e.db.Exec(sql)
	return err
}

// GetDatabaseInfo 获取数据库信息
func (e *PostgreSQLExecutor) GetDatabaseInfo(name string) (map[string]interface{}, error) {
	if err := utils.ValidateDatabaseName(name); err != nil {
		return nil, err
	}
	// 注意：需要在目标数据库中查询表数量，这里简化处理
	query := `
		SELECT 
			d.datname,
			pg_encoding_to_char(d.encoding) AS encoding,
			d.datcollate,
			pg_database_size(d.datname) AS size
		FROM pg_database d
		WHERE d.datname = $1
	`

	var datname, encoding, datcollate sql.NullString
	var size sql.NullInt64

	err := e.db.QueryRow(query, name).Scan(&datname, &encoding, &datcollate, &size)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"name": datname.String,
	}

	if encoding.Valid {
		result["encoding"] = encoding.String
	}
	if datcollate.Valid {
		result["collation"] = datcollate.String
	}
	if size.Valid {
		result["size"] = size.Int64
	}
	// table_count需要切换到目标数据库查询，这里暂时不提供

	return result, nil
}

// RenameDatabase 重命名数据库
func (e *PostgreSQLExecutor) RenameDatabase(oldName, newName string) error {
	if err := utils.ValidateDatabaseName(oldName); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(newName); err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER DATABASE \"%s\" RENAME TO \"%s\"", oldName, newName)
	_, err := e.db.Exec(sql)
	return err
}

// GetUsers 获取用户列表
func (e *PostgreSQLExecutor) GetUsers() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			u.usename,
			u.usesuper,
			COALESCE(
				STRING_AGG(DISTINCT p.perm::text, ', '),
				''
			) AS privileges
		FROM pg_user u
		LEFT JOIN (
			SELECT 
				grantee,
				privilege_type AS perm
			FROM information_schema.role_table_grants
			WHERE grantee != 'PUBLIC'
			UNION
			SELECT 
				grantee,
				privilege_type AS perm
			FROM information_schema.role_usage_grants
			WHERE grantee != 'PUBLIC'
		) p ON u.usename = p.grantee
		GROUP BY u.usename, u.usesuper
		ORDER BY u.usename
	`

	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var username string
		var isSuperuser bool
		var privileges sql.NullString

		if err := rows.Scan(&username, &isSuperuser, &privileges); err != nil {
			continue
		}

		userInfo := map[string]interface{}{
			"username": username,
			"role":     "superuser",
		}
		if !isSuperuser {
			userInfo["role"] = "user"
		}

		if privileges.Valid {
			userInfo["privileges"] = privileges.String
		} else {
			userInfo["privileges"] = ""
		}

		users = append(users, userInfo)
	}

	return users, nil
}

// CreateUser 创建用户
func (e *PostgreSQLExecutor) CreateUser(username, host, password string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	sql := fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD $1", username)
	_, err := e.db.Exec(sql, password)
	return err
}

// DropUser 删除用户
func (e *PostgreSQLExecutor) DropUser(username, host string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP USER \"%s\"", username)
	_, err := e.db.Exec(sql)
	return err
}

// AlterUserPassword 修改用户密码
func (e *PostgreSQLExecutor) AlterUserPassword(username, host, newPassword string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER USER \"%s\" WITH PASSWORD $1", username)
	_, err := e.db.Exec(sql, newPassword)
	return err
}

// GetUserGrants 获取用户权限
func (e *PostgreSQLExecutor) GetUserGrants(username, host string) ([]string, error) {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return nil, err
	}
	query := `
		SELECT 
			'GRANT ' || privilege_type || ' ON ' || table_schema || '.' || table_name || ' TO ' || grantee AS grant_statement
		FROM information_schema.role_table_grants
		WHERE grantee = $1
		UNION
		SELECT 
			'GRANT ' || privilege_type || ' ON DATABASE ' || object_name || ' TO ' || grantee AS grant_statement
		FROM information_schema.role_usage_grants
		WHERE grantee = $1
	`

	rows, err := e.db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var grant string
		if err := rows.Scan(&grant); err != nil {
			continue
		}
		grants = append(grants, grant)
	}

	return grants, nil
}

// GrantDatabasePrivileges 授予数据库权限
func (e *PostgreSQLExecutor) GrantDatabasePrivileges(username, host, database string, privileges []string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("GRANT %s ON DATABASE \"%s\" TO \"%s\"", privs, database, username)
	_, err := e.db.Exec(sql)
	return err
}

// RevokeDatabasePrivileges 撤销数据库权限
func (e *PostgreSQLExecutor) RevokeDatabasePrivileges(username, host, database string, privileges []string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("REVOKE %s ON DATABASE \"%s\" FROM \"%s\"", privs, database, username)
	_, err := e.db.Exec(sql)
	return err
}

// GrantTablePrivileges 授予表权限
func (e *PostgreSQLExecutor) GrantTablePrivileges(username, host, database, table string, privileges []string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("GRANT %s ON TABLE \"%s\".\"%s\" TO \"%s\"", privs, "public", table, username)
	_, err := e.db.Exec(sql)
	return err
}

// RevokeTablePrivileges 撤销表权限
func (e *PostgreSQLExecutor) RevokeTablePrivileges(username, host, database, table string, privileges []string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("REVOKE %s ON TABLE \"%s\".\"%s\" FROM \"%s\"", privs, "public", table, username)
	_, err := e.db.Exec(sql)
	return err
}

// GrantColumnPrivileges 授予列权限
func (e *PostgreSQLExecutor) GrantColumnPrivileges(username, host, database, table string, columnPrivileges map[string][]string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	for col := range columnPrivileges {
		if err := utils.ValidateColumnName(col); err != nil {
			return err
		}
	}
	for col, privs := range columnPrivileges {
		privsStr := strings.Join(privs, ", ")
		sql := fmt.Sprintf("GRANT %s (%s) ON TABLE \"%s\".\"%s\" TO \"%s\"", privsStr, col, "public", table, username)
		_, err := e.db.Exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

// RevokeColumnPrivileges 撤销列权限
func (e *PostgreSQLExecutor) RevokeColumnPrivileges(username, host, database, table string, columnPrivileges map[string][]string) error {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	for col := range columnPrivileges {
		if err := utils.ValidateColumnName(col); err != nil {
			return err
		}
	}
	for col, privs := range columnPrivileges {
		privsStr := strings.Join(privs, ", ")
		sql := fmt.Sprintf("REVOKE %s (%s) ON TABLE \"%s\".\"%s\" FROM \"%s\"", privsStr, col, "public", table, username)
		_, err := e.db.Exec(sql)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPermissionTree 获取权限树
func (e *PostgreSQLExecutor) GetPermissionTree(username, host string) ([]map[string]interface{}, error) {
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return nil, err
	}
	// 获取用户的所有权限
	grants, err := e.GetUserGrants(username, host)
	if err != nil {
		return nil, err
	}

	// 获取所有数据库
	databases, err := e.GetDatabases()
	if err != nil {
		return nil, err
	}

	tree := make([]map[string]interface{}, 0)

	for _, db := range databases {
		dbNode := map[string]interface{}{
			"type":      "database",
			"name":      db,
			"path":      db,
			"privileges": []string{},
			"children":  []map[string]interface{}{},
		}

		// 检查该数据库的权限
		for _, grant := range grants {
			if strings.Contains(grant, db) {
				// 解析权限
				if strings.Contains(strings.ToUpper(grant), "SELECT") {
					privs := dbNode["privileges"].([]string)
					privs = append(privs, "SELECT")
					dbNode["privileges"] = privs
				}
				// ... 其他权限解析
			}
		}

		tree = append(tree, dbNode)
	}

	return tree, nil
}

// CreateTable 在当前数据库中建表（PostgreSQL 连接已绑定库）
func (e *PostgreSQLExecutor) CreateTable(database string, tableName string, createDDL string) error {
	if err := utils.ValidateTableName(tableName); err != nil {
		return err
	}
	if createDDL != "" {
		_, err := e.db.Exec(createDDL)
		return err
	}
	// 默认简单表：仅主键 id
	sql := fmt.Sprintf("CREATE TABLE \"%s\" ( id SERIAL PRIMARY KEY )", tableName)
	_, err := e.db.Exec(sql)
	return err
}

// DropTable 在当前数据库中删表
func (e *PostgreSQLExecutor) DropTable(database string, tableName string) error {
	if err := utils.ValidateTableName(tableName); err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
	_, err := e.db.Exec(sql)
	return err
}
