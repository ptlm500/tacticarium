# Tacticarium — Architecture Documentation

A real-time, mobile-first turn tracker for Warhammer 40K 10th Edition. Two players on separate devices track command points, victory points, phases, and stratagems across 5 battle rounds.

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Frontend | React 18 + TypeScript | UI framework |
| Build | Vite 5 | Dev server + bundler |
| Styling | Tailwind CSS 4 | Utility-first CSS |
| State | Zustand | Lightweight state management |
| Routing | React Router v6 | Client-side routing |
| Backend | Go 1.22 | API server |
| Router | chi v5 | HTTP routing + middleware |
| WebSocket | nhooyr.io/websocket | Real-time bidirectional comms |
| Database | PostgreSQL 16 | Persistent storage |
| DB Driver | pgx v5 + pgxpool | Connection pooling |
| Auth | Discord OAuth2 → JWT | User authentication |
| Deploy | Railway | 3-service deployment |

---

## Project Structure

```
tacticarium/
├── Factions.csv                   # Source data: 27 factions (pipe-delimited)
├── Stratagems.csv                 # Source data: 1,397 stratagems (pipe-delimited)
├── docker-compose.yml             # Local Postgres 16
├── Makefile                       # Dev commands
│
├── backend/
│   ├── Dockerfile                 # Multi-stage: Go build → alpine
│   ├── go.mod
│   ├── cmd/
│   │   ├── server/main.go         # HTTP + WS server entrypoint
│   │   └── seed/main.go           # Data seeding CLI
│   └── internal/
│       ├── config/config.go       # Environment configuration
│       ├── auth/
│       │   ├── discord.go         # Discord OAuth flow
│       │   ├── jwt.go             # JWT generation + validation
│       │   └── middleware.go      # Auth middleware (header or cookie)
│       ├── db/
│       │   ├── db.go              # pgxpool + embedded migrations
│       │   └── migrations/        # SQL schema files
│       ├── models/models.go       # Shared data types
│       ├── handler/
│       │   ├── auth_handler.go    # Discord login, /me, logout
│       │   ├── faction_handler.go # Factions, detachments, stratagems
│       │   ├── mission_handler.go # Mission packs, missions, secondaries
│       │   └── game_handler.go    # Game CRUD, WS upgrade, persistence
│       ├── game/
│       │   ├── engine.go          # Core state machine (Apply)
│       │   ├── state.go           # GameState, PlayerState types
│       │   ├── actions.go         # Action + Event type definitions
│       │   └── rules.go           # 10th edition constants + phase logic
│       ├── ws/
│       │   ├── hub.go             # Global room manager
│       │   ├── room.go            # Per-game room (goroutine)
│       │   ├── client.go          # Per-connection read/write pumps
│       │   └── protocol.go        # Message constructors
│       ├── seed/                  # CSV import logic
│       └── pkg/invite/code.go     # Invite code generation
│
├── frontend/
│   ├── Dockerfile                 # Multi-stage: Vite build → nginx
│   ├── nginx.conf                 # SPA fallback config
│   ├── src/
│   │   ├── App.tsx                # Router + AuthGuard
│   │   ├── api/                   # REST client (auth, games, factions, missions)
│   │   ├── hooks/
│   │   │   ├── useAuth.ts         # Auth context + Discord login
│   │   │   ├── useWebSocket.ts    # WS connection + reconnect
│   │   │   └── useGameState.ts    # WS → Zustand bridge
│   │   ├── stores/gameStore.ts    # Zustand game state
│   │   ├── types/                 # TypeScript type definitions
│   │   ├── pages/                 # Login, Lobby, Setup, Game, History
│   │   └── components/
│   │       ├── game/              # PhaseTracker, CPCounter, VPCounter, etc.
│   │       └── setup/             # FactionPicker, DetachmentPicker
│
└── scraper/                       # (Planned) Playwright mission scraper
```

---

## Key Architectural Decisions

### 1. Server-Authoritative Game Engine

All game state mutations flow through `engine.Apply(action) → ([]events, error)`. The engine validates every action against the current state and rejects invalid ones. Clients receive the full authoritative state after each action — there is no client-side game logic.

**Why**: Prevents cheating and state desync between players. The server is the single source of truth.

### 2. Actor Model for WebSocket Rooms

Each active game runs in its own goroutine via `Room.Run()`. The room processes actions sequentially through channels, eliminating race conditions without fine-grained locking.

```
Hub (1 instance) ──manages──> Room (1 per game) ──contains──> Client (1 per connection)
```

**Why**: Simple concurrency model. Sequential action processing means the engine never needs to be thread-safe.

### 3. Full State Broadcast (not Delta)

After every action, the server broadcasts the complete `GameState` to all connected clients, not just a delta.

**Why**: Simplifies reconnection (client just receives full state), eliminates delta-application bugs, and GameState is small enough (~2KB) that bandwidth is not a concern.

### 4. Dual Persistence

Game state lives in memory (in the Engine) during active play, but is also persisted to PostgreSQL after every action via the `OnStateChange` callback.

**Why**: Fast in-memory reads during gameplay, but survives server restarts. When a player reconnects, state is loaded from DB if the room was garbage-collected.

### 5. Event Sourcing (Audit Trail)

Every action produces typed `GameEvent` records that are both broadcast to clients and persisted to the `game_events` table with JSONB payloads.

**Why**: Complete game replay capability, debugging, and the game log UI.

### 6. JWT via Cookie + Query Param

- REST endpoints: JWT in `token` cookie (HttpOnly, SameSite=Lax) or `Authorization: Bearer` header
- WebSocket: JWT in query parameter (`?token=...`) because browsers don't support custom headers on WebSocket upgrade

**Why**: Cookies for seamless browser auth; query param as the only option for WS.

### 7. Invite Codes (not User Lookup)

Games are joined via 6-character alphanumeric codes (excluding ambiguous chars like O/0/I/1/L). No friends list or user search.

**Why**: Simple, privacy-preserving. Works across any communication channel (Discord, text, in-person).

---

## REST API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/health` | Health check → `{"status":"ok"}` |
| `GET` | `/api/auth/discord` | Redirect to Discord OAuth |
| `GET` | `/api/auth/discord/callback` | OAuth callback, sets JWT cookie |

### Authenticated (require JWT)

**Auth:**

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/auth/me` | Current user profile |
| `POST` | `/api/auth/logout` | Clear auth cookie |

**Reference Data:**

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/factions` | All 27 factions |
| `GET` | `/api/factions/{factionId}/detachments` | Detachments for a faction |
| `GET` | `/api/factions/{factionId}/stratagems` | All stratagems for a faction |
| `GET` | `/api/detachments/{detachmentId}/stratagems` | Stratagems for a specific detachment |
| `GET` | `/api/mission-packs` | All mission packs |
| `GET` | `/api/mission-packs/{packId}/missions` | Missions in a pack |
| `GET` | `/api/mission-packs/{packId}/secondaries` | Secondary objectives |
| `GET` | `/api/mission-packs/{packId}/gambits` | Gambit cards |

**Games:**

| Method | Path | Description | Returns |
|--------|------|-------------|---------|
| `POST` | `/api/games` | Create game | `{id, inviteCode}` |
| `GET` | `/api/games` | List user's games (max 50) | `GameSummary[]` |
| `POST` | `/api/games/join/{code}` | Join by invite code | `{id, inviteCode}` |
| `GET` | `/api/games/{gameId}` | Full game state | `GameState` |
| `GET` | `/api/games/{gameId}/events` | Event history | `GameEvent[]` |
| `GET` | `/api/users/me/history` | Completed games | `GameSummary[]` |

**WebSocket:**

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/ws/game/{gameId}?token={jwt}` | Upgrade to WebSocket |

---

## WebSocket Protocol

### Client → Server

| Message Type | Payload | Description |
|-------------|---------|-------------|
| `action` | `{type: ActionType, ...data}` | Game action |
| `ping` | (none) | Keepalive |
| `sync_request` | (none) | Request full state |

### Server → Client

| Message Type | Payload | Description |
|-------------|---------|-------------|
| `state_update` | `GameState` | Full authoritative state |
| `event` | `GameEvent` | Individual event for game log |
| `error` | `{message, code}` | Action rejected |
| `pong` | (none) | Keepalive response |
| `player_connected` | `{playerNumber, username}` | Opponent connected |
| `player_disconnected` | `{playerNumber}` | Opponent disconnected |

### Action Types

| Action | Data | Phase Restriction |
|--------|------|-------------------|
| `select_faction` | `{factionId, factionName}` | Setup only |
| `select_detachment` | `{detachmentId, detachmentName}` | Setup only |
| `select_mission` | `{missionPackId, missionId, missionName}` | Setup only |
| `select_secondary` | `{secondaryId?, customName?, customMaxVp?}` | Setup only |
| `remove_secondary` | `{secondaryId}` | Setup only |
| `set_ready` | `{ready: bool}` | Setup only |
| `advance_phase` | (none) | Active, active player only |
| `adjust_cp` | `{delta: int}` | Active |
| `score_vp` | `{category: "primary"\|"secondary"\|"gambit", delta: int}` | Active |
| `use_stratagem` | `{stratagemId, stratagemName, cpCost}` | Active |
| `declare_gambit` | `{gambitId}` | Active, round ≥ 3 |
| `concede` | (none) | Active |
| `set_paint_score` | `{score: int}` | Any |

---

## Database Schema

### Reference Data Tables

| Table | PK | Key Columns |
|-------|-----|------------|
| `users` | UUID | discord_id (unique), discord_username, discord_avatar |
| `factions` | TEXT | name, wahapedia_link |
| `detachments` | TEXT | faction_id → factions, name |
| `stratagems` | TEXT | faction_id, detachment_id, name, type, cp_cost, turn, phase, description |
| `mission_packs` | TEXT | name, description |
| `missions` | UUID | mission_pack_id → mission_packs, name, deployment_map, rules_text |
| `secondaries` | UUID | mission_pack_id, name, category, description, max_vp |
| `gambits` | UUID | mission_pack_id, name, description, vp_value |

### Game Tables

| Table | PK | Key Columns |
|-------|-----|------------|
| `games` | UUID | invite_code (unique), status, current_round, current_phase, active_player, winner_id |
| `game_players` | UUID | game_id, user_id, player_number (1\|2), faction_id, cp, vp_primary/secondary/gambit/paint, is_ready |
| `game_player_secondaries` | UUID | game_player_id, secondary_id, custom_name, vp_scored |
| `game_events` | BIGSERIAL | game_id, player_number, event_type, event_data (JSONB), round, phase |
| `stratagem_usage` | BIGSERIAL | game_id, game_player_id, stratagem_id, round, phase, cp_spent |

---

## Game Flow

### 1. Setup Phase

```
Player 1: Create Game → POST /api/games → {id, inviteCode}
Player 1: Share invite code to Player 2
Player 2: Join Game → POST /api/games/join/{code}
Both:     Connect WebSocket → GET /ws/game/{id}?token={jwt}
Both:     Select faction → action: select_faction
Both:     Select detachment → action: select_detachment
Both:     Ready up → action: set_ready
          When both ready → Game starts (status=active, round=1, phase=command)
```

### 2. Active Game (per round)

```
Round N (1-5):
  Player A's Turn:
    Command Phase  → (auto +1 CP if round ≥ 2)
    Movement Phase
    Shooting Phase
    Charge Phase
    Fight Phase    → Turn ends, switch to Player B
  Player B's Turn:
    Command Phase  → (auto +1 CP if round ≥ 2)
    Movement Phase
    Shooting Phase
    Charge Phase
    Fight Phase    → Round ends, advance to round N+1

During any phase:
  - Active player can: advance_phase, adjust_cp, score_vp, use_stratagem
  - Reactive player can: adjust_cp, score_vp, use_stratagem (opponent's turn stratagems)
```

### 3. Game End

- **Round 5 complete**: Winner = highest total VP. Tie if equal.
- **Concede**: Winner = opponent.
- Final state persisted, game appears in history.

### VP Caps

| Category | Max |
|----------|-----|
| Primary | 50 |
| Secondary | 40 |
| Gambit | 12 |
| Paint | 10 |

---

## Stratagem Filtering Logic

When displaying available stratagems for a player during a specific phase:

1. Include stratagems matching the player's `faction_id` (or no faction = core)
2. Include stratagems matching the player's `detachment_id` (or no detachment = faction-wide)
3. **Phase filter**: Match current phase name, or `"Any phase"`, or compound phases like `"Shooting or Fight phase"`
4. **Turn filter**:
   - Active player sees: `"Your turn"` + `"Either player's turn"`
   - Reactive player sees: `"Opponent's turn"` + `"Either player's turn"`
5. Validate CP affordability (UI disables "Use" button if insufficient)

---

## Frontend Architecture

### Page Flow

```
LoginPage → LobbyPage → GameSetupPage → GamePage
               │                            │
               ↓                            ↓
         GameHistoryPage            (game ends → result screen)
```

### State Management

- **Zustand store** (`gameStore.ts`): Holds `gameState`, `events[]`, `error`, `opponentConnected`
- **useWebSocket hook**: Manages WS connection lifecycle, exponential backoff reconnection (1s → 30s), 30s ping interval
- **useGameConnection hook**: Routes incoming WS messages to the Zustand store

### Data Flow

```
WebSocket message → useWebSocket.onMessage → useGameConnection.handleMessage → Zustand store → React components
User action → sendAction() → WebSocket → Server engine → broadcast state_update → all clients
```

---

## Deployment (Railway)

| Service | Build | Port | Health Check |
|---------|-------|------|-------------|
| backend | `backend/Dockerfile` (Go → alpine) | 8080 | `GET /api/health` |
| frontend | `frontend/Dockerfile` (Vite → nginx) | 80 | — |
| PostgreSQL | Railway managed plugin | 5432 | — |

### Environment Variables

**Backend:**
- `DATABASE_URL` — Postgres connection string (auto-injected by Railway)
- `DISCORD_CLIENT_ID` — Discord application client ID
- `DISCORD_CLIENT_SECRET` — Discord application secret
- `DISCORD_REDIRECT_URI` — e.g. `https://api.myapp.railway.app/api/auth/discord/callback`
- `JWT_SECRET` — Random secret for signing JWTs
- `FRONTEND_URL` — e.g. `https://myapp.railway.app` (for CORS)
- `PORT` — `8080`

**Frontend (build-time):**
- `VITE_API_URL` — Backend URL, e.g. `https://api.myapp.railway.app`
- `VITE_WS_URL` — Backend WS URL, e.g. `wss://api.myapp.railway.app`

---

## Local Development

```bash
# 1. Start Postgres
make dev-db

# 2. Copy and configure env
cp backend/.env.example backend/.env
# Edit .env with Discord credentials

# 3. Seed database
make seed

# 4. Start backend (terminal 1)
make dev-backend

# 5. Start frontend (terminal 2)
make dev-frontend
```

Prerequisites: Go 1.22+, Node 20+, Docker
