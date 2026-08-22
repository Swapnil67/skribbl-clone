# 🚀 Project Blueprint: Distributed Real-Time Multiplayer Skribbl Clone

## 📌 Project Overview
A production-grade, highly scalable, distributed real-time multiplayer drawing and guessing game engineered in Golang. This project acts as a senior backend engineering showcase demonstrating high-throughput network architectures, distributed memory caching, and strict concurrency primitives.

---

## 🛠️ High-Level System Architecture
* **Networking:** Clean REST APIs + High-performance, non-blocking WebSockets.
* **Concurrency Control:** Thread-safe state machines managed via Go Mutexes and Channels.
* **Distributed Layer:** Horizontal scaling capability utilizing Redis Pub/Sub for cross-server message broadcasting.
* **Persistence Tier:** PostgreSQL for persistent master metrics and lifetime user authentication.
* **Observability:** Structured logging and Prometheus metrics exposure for production debugging.

---

## 🗺️ Implementation Roadmap

### 📦 Phase 1: Foundations & Core Networking (The Skeleton)
* [✓] Initialize standard Go directory layout (`cmd/`, `internal/game/`, `internal/transport/`).
* [✓] Setup standard HTTP router using `gin-gonic/gin` with health checks and CORS.
* [✓] Implement ephemeral guest session handler issuing lightweight tokenized session IDs.
* [✓] Establish low-overhead WebSocket handshake mechanics and origin validation using `gorilla/websocket`.
* [✓] Construct thread-safe dual-goroutine client pipelines (`readPump`/`writePump`) with Ping/Pong keep-alives.
* [✓] Build a channel-driven in-memory `Hub` managing thread-safe client registration and event dispatching.
* [✓] **Canvas Sync Protocol:** Define strongly-typed event structs for real-time stroke coordinates, colors, actions, and chat messages.
* [✓] Deploy a minimal static HTML5 canvas harness to verify multi-client real-time synchronization.

### 🎮 Phase 2: In-Memory Game Loops & Room Concurrency (The Brain)

* [✓] **Multi-Room Manager & Partitioned Hubs:** Refactor the global Hub into isolated per-room hubs using `sync.RWMutex` to partition drawing traffic and lifecycle events.
* [✓] **Channel Matchmaking Engine:** Build a centralized pool worker goroutine routing incoming guest sessions into 5-player public rooms or private invite codes.
* [] **State Machine & Ticker Game Loop:** Manage per-room match phases (`LOBBY`, `WORD_SELECTION`, `DRAWING`, `ROUND_SUMMARY`) using non-blocking Go `time.Ticker` bindings.
* [] **Fuzzy Guess Matcher:** Implement a Levenshtein distance string similarity engine to validate chat guesses and send private "close guess" alerts.
* [] **Dynamic Scoring & Turn Arbiter:** Calculate decay scores based on response latency, rotate active drawer permissions, and censor correct guesses in public room chat.

### 🌐 Phase 3: Distributed State & Scale (The Senior Showcase)

* [] **Redis Pub/Sub Transport Bus:** Decouple room hubs by routing drawing streams and chat events through per-room Redis channels (`room:{id}:events`).
* [] **Distributed Room State in Redis:** Persist active room metadata, player scores, and round states into Redis Hashes with TTLs for cross-server visibility.
* [] **Ephemeral Session Reconnection:** Cache disconnected guest state in Redis for 30 seconds to allow seamless browser refresh recovery without point loss.
* [] **Token-Bucket Chat Rate Limiter:** Implement a per-socket rate limiter in Go to throttle spam and mitigate automated dictionary attacks on guesses.

### 📊 Phase 4: Production Diagnostics & Observability (The Production Ready Check)

* [] **Structured Logging with `log/slog`:** Implement context-aware JSON structured logging tagged with `session_id` and `room_id`.
* [] **Prometheus Metrics Scrape Target:** Expose `/metrics` tracking active rooms, connected WebSocket clients, draw frame rates, and guess evaluation latencies.
* [] **Refined Graceful Draining:** Extend existing signal traps to broadcast match-ending notifications over sockets and wait for game tickers to cleanly exit before process termination.

---

## 📝 Resume Impact Narrative
*Copy-paste snippet for application pipelines:*

> **Distributed Real-time Multiplayer Engine (Golang, WebSockets, Redis)**
> * Designed and implemented a real-time multiplayer backend in Go capable of handling horizontal scaling across multiple nodes using **Redis Pub/Sub** for room state synchronization.
> * Architected a concurrent matchmaking system using **Go Goroutines and Channels**, reducing player wait times through an efficient background worker pool.
> * Mitigated data-race vulnerabilities in multi-threaded game state mutations by implementing strict memory locking mechanisms with **sync.Mutex**.
> * Built highly optimized WebSocket event handlers utilizing `gorilla/websocket`, minimizing memory overhead per active connection.