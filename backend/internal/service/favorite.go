package service

import (
	"errors"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
)

type FavoriteService struct {
	favoriteDAO    *sqlite.FavoriteDAO
	connectionDAO  *sqlite.ConnectionDAO
}

func NewFavoriteService(favoriteDAO *sqlite.FavoriteDAO, connectionDAO *sqlite.ConnectionDAO) *FavoriteService {
	return &FavoriteService{
		favoriteDAO:   favoriteDAO,
		connectionDAO: connectionDAO,
	}
}

// Create 创建收藏
func (s *FavoriteService) Create(userID int, req *models.FavoriteCreateRequest) (*models.Favorite, error) {
	// 验证连接是否属于当前用户
	conn, err := s.connectionDAO.GetByID(req.ConnectionID)
	if err != nil {
		return nil, err
	}
	if conn.UserID != userID {
		return nil, errors.New("无权访问此连接")
	}

	fav := &models.Favorite{
		UserID:       userID,
		ConnectionID: req.ConnectionID,
		Database:     req.Database,
		Name:         req.Name,
		SQLQuery:     req.SQLQuery,
		Description:  req.Description,
	}

	if err := s.favoriteDAO.Create(fav); err != nil {
		return nil, err
	}

	return fav, nil
}

// GetByUserID 获取用户的收藏列表，支持 connectionID、database 筛选与 sort 排序
func (s *FavoriteService) GetByUserID(userID int, connectionID *int, database *string, sort string) ([]*models.Favorite, error) {
	return s.favoriteDAO.GetByUserID(userID, connectionID, database, sort)
}

// GetConnectionName 返回连接名称，用于列表展示
func (s *FavoriteService) GetConnectionName(connectionID int) (string, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return "", err
	}
	return conn.Name, nil
}

// RecordUse 记录收藏被使用，更新 last_used_at
func (s *FavoriteService) RecordUse(id, userID int) error {
	fav, err := s.favoriteDAO.GetByID(id)
	if err != nil {
		return err
	}
	if fav.UserID != userID {
		return errors.New("无权访问此收藏")
	}
	return s.favoriteDAO.UpdateLastUsedAt(id, userID, time.Now())
}

// GetByID 获取收藏
func (s *FavoriteService) GetByID(id, userID int) (*models.Favorite, error) {
	fav, err := s.favoriteDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	if fav.UserID != userID {
		return nil, errors.New("无权访问此收藏")
	}

	return fav, nil
}

// Update 更新收藏
func (s *FavoriteService) Update(id, userID int, req *models.FavoriteUpdateRequest) (*models.Favorite, error) {
	fav, err := s.favoriteDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	if fav.UserID != userID {
		return nil, errors.New("无权访问此收藏")
	}

	if req.Name != "" {
		fav.Name = req.Name
	}
	if req.SQLQuery != "" {
		fav.SQLQuery = req.SQLQuery
	}
	if req.Description != "" {
		fav.Description = req.Description
	}

	if err := s.favoriteDAO.Update(fav); err != nil {
		return nil, err
	}

	return fav, nil
}

// Delete 删除收藏
func (s *FavoriteService) Delete(id, userID int) error {
	fav, err := s.favoriteDAO.GetByID(id)
	if err != nil {
		return err
	}

	if fav.UserID != userID {
		return errors.New("无权访问此收藏")
	}

	return s.favoriteDAO.Delete(id, userID)
}
