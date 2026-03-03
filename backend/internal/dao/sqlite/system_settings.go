package sqlite

import (
	"database/sql"
	"time"

	"lingosql/internal/models"
)

type SystemSettingsDAO struct {
	db *sql.DB
}

func NewSystemSettingsDAO(db *sql.DB) *SystemSettingsDAO {
	return &SystemSettingsDAO{db: db}
}

// Set 设置配置项
func (dao *SystemSettingsDAO) Set(key, value string) error {
	query := `INSERT INTO system_settings (key, value, updated_at) VALUES (?, ?, ?)
	          ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := dao.db.Exec(query, key, value, time.Now())
	return err
}

// SetWithTx 在事务中设置配置项
func (dao *SystemSettingsDAO) SetWithTx(tx *sql.Tx, key, value string) error {
	query := `INSERT INTO system_settings (key, value, updated_at) VALUES (?, ?, ?)
	          ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`
	_, err := tx.Exec(query, key, value, time.Now())
	return err
}

// Get 获取配置项
func (dao *SystemSettingsDAO) Get(key string) (string, error) {
	var value string
	query := `SELECT value FROM system_settings WHERE key = ?`
	err := dao.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// GetAll 获取所有配置
func (dao *SystemSettingsDAO) GetAll() (map[string]string, error) {
	query := `SELECT key, value FROM system_settings`
	rows, err := dao.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var s models.SystemSetting
		if err := rows.Scan(&s.Key, &s.Value); err != nil {
			return nil, err
		}
		result[s.Key] = s.Value
	}
	return result, rows.Err()
}
