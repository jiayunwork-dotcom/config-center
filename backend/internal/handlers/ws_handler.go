package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

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

type WSHandler struct{}

var wsHandler *WSHandler

func NewWSHandler() *WSHandler {
	if wsHandler == nil {
		wsHandler = &WSHandler{}
	}
	return wsHandler
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

	push.Engine.AddWS(clientID, uint(namespaceID), c.ClientIP(), conn)
	defer push.Engine.RemoveWS(uint(namespaceID), clientID)

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

func (h *WSHandler) GetConnectionStats(c *gin.Context) {
	counts := push.Engine.GetAllConnectionCounts()
	c.JSON(http.StatusOK, counts)
}
