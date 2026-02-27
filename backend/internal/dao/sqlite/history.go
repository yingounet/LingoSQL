package sqlite

import (
	"database/sql"
	"time"

	"lingosql/internal/models"
)

type HistoryDAO struct {
	db *sql.DB
}

func NewHistoryDAO(db *sql.DB) *HistoryDAO {
	return &HistoryDAO{db: db}
}

// Create 创建历史记录
func (dao *HistoryDAO) Create(history *models.QueryHistory) error {
	query := `INSERT INTO query_history 
	          (user_id, connection_id, sql_query, execution_time_ms, rows_affected, success, error_message, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := dao.db.Exec(query, history.UserID, history.ConnectionID, history.SQLQuery,
		history.ExecutionTimeMs, history.RowsAffected, history.Success, history.ErrorMessage, time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	history.ID = int(id)
	return nil
}

// GetByID 根据 ID 获取历史记录
func (dao *HistoryDAO) GetByID(id int) (*models.QueryHistory, error) {
	history := &models.QueryHistory{}
	query := `SELECT id, user_id, connection_id, sql_query, execution_time_ms, 
	          rows_affected, success, error_message, created_at 
	          FROM query_history WHERE id = ?`
	
	err := dao.db.QueryRow(query, id).Scan(
		&history.ID, &history.UserID, &history.ConnectionID, &history.SQLQuery,
		&history.ExecutionTimeMs, &history.RowsAffected, &history.Success,
		&history.ErrorMessage, &history.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetByUserID 获取用户的历史记录（分页）
func (dao *HistoryDAO) GetByUserID(userID int, connectionID *int, page, pageSize int) ([]*models.QueryHistory, int, error) {
	var countQuery string
	var listQuery string
	var args []interface{}

	offset := (page - 1) * pageSize

	if connectionID != nil {
		countQuery = `SELECT COUNT(*) FROM query_history WHERE user_id = ? AND connection_id = ?`
		listQuery = `SELECT id, user_id, connection_id, sql_query, execution_time_ms, 
		             rows_affected, success, error_message, created_at 
		             FROM query_history WHERE user_id = ? AND connection_id = ? 
		             ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{userID, *connectionID, pageSize, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM query_history WHERE user_id = ?`
		listQuery = `SELECT id, user_id, connection_id, sql_query, execution_time_ms, 
		             rows_affected, success, error_message, created_at 
		             FROM query_history WHERE user_id = ? 
		             ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{userID, pageSize, offset}
	}

	// 获取总数
	var total int
	countArgs := args[:len(args)-2] // 移除 LIMIT 和 OFFSET 参数
	err := dao.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	rows, err := dao.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var histories []*models.QueryHistory
	for rows.Next() {
		history := &models.QueryHistory{}
		err := rows.Scan(
			&history.ID, &history.UserID, &history.ConnectionID, &history.SQLQuery,
			&history.ExecutionTimeMs, &history.RowsAffected, &history.Success,
			&history.ErrorMessage, &history.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		histories = append(histories, history)
	}

	return histories, total, nil
}

// Delete 删除历史记录
func (dao *HistoryDAO) Delete(id, userID int) error {
	query := `DELETE FROM query_history WHERE id = ? AND user_id = ?`
	_, err := dao.db.Exec(query, id, userID)
	return err
}

// DeleteByUserID 删除用户的所有历史记录
func (dao *HistoryDAO) DeleteByUserID(userID int, connectionID *int) error {
	if connectionID != nil {
		query := `DELETE FROM query_history WHERE user_id = ? AND connection_id = ?`
		_, err := dao.db.Exec(query, userID, *connectionID)
		return err
	} else {
		query := `DELETE FROM query_history WHERE user_id = ?`
		_, err := dao.db.Exec(query, userID)
		return err
	}
}
