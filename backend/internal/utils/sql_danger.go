package utils

import "strings"

// IsDangerousSQL 判断是否包含危险 SQL（DROP/TRUNCATE/DELETE 无 WHERE）
func IsDangerousSQL(sql string) (bool, string) {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(upper, "DROP ") {
		return true, "检测到 DROP 语句"
	}
	if strings.HasPrefix(upper, "TRUNCATE ") {
		return true, "检测到 TRUNCATE 语句"
	}
	if strings.HasPrefix(upper, "DELETE ") && !strings.Contains(upper, " WHERE ") {
		return true, "检测到无 WHERE 的 DELETE 语句"
	}
	return false, ""
}
