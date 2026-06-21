package services

import (
	"errors"

	"config-center/internal/database"
	"config-center/internal/models"

	"gorm.io/gorm"
)

type ApprovalService struct{}

func NewApprovalService() *ApprovalService {
	return &ApprovalService{}
}

func (s *ApprovalService) CreateApproval(applicantID uint, applicant string, configItemID uint, configKey string, newValue string, oldValue string, environment string, description string) (*models.PendingApproval, error) {
	approval := &models.PendingApproval{
		ApplicantID:  applicantID,
		Applicant:    applicant,
		ConfigItemID: configItemID,
		ConfigKey:    configKey,
		NewValue:     newValue,
		OldValue:     oldValue,
		Environment:  environment,
		Description:  description,
		Status:       "pending",
	}
	if err := database.DB.Create(approval).Error; err != nil {
		return nil, err
	}
	return approval, nil
}

func (s *ApprovalService) ListApprovals(status string, page, pageSize int) ([]models.PendingApproval, int64, error) {
	var approvals []models.PendingApproval
	var total int64

	query := database.DB.Model(&models.PendingApproval{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&approvals).Error

	return approvals, total, err
}

func (s *ApprovalService) GetApproval(id uint) (*models.PendingApproval, error) {
	var approval models.PendingApproval
	if err := database.DB.First(&approval, id).Error; err != nil {
		return nil, errors.New("approval not found")
	}
	return &approval, nil
}

func (s *ApprovalService) Approve(id uint, reviewerID uint, reviewer string) (*models.ConfigItem, error) {
	var approval models.PendingApproval
	if err := database.DB.First(&approval, id).Error; err != nil {
		return nil, errors.New("approval not found")
	}

	if approval.Status != "pending" {
		return nil, errors.New("approval is not in pending status")
	}

	approval.Status = "approved"
	approval.ReviewerID = &reviewerID
	approval.Reviewer = reviewer
	database.DB.Save(&approval)

	configService := NewConfigService()
	result, err := configService.UpdateConfigItem(approval.ConfigItemID, approval.NewValue, approval.Applicant, approval.Description)
	if err != nil {
		approval.Status = "pending"
		approval.ReviewerID = nil
		approval.Reviewer = ""
		database.DB.Save(&approval)
		return nil, err
	}

	return result, nil
}

func (s *ApprovalService) Reject(id uint, reviewerID uint, reviewer string, reviewNote string) error {
	var approval models.PendingApproval
	if err := database.DB.First(&approval, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("approval not found")
		}
		return err
	}

	if approval.Status != "pending" {
		return errors.New("approval is not in pending status")
	}

	approval.Status = "rejected"
	approval.ReviewerID = &reviewerID
	approval.Reviewer = reviewer
	approval.ReviewNote = reviewNote
	return database.DB.Save(&approval).Error
}

func (s *ApprovalService) HasPendingApproval(configItemID uint) (bool, error) {
	var count int64
	database.DB.Model(&models.PendingApproval{}).
		Where("config_item_id = ? AND status = ?", configItemID, "pending").
		Count(&count)
	return count > 0, nil
}
