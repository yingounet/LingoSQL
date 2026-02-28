package sqlite

import (
	"database/sql"
	"time"

	"lingosql/internal/models"
)

type AuditDAO struct {
	db *sql.DB
}

func NewAuditDAO(db *sql.DB) *AuditDAO {
	return &AuditDAO{db: db}
}

func (dao *AuditDAO) Create(log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (user_id, action, resource_type, resource_id, success, error_message, details, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := dao.db.Exec(query, log.UserID, log.Action, log.ResourceType, log.ResourceID, log.Success, log.ErrorMessage, log.Details, time.Now())
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = int(id)
	return nil
}

func (dao *AuditDAO) ListByUser(userID int, action string, page, pageSize int) ([]models.AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	if action == "" {
		if err := dao.db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE user_id = ?`, userID).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := dao.db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE user_id = ? AND action = ?`, userID, action).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	var rows *sql.Rows
	var err error
	if action == "" {
		rows, err = dao.db.Query(`SELECT id, user_id, action, resource_type, resource_id, success, error_message, details, created_at
			FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, userID, pageSize, offset)
	} else {
		rows, err = dao.db.Query(`SELECT id, user_id, action, resource_type, resource_id, success, error_message, details, created_at
			FROM audit_logs WHERE user_id = ? AND action = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, userID, action, pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]models.AuditLog, 0)
	for rows.Next() {
		var log models.AuditLog
		var resourceID sql.NullInt64
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.ResourceType, &resourceID,
			&log.Success, &log.ErrorMessage, &log.Details, &log.CreatedAt,
		); err != nil {
			continue
		}
		if resourceID.Valid {
			id := int(resourceID.Int64)
			log.ResourceID = &id
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
