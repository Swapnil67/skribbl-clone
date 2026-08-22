package ws

import (
	"log"
	"sync"
)

type RoomState string

const (
	StateLobby         RoomState = "LOBBY"
	StateWordSelection RoomState = "WORD_SELECTION"
	StateDrawing       RoomState = "DRAWING"
	StateRoundSummary  RoomState = "ROUND_SUMMARY"
	StateGameOver      RoomState = "GAME_OVER"
)

type Room struct {
	ID         string
	State      RoomState
	MaxPlayers int

	// * Thread safe client map
	mu      sync.RWMutex
	clients map[*Client]bool

	// * Channels for client lifecycle & messaging
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client

	// * RoomManager reference to notify when room becomes empty
	manager *RoomManager
}

func NewRoom(id string, maxPlayers int, manager *RoomManager) *Room {
	return &Room{
		ID:         id,
		State:      StateLobby,
		MaxPlayers: maxPlayers,
		clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 128),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		manager:    manager,
	}
}

// * Run executes the room's isolated event loop
func (r *Room) Run() {
	log.Printf("🎮 Room [%s] event loop started", r.ID)
	for {
		select {
		// * Register the client from room
		case client := <-r.Register:
			r.mu.Lock()
			r.clients[client] = true
			total := len(r.clients)
			r.mu.Unlock()

			log.Printf("User [%s] joined Room [%s] (Total: %d)", client.SessionID, r.ID, total)

		// * Unregister the client from room
		case client := <-r.Unregister:
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.Send)
			}
			remaining := len(r.clients)
			r.mu.Unlock()

			log.Printf("User [%s] left Room [%s] (Remaining: %d)", client.SessionID, r.ID, remaining)

			// * Clean up the room when the last player leaves
			if remaining == 0 {
				log.Printf("🧹 Room [%s] is empty. Cleaning up...", r.ID)
				// r.manager.DeleteRoom(r.ID) // TODO
				return // Terminate the room goroutine
			}

		case message := <-r.Broadcast:
			r.mu.RLock()
			for client := range r.clients {
				select {
				case client.Send <- message:
				default:
					// * Slow consumer: channel full, drop client to prevent blocking others
					close(client.Send)
					delete(r.clients, client)
				}
			}
			r.mu.RUnlock()
		}
	}
}

// * ClientCount returns the current number of active players (Read lock)
func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}
