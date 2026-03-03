package sqlite

import (
	"database/sql"
	"time"

	"lingosql/internal/models"
)

type UserDAO struct {
	db *sql.DB
}

func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{db: db}
}

// Create 创建用户
func (dao *UserDAO) Create(user *models.User) error {
	query := `INSERT INTO users (username, email, password_hash, failed_login_count, last_failed_login_at, locked_until, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := dao.db.Exec(query, user.Username, user.Email, user.PasswordHash, 
		0, nil, nil, time.Now(), time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

// CreateWithTx 在事务中创建用户
func (dao *UserDAO) CreateWithTx(tx *sql.Tx, user *models.User) error {
	query := `INSERT INTO users (username, email, password_hash, failed_login_count, last_failed_login_at, locked_until, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.Exec(query, user.Username, user.Email, user.PasswordHash,
		0, nil, nil, time.Now(), time.Now())
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)
	return nil
}

// GetByID 根据 ID 获取用户
func (dao *UserDAO) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, failed_login_count, last_failed_login_at, locked_until, created_at, updated_at 
	          FROM users WHERE id = ?`
	
	var lastFailed sql.NullTime
	var lockedUntil sql.NullTime
	err := dao.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FailedLoginCount, &lastFailed, &lockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastFailed.Valid {
		user.LastFailedLoginAt = &lastFailed.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (dao *UserDAO) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, failed_login_count, last_failed_login_at, locked_until, created_at, updated_at 
	          FROM users WHERE username = ?`
	
	var lastFailed sql.NullTime
	var lockedUntil sql.NullTime
	err := dao.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FailedLoginCount, &lastFailed, &lockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastFailed.Valid {
		user.LastFailedLoginAt = &lastFailed.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (dao *UserDAO) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, failed_login_count, last_failed_login_at, locked_until, created_at, updated_at 
	          FROM users WHERE email = ?`
	
	var lastFailed sql.NullTime
	var lockedUntil sql.NullTime
	err := dao.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FailedLoginCount, &lastFailed, &lockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if lastFailed.Valid {
		user.LastFailedLoginAt = &lastFailed.Time
	}
	if lockedUntil.Valid {
		user.LockedUntil = &lockedUntil.Time
	}

	return user, nil
}

// Count 返回用户总数
func (dao *UserDAO) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	err := dao.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ExistsByUsername 检查用户名是否存在
func (dao *UserDAO) ExistsByUsername(username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = ?`
	err := dao.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (dao *UserDAO) ExistsByEmail(email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = ?`
	err := dao.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLoginFailure 更新登录失败信息
func (dao *UserDAO) UpdateLoginFailure(userID int, failedCount int, lastFailed time.Time, lockedUntil *time.Time) error {
	query := `UPDATE users 
	          SET failed_login_count = ?, last_failed_login_at = ?, locked_until = ?, updated_at = ?
	          WHERE id = ?`
	_, err := dao.db.Exec(query, failedCount, lastFailed, lockedUntil, time.Now(), userID)
	return err
}

// UpdateLoginSuccess 重置登录失败信息
func (dao *UserDAO) UpdateLoginSuccess(userID int) error {
	query := `UPDATE users 
	          SET failed_login_count = 0, last_failed_login_at = NULL, locked_until = NULL, updated_at = ?
	          WHERE id = ?`
	_, err := dao.db.Exec(query, time.Now(), userID)
	return err
}
