package services

import (
	"errors"
	"fmt"
	"time"

	"config-center/internal/database"
	"config-center/internal/diffutil"
	"config-center/internal/models"
	"config-center/internal/push"
	"config-center/internal/validator"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) CreateConfigItem(item *models.ConfigItem, operator string) (*models.ConfigItem, error) {
	if err := validator.ValidateFormat(item.Value, item.Format); err != nil {
		return nil, fmt.Errorf("format validation failed: %w", err)
	}

	if item.Schema != nil && *item.Schema != "" {
		if item.Format == "json" {
			if err := validator.ValidateWithSchema(item.Value, *item.Schema); err != nil {
				return nil, fmt.Errorf("schema validation failed: %w", err)
			}
		}
	}

	item.CurrentVersion = 1

	result := database.DB.Create(item)
	if result.Error != nil {
		return nil, result.Error
	}

	version := &models.ConfigVersion{
		TenantID:     item.TenantID,
		ConfigItemID: item.ID,
		Version:      1,
		Value:        item.Value,
		Operator:     operator,
		ChangeType:   "create",
		Description:  "Initial version",
	}
	database.DB.Create(version)

	s.publishChange(item)

	return item, nil
}

func (s *ConfigService) UpdateConfigItem(id uint, newValue string, operator string, description string) (*models.ConfigItem, error) {
	var item models.ConfigItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return nil, errors.New("config item not found")
	}

	if err := validator.ValidateFormat(newValue, item.Format); err != nil {
		return nil, fmt.Errorf("format validation failed: %w", err)
	}

	if item.Schema != nil && *item.Schema != "" {
		if item.Format == "json" {
			if err := validator.ValidateWithSchema(newValue, *item.Schema); err != nil {
				return nil, fmt.Errorf("schema validation failed: %w", err)
			}
		}
	}

	oldValue := item.Value
	diffStr, _ := diffutil.ComputeUnifiedDiff(oldValue, newValue)

	newVersion := item.CurrentVersion + 1

	item.Value = newValue
	item.CurrentVersion = newVersion
	database.DB.Save(&item)

	version := &models.ConfigVersion{
		TenantID:     item.TenantID,
		ConfigItemID: item.ID,
		Version:      newVersion,
		Value:        newValue,
		Operator:     operator,
		ChangeType:   "update",
		Diff:         diffStr,
		Description:  description,
	}
	database.DB.Create(version)

	s.publishChange(&item)

	return &item, nil
}

func (s *ConfigService) GetConfigItem(id uint) (*models.ConfigItem, error) {
	var item models.ConfigItem
	if err := database.DB.First(&item, id).Error; err != nil {
		return nil, errors.New("config item not found")
	}
	return &item, nil
}

func (s *ConfigService) GetConfigItems(namespaceID, groupID uint, environment string) ([]models.ConfigItem, error) {
	var items []models.ConfigItem
	query := database.DB.Where("namespace_id = ? AND environment = ?", namespaceID, environment)
	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *ConfigService) DeleteConfigItem(id uint) error {
	result := database.DB.Delete(&models.ConfigItem{}, id)
	return result.Error
}

func (s *ConfigService) RollbackToVersion(configItemID uint, targetVersion int, operator string) (*models.ConfigItem, error) {
	var item models.ConfigItem
	if err := database.DB.First(&item, configItemID).Error; err != nil {
		return nil, errors.New("config item not found")
	}

	var targetVersionRec models.ConfigVersion
	if err := database.DB.Where("config_item_id = ? AND version = ?", configItemID, targetVersion).First(&targetVersionRec).Error; err != nil {
		return nil, errors.New("target version not found")
	}

	oldValue := item.Value
	newValue := targetVersionRec.Value

	diffStr, _ := diffutil.ComputeUnifiedDiff(oldValue, newValue)

	newVersionNum := item.CurrentVersion + 1
	item.Value = newValue
	item.CurrentVersion = newVersionNum
	database.DB.Save(&item)

	version := &models.ConfigVersion{
		TenantID:     item.TenantID,
		ConfigItemID: item.ID,
		Version:      newVersionNum,
		Value:        newValue,
		Operator:     operator,
		ChangeType:   "rollback",
		Diff:         diffStr,
		Description:  fmt.Sprintf("Rollback to version %d", targetVersion),
	}
	database.DB.Create(version)

	s.publishChange(&item)

	return &item, nil
}

func (s *ConfigService) publishChange(item *models.ConfigItem) {
	event := push.ConfigChangeEvent{
		TenantID:    item.TenantID,
		NamespaceID: item.NamespaceID,
		GroupID:     item.GroupID,
		Key:         item.Key,
		Version:     item.CurrentVersion,
		Value:       item.Value,
		Format:      item.Format,
		Environment: item.Environment,
		Timestamp:   time.Now().Unix(),
	}
	push.PublishConfigChange(event)
}

func (s *ConfigService) GetVersionHistory(configItemID uint, page, pageSize int) ([]models.ConfigVersion, int64, error) {
	var versions []models.ConfigVersion
	var total int64

	database.DB.Model(&models.ConfigVersion{}).Where("config_item_id = ?", configItemID).Count(&total)

	offset := (page - 1) * pageSize
	err := database.DB.Where("config_item_id = ?", configItemID).
		Order("version DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&versions).Error

	return versions, total, err
}

func (s *ConfigService) CompareVersions(configItemID uint, version1, version2 int) ([]diffutil.DiffLine, error) {
	var v1, v2 models.ConfigVersion

	if err := database.DB.Where("config_item_id = ? AND version = ?", configItemID, version1).First(&v1).Error; err != nil {
		return nil, errors.New("version1 not found")
	}
	if err := database.DB.Where("config_item_id = ? AND version = ?", configItemID, version2).First(&v2).Error; err != nil {
		return nil, errors.New("version2 not found")
	}

	diffLines := diffutil.ComputeLineDiff(v1.Value, v2.Value)
	return diffLines, nil
}
