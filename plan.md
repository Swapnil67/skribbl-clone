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
* [ ] Establish low-overhead WebSocket handshake mechanics and origin validation using `gorilla/websocket`.
* [ ] Construct thread-safe dual-goroutine client pipelines (`readPump`/`writePump`) with Ping/Pong keep-alives.
* [ ] Build a channel-driven in-memory `Hub` managing thread-safe client registration and event dispatching.
* [ ] **Canvas Sync Protocol:** Define strongly-typed event structs for real-time stroke coordinates, colors, actions, and chat messages.
* [ ] Deploy a minimal static HTML5 canvas harness to verify multi-client real-time synchronization.

### 🎮 Phase 2: In-Memory Game Loops & Concurrency (The Brain)
- [ ] **Thread-Safe State Engine:** Map room metadata using strict `sync.RWMutex` bindings to prevent race conditions during heavy write phases.
- [ ] **Channel Matchmaking Queue:** Construct background Goroutines to continually ingest active players from a pool into static 5-person rooms.
- [ ] **Non-Blocking Game Loops:** Manage dedicated per-room rounds via clean Go `time.Ticker` bindings.
- [ ] **Fuzzy Text Matcher:** Implement a Levenshtein distance string similarity check to calculate and flag proximity triggers (e.g., alert privately if a user guesses "elliphant").
- [ ] **Dynamic Score Matrix:** Code algorithmic logic weighting scores based on response times and active turn feedback.

### 🌐 Phase 3: Distributed State Scaling (The Senior Showcase)
- [ ] **Redis Pub/Sub Sync Layer:** Decouple connection tracking. Broadcast room state changes horizontally to allow Server A clients to natively view drawings initialized on Server B.
- [ ] **Resilient Client Recon Recovery:** Cache temporary state payloads in Redis using an expiration TTL. Allow dropped socket streams to reconnect seamlessly within 30 seconds without wiping point histories.
- [ ] **Token-Bucket Rate Limiter:** Build socket-level network interceptors to stop script injections from executing brute-force dictionary attacks.
- [ ] **Ping/Pong Heartbeats:** Run automated runtime ticker sweeps to quickly purge orphaned connections and keep memory overhead lean.

### 📊 Phase 4: Production Operations & Diagnostics (The Production Ready Check)
- [ ] **Structured Telemetry:** Hook up standard `slog` or `uber-go/zap` modules to emit raw JSON tracking logs marked with custom `correlation_id` tags.
- [ ] **Prometheus Exporters:** Configure a `/metrics` scrape target tracking operational performance data like active server room footprints, payload transmission delays, and active user drops.
- [ ] **Graceful Evacuation Triggers:** Intercept standard terminal kills (`SIGTERM`, `SIGINT`). Allow existing match processes to persist to completion before executing a clean thread shutdown.

---

## 📝 Resume Impact Narrative
*Copy-paste snippet for application pipelines:*

> **Distributed Real-time Multiplayer Engine (Golang, WebSockets, Redis)**
> * Designed and implemented a real-time multiplayer backend in Go capable of handling horizontal scaling across multiple nodes using **Redis Pub/Sub** for room state synchronization.
> * Architected a concurrent matchmaking system using **Go Goroutines and Channels**, reducing player wait times through an efficient background worker pool.
> * Mitigated data-race vulnerabilities in multi-threaded game state mutations by implementing strict memory locking mechanisms with **sync.Mutex**.
> * Built highly optimized WebSocket event handlers utilizing `gorilla/websocket`, minimizing memory overhead per active connection.
