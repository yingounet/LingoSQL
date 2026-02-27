package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"lingosql/internal/models"
)

type ConnectionDAO struct {
	db *sql.DB
}

func NewConnectionDAO(db *sql.DB) *ConnectionDAO {
	return &ConnectionDAO{db: db}
}

// Create 创建连接
func (dao *ConnectionDAO) Create(conn *models.Connection) error {
	query := `INSERT INTO connections 
	          (user_id, name, db_type, connection_type, db_config, ssh_config, is_default, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := dao.db.Exec(query, conn.UserID, conn.Name, conn.DBType, conn.ConnectionType,
		conn.DbConfigJSON, conn.SshConfigJSON, conn.IsDefault, time.Now(), time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	conn.ID = int(id)
	conn.CreatedAt = time.Now()
	conn.UpdatedAt = time.Now()
	return nil
}

// GetByID 根据 ID 获取连接
func (dao *ConnectionDAO) GetByID(id int) (*models.Connection, error) {
	conn := &models.Connection{}
	query := `SELECT id, user_id, name, db_type, connection_type, db_config, ssh_config, 
	          is_default, last_used_at, created_at, updated_at 
	          FROM connections WHERE id = ?`

	var sshConfig sql.NullString
	var lastUsedAt sql.NullTime
	err := dao.db.QueryRow(query, id).Scan(
		&conn.ID, &conn.UserID, &conn.Name, &conn.DBType, &conn.ConnectionType,
		&conn.DbConfigJSON, &sshConfig, &conn.IsDefault,
		&lastUsedAt, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if sshConfig.Valid {
		conn.SshConfigJSON = &sshConfig.String
	}
	if lastUsedAt.Valid {
		conn.LastUsedAt = &lastUsedAt.Time
	}

	return conn, nil
}

// GetByUserID 获取用户的所有连接（带分页和筛选）
func (dao *ConnectionDAO) GetByUserID(userID int, params *models.ConnectionListParams) ([]*models.Connection, int, error) {
	// 构建查询条件
	conditions := []string{"user_id = ?"}
	args := []interface{}{userID}

	if params.DbType != "" {
		conditions = append(conditions, "db_type = ?")
		args = append(args, params.DbType)
	}

	if params.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR db_config LIKE ?)")
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	whereClause := strings.Join(conditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM connections WHERE %s`, whereClause)
	var total int
	if err := dao.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页参数
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 获取数据，按最后连接时间倒序（从未连接的排最后）
	query := fmt.Sprintf(`SELECT id, user_id, name, db_type, connection_type, db_config, ssh_config, 
	          is_default, last_used_at, created_at, updated_at 
	          FROM connections WHERE %s 
	          ORDER BY COALESCE(last_used_at, '1970-01-01') DESC, is_default DESC, updated_at DESC 
	          LIMIT ? OFFSET ?`, whereClause)

	args = append(args, pageSize, offset)
	rows, err := dao.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var connections []*models.Connection
	for rows.Next() {
		conn := &models.Connection{}
		var sshConfig sql.NullString
		var lastUsedAt sql.NullTime
		err := rows.Scan(
			&conn.ID, &conn.UserID, &conn.Name, &conn.DBType, &conn.ConnectionType,
			&conn.DbConfigJSON, &sshConfig, &conn.IsDefault,
			&lastUsedAt, &conn.CreatedAt, &conn.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if sshConfig.Valid {
			conn.SshConfigJSON = &sshConfig.String
		}
		if lastUsedAt.Valid {
			conn.LastUsedAt = &lastUsedAt.Time
		}
		connections = append(connections, conn)
	}

	return connections, total, nil
}

// Update 更新连接
func (dao *ConnectionDAO) Update(conn *models.Connection) error {
	query := `UPDATE connections 
	          SET name = ?, db_type = ?, connection_type = ?, db_config = ?, 
	              ssh_config = ?, is_default = ?, updated_at = ? 
	          WHERE id = ? AND user_id = ?`

	_, err := dao.db.Exec(query, conn.Name, conn.DBType, conn.ConnectionType,
		conn.DbConfigJSON, conn.SshConfigJSON, conn.IsDefault,
		time.Now(), conn.ID, conn.UserID)
	if err != nil {
		return err
	}

	conn.UpdatedAt = time.Now()
	return nil
}

// Delete 删除连接
func (dao *ConnectionDAO) Delete(id, userID int) error {
	query := `DELETE FROM connections WHERE id = ? AND user_id = ?`
	_, err := dao.db.Exec(query, id, userID)
	return err
}

// SetDefault 设置默认连接
func (dao *ConnectionDAO) SetDefault(id, userID int) error {
	// 开始事务
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 先清除该用户的所有默认连接
	_, err = tx.Exec(`UPDATE connections SET is_default = 0, updated_at = ? WHERE user_id = ?`,
		time.Now(), userID)
	if err != nil {
		return err
	}

	// 设置新的默认连接
	_, err = tx.Exec(`UPDATE connections SET is_default = 1, updated_at = ? WHERE id = ? AND user_id = ?`,
		time.Now(), id, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Exists 检查连接是否存在
func (dao *ConnectionDAO) Exists(id, userID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM connections WHERE id = ? AND user_id = ?`
	err := dao.db.QueryRow(query, id, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLastUsedAt 更新连接的最后使用时间
func (dao *ConnectionDAO) UpdateLastUsedAt(id, userID int, t time.Time) error {
	query := `UPDATE connections SET last_used_at = ? WHERE id = ? AND user_id = ?`
	_, err := dao.db.Exec(query, t, id, userID)
	return err
}

// GetUserConnectionCount 获取用户连接数量
func (dao *ConnectionDAO) GetUserConnectionCount(userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM connections WHERE user_id = ?`
	err := dao.db.QueryRow(query, userID).Scan(&count)
	return count, err
}
