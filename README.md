# Skribbl Clone

### 1. High-Level Multi-Node Architecture

```
                       [ Client A ]          [ Client B ]
                            │                     │
                            ▼                     ▼
                     ┌─────────────┐       ┌─────────────┐
                     │ Go Server 1 │       │ Go Server 2 │
                     └──────┬──────┘       └──────┬──────┘
                            │                     │
             Publish Events │                     │ Subscribe & Fan-Out
                            ▼                     ▼
             ┌──────────────────────────────────────────────────┐
             │                   REDIS CLUSTER                  │
             │                                                  │
             │  • Pub/Sub Bus    (room:{id}:events)             │
             │  • Hashes         (room:{id}:meta)               │
             │  • Sorted Sets    (room:{id}:leaderboard)        │
             │  • Lists/Streams  (room:{id}:strokes)            │
             │  • Sets           (room:{id}:members)            │
             └──────────────────────────────────────────────────┘

```

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
