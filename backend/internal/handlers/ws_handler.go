package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"config-center/internal/database"
	"config-center/internal/models"
	"config-center/internal/push"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	clients map[uint]map[string]*websocket.Conn
	mu      sync.RWMutex
}

var wsHandler *WSHandler

func NewWSHandler() *WSHandler {
	if wsHandler == nil {
		wsHandler = &WSHandler{
			clients: make(map[uint]map[string]*websocket.Conn),
		}
		go wsHandler.listenRedis()
	}
	return wsHandler
}

func (h *WSHandler) listenRedis() {
	for {
		time.Sleep(1 * time.Second)
	}
}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket: %v", err)
		return
	}
	defer conn.Close()

	namespaceID, _ := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
	clientID := c.Query("client_id")

	if namespaceID == 0 || clientID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"namespace_id and client_id are required"}`))
		return
	}

	h.addClient(uint(namespaceID), clientID, conn)
	defer h.removeClient(uint(namespaceID), clientID)

	go h.updateClientConnection(uint(namespaceID), clientID, c.ClientIP())

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *WSHandler) addClient(namespaceID uint, clientID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[namespaceID]; !ok {
		h.clients[namespaceID] = make(map[string]*websocket.Conn)
	}
	h.clients[namespaceID][clientID] = conn
}

func (h *WSHandler) removeClient(namespaceID uint, clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[namespaceID]; ok {
		delete(clients, clientID)
	}
}

func (h *WSHandler) Broadcast(namespaceID uint, event push.ConfigChangeEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[namespaceID]
	if !ok {
		return
	}

	msg, _ := json.Marshal(gin.H{
		"type":  "config_change",
		"event": event,
	})

	for _, conn := range clients {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (h *WSHandler) GetWSConnectionCount(namespaceID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[namespaceID]; ok {
		return len(clients)
	}
	return 0
}

func (h *WSHandler) updateClientConnection(namespaceID uint, clientID string, ip string) {
	var conn models.ClientConnection
	result := database.DB.Where("namespace_id = ? AND client_id = ?", namespaceID, clientID).First(&conn)

	now := time.Now()
	if result.Error != nil {
		conn = models.ClientConnection{
			TenantID:    1,
			NamespaceID: namespaceID,
			ClientID:    clientID,
			IPAddress:   ip,
			ConnectType: "websocket",
			LastPullAt:  &now,
		}
		database.DB.Create(&conn)
	} else {
		conn.LastPullAt = &now
		conn.IPAddress = ip
		database.DB.Save(&conn)
	}
}

func (h *WSHandler) GetConnectionStats(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[uint]int)
	for nsID, clients := range h.clients {
		stats[nsID] = len(clients)
	}
	c.JSON(http.StatusOK, stats)
}
