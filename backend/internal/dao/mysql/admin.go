package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"lingosql/internal/utils"
)

// CheckAdminPermissions 检查管理员权限
func (e *MySQLExecutor) CheckAdminPermissions() (map[string]bool, error) {
	permissions := make(map[string]bool)

	// 首先检查是否是超级用户（Super_priv = 'Y'）
	// 使用 SUBSTRING_INDEX(USER(), '@', 1) 获取用户名部分
	var superPriv string
	var createUserPriv string
	var createPriv string
	var grantPriv string
	
	err := e.db.QueryRow(`
		SELECT 
			MAX(CASE WHEN Super_priv = 'Y' THEN 'Y' ELSE 'N' END) AS super_priv,
			MAX(CASE WHEN Create_user_priv = 'Y' THEN 'Y' ELSE 'N' END) AS create_user_priv,
			MAX(CASE WHEN Create_priv = 'Y' THEN 'Y' ELSE 'N' END) AS create_priv,
			MAX(CASE WHEN Grant_priv = 'Y' THEN 'Y' ELSE 'N' END) AS grant_priv
		FROM mysql.user 
		WHERE User = SUBSTRING_INDEX(USER(), '@', 1)
	`).Scan(&superPriv, &createUserPriv, &createPriv, &grantPriv)
	
	if err != nil {
		// 如果查询失败，尝试使用 SHOW GRANTS 方式
		rows, err2 := e.db.Query("SHOW GRANTS FOR CURRENT_USER()")
		if err2 != nil {
			return permissions, fmt.Errorf("检查权限失败: %w", err2)
		}
		defer rows.Close()

		var grants string
		for rows.Next() {
			if err := rows.Scan(&grants); err != nil {
				continue
			}
			grantUpper := strings.ToUpper(grants)
			if strings.Contains(grantUpper, "GRANT OPTION") || strings.Contains(grantUpper, "WITH GRANT OPTION") {
				permissions["has_permission_admin"] = true
			}
			if strings.Contains(grantUpper, "CREATE USER") {
				permissions["has_user_admin"] = true
			}
			if strings.Contains(grantUpper, "CREATE") && strings.Contains(grantUpper, "DATABASE") {
				permissions["has_database_admin"] = true
			}
			// 检查是否是 ALL PRIVILEGES 或 SUPER 权限
			if strings.Contains(grantUpper, "ALL PRIVILEGES") || strings.Contains(grantUpper, "SUPER") {
				permissions["has_database_admin"] = true
				permissions["has_user_admin"] = true
				permissions["has_permission_admin"] = true
			}
		}

		return permissions, nil
	}

	// 如果是超级用户，拥有所有权限
	if superPriv == "Y" {
		permissions["has_database_admin"] = true
		permissions["has_user_admin"] = true
		permissions["has_permission_admin"] = true
		return permissions, nil
	}

	// 检查具体权限
	if createUserPriv == "Y" {
		permissions["has_user_admin"] = true
	}

	if createPriv == "Y" {
		permissions["has_database_admin"] = true
	}

	if grantPriv == "Y" {
		permissions["has_permission_admin"] = true
	}

	return permissions, nil
}

// CreateDatabase 创建数据库
func (e *MySQLExecutor) CreateDatabase(name string, charset string, collation string) error {
	if err := utils.ValidateDatabaseName(name); err != nil {
		return err
	}
	if err := utils.ValidateCharsetOrCollation(charset); err != nil {
		return err
	}
	if err := utils.ValidateCharsetOrCollation(collation); err != nil {
		return err
	}
	var sql string
	if charset != "" && collation != "" {
		sql = fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s", name, charset, collation)
	} else if charset != "" {
		sql = fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s", name, charset)
	} else {
		sql = fmt.Sprintf("CREATE DATABASE `%s`", name)
	}

	_, err := e.db.Exec(sql)
	return err
}

// DropDatabase 删除数据库
func (e *MySQLExecutor) DropDatabase(name string) error {
	if err := utils.ValidateDatabaseName(name); err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP DATABASE `%s`", name)
	_, err := e.db.Exec(sql)
	return err
}

// GetDatabaseInfo 获取数据库信息
func (e *MySQLExecutor) GetDatabaseInfo(name string) (map[string]interface{}, error) {
	if err := utils.ValidateDatabaseName(name); err != nil {
		return nil, err
	}
	// 获取数据库基本信息
	query1 := fmt.Sprintf(`
		SELECT 
			SCHEMA_NAME,
			DEFAULT_CHARACTER_SET_NAME,
			DEFAULT_COLLATION_NAME
		FROM information_schema.SCHEMATA
		WHERE SCHEMA_NAME = '%s'
	`, name)

	var schemaName, charset, collation sql.NullString
	err := e.db.QueryRow(query1).Scan(&schemaName, &charset, &collation)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"name": schemaName.String,
	}

	if charset.Valid {
		result["charset"] = charset.String
	}
	if collation.Valid {
		result["collation"] = collation.String
	}

	// 获取数据库大小和表数量
	query2 := fmt.Sprintf(`
		SELECT 
			ROUND(SUM(DATA_LENGTH + INDEX_LENGTH), 0) AS size_bytes,
			COUNT(DISTINCT TABLE_NAME) AS table_count
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = '%s'
	`, name)

	var size sql.NullInt64
	var tableCount sql.NullInt64
	err = e.db.QueryRow(query2).Scan(&size, &tableCount)
	if err == nil {
		if size.Valid {
			result["size"] = size.Int64
		}
		if tableCount.Valid {
			result["table_count"] = tableCount.Int64
		}
	}

	return result, nil
}

// RenameDatabase MySQL不支持重命名数据库
func (e *MySQLExecutor) RenameDatabase(oldName, newName string) error {
	return fmt.Errorf("MySQL不支持重命名数据库，请使用导出导入方式")
}

// GetUsers 获取用户列表
func (e *MySQLExecutor) GetUsers() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			User,
			Host,
			GROUP_CONCAT(DISTINCT 
				CASE 
					WHEN Select_priv = 'Y' THEN 'SELECT'
					WHEN Insert_priv = 'Y' THEN 'INSERT'
					WHEN Update_priv = 'Y' THEN 'UPDATE'
					WHEN Delete_priv = 'Y' THEN 'DELETE'
					WHEN Create_priv = 'Y' THEN 'CREATE'
					WHEN Drop_priv = 'Y' THEN 'DROP'
				END
			) AS privileges
		FROM mysql.user
		WHERE User != ''
		GROUP BY User, Host
		ORDER BY User, Host
	`

	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var user, host string
		var privileges sql.NullString

		if err := rows.Scan(&user, &host, &privileges); err != nil {
			continue
		}

		userInfo := map[string]interface{}{
			"username": user,
			"host":     host,
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
func (e *MySQLExecutor) CreateUser(username, host, password string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	sql := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, password)
	_, err := e.db.Exec(sql)
	return err
}

// DropUser 删除用户
func (e *MySQLExecutor) DropUser(username, host string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP USER '%s'@'%s'", username, host)
	_, err := e.db.Exec(sql)
	return err
}

// AlterUserPassword 修改用户密码
func (e *MySQLExecutor) AlterUserPassword(username, host, newPassword string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	sql := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", username, host, newPassword)
	_, err := e.db.Exec(sql)
	return err
}

// GetUserGrants 获取用户权限
func (e *MySQLExecutor) GetUserGrants(username, host string) ([]string, error) {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return nil, err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", username, host)

	rows, err := e.db.Query(query)
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
func (e *MySQLExecutor) GrantDatabasePrivileges(username, host, database string, privileges []string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("GRANT %s ON `%s`.* TO '%s'@'%s'", privs, database, username, host)
	_, err := e.db.Exec(sql)
	return err
}

// RevokeDatabasePrivileges 撤销数据库权限
func (e *MySQLExecutor) RevokeDatabasePrivileges(username, host, database string, privileges []string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("REVOKE %s ON `%s`.* FROM '%s'@'%s'", privs, database, username, host)
	_, err := e.db.Exec(sql)
	return err
}

// GrantTablePrivileges 授予表权限
func (e *MySQLExecutor) GrantTablePrivileges(username, host, database, table string, privileges []string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("GRANT %s ON `%s`.`%s` TO '%s'@'%s'", privs, database, table, username, host)
	_, err := e.db.Exec(sql)
	return err
}

// RevokeTablePrivileges 撤销表权限
func (e *MySQLExecutor) RevokeTablePrivileges(username, host, database, table string, privileges []string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	privs := strings.Join(privileges, ", ")
	sql := fmt.Sprintf("REVOKE %s ON `%s`.`%s` FROM '%s'@'%s'", privs, database, table, username, host)
	_, err := e.db.Exec(sql)
	return err
}

// GrantColumnPrivileges 授予列权限
func (e *MySQLExecutor) GrantColumnPrivileges(username, host, database, table string, columnPrivileges map[string][]string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
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

	var parts []string
	for col, privs := range columnPrivileges {
		privsStr := strings.Join(privs, ", ")
		parts = append(parts, fmt.Sprintf("%s (%s)", privsStr, col))
	}

	privsStr := strings.Join(parts, ", ")
	sql := fmt.Sprintf("GRANT %s ON `%s`.`%s` TO '%s'@'%s'", privsStr, database, table, username, host)
	_, err := e.db.Exec(sql)
	return err
}

// RevokeColumnPrivileges 撤销列权限
func (e *MySQLExecutor) RevokeColumnPrivileges(username, host, database, table string, columnPrivileges map[string][]string) error {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return err
	}
	if err := utils.ValidateDatabaseName(database); err != nil {
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

	var parts []string
	for col, privs := range columnPrivileges {
		privsStr := strings.Join(privs, ", ")
		parts = append(parts, fmt.Sprintf("%s (%s)", privsStr, col))
	}

	privsStr := strings.Join(parts, ", ")
	sql := fmt.Sprintf("REVOKE %s ON `%s`.`%s` FROM '%s'@'%s'", privsStr, database, table, username, host)
	_, err := e.db.Exec(sql)
	return err
}

// GetPermissionTree 获取权限树
func (e *MySQLExecutor) GetPermissionTree(username, host string) ([]map[string]interface{}, error) {
	if host == "" {
		host = "localhost"
	}
	if err := utils.ValidateMySQLUsername(username); err != nil {
		return nil, err
	}
	if err := utils.ValidateMySQLHost(host); err != nil {
		return nil, err
	}
	// 获取用户的所有权限
	grants, err := e.GetUserGrants(username, host)
	if err != nil {
		return nil, err
	}

	// 解析权限并构建树结构
	// 这里简化实现，返回一个扁平结构
	// 实际应该解析GRANT语句并构建树形结构
	tree := make([]map[string]interface{}, 0)

	// 获取所有数据库
	databases, err := e.GetDatabases()
	if err != nil {
		return nil, err
	}

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
			if strings.Contains(grant, db+".") || strings.Contains(grant, db+".*") {
				// 解析权限
				// 简化处理，实际需要更复杂的解析
				if strings.Contains(strings.ToUpper(grant), "SELECT") {
					dbNode["privileges"] = append(dbNode["privileges"].([]string), "SELECT")
				}
				// ... 其他权限
			}
		}

		tree = append(tree, dbNode)
	}

	return tree, nil
}

// CreateTable 在指定数据库中建表（调用方已用该 database 建连时可直接执行；否则需带库名）
func (e *MySQLExecutor) CreateTable(database string, tableName string, createDDL string) error {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	if err := utils.ValidateTableName(tableName); err != nil {
		return err
	}
	if createDDL != "" {
		_, err := e.db.Exec(createDDL)
		return err
	}
	sql := fmt.Sprintf("CREATE TABLE `%s`.`%s` ( `id` INT NOT NULL AUTO_INCREMENT, PRIMARY KEY (`id`)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4", database, tableName)
	_, err := e.db.Exec(sql)
	return err
}

// DropTable 在指定数据库中删表
func (e *MySQLExecutor) DropTable(database string, tableName string) error {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return err
	}
	if err := utils.ValidateTableName(tableName); err != nil {
		return err
	}
	sql := fmt.Sprintf("DROP TABLE `%s`.`%s`", database, tableName)
	_, err := e.db.Exec(sql)
	return err
}
