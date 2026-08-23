package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"skribbl-clone/internal/game"
	"sync"
	"time"
)

type Room struct {
	ID         string
	State      game.RoundPhase
	MaxPlayers int

	// * Thread safe client & players map
	mu      sync.RWMutex
	clients map[*Client]bool
	players map[string]*game.PlayerState

	// * Game Loop Turn Tracking
	currentRound  int
	drawingOrder  []string
	drawerIndex   int
	currentWord   string
	timeRemaning  int
	isLoopRunning bool
	stopLoop      chan struct{}

	// * Channels for client lifecycle & messaging
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	manager    *RoomManager
}

func NewRoom(id string, maxPlayers int, manager *RoomManager) *Room {
	return &Room{
		ID:           id,
		State:        game.PhaseLobby,
		MaxPlayers:   maxPlayers,
		clients:      make(map[*Client]bool),
		players:      make(map[string]*game.PlayerState),
		drawingOrder: make([]string, 0),
		Broadcast:    make(chan []byte, 128),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		manager:      manager,
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
			// * Check if player exists
			if _, exists := r.players[client.SessionID]; !exists {
				r.players[client.SessionID] = &game.PlayerState{
					SessionID: client.SessionID,
					Username:  "Player-" + client.SessionID[:4],
					Score:     0,
				}
				r.drawingOrder = append(r.drawingOrder, client.SessionID)
			}
			total := len(r.clients)
			r.mu.Unlock()

			log.Printf("User [%s] joined Room [%s] (Total: %d)", client.SessionID, r.ID, total)

			// ! Auto-start loop when at least 2 players join and room is in LOBBY
			r.mu.Lock()
			if total >= 2 && r.State == game.PhaseLobby && !r.isLoopRunning {
				r.isLoopRunning = true
				go r.startGameLoop()
			}
			r.mu.Unlock()

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
				log.Printf("🧹 Room [%s] empty. Stopping loop...", r.ID)
				r.mu.Lock()
				if r.isLoopRunning {
					close(r.stopLoop)
					r.isLoopRunning = false
				}
				r.mu.Unlock()

				r.manager.DeleteRoom(r.ID)
				return // * Terminate the room goroutine
			}

		case message := <-r.Broadcast:
			r.broadcastRaw(message)
			// r.mu.RLock()
			// for client := range r.clients {
			// 	select {
			// 	case client.Send <- message:
			// 	default:
			// 		// * Slow consumer: channel full, drop client to prevent blocking others
			// 		close(client.Send)
			// 		delete(r.clients, client)
			// 	}
			// }
			// r.mu.RUnlock()
		}
	}
}

// * ----------------------------------------------------
// * State Machine Engine
// * ----------------------------------------------------

func (r *Room) startGameLoop() {
	log.Printf("🚀 Starting Game Loop for Room [%s]", r.ID)

	// * 1. Outer Loop: Cycles through rounds 1, 2, and 3
	for round := 1; round <= game.TotalRounds; round++ {
		r.mu.Lock()
		r.currentRound = round
		r.drawerIndex = 0 // * Reset to the first player in the queue
		r.mu.Unlock()

		// * 2. Inner Loop: Iterates through the player queue for this round
		for {
			r.mu.RLock()
			totalDrawers := len(r.drawingOrder)
			curIdx := r.drawerIndex
			r.mu.RUnlock()

			fmt.Printf("curIdx: %d\n", curIdx)
			fmt.Printf("totalDrawers: %d\n", totalDrawers)

			// * Everyone has drawn this round
			// * Players dropped below 2 (can't play alone)
			if curIdx >= totalDrawers || totalDrawers < 2 {
				break
			}

			// * 3. Executes the 3 phases: Select Word -> Draw -> Summary
			r.runTurnCycle()

			// * 4. Advance pointer to the next player
			r.mu.Lock()
			r.drawerIndex++
			r.mu.Unlock()
		}
	}

	// * 5. Game finished: declare winner & set state to GAME_OVER
	r.setPhase(game.PhaseGameOver)
	r.broadcastEvent(game.EventPhaseChange, game.PhaseChangePayload{
		Phase:       game.PhaseGameOver,
		RoundNumber: game.TotalRounds,
		MaxRounds:   game.TotalRounds,
	})
}

func (r *Room) runTurnCycle() {
	r.mu.Lock()
	currentDrawerID := r.drawingOrder[r.drawerIndex]
	// * Reset previous turn's guess status for every player
	for _, p := range r.players {
		p.HasGuessed = false
	}
	r.mu.Unlock()

	fmt.Printf("1. WORD SELECTION PHASE (15s)")

	// * 1. WORD SELECTION PHASE (15s)
	r.setPhase(game.PhaseWordSelection)
	choices := r.pickWordChoices(3)
	r.broadcastEvent(game.EventPhaseChange, game.PhaseChangePayload{
		Phase:       game.PhaseWordSelection,
		CurrentTurn: currentDrawerID,
		RoundNumber: r.currentRound,
		MaxRounds:   game.TotalRounds,
	})
	r.mu.Lock()
	r.currentWord = choices[0]
	fmt.Printf("currentWord: %s\n", r.currentWord)
	r.mu.Unlock()
	if !r.runCountdown(int(game.DurationWordSelect.Seconds())) {
		return
	}

	fmt.Printf("2. DRAWING PHASE (60s)")

	// * 2. DRAWING PHASE (60s)
	r.setPhase(game.PhaseDrawing)
	r.broadcastEvent(game.EventPhaseChange, game.PhaseChangePayload{
		Phase:       game.PhaseDrawing,
		CurrentTurn: currentDrawerID,
		RoundNumber: r.currentRound,
		MaxRounds:   game.TotalRounds,
	})
	r.broadcastEvent(game.EventWordSelected, game.WordSelectedPayload{
		DrawerID:   currentDrawerID,
		WordLength: len(r.currentWord),
	})
	if !r.runCountdown(int(game.DurationDrawing.Seconds())) {
		return
	}

	// * 3. ROUND SUMMARY PHASE (5s)
	r.setPhase(game.PhaseRoundSummary)
	r.mu.RLock()
	revealed := r.currentWord
	storeCopy := make(map[string]game.PlayerState)
	for k, v := range r.players {
		storeCopy[k] = *v
	}
	r.mu.RUnlock()
	r.broadcastEvent(game.EventRoundEnd, game.RoundEndPayload{
		RevealedWord: revealed,
		Scores:       storeCopy,
	})
	// * Wipe canvas at the end of every turn
	r.broadcastEvent(game.EventClearCanvas, map[string]interface{}{})

	_ = r.runCountdown(int(game.DurationSummary.Seconds()))
}

// * runCountdown ticks every 1 second and checks for room cancellation
func (r *Room) runCountdown(seconds int) bool {
	// * 1. Create a 1-second ticker and ensure it stops to prevent goroutine leaks
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// * 2. Loop backwards from 'seconds' down to 0
	for remaining := seconds; remaining >= 0; remaining-- {
		// * 3. Broadcast the current time to all clients in the room
		r.broadcastEvent(game.EventTimerTick, game.TimerTickPayload{
			RemainingSeconds: remaining,
		})

		select {
		case <-r.stopLoop:
			return false // * Abort ticker if players leave
		case <-ticker.C:
			// * 1 second passed -> loop continues to the next second
		}
	}

	return true
}

func (r *Room) setPhase(phase game.RoundPhase) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.State = phase
}

func (r *Room) pickWordChoices(count int) []string {
	shuffled := make([]string, len(game.WordBank))
	copy(shuffled, game.WordBank)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:count]
}

func (r *Room) broadcastEvent(eventType game.EventType, payload interface{}) {
	event := game.OutboundEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
	fmt.Printf("broadcast: %s\n", eventType)
	raw, err := json.Marshal(event)
	if err == nil {
		r.broadcastRaw(raw)
	}
}

func (r *Room) broadcastRaw(message []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for client := range r.clients {
		select {
		case client.Send <- message:
		default:
		}
	}
}

// * ClientCount returns the current number of active players (Read lock)
func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}
