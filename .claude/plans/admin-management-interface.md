# Admin Management Interface Plan

## Overview

Add a web-based admin interface for managing all reference data (factions, stratagems, missions, secondaries, gambits, challenger cards, mission rules). Authenticated via GitHub OAuth (separate from the player Discord auth). Same Go backend, separate React frontend.

---

## Architecture

```
┌─────────────────────┐     ┌──────────────────────────────────┐
│  Admin Frontend      │     │  Player Frontend                  │
│  (Vite + React)      │     │  (Vite + React)                   │
│  Port 5174           │     │  Port 5173                        │
│  /admin/...          │     │  /...                             │
└────────┬────────────┘     └──────────┬───────────────────────┘
         │                             │
         │  POST/GET /api/admin/...    │  GET /api/... + WS
         ▼                             ▼
┌──────────────────────────────────────────────────────────────┐
│  Go Backend (chi router, port 8080)                          │
│                                                              │
│  /api/auth/github          ← GitHub OAuth (admin)            │
│  /api/auth/discord         ← Discord OAuth (players)         │
│  /api/admin/...            ← Admin CRUD (GitHub JWT required) │
│  /api/...                  ← Player read-only (Discord JWT)   │
└──────────────────────────────────────────────────────────────┘
```

### Key Decisions
- **Same backend server** — admin routes live under `/api/admin/` with separate middleware
- **Separate JWT namespace** — admin JWTs include a `role: "admin"` claim to distinguish from player JWTs (same secret is fine)
- **GitHub user allowlist** — env var `ADMIN_GITHUB_IDS` (comma-separated GitHub user IDs) controls who gets admin access
- **Separate frontend** — new `admin/` directory at repo root, independent Vite app with its own `package.json`
- **CORS** — backend accepts both `FRONTEND_URL` and `ADMIN_FRONTEND_URL` origins

---

## Phase 1: GitHub OAuth & Admin Auth Infrastructure

### 1a. Backend — GitHub OAuth provider

New file: `backend/internal/auth/github.go`
- `GitHubConfig` struct (ClientID, ClientSecret, RedirectURI) — mirrors `DiscordConfig`
- `AuthURL(state)` → GitHub authorize URL with `read:user` scope
- `ExchangeCode(code)` → exchange for access token
- `FetchGitHubUser(accessToken)` → GET `https://api.github.com/user`, return `{ID, Login, AvatarURL}`

### 1b. Backend — Config changes

Update `backend/internal/config/config.go`:
- Add fields: `GitHubClientID`, `GitHubClientSecret`, `GitHubRedirectURI`, `AdminGitHubIDs`, `AdminFrontendURL`
- Env vars: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URI`, `ADMIN_GITHUB_IDS`, `ADMIN_FRONTEND_URL`

### 1c. Backend — Admin auth handler

New file: `backend/internal/handler/admin_auth_handler.go`
- `HandleGitHubRedirect` — set state cookie, redirect to GitHub
- `HandleGitHubCallback` — exchange code, fetch user, check GitHub ID against allowlist, generate admin JWT (with `role: "admin"` claim), redirect to admin frontend with token
- `HandleAdminMe` — return current admin user info

### 1d. Backend — Admin auth middleware

New file: `backend/internal/auth/admin_middleware.go`
- Reuses existing JWT validation
- Additionally checks that the `role` claim is `"admin"`
- Injects `AdminContext` into request context

### 1e. Backend — Router changes

Update `backend/internal/server/router.go`:
- Add CORS origin for `AdminFrontendURL`
- Add public routes: `GET /api/auth/github`, `GET /api/auth/github/callback`
- Add admin route group with admin middleware:
  ```
  r.Group(func(r chi.Router) {
      r.Use(auth.AdminMiddleware(cfg.JWTSecret))
      // ... admin CRUD routes (Phase 2)
  })
  ```

---

## Phase 2: Admin CRUD API Endpoints

All routes prefixed with `/api/admin/`, protected by admin middleware.

### 2a. Factions handler

New file: `backend/internal/handler/admin_faction_handler.go`

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/admin/factions` | List all factions |
| GET | `/api/admin/factions/{id}` | Get single faction |
| POST | `/api/admin/factions` | Create faction |
| PUT | `/api/admin/factions/{id}` | Update faction |
| DELETE | `/api/admin/factions/{id}` | Delete faction |

### 2b. Detachments handler

New file: `backend/internal/handler/admin_detachment_handler.go`

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/admin/detachments` | List all (filterable by `?faction_id=`) |
| GET | `/api/admin/detachments/{id}` | Get single |
| POST | `/api/admin/detachments` | Create |
| PUT | `/api/admin/detachments/{id}` | Update |
| DELETE | `/api/admin/detachments/{id}` | Delete |

### 2c. Stratagems handler

New file: `backend/internal/handler/admin_stratagem_handler.go`

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/admin/stratagems` | List all (filterable by `?faction_id=`, `?detachment_id=`) |
| GET | `/api/admin/stratagems/{id}` | Get single |
| POST | `/api/admin/stratagems` | Create |
| PUT | `/api/admin/stratagems/{id}` | Update |
| DELETE | `/api/admin/stratagems/{id}` | Delete |

### 2d. Mission packs handler

New file: `backend/internal/handler/admin_mission_pack_handler.go`

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/admin/mission-packs` | List all |
| POST | `/api/admin/mission-packs` | Create |
| PUT | `/api/admin/mission-packs/{id}` | Update |
| DELETE | `/api/admin/mission-packs/{id}` | Delete (cascade to missions, secondaries, etc.) |

### 2e. Missions handler

New file: `backend/internal/handler/admin_mission_handler.go`

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/admin/missions` | List all (filterable by `?pack_id=`) |
| GET | `/api/admin/missions/{id}` | Get single (includes scoring_rules JSON) |
| POST | `/api/admin/missions` | Create |
| PUT | `/api/admin/missions/{id}` | Update (including scoring_rules, scoring_timing) |
| DELETE | `/api/admin/missions/{id}` | Delete |

### 2f. Secondaries handler

New file: `backend/internal/handler/admin_secondary_handler.go`

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/admin/secondaries` | List all (filterable by `?pack_id=`) |
| GET | `/api/admin/secondaries/{id}` | Get single (includes scoring_options JSON) |
| POST | `/api/admin/secondaries` | Create |
| PUT | `/api/admin/secondaries/{id}` | Update |
| DELETE | `/api/admin/secondaries/{id}` | Delete |

### 2g. Gambits, Challenger Cards, Mission Rules

Same CRUD pattern for each:

| Entity | Route prefix |
|--------|-------------|
| Gambits | `/api/admin/gambits` |
| Challenger Cards | `/api/admin/challenger-cards` |
| Mission Rules | `/api/admin/mission-rules` |

### 2h. Bulk import endpoint

| Method | Path | Action |
|--------|------|--------|
| POST | `/api/admin/import/factions` | Upload CSV, upsert factions |
| POST | `/api/admin/import/stratagems` | Upload CSV, upsert stratagems + detachments |
| POST | `/api/admin/import/missions` | Upload JSON, upsert missions/secondaries/gambits/etc. |

These reuse the existing seed logic from `backend/internal/seed/` but accept file upload via `multipart/form-data` instead of reading from disk. Returns a summary of what was created/updated.

---

## Phase 3: Admin Frontend — Scaffold & Auth

### 3a. Project setup

New directory: `admin/` at repo root

```
admin/
├── src/
│   ├── api/
│   │   ├── client.ts          # API client (same pattern as frontend)
│   │   └── auth.ts            # GitHub auth API
│   ├── components/
│   │   ├── Layout.tsx         # Sidebar nav + main content area
│   │   ├── DataTable.tsx      # Reusable table with sort/filter/pagination
│   │   ├── FormField.tsx      # Labeled input wrapper
│   │   ├── ConfirmDialog.tsx  # Delete confirmation modal
│   │   └── ImportDialog.tsx   # File upload modal with preview
│   ├── hooks/
│   │   └── useAuth.ts         # GitHub auth context (same pattern)
│   ├── pages/
│   │   ├── LoginPage.tsx
│   │   ├── AuthCallbackPage.tsx
│   │   ├── DashboardPage.tsx  # Overview / landing
│   │   ├── factions/
│   │   │   ├── FactionListPage.tsx
│   │   │   └── FactionEditPage.tsx  # Create + edit (shared form)
│   │   ├── detachments/
│   │   ├── stratagems/
│   │   ├── mission-packs/
│   │   ├── missions/
│   │   ├── secondaries/
│   │   ├── gambits/
│   │   ├── challenger-cards/
│   │   └── mission-rules/
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── package.json               # React, React Router, Tailwind, Vite
├── vite.config.ts
├── tsconfig.json
└── tsconfig.app.json
```

Tech stack mirrors the player frontend: React 18 + TypeScript + Vite + Tailwind v4 + React Router v7.

### 3b. Auth flow

Same pattern as player frontend:
1. Login page with "Sign in with GitHub" button
2. Redirects to `/api/auth/github`
3. GitHub OAuth flow → callback → redirect to `/auth/callback?token=...`
4. Store admin JWT in localStorage (separate key: `admin_token`)
5. `AuthGuard` protects all admin routes

---

## Phase 4: Admin Frontend — Entity Pages

Each entity follows the same UI pattern:

### List page
- Table with columns for key fields
- Search/filter bar (e.g., filter stratagems by faction)
- "Create new" button
- Edit/Delete action buttons per row
- "Import" button (opens file upload dialog)
- Pagination for large datasets (stratagems)

### Edit/Create page
- Form with labeled fields matching the DB schema
- **Structured sub-forms** for JSONB fields:
  - Mission `scoring_rules`: repeatable section with fields for label, vp, minRound, description, scoringTiming. Add/remove buttons.
  - Secondary `scoring_options`: repeatable section with fields for label, vp, mode (dropdown: fixed/tactical/both). Add/remove buttons.
- Dropdowns for foreign keys (e.g., faction picker when editing a stratagem)
- Save + Cancel buttons
- Validation (required fields, numeric ranges)

### Entity-specific notes

| Entity | Filter/Group by | Special fields |
|--------|----------------|----------------|
| Factions | — | wahapedia_link |
| Detachments | faction | — |
| Stratagems | faction, detachment | cp_cost (int), type/turn/phase (dropdowns) |
| Mission Packs | — | — |
| Missions | mission pack | scoring_rules (structured sub-form), scoring_timing (dropdown) |
| Secondaries | mission pack | max_vp (int), is_fixed (checkbox), scoring_options (sub-form) |
| Gambits | mission pack | vp_value (int) |
| Challenger Cards | mission pack | lore, description (textarea) |
| Mission Rules | mission pack | lore, description (textarea) |

### Import dialog
- File input (accepts `.csv` for factions/stratagems, `.json` for missions)
- Preview of parsed rows before confirming
- Submit calls the bulk import endpoint
- Shows results summary (N created, M updated, K errors)

---

## Phase 5: Deployment & Dev Setup

### 5a. Makefile additions

```makefile
dev-admin:    cd admin && npm run dev    # Vite on port 5174
```

### 5b. Environment variables (new)

| Variable | Default | Description |
|----------|---------|-------------|
| `GITHUB_CLIENT_ID` | `""` | GitHub OAuth app client ID |
| `GITHUB_CLIENT_SECRET` | `""` | GitHub OAuth app client secret |
| `GITHUB_REDIRECT_URI` | `http://localhost:8080/api/auth/github/callback` | GitHub OAuth callback URL |
| `ADMIN_GITHUB_IDS` | `""` | Comma-separated GitHub user IDs allowed admin access |
| `ADMIN_FRONTEND_URL` | `http://localhost:5174` | Admin frontend URL (for CORS + redirect) |

### 5c. Railway deployment

- Add a third service for the admin frontend (Docker/nginx, same pattern as player frontend)
- Set `ADMIN_FRONTEND_URL` on the backend service
- Create a GitHub OAuth App pointing to the production admin callback URL

---

## Implementation Order

1. **Phase 1** — GitHub OAuth + admin auth (backend only). Testable with curl/browser.
2. **Phase 2** — Admin CRUD endpoints (backend). Testable with curl.
3. **Phase 3** — Admin frontend scaffold + auth flow. Can log in and see dashboard.
4. **Phase 4** — Entity pages, one at a time. Start with factions (simplest), then mission packs → missions → secondaries (increasingly complex forms), then stratagems (largest dataset), then remaining entities.
5. **Phase 5** — Deployment config + import feature.

Estimated file count: ~15 new Go files, ~30 new TypeScript files, 1 new migration (admin_users table — optional, can start without).
