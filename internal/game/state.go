package game

import "time"

type RoundPhase string

const (
	PhaseLobby         RoundPhase = "LOBBY"
	PhaseWordSelection RoundPhase = "WORD_SELECTION"
	PhaseDrawing       RoundPhase = "DRAWING"
	PhaseRoundSummary  RoundPhase = "ROUND_SUMMARY"
	PhaseGameOver      RoundPhase = "GAME_OVER"
)

const (
	DurationWordSelect = 15 * time.Second
	DurationDrawing    = 60 * time.Second
	DurationSummary    = 5 * time.Second
	TotalRounds        = 3
)

var WordBank = []string{
	"apple", "guitar", "rocket", "elephant", "bicycle",
	"pyramid", "sunflower", "castle", "tornado", "camera",
}

type PlayerState struct {
	SessionID  string `json:"session_id"`
	Username   string `json:"username"`
	Score      int    `json:"score"`
	HasGuessed bool   `json:"has_guessed"`
}
