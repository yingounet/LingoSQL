package utils

import (
	"fmt"
	"regexp"
)

// 数据库标识符（库名、表名、列名）允许：字母、数字、下划线，长度 1~64（与 MySQL 一致）
var safeIdentRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,64}$`)

// IsSafeDatabaseName 校验数据库名是否安全（防 SQL 注入）
func IsSafeDatabaseName(name string) bool {
	return safeIdentRegex.MatchString(name)
}

// IsSafeTableName 校验表名是否安全
func IsSafeTableName(name string) bool {
	return safeIdentRegex.MatchString(name)
}

// IsSafeIdentifier 校验通用标识符（列名等）
func IsSafeIdentifier(name string) bool {
	return safeIdentRegex.MatchString(name)
}

// ValidateDatabaseName 校验数据库名，不安全则返回 error
func ValidateDatabaseName(name string) error {
	if name == "" {
		return fmt.Errorf("数据库名不能为空")
	}
	if !IsSafeDatabaseName(name) {
		return fmt.Errorf("数据库名仅允许字母、数字、下划线且长度 1~64")
	}
	return nil
}

// ValidateTableName 校验表名
func ValidateTableName(name string) error {
	if name == "" {
		return fmt.Errorf("表名不能为空")
	}
	if !IsSafeTableName(name) {
		return fmt.Errorf("表名仅允许字母、数字、下划线且长度 1~64")
	}
	return nil
}

// ValidateColumnName 校验列名
func ValidateColumnName(name string) error {
	if name == "" {
		return fmt.Errorf("列名不能为空")
	}
	if !IsSafeIdentifier(name) {
		return fmt.Errorf("列名仅允许字母、数字、下划线且长度 1~64")
	}
	return nil
}

// MySQL 用户名仅允许字母数字下划线；host 允许 localhost、%、IP 等
var mysqlUserRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,32}$`)
var mysqlHostRegex = regexp.MustCompile(`^[a-zA-Z0-9_.%\-]{1,255}$`)

// ValidateMySQLUsername 校验 MySQL 用户名（用于 CREATE USER 等）
func ValidateMySQLUsername(name string) error {
	if name == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if !mysqlUserRegex.MatchString(name) {
		return fmt.Errorf("用户名仅允许字母、数字、下划线且长度 1~32")
	}
	return nil
}

// ValidateMySQLHost 校验 MySQL 主机（用于 CREATE USER 等）
func ValidateMySQLHost(host string) error {
	if host == "" {
		return fmt.Errorf("主机不能为空")
	}
	if !mysqlHostRegex.MatchString(host) {
		return fmt.Errorf("主机名仅允许字母、数字、下划线、点、%%、横线且长度 1~255")
	}
	return nil
}

// ValidateCharsetOrCollation 校验字符集/排序规则名
func ValidateCharsetOrCollation(name string) error {
	if name == "" {
		return nil
	}
	if !safeIdentRegex.MatchString(name) {
		return fmt.Errorf("字符集/排序规则仅允许字母、数字、下划线且长度 1~64")
	}
	return nil
}
