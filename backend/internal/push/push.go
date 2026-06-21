package push

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"config-center/internal/config"
	"config-center/internal/database"
	"config-center/internal/models"
	redisclient "config-center/internal/redisclient"

	"github.com/gorilla/websocket"
)

type ConfigChangeEvent struct {
	TenantID    uint   `json:"tenant_id"`
	NamespaceID uint   `json:"namespace_id"`
	GroupID     uint   `json:"group_id"`
	Key         string `json:"key"`
	Version     int    `json:"version"`
	Value       string `json:"value"`
	Format      string `json:"format"`
	Environment string `json:"environment"`
	Timestamp   int64  `json:"timestamp"`
}

type LongPollConn struct {
	ClientID    string
	NamespaceID uint
	Version     int
	IP          string
	CreatedAt   time.Time
	RespChan    chan ConfigChangeEvent
}

type WSConn struct {
	ClientID    string
	NamespaceID uint
	IP          string
	Conn        *websocket.Conn
}

type PushEngine struct {
	cfg           *config.Config
	longPollConns map[uint]map[string]*LongPollConn
	wsConns       map[uint]map[string]*WSConn
	mu            sync.RWMutex
}

type ClientInfo struct {
	ClientID string
	IP       string
}

var Engine *PushEngine

func Init(cfg *config.Config) {
	Engine = &PushEngine{
		cfg:           cfg,
		longPollConns: make(map[uint]map[string]*LongPollConn),
		wsConns:       make(map[uint]map[string]*WSConn),
	}

	go Engine.subscribeRedis()
}

func (e *PushEngine) subscribeRedis() {
	ctx := context.Background()
	pubsub := redisclient.Subscribe(ctx, "config_changes", "gray_config_changes")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var event ConfigChangeEvent
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("Failed to unmarshal config change event: %v", err)
			continue
		}
		log.Printf("Config change received on channel %s: namespace=%d, key=%s, version=%d",
			msg.Channel, event.NamespaceID, event.Key, event.Version)
		e.broadcastToLongPoll(event)
		e.broadcastToWS(event)
	}
}

func (e *PushEngine) broadcastToLongPoll(event ConfigChangeEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	conns, ok := e.longPollConns[event.NamespaceID]
	if !ok {
		return
	}

	for _, conn := range conns {
		if conn.Version < event.Version {
			select {
			case conn.RespChan <- event:
			default:
			}
		}
	}
}

func (e *PushEngine) broadcastToWS(event ConfigChangeEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	conns, ok := e.wsConns[event.NamespaceID]
	if !ok {
		return
	}

	msg, _ := json.Marshal(map[string]interface{}{
		"type":  "config_change",
		"event": event,
	})

	for _, conn := range conns {
		conn.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (e *PushEngine) AddLongPoll(clientID string, namespaceID uint, version int, ip string) <-chan ConfigChangeEvent {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.longPollConns[namespaceID]; !ok {
		e.longPollConns[namespaceID] = make(map[string]*LongPollConn)
	}

	respChan := make(chan ConfigChangeEvent, 1)
	conn := &LongPollConn{
		ClientID:    clientID,
		NamespaceID: namespaceID,
		Version:     version,
		IP:          ip,
		CreatedAt:   time.Now(),
		RespChan:    respChan,
	}

	e.longPollConns[namespaceID][clientID] = conn

	go func() {
		time.Sleep(time.Duration(e.cfg.LongPollTimeout) * time.Second)
		e.removeLongPoll(namespaceID, clientID)
	}()

	return respChan
}

func (e *PushEngine) removeLongPoll(namespaceID uint, clientID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if conns, ok := e.longPollConns[namespaceID]; ok {
		if conn, ok := conns[clientID]; ok {
			close(conn.RespChan)
			delete(conns, clientID)
		}
	}
}

func (e *PushEngine) RemoveLongPoll(namespaceID uint, clientID string) {
	e.removeLongPoll(namespaceID, clientID)
}

func (e *PushEngine) AddWS(clientID string, namespaceID uint, ip string, conn *websocket.Conn) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.wsConns[namespaceID]; !ok {
		e.wsConns[namespaceID] = make(map[string]*WSConn)
	}

	e.wsConns[namespaceID][clientID] = &WSConn{
		ClientID:    clientID,
		NamespaceID: namespaceID,
		IP:          ip,
		Conn:        conn,
	}

	go e.updateClientConnection(namespaceID, clientID, ip, "websocket")
}

func (e *PushEngine) RemoveWS(namespaceID uint, clientID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if conns, ok := e.wsConns[namespaceID]; ok {
		delete(conns, clientID)
	}
}

func (e *PushEngine) GetConnectionCount(namespaceID uint) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	count := 0
	if conns, ok := e.longPollConns[namespaceID]; ok {
		count += len(conns)
	}
	if conns, ok := e.wsConns[namespaceID]; ok {
		count += len(conns)
	}
	return count
}

func (e *PushEngine) GetAllConnectionCounts() map[uint]int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[uint]int)
	for nsID, conns := range e.longPollConns {
		result[nsID] += len(conns)
	}
	for nsID, conns := range e.wsConns {
		result[nsID] += len(conns)
	}
	return result
}

func (e *PushEngine) GetAllClients(namespaceID uint) []ClientInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	clientMap := make(map[string]bool)
	var clients []ClientInfo

	if conns, ok := e.longPollConns[namespaceID]; ok {
		for _, conn := range conns {
			if !clientMap[conn.ClientID] {
				clientMap[conn.ClientID] = true
				clients = append(clients, ClientInfo{
					ClientID: conn.ClientID,
					IP:       conn.IP,
				})
			}
		}
	}

	if conns, ok := e.wsConns[namespaceID]; ok {
		for _, conn := range conns {
			if !clientMap[conn.ClientID] {
				clientMap[conn.ClientID] = true
				clients = append(clients, ClientInfo{
					ClientID: conn.ClientID,
					IP:       conn.IP,
				})
			}
		}
	}

	return clients
}

func (e *PushEngine) PushToClient(namespaceID uint, clientID string, event ConfigChangeEvent) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pushed := false

	if conns, ok := e.longPollConns[namespaceID]; ok {
		if conn, ok := conns[clientID]; ok {
			select {
			case conn.RespChan <- event:
				pushed = true
			default:
			}
		}
	}

	if conns, ok := e.wsConns[namespaceID]; ok {
		if conn, ok := conns[clientID]; ok {
			msg, _ := json.Marshal(map[string]interface{}{
				"type":  "config_change",
				"event": event,
			})
			err := conn.Conn.WriteMessage(websocket.TextMessage, msg)
			if err == nil {
				pushed = true
			}
		}
	}

	return pushed
}

func (e *PushEngine) updateClientConnection(namespaceID uint, clientID string, ip string, connType string) {
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

func PublishConfigChange(event ConfigChangeEvent) error {
	ctx := context.Background()
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return redisclient.Publish(ctx, "config_changes", string(payload))
}

func PublishGrayConfigChange(event ConfigChangeEvent) error {
	ctx := context.Background()
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return redisclient.Publish(ctx, "gray_config_changes", string(payload))
}
