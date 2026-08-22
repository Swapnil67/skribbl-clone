package game

import "time"

// * struct to represent the ephemeral session in memory.

type Session struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	AvatarId  int       `json:"avatar_id"`
	CreatedAt time.Time `json:"created_id"`
}
