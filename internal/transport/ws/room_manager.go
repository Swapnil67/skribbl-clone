package ws

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
)

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomFull     = errors.New("room is full")
)

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

// * CreateRoom generates a unique room code and initializes its event loop
func (rm *RoomManager) CreateRoom(maxPlayers int) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	code := rm.generateUniqueCode()
	room := NewRoom(code, maxPlayers, rm)
	rm.rooms[code] = room

	// * Start the room's isolated goroutine
	go room.Run()
	return room
}

// * GetRoom returns a pointer to the room if it exists (Read lock)
func (rm *RoomManager) GetRoom(roomID string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// * DeleteRoom removes a room from memory (Write lock)
func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.rooms, roomID)
}

// * RoomCount returns the total number of active rooms
func (rm *RoomManager) RoomCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.rooms)
}

// * Generates a random 6-character room code (e.g., "K9X2W7")
func (rm *RoomManager) generateUniqueCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excludes ambiguous letters (I, O, 1, 0)
	for {
		b := make([]byte, 6)
		for i := range b {
			idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			b[i] = charset[idx.Int64()]
		}
		code := string(b)
		if _, exists := rm.rooms[code]; !exists {
			return code
		}
	}
}
