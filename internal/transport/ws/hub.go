package ws

import "log"

type Hub struct {
	// * Registered active clients mapped by pointer reference
	clients map[*Client]bool

	// * Inbound messages from clients to fan out
	Broadcast chan []byte

	// * Register requests from newly connected clients
	Register chan *Client

	// * Unregister requests from disconnecting clients
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	log.Println("⚡ In-Memory WebSocket Hub running")
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			log.Printf("Client registered: session_id=%s (Active: %d)", client.SessionID, len(h.clients))

		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("Client unregistered: session_id=%s (Active: %d)", client.SessionID, len(h.clients))
			}

		case message := <-h.Broadcast:
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// * Slow consumer: channel full, drop client to prevent blocking others
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}
