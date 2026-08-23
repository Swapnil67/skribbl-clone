package ws

import (
	"fmt"
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
	// hub *Hub // * hub will manage client lifecycle
	Manager *RoomManager // * RoomManager for Room lifecycle
}

func NewHandler(manager *RoomManager) *Handler {
	return &Handler{
		Manager: manager,
	}
}

func (h *Handler) HandleWebSocket(c *gin.Context) {
	// * 1. Extract session_id from Cookie or Query Param (?session_id=...)
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = c.Query("session_id")
	}

	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing session_id. Create a session first."})
		return
	}

	// * 2. Extract room_id
	roomID := c.Query("room_id")
	fmt.Printf("roomID: %s\n", roomID)
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing room_id query parameter"})
		return
	}

	// * 3. Lookup Room
	room, err := h.Manager.GetRoom(roomID)
	fmt.Printf("err: %v\n", err)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room does not exist or has expired"})
		return
	}

	// * 4. Validate capacity
	if room.ClientCount() >= room.MaxPlayers {
		c.JSON(http.StatusForbidden, gin.H{"error": "Room is already full"})
		return
	}

	// * 5. Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	log.Printf("Client connected: session_id=%s, remote_addr=%s", sessionID, conn.RemoteAddr())

	client := NewClient(room, conn, sessionID)
	room.Register <- client

	// * Start read and write pumps in dedicated goroutines
	go client.WritePump()
	go client.ReadPump()
}
