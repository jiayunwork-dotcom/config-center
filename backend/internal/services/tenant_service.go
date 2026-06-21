package services

import (
	"errors"

	"config-center/internal/database"
	"config-center/internal/models"
)

type TenantService struct{}

func NewTenantService() *TenantService {
	return &TenantService{}
}

func (s *TenantService) CreateTenant(tenant *models.Tenant) (*models.Tenant, error) {
	result := database.DB.Create(tenant)
	if result.Error != nil {
		return nil, result.Error
	}
	return tenant, nil
}

func (s *TenantService) GetTenants() ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := database.DB.Find(&tenants).Error
	return tenants, err
}

func (s *TenantService) GetTenant(id uint) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := database.DB.First(&tenant, id).Error; err != nil {
		return nil, errors.New("tenant not found")
	}
	return &tenant, nil
}

func (s *TenantService) UpdateTenant(id uint, name, displayName string, maxNS, maxItems, maxVersions int) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := database.DB.First(&tenant, id).Error; err != nil {
		return nil, errors.New("tenant not found")
	}
	tenant.Name = name
	tenant.DisplayName = displayName
	tenant.MaxNamespaces = maxNS
	tenant.MaxConfigItems = maxItems
	tenant.MaxVersions = maxVersions
	result := database.DB.Save(&tenant)
	if result.Error != nil {
		return nil, result.Error
	}
	return &tenant, nil
}

func (s *TenantService) DeleteTenant(id uint) error {
	result := database.DB.Delete(&models.Tenant{}, id)
	return result.Error
}

func (s *TenantService) CheckNamespaceQuota(tenantID uint) (bool, error) {
	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		return false, err
	}

	var count int64
	database.DB.Model(&models.Namespace{}).Where("tenant_id = ?", tenantID).Count(&count)

	return count < int64(tenant.MaxNamespaces), nil
}

func (s *TenantService) CheckConfigItemQuota(tenantID uint) (bool, error) {
	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		return false, err
	}

	var count int64
	database.DB.Model(&models.ConfigItem{}).Where("tenant_id = ?", tenantID).Count(&count)

	return count < int64(tenant.MaxConfigItems), nil
}
