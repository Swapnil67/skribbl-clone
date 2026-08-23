package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"skribbl-clone/internal/game"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// * Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// * Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// * Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10 // * 54 seconds

	// * Maximum message size allowed from peer (512 KB).
	maxMessageSize = 512 * 1024
)

type Client struct {
	Hub       *Hub
	Room      *Room
	Conn      *websocket.Conn
	SessionID string
	Send      chan []byte
}

func NewClient(room *Room, conn *websocket.Conn, sessionID string) *Client {
	return &Client{
		Room:      room,
		Conn:      conn,
		SessionID: sessionID,
		Send:      make(chan []byte, 256), // * Buffered channel of outbound messages,
	}
}

// * readPump pumps messages from the websocket connection to the room.
func (c *Client) ReadPump() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	// * 1. Initial 60-second read deadline
	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	// * 2. Pong Handler: Extends deadline whenever a Pong arrives
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// * Infinite loop for reading client ws messages
	for {
		_, rawMessage, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error (session_id=%s): %v", c.SessionID, err)
			}
			break
		}

		// * 1. Unmarshal into the general InboundEvent envelope
		var env game.InboundEvent
		if err := json.Unmarshal(rawMessage, &env); err != nil {
			log.Printf("Malformed packet from %s: %v", c.SessionID, err)
			continue
		}

		switch env.Type {

		// * Intercept chat to validate guesses against target word
		case game.EventChatMessage:
			var chat game.ChatMessagePayload
			if err := json.Unmarshal(env.Payload, &chat); err == nil {
				c.Room.HandleChatMessage(c, chat.Text)
			}

		case game.EventDrawStroke, game.EventClearCanvas, game.EventUndoStroke:
			c.Room.mu.RLock()
			isDrawingPhase := c.Room.State == game.PhaseDrawing
			isDrawer := len(c.Room.drawerOrder) > c.Room.drawerIndex && c.Room.drawerOrder[c.Room.drawerIndex] == c.SessionID
			c.Room.mu.RUnlock()

			if !isDrawingPhase || !isDrawer {
				continue // * Drop unauthorized draw packets
			}

			fmt.Printf("%s drawing\n", c.SessionID)

			// * Forward valid draw actions to room clients
			c.Room.Broadcast <- rawMessage
		}
	}
}

// * writePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// * The hub closed the channel.
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, _ = w.Write(message)

			// Drain any queued messages in the channel and batch them into the frame
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// * ⏰ Ticker fires: write a low-level Ping frame to the client
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// * If writing ping fails, socket is broken -> trigger cleanup
				log.Println("Ping Failed")
				return
			}
		}
	}
}

func (c *Client) SendEvent(eventType game.EventType, payload interface{}) {
	event := game.OutboundEvent{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}

	raw, err := json.Marshal(event)
	if err != nil {
		return
	}
	select {
	case c.Send <- raw:
	default:
	}
}
