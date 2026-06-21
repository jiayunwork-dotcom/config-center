package services

import (
	"errors"

	"config-center/internal/database"
	"config-center/internal/models"
)

type NamespaceService struct{}

func NewNamespaceService() *NamespaceService {
	return &NamespaceService{}
}

func (s *NamespaceService) CreateNamespace(ns *models.Namespace) (*models.Namespace, error) {
	result := database.DB.Create(ns)
	if result.Error != nil {
		return nil, result.Error
	}
	return ns, nil
}

func (s *NamespaceService) GetNamespaces(tenantID uint, accessibleIDs []uint) ([]models.Namespace, error) {
	var namespaces []models.Namespace
	query := database.DB.Where("tenant_id = ?", tenantID)
	if accessibleIDs != nil {
		if len(accessibleIDs) == 0 {
			return []models.Namespace{}, nil
		}
		query = query.Where("id IN ?", accessibleIDs)
	}
	err := query.Find(&namespaces).Error
	return namespaces, err
}

func (s *NamespaceService) GetNamespace(id uint) (*models.Namespace, error) {
	var ns models.Namespace
	if err := database.DB.First(&ns, id).Error; err != nil {
		return nil, errors.New("namespace not found")
	}
	return &ns, nil
}

func (s *NamespaceService) UpdateNamespace(id uint, name, description string) (*models.Namespace, error) {
	var ns models.Namespace
	if err := database.DB.First(&ns, id).Error; err != nil {
		return nil, errors.New("namespace not found")
	}
	ns.Name = name
	ns.Description = description
	result := database.DB.Save(&ns)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ns, nil
}

func (s *NamespaceService) DeleteNamespace(id uint) error {
	result := database.DB.Delete(&models.Namespace{}, id)
	return result.Error
}

type GroupService struct{}

func NewGroupService() *GroupService {
	return &GroupService{}
}

func (s *GroupService) CreateGroup(group *models.Group) (*models.Group, error) {
	result := database.DB.Create(group)
	if result.Error != nil {
		return nil, result.Error
	}
	return group, nil
}

func (s *GroupService) GetGroups(namespaceID uint) ([]models.Group, error) {
	var groups []models.Group
	err := database.DB.Where("namespace_id = ?", namespaceID).Find(&groups).Error
	return groups, err
}

func (s *GroupService) GetGroup(id uint) (*models.Group, error) {
	var group models.Group
	if err := database.DB.First(&group, id).Error; err != nil {
		return nil, errors.New("group not found")
	}
	return &group, nil
}

func (s *GroupService) UpdateGroup(id uint, name, description string) (*models.Group, error) {
	var group models.Group
	if err := database.DB.First(&group, id).Error; err != nil {
		return nil, errors.New("group not found")
	}
	group.Name = name
	group.Description = description
	result := database.DB.Save(&group)
	if result.Error != nil {
		return nil, result.Error
	}
	return &group, nil
}

func (s *GroupService) DeleteGroup(id uint) error {
	result := database.DB.Delete(&models.Group{}, id)
	return result.Error
}
