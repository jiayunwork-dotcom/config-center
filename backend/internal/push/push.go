package push

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"config-center/internal/config"
	redisclient "config-center/internal/redisclient"
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

type PushEngine struct {
	cfg           *config.Config
	longPollConns map[uint]map[string]*LongPollConn
	mu            sync.RWMutex
}

var Engine *PushEngine

func Init(cfg *config.Config) {
	Engine = &PushEngine{
		cfg:           cfg,
		longPollConns: make(map[uint]map[string]*LongPollConn),
	}

	go Engine.subscribeRedis()
}

func (e *PushEngine) subscribeRedis() {
	ctx := context.Background()
	pubsub := redisclient.Subscribe(ctx, "config_changes")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var event ConfigChangeEvent
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("Failed to unmarshal config change event: %v", err)
			continue
		}
		e.broadcastToLongPoll(event)
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

func (e *PushEngine) GetConnectionCount(namespaceID uint) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if conns, ok := e.longPollConns[namespaceID]; ok {
		return len(conns)
	}
	return 0
}

func (e *PushEngine) GetAllConnectionCounts() map[uint]int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[uint]int)
	for nsID, conns := range e.longPollConns {
		result[nsID] = len(conns)
	}
	return result
}

func PublishConfigChange(event ConfigChangeEvent) error {
	ctx := context.Background()
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return redisclient.Publish(ctx, "config_changes", string(payload))
}
