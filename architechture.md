
### Summary of Structure

```
                  ┌────────────────────────────────────────┐
                  │              RoomManager               │
                  │   - Manages map[room_id]*Room          │
                  │   - Thread-safe lookup with RWMutex    │
                  └──────────────────┬─────────────────────┘
                                     │
                     ┌───────────────┴───────────────┐
                     ▼                               ▼
            ┌──────────────────┐            ┌──────────────────┐
            │      Room A      │            │      Room B      │
            │  - Own Event Loop│            │  - Own Event Loop│
            │  - In-Room Hub   │            │  - In-Room Hub   │
            │  - State Machine │            │  - State Machine │
            └────────┬─────────┘            └──────────────────┘
                     │
           ┌─────────┴─────────┐
           ▼                   ▼
    ┌─────────────┐     ┌─────────────┐
    │  Client 1   │     │  Client 2   │
    │  (read/write│     │  (read/write│
    │    pumps)   │     │    pumps)   │
    └─────────────┘     └─────────────┘

```

### The Core Problem: Why Do We Need an Autonomous Game Loop?

A multiplayer game like Skribbl requires the server to be **proactive**:

1. It needs an internal clock to tick down from 60 to 0 seconds.
2. It must transition through states on its own (e.g., when time runs out $\rightarrow$ reveal the word $\rightarrow$ award points $\rightarrow$ switch the drawer).
3. It must enforce server-side authority (e.g., non-drawers cannot send stroke coordinates).

---

### 1. The Finite State Machine (FSM) Lifecycle

A game round moves through a strict cycle of phases:

```
                  ┌───────────────────────────────┐
                  │             LOBBY             │ ◄── Waiting for ≥ 2 players
                  └───────────────┬───────────────┘
                                  │ (2 players join)
                                  ▼
                  ┌───────────────────────────────┐
                  │        WORD_SELECTION         │ ◄── 15s: Active drawer picks 1 of 3 words
                  └───────────────┬───────────────┘
                                  │ (Word picked or 15s expires)
                                  ▼
                  ┌───────────────────────────────┐
                  │            DRAWING            │ ◄── 60s: Drawer draws, guessers guess in chat
                  └───────────────┬───────────────┘
                                  │ (60s expires OR everyone guesses correctly)
                                  ▼
                  ┌───────────────────────────────┐
                  │         ROUND_SUMMARY         │ ◄── 5s: Show word & scoreboard, wipe canvas
                  └───────────────┬───────────────┘
                                  │
                   ┌──────────────┴──────────────┐
     More players/rounds left?                   All 3 rounds finished?
                   ▼                                       ▼
       (Loop back to WORD_SELECTION)                  [ GAME_OVER ]

```

---

### 2. How `time.Ticker` Runs Without Blocking WebSocket Traffic

In Go, if you use `time.Sleep(60 * time.Second)` in the main thread, the entire server freezes and stops handling incoming network messages.

Instead, we run `startGameLoop()` inside its own **independent background goroutine**:

* **Thread 1 (`room.Run()`):** Continuously handles incoming client connects, disconnects, and chat events.
* **Thread 2 (`room.startGameLoop()`):** Manages the game clock and state transitions using `time.NewTicker(1 * time.Second)`.

```
Goroutine 1: room.Run()            Goroutine 2: startGameLoop()
───────────────────────            ────────────────────────────
Listens on channels:               Ticks every 1 second:
• Client joins                     • 59s... emit TIMER_TICK
• Client leaves                    • 58s... emit TIMER_TICK
• Chat message arrives             • ...
                                   • 00s... emit PHASE_CHANGE (ROUND_SUMMARY)

```

---

### 3. Server-Side Permission Checks (Anti-Cheat)

In client-side JavaScript, anyone can open DevTools and trigger `sendEvent("DRAW_STROKE", ...)`.

Step 3 enforces **server-side authorization**:
When a `DRAW_STROKE` frame arrives in `ReadPump`:

1. The server checks the room's current state: is it `DRAWING`?
2. The server checks the active turn: does `SessionID` match `drawerOrder[drawerIndex]`?
3. If **no**, the server discards the packet immediately without broadcasting it to other players.

---

### 4. Cancellation & Leak Prevention (`stopLoop`)

If all players close their browser tabs while a round is active:

* `stopLoop := make(chan struct{})` is closed.
* The `select` block inside `runCountdown()` immediately intercepts the cancellation signal and terminates the timer goroutine, preventing orphaned memory leaks in the Go runtime.


### Routes

```bash
# Create New Room
curl -i -X POST http://localhost:8080/api/v1/rooms \
  -H "Content-Type: application/json" \
  -d '{"max_players": 6}'

# Get Room Info
curl -i -X GET http://localhost:8080/api/v1/rooms/5ZHN48 \
  -H "Content-Type: application/json" 
```