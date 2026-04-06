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
| Auth | Discord OAuth2 → JWT | Player authentication |
| Admin Auth | GitHub OAuth2 → JWT | Admin authentication |
| Deploy | Railway | 4-service deployment |

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
│       │   ├── github.go          # GitHub OAuth flow (admin)
│       │   ├── jwt.go             # JWT generation + validation (with role claim)
│       │   ├── middleware.go      # Player auth middleware (header or cookie)
│       │   └── admin_middleware.go # Admin auth middleware (checks role=admin)
│       ├── db/
│       │   ├── db.go              # pgxpool + embedded migrations
│       │   └── migrations/        # SQL schema files
│       ├── models/models.go       # Shared data types
│       ├── handler/
│       │   ├── auth_handler.go    # Discord login, /me, logout
│       │   ├── admin_auth_handler.go # GitHub login, admin /me
│       │   ├── admin_handler.go   # Admin CRUD for all reference data
│       │   ├── admin_import_handler.go # Bulk CSV/JSON import
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
│       ├── seed/                  # CSV/JSON import logic (reused by admin)
│       └── pkg/invite/code.go     # Invite code generation
│
├── admin/                         # Admin management frontend
│   ├── package.json               # React 18 + Vite 5 + Tailwind 4
│   ├── vite.config.ts
│   ├── src/
│   │   ├── App.tsx                # Router + AuthGuard (GitHub auth)
│   │   ├── api/
│   │   │   ├── client.ts          # REST client with admin_token
│   │   │   ├── auth.ts            # GitHub auth API
│   │   │   └── admin.ts           # CRUD + import API for all entities
│   │   ├── hooks/useAuth.ts       # GitHub auth context
│   │   ├── components/
│   │   │   ├── Layout.tsx         # Sidebar nav + content area
│   │   │   ├── DataTable.tsx      # Reusable table with search/actions
│   │   │   └── ImportDialog.tsx   # File upload with preview
│   │   └── pages/                 # List + Edit pages per entity:
│   │       ├── factions/          #   Factions
│   │       ├── detachments/       #   Detachments
│   │       ├── stratagems/        #   Stratagems
│   │       ├── mission-packs/     #   Mission Packs
│   │       ├── missions/          #   Missions (with scoring rules sub-form)
│   │       ├── secondaries/       #   Secondaries (with scoring options sub-form)
│   │       ├── gambits/           #   Gambits
│   │       ├── challenger-cards/  #   Challenger Cards
│   │       └── mission-rules/     #   Mission Rules
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

### 8. Separate Admin Auth via GitHub OAuth

The admin management interface uses GitHub OAuth2, completely separate from the player Discord OAuth. Admin JWTs carry a `role: "admin"` claim. Access is controlled via a `ADMIN_GITHUB_IDS` environment variable (comma-separated GitHub user IDs).

**Why**: Decouples admin access from the player identity system. GitHub is a natural fit for developers managing the data. The allowlist approach is simple and sufficient for a single admin.

### 7. Invite Codes (not User Lookup)

Games are joined via 6-character alphanumeric codes (excluding ambiguous chars like O/0/I/1/L). No friends list or user search.

**Why**: Simple, privacy-preserving. Works across any communication channel (Discord, text, in-person).

---

## REST API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/health` | Health check → `{"status":"ok"}` |
| `GET` | `/api/auth/discord` | Redirect to Discord OAuth (players) |
| `GET` | `/api/auth/discord/callback` | Discord OAuth callback |
| `GET` | `/api/auth/github` | Redirect to GitHub OAuth (admin) |
| `GET` | `/api/auth/github/callback` | GitHub OAuth callback |

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

### Admin (require admin JWT with `role: "admin"`)

All admin endpoints follow the same CRUD pattern per entity:

| Method | Path Pattern | Description |
|--------|-------------|-------------|
| `GET` | `/api/admin/{entity}` | List all (with optional `?faction_id=`, `?pack_id=` filters) |
| `GET` | `/api/admin/{entity}/{id}` | Get single |
| `POST` | `/api/admin/{entity}` | Create |
| `PUT` | `/api/admin/{entity}/{id}` | Update |
| `DELETE` | `/api/admin/{entity}/{id}` | Delete |

**Entities:** `factions`, `detachments`, `stratagems`, `mission-packs`, `missions`, `secondaries`, `gambits`, `challenger-cards`, `mission-rules`

**Admin Auth:**

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/me` | Current admin user info |

**Bulk Import:**

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/admin/import/factions` | Upload CSV, upsert factions |
| `POST` | `/api/admin/import/stratagems` | Upload CSV, upsert stratagems + detachments |
| `POST` | `/api/admin/import/missions` | Upload JSON, upsert all mission data |

Import endpoints accept `multipart/form-data` with a `file` field and reuse the existing seed logic.

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
| `missions` | TEXT | mission_pack_id, name, lore, description, scoring_rules (JSONB), scoring_timing |
| `secondaries` | TEXT | mission_pack_id, name, lore, description, max_vp, is_fixed, scoring_options (JSONB) |
| `gambits` | TEXT | mission_pack_id, name, description, vp_value |
| `mission_rules` | TEXT | mission_pack_id, name, lore, description |
| `challenger_cards` | TEXT | mission_pack_id, name, lore, description |

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

### Player Frontend (`frontend/`)

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

### Admin Frontend (`admin/`)

Same tech stack as the player frontend (React 18 + TypeScript + Vite 5 + Tailwind CSS 4) but as a separate standalone Vite application.

**Page Flow:**
```
LoginPage → DashboardPage → EntityListPage → EntityEditPage
```

**Key Components:**
- `Layout` — Sidebar navigation with links to all entity types
- `DataTable` — Generic table component with search, inline delete confirmation
- `ImportDialog` — File upload modal with progress and result summary

**API Pattern:**
- `adminApi` — Generic CRUD factory (`list`, `get`, `create`, `update`, `delete`) per entity
- `uploadFile` — Separate function for `multipart/form-data` imports
- Auth token stored in `localStorage` under `admin_token` (separate from player `token`)

---

## Deployment (Railway)

| Service | Build | Port | Health Check |
|---------|-------|------|-------------|
| backend | `backend/Dockerfile` (Go → alpine) | 8080 | `GET /api/health` |
| frontend | `frontend/Dockerfile` (Vite → nginx) | 80 | — |
| admin | `admin/` (Vite build → nginx) | 80 | — |
| PostgreSQL | Railway managed plugin | 5432 | — |

### Environment Variables

**Backend:**
- `DATABASE_URL` — Postgres connection string (auto-injected by Railway)
- `DISCORD_CLIENT_ID` — Discord application client ID
- `DISCORD_CLIENT_SECRET` — Discord application secret
- `DISCORD_REDIRECT_URI` — e.g. `https://api.myapp.railway.app/api/auth/discord/callback`
- `JWT_SECRET` — Random secret for signing JWTs (shared by player + admin auth)
- `FRONTEND_URL` — e.g. `https://myapp.railway.app` (for CORS)
- `PORT` — `8080`
- `GITHUB_CLIENT_ID` — GitHub OAuth application client ID (admin auth)
- `GITHUB_CLIENT_SECRET` — GitHub OAuth application secret (admin auth)
- `GITHUB_REDIRECT_URI` — e.g. `https://api.myapp.railway.app/api/auth/github/callback`
- `ADMIN_GITHUB_IDS` — Comma-separated GitHub user IDs allowed admin access
- `ADMIN_FRONTEND_URL` — e.g. `https://admin.myapp.railway.app` (for CORS + redirect)

**Frontend (build-time):**
- `VITE_API_URL` — Backend URL, e.g. `https://api.myapp.railway.app`
- `VITE_WS_URL` — Backend WS URL, e.g. `wss://api.myapp.railway.app`

**Admin Frontend (build-time):**
- `VITE_API_URL` — Backend URL (same as player frontend)

---

## Local Development

```bash
# 1. Start Postgres
make db-start

# 2. Copy and configure env
cp backend/.env.example backend/.env
# Edit .env with Discord + GitHub OAuth credentials + ADMIN_GITHUB_IDS

# 3. Seed database
make seed

# 4. Start backend (terminal 1)
make dev-backend

# 5. Start player frontend (terminal 2)
make dev-frontend

# 6. Start admin frontend (terminal 3, optional)
make dev-admin
```

The player frontend runs on port 5173, the admin frontend on port 5174, and the backend on port 8080.

Prerequisites: Go 1.25+, Node 20+, Docker

---

## Admin Management Interface

The admin interface is a standalone React application (`admin/`) for managing all reference data. It uses the same Go backend but authenticates via GitHub OAuth instead of Discord.

### Auth Flow

1. Admin visits admin frontend → clicks "Sign in with GitHub"
2. Redirects to `GET /api/auth/github` → GitHub OAuth consent screen
3. GitHub redirects back to `GET /api/auth/github/callback`
4. Backend verifies GitHub user ID is in `ADMIN_GITHUB_IDS` allowlist
5. Generates JWT with `role: "admin"` claim → redirects to admin frontend with token
6. Admin frontend stores token in `localStorage` as `admin_token`

Admin JWTs are validated by `AdminMiddleware`, which checks for the `role: "admin"` claim. This is separate from the player `Middleware` — player tokens cannot access admin routes and vice versa.

### Managed Entities

| Entity | Key Fields | Notes |
|--------|-----------|-------|
| Factions | id, name, wahapediaLink | Importable via CSV |
| Detachments | id, factionId, name | Filterable by faction |
| Stratagems | id, factionId, detachmentId, name, type, cpCost, phase... | Importable via CSV, filterable by faction |
| Mission Packs | id, name, description | Parent entity for missions/secondaries/gambits |
| Missions | id, missionPackId, name, description, scoringRules (JSONB), scoringTiming | Structured sub-form for scoring rules |
| Secondaries | id, missionPackId, name, maxVp, isFixed, scoringOptions (JSONB) | Structured sub-form for scoring options |
| Gambits | id, missionPackId, name, vpValue | |
| Challenger Cards | id, missionPackId, name, lore, description | |
| Mission Rules | id, missionPackId, name, lore, description | |

### Import Feature

The admin UI supports bulk import via file upload. Import endpoints accept `multipart/form-data` and reuse the existing `seed` package logic:

- **Factions CSV**: Pipe-delimited (`id|name|wahapedia_link`)
- **Stratagems CSV**: Pipe-delimited, auto-creates detachments from columns 8-9
- **Missions JSON**: Array of `{id, lore, body}` entries, classifies by ID prefix into missions/secondaries/gambits/challenger cards/mission rules
