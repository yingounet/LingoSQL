package service

import (
	"encoding/json"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
)

type AuditService struct {
	auditDAO *sqlite.AuditDAO
}

func NewAuditService(auditDAO *sqlite.AuditDAO) *AuditService {
	return &AuditService{auditDAO: auditDAO}
}

func (s *AuditService) Record(userID int, action string, resourceType string, resourceID *int, success bool, errMsg string, detail interface{}) {
	detailJSON := ""
	if detail != nil {
		if b, err := json.Marshal(detail); err == nil {
			detailJSON = string(b)
		}
	}
	log := &models.AuditLog{
		UserID:      userID,
		Action:      action,
		ResourceType: resourceType,
		ResourceID:  resourceID,
		Success:     success,
		ErrorMessage: errMsg,
		Details:     detailJSON,
	}
	_ = s.auditDAO.Create(log)
}

func (s *AuditService) List(userID int, action string, page, pageSize int) (*models.AuditLogListResponse, error) {
	logs, total, err := s.auditDAO.ListByUser(userID, action, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &models.AuditLogListResponse{
		List:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
