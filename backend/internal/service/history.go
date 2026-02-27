package service

import (
	"errors"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
)

type HistoryService struct {
	historyDAO     *sqlite.HistoryDAO
	connectionDAO  *sqlite.ConnectionDAO
}

func NewHistoryService(historyDAO *sqlite.HistoryDAO, connectionDAO *sqlite.ConnectionDAO) *HistoryService {
	return &HistoryService{
		historyDAO:    historyDAO,
		connectionDAO: connectionDAO,
	}
}

// GetByUserID 获取用户的历史记录
func (s *HistoryService) GetByUserID(userID int, connectionID *int, page, pageSize int) (*models.QueryHistoryListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	histories, total, err := s.historyDAO.GetByUserID(userID, connectionID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	list := make([]models.QueryHistoryResponse, len(histories))
	for i, h := range histories {
		// 获取连接名称
		connName := ""
		if conn, err := s.connectionDAO.GetByID(h.ConnectionID); err == nil {
			connName = conn.Name
		}

		list[i] = models.QueryHistoryResponse{
			ID:              h.ID,
			ConnectionID:    h.ConnectionID,
			ConnectionName:  connName,
			SQLQuery:        h.SQLQuery,
			ExecutionTimeMs: h.ExecutionTimeMs,
			RowsAffected:    h.RowsAffected,
			Success:         h.Success,
			ErrorMessage:    h.ErrorMessage,
			CreatedAt:       h.CreatedAt,
		}
	}

	return &models.QueryHistoryListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetByID 获取单条历史记录
func (s *HistoryService) GetByID(id, userID int) (*models.QueryHistoryResponse, error) {
	history, err := s.historyDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	if history.UserID != userID {
		return nil, errors.New("无权访问此历史记录")
	}

	// 获取连接名称
	connName := ""
	if conn, err := s.connectionDAO.GetByID(history.ConnectionID); err == nil {
		connName = conn.Name
	}

	return &models.QueryHistoryResponse{
		ID:              history.ID,
		ConnectionID:    history.ConnectionID,
		ConnectionName:  connName,
		SQLQuery:        history.SQLQuery,
		ExecutionTimeMs: history.ExecutionTimeMs,
		RowsAffected:    history.RowsAffected,
		Success:         history.Success,
		ErrorMessage:    history.ErrorMessage,
		CreatedAt:       history.CreatedAt,
	}, nil
}

// Delete 删除历史记录
func (s *HistoryService) Delete(id, userID int) error {
	history, err := s.historyDAO.GetByID(id)
	if err != nil {
		return err
	}

	if history.UserID != userID {
		return errors.New("无权访问此历史记录")
	}

	return s.historyDAO.Delete(id, userID)
}
