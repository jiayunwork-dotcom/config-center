package services

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"config-center/internal/database"
	"config-center/internal/models"
	"config-center/internal/push"
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

	go s.executeGrayPush(&release)

	return &release, nil
}

func (s *GrayReleaseService) executeGrayPush(release *models.GrayRelease) {
	log.Printf("Starting gray push: id=%d, strategy=%s, target_version=%d",
		release.ID, release.Strategy, release.TargetVersion)

	var configItem models.ConfigItem
	if err := database.DB.First(&configItem, release.ConfigItemID).Error; err != nil {
		log.Printf("Config item not found: %v", err)
		return
	}

	var targetVersion models.ConfigVersion
	if err := database.DB.Where("config_item_id = ? AND version = ?",
		release.ConfigItemID, release.TargetVersion).First(&targetVersion).Error; err != nil {
		log.Printf("Target version not found: %v", err)
		return
	}

	event := push.ConfigChangeEvent{
		TenantID:    configItem.TenantID,
		NamespaceID: configItem.NamespaceID,
		GroupID:     configItem.GroupID,
		Key:         configItem.Key,
		Version:     targetVersion.Version,
		Value:       targetVersion.Value,
		Format:      configItem.Format,
		Environment: configItem.Environment,
		Timestamp:   time.Now().Unix(),
	}

	allClients := s.getAllClients(configItem.NamespaceID)
	release.TotalCount = len(allClients)
	log.Printf("Total clients found: %d", len(allClients))

	targetClients := s.filterClients(allClients, release)
	log.Printf("Target clients after filter: %d", len(targetClients))

	pushedCount := 0
	for _, client := range targetClients {
		if s.pushToClient(configItem.NamespaceID, client.ClientID, event) {
			pushedCount++
		}
	}

	release.PushedCount = pushedCount
	database.DB.Save(release)
	log.Printf("Gray push completed: pushed=%d/%d", pushedCount, release.TotalCount)

	if err := push.PublishGrayConfigChange(event); err != nil {
		log.Printf("Failed to publish gray config change: %v", err)
	}
}

func (s *GrayReleaseService) getAllClients(namespaceID uint) []push.ClientInfo {
	allClients := push.Engine.GetAllClients(namespaceID)
	log.Printf("GetAllClients returned %d clients for namespace %d", len(allClients), namespaceID)
	return allClients
}

func (s *GrayReleaseService) filterClients(clients []push.ClientInfo, release *models.GrayRelease) []push.ClientInfo {
	if release.Strategy == "ip_list" {
		return s.filterByIPList(clients, release.IPList)
	} else if release.Strategy == "percentage" {
		return s.filterByPercentage(clients, release.Percentage)
	}
	return clients
}

func (s *GrayReleaseService) filterByIPList(clients []push.ClientInfo, ipList []string) []push.ClientInfo {
	if len(ipList) == 0 {
		return nil
	}

	ipSet := make(map[string]bool)
	for _, ip := range ipList {
		ipSet[ip] = true
	}

	var result []push.ClientInfo
	for _, client := range clients {
		if ipSet[client.IP] {
			result = append(result, client)
		}
	}
	return result
}

func (s *GrayReleaseService) filterByPercentage(clients []push.ClientInfo, percentage int) []push.ClientInfo {
	if percentage <= 0 || len(clients) == 0 {
		return nil
	}
	if percentage >= 100 {
		return clients
	}

	targetCount := (len(clients) * percentage) / 100
	if targetCount == 0 {
		targetCount = 1
	}

	shuffled := make([]push.ClientInfo, len(clients))
	copy(shuffled, clients)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:targetCount]
}

func (s *GrayReleaseService) pushToClient(namespaceID uint, clientID string, event push.ConfigChangeEvent) bool {
	pushed := push.Engine.PushToClient(namespaceID, clientID, event)

	if pushed {
		log.Printf("Pushed config to client: %s (namespace=%d, key=%s)",
			clientID, namespaceID, event.Key)
	} else {
		log.Printf("Failed to push to client: %s", clientID)
	}

	return pushed
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
		oldValue := configItem.Value
		configItem.Value = targetVersion.Value
		configItem.CurrentVersion = release.TargetVersion
		database.DB.Save(&configItem)

		event := push.ConfigChangeEvent{
			TenantID:    configItem.TenantID,
			NamespaceID: configItem.NamespaceID,
			GroupID:     configItem.GroupID,
			Key:         configItem.Key,
			Version:     targetVersion.Version,
			Value:       targetVersion.Value,
			Format:      configItem.Format,
			Environment: configItem.Environment,
			Timestamp:   time.Now().Unix(),
		}
		push.PublishConfigChange(event)

		log.Printf("Full push completed: config=%s, version=%d, diff from v%d to v%d",
			configItem.Key, targetVersion.Version,
			configItem.CurrentVersion-1, targetVersion.Version)
		_ = oldValue
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

	log.Printf("Gray release rolled back: id=%d", id)

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
