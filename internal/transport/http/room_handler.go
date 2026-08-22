package http

import (
	"errors"
	"net/http"
	"skribbl-clone/internal/transport/ws"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	manager *ws.RoomManager
}

func NewRoomHandler(manager *ws.RoomManager) *RoomHandler {
	return &RoomHandler{
		manager: manager,
	}
}

// * Request payload for POST /api/v1/rooms
type CreateRoomRequest struct {
	MaxPlayers int `json:"max_players" binding:"omitempty,min=2,max=12"`
}

// * Response payload for POST /api/v1/rooms
type CreateRoomResponse struct {
	RoomID     string `json:"room_id"`
	MaxPlayers int    `json:"max_players"`
	State      string `json:"state"`
}

// * Response payload for GET /api/v1/rooms/:code
type RoomDetailResponse struct {
	RoomID      string `json:"room_id"`
	PlayerCount int    `json:"player_count"`
	MaxPlayers  int    `json:"max_players"`
	State       string `json:"state"`
	CanJoin     bool   `json:"can_join"`
}

func (h *RoomHandler) HandleCreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid room payload",
			"details": err.Error(),
		})
		return
	}

	// * Default to 8 players if not specified
	maxPlayers := req.MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = 8
	}

	room := h.manager.CreateRoom(maxPlayers)

	c.JSON(http.StatusOK, CreateRoomResponse{
		RoomID:     room.ID,
		MaxPlayers: maxPlayers,
		State:      string(room.State),
	})
}

// * GET /api/v1/rooms/:code
func (h *RoomHandler) HandleGetRoom(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room code is required"})
		return
	}

	room, err := h.manager.GetRoom(code)
	if err != nil {
		if errors.Is(err, ws.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Room not found or expired",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve room"})
		return
	}

	currentPlayers := room.ClientCount()
	canJoin := currentPlayers < room.MaxPlayers && room.State == ws.StateLobby

	c.JSON(http.StatusOK, RoomDetailResponse{
		RoomID:      room.ID,
		PlayerCount: currentPlayers,
		MaxPlayers:  room.MaxPlayers,
		State:       string(room.State),
		CanJoin:     canJoin,
	})
}
