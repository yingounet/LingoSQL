package sqlite

import (
	"database/sql"
	"time"

	"lingosql/internal/models"
)

type SystemHistoryDAO struct {
	db *sql.DB
}

func NewSystemHistoryDAO(db *sql.DB) *SystemHistoryDAO {
	return &SystemHistoryDAO{db: db}
}

// Create 创建系统执行记录
func (dao *SystemHistoryDAO) Create(history *models.SystemQueryHistory) error {
	query := `INSERT INTO system_query_history 
	          (connection_id, user_id, sql_query, operation_type, execution_time_ms, rows_affected, success, error_message, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := dao.db.Exec(query, history.ConnectionID, history.UserID, history.SQLQuery,
		history.OperationType, history.ExecutionTimeMs, history.RowsAffected, history.Success, history.ErrorMessage, time.Now())
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

// GetByID 根据 ID 获取系统执行记录
func (dao *SystemHistoryDAO) GetByID(id int) (*models.SystemQueryHistory, error) {
	history := &models.SystemQueryHistory{}
	query := `SELECT id, connection_id, user_id, sql_query, operation_type, execution_time_ms, 
	          rows_affected, success, error_message, created_at 
	          FROM system_query_history WHERE id = ?`
	
	err := dao.db.QueryRow(query, id).Scan(
		&history.ID, &history.ConnectionID, &history.UserID, &history.SQLQuery,
		&history.OperationType, &history.ExecutionTimeMs, &history.RowsAffected, &history.Success,
		&history.ErrorMessage, &history.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetByConnectionID 根据 connection_id 获取系统执行记录（分页）
func (dao *SystemHistoryDAO) GetByConnectionID(connectionID, userID int, page, pageSize int) ([]*models.SystemQueryHistory, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM system_query_history WHERE connection_id = ? AND user_id = ?`
	listQuery := `SELECT id, connection_id, user_id, sql_query, operation_type, execution_time_ms, 
	              rows_affected, success, error_message, created_at 
	              FROM system_query_history WHERE connection_id = ? AND user_id = ? 
	              ORDER BY created_at DESC LIMIT ? OFFSET ?`

	// 获取总数
	var total int
	err := dao.db.QueryRow(countQuery, connectionID, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	rows, err := dao.db.Query(listQuery, connectionID, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var histories []*models.SystemQueryHistory
	for rows.Next() {
		history := &models.SystemQueryHistory{}
		err := rows.Scan(
			&history.ID, &history.ConnectionID, &history.UserID, &history.SQLQuery,
			&history.OperationType, &history.ExecutionTimeMs, &history.RowsAffected, &history.Success,
			&history.ErrorMessage, &history.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		histories = append(histories, history)
	}

	return histories, total, nil
}

// Delete 删除系统执行记录
func (dao *SystemHistoryDAO) Delete(id, userID int) error {
	query := `DELETE FROM system_query_history WHERE id = ? AND user_id = ?`
	_, err := dao.db.Exec(query, id, userID)
	return err
}
