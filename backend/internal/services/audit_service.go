package services

import (
	"config-center/internal/database"
	"config-center/internal/models"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

func truncateValue(s string) string {
	if len(s) > 100 {
		return s[:100]
	}
	return s
}

func (s *AuditService) CreateLog(
	userID *uint,
	username string,
	action string,
	resourceType string,
	resourceID *uint,
	resourceName string,
	oldValue string,
	newValue string,
	ipAddress string,
) error {
	log := &models.AuditLog{
		TenantID:     1,
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		OldValue:     truncateValue(oldValue),
		NewValue:     truncateValue(newValue),
		IPAddress:    ipAddress,
	}
	return database.DB.Create(log).Error
}

func (s *AuditService) ListLogs(page, pageSize int, userID *uint, action string) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := database.DB.Model(&models.AuditLog{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}
