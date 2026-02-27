package sqlite

import (
	"database/sql"
	"time"

	"lingosql/internal/models"
)

type FavoriteDAO struct {
	db *sql.DB
}

func NewFavoriteDAO(db *sql.DB) *FavoriteDAO {
	return &FavoriteDAO{db: db}
}

// Create 创建收藏
func (dao *FavoriteDAO) Create(fav *models.Favorite) error {
	query := `INSERT INTO favorites (user_id, connection_id, database, name, sql_query, description, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := dao.db.Exec(query, fav.UserID, fav.ConnectionID, nullString(fav.Database), fav.Name,
		fav.SQLQuery, fav.Description, time.Now())
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	fav.ID = int(id)
	return nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// GetByID 根据 ID 获取收藏
func (dao *FavoriteDAO) GetByID(id int) (*models.Favorite, error) {
	fav := &models.Favorite{}
	query := `SELECT id, user_id, connection_id, database, name, sql_query, description, created_at, last_used_at 
	          FROM favorites WHERE id = ?`
	var dbNullable sql.NullString
	var lastUsedNullable sql.NullTime
	err := dao.db.QueryRow(query, id).Scan(
		&fav.ID, &fav.UserID, &fav.ConnectionID, &dbNullable, &fav.Name,
		&fav.SQLQuery, &fav.Description, &fav.CreatedAt, &lastUsedNullable,
	)
	if err != nil {
		return nil, err
	}
	if dbNullable.Valid {
		fav.Database = dbNullable.String
	}
	if lastUsedNullable.Valid {
		fav.LastUsedAt = &lastUsedNullable.Time
	}
	return fav, nil
}

// GetByUserID 获取用户的收藏列表，支持按 connection_id、database 筛选，按 sort 排序（created_at | last_used_at）
func (dao *FavoriteDAO) GetByUserID(userID int, connectionID *int, database *string, sort string) ([]*models.Favorite, error) {
	query := `SELECT id, user_id, connection_id, database, name, sql_query, description, created_at, last_used_at 
	          FROM favorites WHERE user_id = ?`
	args := []interface{}{userID}
	if connectionID != nil {
		query += ` AND connection_id = ?`
		args = append(args, *connectionID)
	}
	if database != nil {
		if *database == "" {
			query += ` AND (database IS NULL OR database = '')`
		} else {
			query += ` AND database = ?`
			args = append(args, *database)
		}
	}
	switch sort {
	case "last_used_at":
		query += ` ORDER BY COALESCE(last_used_at, '1970-01-01') DESC, created_at DESC`
	default:
		query += ` ORDER BY created_at DESC`
	}

	rows, err := dao.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []*models.Favorite
	for rows.Next() {
		fav := &models.Favorite{}
		var dbNullable sql.NullString
		var lastUsedNullable sql.NullTime
		err := rows.Scan(
			&fav.ID, &fav.UserID, &fav.ConnectionID, &dbNullable, &fav.Name,
			&fav.SQLQuery, &fav.Description, &fav.CreatedAt, &lastUsedNullable,
		)
		if err != nil {
			return nil, err
		}
		if dbNullable.Valid {
			fav.Database = dbNullable.String
		}
		if lastUsedNullable.Valid {
			fav.LastUsedAt = &lastUsedNullable.Time
		}
		favorites = append(favorites, fav)
	}
	return favorites, nil
}

// Update 更新收藏
func (dao *FavoriteDAO) Update(fav *models.Favorite) error {
	query := `UPDATE favorites SET name = ?, sql_query = ?, description = ? 
	          WHERE id = ? AND user_id = ?`
	
	_, err := dao.db.Exec(query, fav.Name, fav.SQLQuery, fav.Description, fav.ID, fav.UserID)
	return err
}

// Delete 删除收藏
func (dao *FavoriteDAO) Delete(id, userID int) error {
	query := `DELETE FROM favorites WHERE id = ? AND user_id = ?`
	_, err := dao.db.Exec(query, id, userID)
	return err
}

// UpdateLastUsedAt 更新最近使用时间
func (dao *FavoriteDAO) UpdateLastUsedAt(id, userID int, t time.Time) error {
	query := `UPDATE favorites SET last_used_at = ? WHERE id = ? AND user_id = ?`
	_, err := dao.db.Exec(query, t, id, userID)
	return err
}
