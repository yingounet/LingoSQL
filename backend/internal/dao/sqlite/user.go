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
	query := `INSERT INTO users (username, email, password_hash, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?)`
	
	result, err := dao.db.Exec(query, user.Username, user.Email, user.PasswordHash, 
		time.Now(), time.Now())
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
	query := `SELECT id, username, email, password_hash, created_at, updated_at 
	          FROM users WHERE id = ?`
	
	err := dao.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (dao *UserDAO) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, created_at, updated_at 
	          FROM users WHERE username = ?`
	
	err := dao.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (dao *UserDAO) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, created_at, updated_at 
	          FROM users WHERE email = ?`
	
	err := dao.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
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
