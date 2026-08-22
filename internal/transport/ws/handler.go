package ws

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// * In development / same-origin setups, return true
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	// * hub will manage client lifecycle
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	// 1. Extract session_id from Cookie or Query Param (?session_id=...)
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = c.Query("session_id")
	}

	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing session_id. Create a session first.",
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	log.Printf("Client connected: session_id=%s, remote_addr=%s", sessionID, conn.RemoteAddr())

	client := NewClient(h.hub, conn, sessionID)
	h.hub.Register <- client

	// * Start read and write pumps in dedicated goroutines
	go client.WritePump()
	go client.ReadPump()
}
