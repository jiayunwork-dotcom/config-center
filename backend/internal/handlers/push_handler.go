package handlers

import (
	"net/http"
	"strconv"
	"time"

	"config-center/internal/database"
	"config-center/internal/models"
	"config-center/internal/push"
	"config-center/internal/services"

	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	configService *services.ConfigService
}

func NewPushHandler() *PushHandler {
	return &PushHandler{
		configService: services.NewConfigService(),
	}
}

func (h *PushHandler) LongPoll(c *gin.Context) {
	clientID := c.Query("client_id")
	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	version, _ := strconv.Atoi(c.DefaultQuery("version", "0"))
	environment := c.DefaultQuery("environment", "dev")

	if clientID == "" || namespaceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id and namespace_id are required"})
		return
	}

	items, _ := h.configService.GetConfigItems(uint(namespaceID), 0, environment)

	hasUpdate := false
	configs := make([]models.ConfigItem, 0)
	for _, item := range items {
		if item.CurrentVersion > version {
			hasUpdate = true
			configs = append(configs, item)
		}
	}

	if hasUpdate {
		go h.updateClientConnection(uint(namespaceID), clientID, c.ClientIP(), "longpoll")
		c.JSON(http.StatusOK, gin.H{
			"changed": true,
			"configs": configs,
		})
		return
	}

	respChan := push.Engine.AddLongPoll(clientID, uint(namespaceID), version, c.ClientIP())
	defer push.Engine.RemoveLongPoll(uint(namespaceID), clientID)

	go h.updateClientConnection(uint(namespaceID), clientID, c.ClientIP(), "longpoll")

	select {
	case event := <-respChan:
		if event.Environment == environment {
			c.JSON(http.StatusOK, gin.H{
				"changed": true,
				"configs": []push.ConfigChangeEvent{event},
			})
			return
		}
	case <-time.After(30 * time.Second):
	}

	c.JSON(http.StatusNotModified, gin.H{
		"changed": false,
		"message": "no changes",
	})
}

func (h *PushHandler) updateClientConnection(namespaceID uint, clientID string, ip string, connType string) {
	var conn models.ClientConnection
	result := database.DB.Where("namespace_id = ? AND client_id = ?", namespaceID, clientID).First(&conn)

	now := time.Now()
	if result.Error != nil {
		conn = models.ClientConnection{
			TenantID:    1,
			NamespaceID: namespaceID,
			ClientID:    clientID,
			IPAddress:   ip,
			ConnectType: connType,
			LastPullAt:  &now,
		}
		database.DB.Create(&conn)
	} else {
		conn.LastPullAt = &now
		conn.IPAddress = ip
		database.DB.Save(&conn)
	}
}

func (h *PushHandler) GetConnections(c *gin.Context) {
	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)

	var conns []models.ClientConnection
	query := database.DB
	if namespaceID > 0 {
		query = query.Where("namespace_id = ?", namespaceID)
	}
	query.Find(&conns)

	c.JSON(http.StatusOK, conns)
}

func (h *PushHandler) GetConnectionStats(c *gin.Context) {
	counts := push.Engine.GetAllConnectionCounts()
	c.JSON(http.StatusOK, counts)
}
