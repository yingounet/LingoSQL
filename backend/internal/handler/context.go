package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/utils"
)

// GetUserID 从上下文安全获取当前用户 ID（由 AuthMiddleware 设置）。
// 若未认证或类型错误，会写入 401 并 abort，返回 (0, false)；否则返回 (userID, true)。
func GetUserID(c *gin.Context) (int, bool) {
	v, exists := c.Get("user_id")
	if !exists || v == nil {
		utils.Unauthorized(c, "未认证")
		c.Abort()
		return 0, false
	}
	userID, ok := v.(int)
	if !ok {
		utils.Unauthorized(c, "未认证")
		c.Abort()
		return 0, false
	}
	return userID, true
}

// ParseConnectionID 从 query 解析 connection_id，无效时写入 400 并返回 (0, false)。
func ParseConnectionID(c *gin.Context) (int, bool) {
	s := c.Query("connection_id")
	if s == "" {
		utils.BadRequest(c, "connection_id 不能为空")
		return 0, false
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		utils.BadRequest(c, "无效的连接 ID")
		return 0, false
	}
	return id, true
}

// ParseDatabaseTable 从 query 解析 database 和 table，缺失时写入 400 并返回 ( "", "", false)。
func ParseDatabaseTable(c *gin.Context) (database, table string, ok bool) {
	database = c.Query("database")
	if database == "" {
		utils.BadRequest(c, "数据库名不能为空")
		return "", "", false
	}
	table = c.Query("table")
	if table == "" {
		utils.BadRequest(c, "表名不能为空")
		return "", "", false
	}
	return database, table, true
}
