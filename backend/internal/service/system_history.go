package service

import (
	"errors"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
)

type SystemHistoryService struct {
	systemHistoryDAO *sqlite.SystemHistoryDAO
	connectionDAO    *sqlite.ConnectionDAO
}

func NewSystemHistoryService(systemHistoryDAO *sqlite.SystemHistoryDAO, connectionDAO *sqlite.ConnectionDAO) *SystemHistoryService {
	return &SystemHistoryService{
		systemHistoryDAO: systemHistoryDAO,
		connectionDAO:     connectionDAO,
	}
}

// Create 创建系统执行记录
func (s *SystemHistoryService) Create(history *models.SystemQueryHistory) error {
	return s.systemHistoryDAO.Create(history)
}

// GetByConnectionID 获取指定连接的系统执行记录
func (s *SystemHistoryService) GetByConnectionID(connectionID, userID int, page, pageSize int) (*models.SystemQueryHistoryListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 验证连接是否属于该用户
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, err
	}
	if conn.UserID != userID {
		return nil, errors.New("无权访问此连接")
	}

	histories, total, err := s.systemHistoryDAO.GetByConnectionID(connectionID, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	list := make([]models.SystemQueryHistoryResponse, len(histories))
	for i, h := range histories {
		// 获取连接名称
		connName := ""
		if conn, err := s.connectionDAO.GetByID(h.ConnectionID); err == nil {
			connName = conn.Name
		}

		list[i] = models.SystemQueryHistoryResponse{
			ID:              h.ID,
			ConnectionID:    h.ConnectionID,
			ConnectionName:  connName,
			SQLQuery:        h.SQLQuery,
			OperationType:   h.OperationType,
			ExecutionTimeMs: h.ExecutionTimeMs,
			RowsAffected:    h.RowsAffected,
			Success:         h.Success,
			ErrorMessage:    h.ErrorMessage,
			CreatedAt:       h.CreatedAt,
		}
	}

	return &models.SystemQueryHistoryListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Delete 删除系统执行记录
func (s *SystemHistoryService) Delete(id, userID int) error {
	history, err := s.systemHistoryDAO.GetByID(id)
	if err != nil {
		return err
	}

	if history.UserID != userID {
		return errors.New("无权访问此历史记录")
	}

	return s.systemHistoryDAO.Delete(id, userID)
}
