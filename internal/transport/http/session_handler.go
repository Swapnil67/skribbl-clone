package http

import (
	"net/http"
	"skribbl-clone/internal/game"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// * Request payload struct with validation rules
type CreateSessionRequest struct {
	// * Username must be between 2 and 20 alphanumeric/underscore characters
	Username string `json:"username" binding:"required,min=2,max=20"`
	// * Optional avatar selection from 1 to 10 (defaults to 1 if omitted)
	AvatarID int `json:"avatar_id" binding:"omitempty,min=1,max=10"`
}

// * Response payload struct
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
	Username  string `json:"username"`
	AvatarID  int    `json:"avatar_id"`
}

func handleCreateGuestSession(c *gin.Context) {
	var req CreateSessionRequest

	// * 1. Parse JSON body & validate tags (returns 400 automatically if invalid)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid session payload",
			"details": err.Error(),
		})
		return
	}

	// * 2. Sanitize username (trim whitespace)
	trimmedUsername := strings.TrimSpace(req.Username)
	if len(trimmedUsername) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username cannot be empty or solely whitespace",
		})
		return
	}

	// * 3. Fallback default for optional avatar
	avatarID := req.AvatarID
	if avatarID == 0 {
		avatarID = 1
	}

	// * 4. Generate unique ephemeral session
	session := game.Session{
		ID:        uuid.NewString(),
		Username:  trimmedUsername,
		AvatarId:  avatarID,
		CreatedAt: time.Now().UTC(),
	}

	// * 5. Set an HTTP-Only cookie for seamless browser sessions
	// * (Cookie: name, value, maxAgeInSeconds, path, domain, secure, httpOnly)
	c.SetCookie("session_id", session.ID, 3600*24, "/", "", false, true)

	// * 6. Return response to the client
	c.JSON(http.StatusCreated, CreateSessionResponse{
		SessionID: session.ID,
		Username:  session.Username,
		AvatarID:  session.AvatarId,
	})

}
