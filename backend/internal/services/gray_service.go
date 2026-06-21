package services

import (
	"errors"
	"time"

	"config-center/internal/database"
	"config-center/internal/models"
)

type GrayReleaseService struct{}

func NewGrayReleaseService() *GrayReleaseService {
	return &GrayReleaseService{}
}

func (s *GrayReleaseService) CreateGrayRelease(release *models.GrayRelease) (*models.GrayRelease, error) {
	result := database.DB.Create(release)
	if result.Error != nil {
		return nil, result.Error
	}
	return release, nil
}

func (s *GrayReleaseService) StartGrayRelease(id uint) (*models.GrayRelease, error) {
	var release models.GrayRelease
	if err := database.DB.First(&release, id).Error; err != nil {
		return nil, errors.New("gray release not found")
	}

	now := time.Now()
	release.Status = "running"
	release.StartedAt = &now
	result := database.DB.Save(&release)
	if result.Error != nil {
		return nil, result.Error
	}

	return &release, nil
}

func (s *GrayReleaseService) FullPush(id uint) (*models.GrayRelease, error) {
	var release models.GrayRelease
	if err := database.DB.First(&release, id).Error; err != nil {
		return nil, errors.New("gray release not found")
	}

	release.Status = "completed"
	result := database.DB.Save(&release)
	if result.Error != nil {
		return nil, result.Error
	}

	var configItem models.ConfigItem
	if err := database.DB.First(&configItem, release.ConfigItemID).Error; err != nil {
		return &release, nil
	}

	var targetVersion models.ConfigVersion
	database.DB.Where("config_item_id = ? AND version = ?", release.ConfigItemID, release.TargetVersion).First(&targetVersion)

	if targetVersion.ID > 0 {
		configItem.Value = targetVersion.Value
		configItem.CurrentVersion = release.TargetVersion
		database.DB.Save(&configItem)
	}

	return &release, nil
}

func (s *GrayReleaseService) RollbackGrayRelease(id uint) (*models.GrayRelease, error) {
	var release models.GrayRelease
	if err := database.DB.First(&release, id).Error; err != nil {
		return nil, errors.New("gray release not found")
	}

	release.Status = "rolled_back"
	result := database.DB.Save(&release)
	if result.Error != nil {
		return nil, result.Error
	}

	return &release, nil
}

func (s *GrayReleaseService) GetGrayReleases(configItemID uint) ([]models.GrayRelease, error) {
	var releases []models.GrayRelease
	err := database.DB.Where("config_item_id = ?", configItemID).Order("created_at DESC").Find(&releases).Error
	return releases, err
}

func (s *GrayReleaseService) GetGrayRelease(id uint) (*models.GrayRelease, error) {
	var release models.GrayRelease
	if err := database.DB.First(&release, id).Error; err != nil {
		return nil, errors.New("gray release not found")
	}
	return &release, nil
}

func (s *GrayReleaseService) UpdatePushedCount(id uint, pushedCount int) error {
	result := database.DB.Model(&models.GrayRelease{}).Where("id = ?", id).Update("pushed_count", pushedCount)
	return result.Error
}
