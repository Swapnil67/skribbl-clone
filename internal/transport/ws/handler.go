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
	// TODO
	// hub will manage client lifecycle
}

func NewHandler() *Handler {
	return &Handler{}
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

	// * Temporary echo loop for Phase 1 smoke testing
	// (Replaced by readPump/writePump in the next step)
	go func(wsConn *websocket.Conn, sID string) {
		defer wsConn.Close()
		for {
			messageType, p, err := wsConn.ReadMessage()
			if err != nil {
				log.Printf("Client disconnected: session_id=%s, err=%v", sID, err)
				break
			}

			// Echo back payload to verify handshake
			if err := wsConn.WriteMessage(messageType, p); err != nil {
				log.Printf("Write error: %v", err)
				break
			}
		}
	}(conn, sessionID)

}
