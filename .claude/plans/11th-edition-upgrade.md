# 11th Edition Upgrade Plan

Upgrade Tacticarium from Warhammer 40K **10th edition** (Chapter Approved 2025-26)
to **11th edition** (`launch` dataslate). 10th edition support is dropped entirely.

## Decisions captured (from clarifying Q&A, 2026-06-13)

1. **Data source** — 11th-edition reference data comes from
   [`wn-mitch/40kdc-data`](https://github.com/wn-mitch/40kdc-data) (`main` branch,
   `edition: "11th"`). We **adopt its schema/format** as the app's data model rather
   than keeping our bespoke CSV/JSON shapes.
2. **Scoring framework** — *redesigned*, not a data swap. 11e changes caps, makes
   primaries asymmetric (Force Dispositions → mission matrix), and replaces the
   tactical/fixed secondary deck and the gambit/challenger catch-up system. We follow
   the 40kdc-data encoding.
3. **Existing games** — **wipe and reseed fresh.** Drop all game + reference data and
   reseed from the new dataset. No 10e history is preserved. (Disposable / pre-launch data.)
4. **New mechanics scope** — **add turn-step tracking.** Model the new explicit
   Start-of-Turn / End-of-Turn steps (they anchor scoring timing). Do **not** model
   battle-shock or any per-unit state — the app stays a scoreboard, not a state tracker.

## What actually changes (and what doesn't)

The provided core-rules PDF (`eng_01-06_..._new40k_core_rules`) is **core rules only**
(sections 01–24). It confirms:

- **Phase order is unchanged**: Command → Movement → Shooting → Charge → Fight (§07.02).
  No phase enum change is needed (40kdc-data reached the same conclusion).
- **Core CP gain unchanged**: both players gain 1 CP each Command phase (§08.02).
- **New explicit Start-of-Turn and End-of-Turn steps** wrap the five phases (§07.02, §29).
  Mission scoring is now explicitly triggered at *end of turn*, *end of command phase*,
  and *end of battle round* — which is exactly the granularity the new scoring data uses.
- New **core stratagem list** (§15.02–15.12) — data, not engine.
- Battle-shock, terrain, cover, combat-timing changes — **out of scope** (tabletop-resolved).

Everything scoring/mission/secondary/faction-related is **not** in the PDF; it comes from
40kdc-data. That is where the bulk of the work lands.

### 10e → 11e scoring model delta

| Concept | 10e (current app) | 11e (40kdc-data) |
|---|---|---|
| Primary objective | One mission, symmetric, `MaxVPPrimary = 50` | **Asymmetric.** Each player picks a **Force Disposition**; a 5×5 **mission-matchup matrix** maps (own, opponent) → the mission *that player* plays. 25 missions. |
| Primary VP cap | 50 / game (global) | **45 / game + 15 / battle-round**, per `mission.vp_per_game_cap` / `vp_per_round_cap` |
| Primary scoring rules | `mission_scoring_rules` rows w/ `scoringTiming` | Lives on a **`secondary-card` with `card_type: "primary"`**, as structured `awards[]` (trigger + when + vp/vp_per) |
| Secondaries | Fixed vs Tactical deck, draw 1, discard-on-fail | **Draw 2 / turn, keep unscored cards** (`when_drawn` deck ops: redraw/swap/draw-extra). Per-card `awards[]`; `mode: fixed|tactical` on awards |
| Secondary VP cap | `MaxVPSecondary = 40` | **45 / game + 15 / round** (same caps as primary; *unconfirmed combined cap* flagged in their migration doc) |
| Gambits | Round 3+, `MaxVPGambit = 12` | **Removed** — no gambit entity in 40kdc-data |
| Challenger cards | Trail-by-6 catch-up, shares gambit cap | **Removed** — no challenger entity |
| Twists (Adapt or Die) | `mission_rules` + `adapt_or_die` action | **Removed** — no twist/mission-rule entity |
| Paint VP | `MaxVPPaint = 10` | App-local concept, not in dataset — **keep as-is** |
| Detachments | `(faction, name, rule)` | Adds `detachment_points` (1–3), `force_dispositions[]` |

> **Risk flag:** 40kdc-data marks several 11e facts as `pre-launch-provisional` /
> `unconfirmed` (the combined 15/round cap, force-disposition card text, stratagem
> `type` categories). We pin to the `launch` dataslate where present and treat
> provisional values as defaults that may be re-seeded later. The dataset is versioned
> (`game_version: {edition, dataslate}`), so re-seeding on a data update is cheap.

---

## Scope decision: Option B — board-state model + DSL auto-scoring

The app becomes a **board-state tracker**, not just a scoreboard. Every card's `awards[]`
is a structured predicate (`trigger` + `when`/`per` + `vp`/`vp_per`). I enumerated the
*actual* vocabulary used across all 43 cards — it splits into two layers:

**Layer 1 — Objective control (the dominant case; auto-scored from a board the players maintain).**
- The deployment pattern gives **5 objective coordinates** + per-player **territory** and
  **deployment-zone** polygons (60×44 board). Each objective's **role** (your-home,
  opponent-home, central, in-your-territory, in-enemy-territory, no-man's-land, expansion)
  is **derived geometrically** from its position vs. those polygons — no manual tagging.
- Players maintain a single piece of state per objective: **who controls it** (none / P1 / P2),
  updated by tapping as control changes on the tabletop. (We do *not* model OC values or
  model positions — control is the player-asserted fact.)
- From that, the engine auto-evaluates the whole objective family used by the cards:
  `controls-objective` (incl. `objective_role`, `count_min`, `scope`, `exclude`),
  `objective-majority`, `new-objective-controlled`, `territory-control`, and the
  `controlled-objective` / `controlled-non-home-objective` / `controlled-objective-in-enemy-territory`
  counters — i.e. **all 25 primary cards and the majority of secondaries auto-score**.

**Layer 2 — Facts the board can't derive (structured player-confirm prompt).**
Some predicates reference unit identity/profile, destruction, or fine positions the app has
no source of truth for: `units-destroyed` / `destroyed-while-on-objective` /
`enemy-character-model-destroyed` and the wounds/starting-strength thresholds;
`engagement-fronts`; `*-wholly-within-opponent-deployment-zone`; `operation-markers`,
`*-has-tag`, `action-completed` (card-action markers).
- For these, the engine still **fires at the correct `trigger`, owns the VP math, and
  enforces caps** — but at trigger time it shows a **DSL-generated structured prompt**
  ("End of your turn — *Bring it Down*: score 2 VP per destroyed enemy unit with Starting
  Strength 13+ — how many?") and the player supplies the irreducible count/boolean.
- This is materially more than manual scoring: the app knows *what to ask, when, and how
  much it's worth* — the player only confirms the one fact off the board model.

**Explicitly out of scope (flagged as future, not built now):** army **rosters** from
`units.json`/`weapons.json` (would make this a list-builder), model coordinates &
measurement, and combat resolution. Card **actions** that place markers (`objective-tag`,
`terrain-area-tag`, `unit-tag`) *may* be modelled as taps in a later iteration to upgrade
some Layer-2 prompts to Layer-1 auto-scoring; v1 treats marker/action predicates as
Layer-2 confirms.

> **The one open fork (see end):** for Layer-2 destruction predicates, do we stay with
> player-confirm prompts (recommended), or invest in roster modelling so destroyed units
> are picked by datasheet and their profile looked up? Roster modelling is a large separate
> surface; I recommend deferring it.

## Architecture decision: data ingestion

40kdc-data ships as (a) JSON files under `data/core/**` and (b) a typed TS package
(`@alpaca-software/40kdc-data`). Our backend is Go and is the seeding authority, so:

- **Backend (Go) seeds from the JSON files.** Vendor a pinned snapshot of the relevant
  `data/core/*.json` (+ per-faction dirs) into the repo (e.g. `data/40kdc/`), or add it as
  a git submodule pinned to a tag. Rewrite `backend/internal/seed/*.go` to read these JSON
  shapes instead of the old pipe-delimited CSVs. Record the source commit/dataslate in a
  `data_source` row for traceability.
- **Frontend (TS) may consume the npm package** for read-only reference rendering
  (card text, mission names) if convenient, but the **authoritative** data path stays
  backend → DB → REST, so the game engine and admin keep a single source of truth.
  Decision: **do not** take a hard frontend dependency on the package initially; render
  from our own API. Revisit later if useful.

We adopt the **entity IDs and field names** from 40kdc-data as our canonical identifiers
(e.g. mission id `battlefield-dominance`, disposition `take-and-hold`). This is the
interoperability contract and makes future re-seeds diff-able.

---

## Work breakdown

### Phase 0 — Spike & vendoring (small)
- Vendor pinned 40kdc-data `data/core` snapshot into the repo (submodule or copy + version file).
- Write a throwaway Go reader to load `missions.json`, `mission-matchups.json`,
  `force-dispositions.json`, `secondary-cards.json`, `stratagems.json`, and a couple of
  faction dirs (`detachments.json`, `factions.json`) — confirm shapes parse and IDs cross-link.
- Decide how much of the secondary-card `awards[]` DSL the app interprets vs. stores opaque
  (see Phase 3 open question).

### Phase 1 — Backend engine (`backend/internal/game`)
- **`rules.go`**: replace VP constants. Remove `MaxVPPrimary=50`, `MaxVPSecondary=40`,
  `MaxVPGambit=12`, `MaxVPCombined=90`. Introduce per-game + per-round caps
  (`PrimaryPerGameCap=45`, `PerRoundCap=15`, etc.) sourced from mission data, not constants.
  Keep `MaxVPPaint=10`. Recompute `MaxVPTotal`.
- **Per-round cap enforcement**: scoring is now bounded *per battle round* (15) as well as
  *per game* (45). Extend the primary-scoring bookkeeping (`VPPrimaryScoredSlots`) to track
  per-round accumulation and reject/clamp against both caps. Same logic for secondaries.
- **Turn-step tracking**: add `StartOfTurn` / `EndOfTurn` steps to the turn model. Two
  options — (a) add them to `PhaseOrder` as pseudo-phases, or (b) add a `TurnStep` field
  orthogonal to `Phase`. **Recommend (b)**: a `currentStep: start_of_turn | phase | end_of_turn`
  so scoring prompts can fire at the right moment without polluting the phase enum.
  `advance_phase` / `revert_phase` walk: StartOfTurn → Command → … → Fight → EndOfTurn → (next turn).
- **Remove**: `declare_gambit`, `score_challenger`, `draw_challenger_card`, `adapt_or_die`
  actions + events; `GambitID`, `IsChallenger`, `ChallengerCardID`, `AdaptOrDieUses`,
  `gambitDeclaredRound` from `PlayerState`. Remove `twists.go`. Strip `VPGambit`.
- **Secondaries rework** (`engine_missions.go`): new deck mechanic — **draw 2 per turn,
  keep unscored**. Replace fixed/tactical-draw-1 + discard-on-fail logic. Implement
  `when_drawn` deck ops (redraw / swap / draw-extra). Awards carry `mode: fixed|tactical`
  for cards that print both tracks — keep a per-game `secondaryMode` (fixed vs tactical) to
  pick which awards apply.
- **Force-disposition / mission selection** (setup): replace single `select_mission` with
  (1) each player selects a Force Disposition, (2) engine resolves each player's mission via
  the matchup matrix (asymmetric — `players[i].missionId` differs). Store per-player mission.
- **Board-state model (new, Option B)** — add a `Board` to `GameState`:
  - `objectives[]`: id, position (from deployment pattern), derived `role`, and
    `controlledBy: 0|1|2`. Roles computed once at setup via point-in-polygon against the
    pattern's territory/zone polygons (`internal/game/board.go`).
  - Actions: `set_objective_control {objectiveId, player}` (either player may update),
    emits an event for the log/replay. Optional `set_objective_tag` for card-action markers.
  - Control history per turn (to evaluate `new-objective-controlled` /
    `objective-newly-controlled-this-turn` — needs the start-of-turn snapshot).
- **Scoring evaluator (new, the heart of Option B)** — `internal/game/scoring/` package:
  - An evaluator that, given a card `award`, the `Board`, turn/round/phase, and any
    player-confirmed inputs, returns the VP to apply. Implement each `when.type` and `per`
    descriptor used by the data (≈14 predicates, ≈22 `per` counters — closed set, table-driven).
  - A **trigger scheduler**: on `advance_phase` / end-of-turn / end-of-battle, find awards
    whose `trigger` fires now; Layer-1 awards auto-apply, Layer-2 awards raise a structured
    prompt to the owning player (`scoring_prompt` event) carrying the human text + the count
    field to fill. Player responds via `confirm_award {cardId, awardIndex, count|bool}`.
  - Apply `exclusive_group` (highest-only), `cumulative`, `vp_per`/`per_max`/`vp_max`, and
    the per-round (15) + per-game (45) caps per category.
- Keep `score_vp` / `adjust_vp_manual` as the manual override/escape hatch.
- Update `engine_test.go` comprehensively (it's 3,800 lines and heavily 10e-coupled); add a
  golden-style test per card award against synthetic board states.

### Phase 2 — Database (`backend/internal/db/migrations`)
Since we wipe and reseed, prefer a **clean schema reset** over incremental ALTERs:
- Add migration(s) that drop the 10e mission/scoring tables (`gambits`, `challenger_cards`,
  `mission_rules`, old `missions`, `secondaries`, `mission_scoring_rules`, `mission_packs`
  as currently shaped) and the per-player gambit/challenger/adapt columns.
- New tables modelled on 40kdc-data entities:
  - `force_dispositions (id, name, text)`
  - `missions (id, name, vp_per_game_cap, vp_per_round_cap, deployment_pattern_ids)`
  - `mission_matchups (id, disposition, opponent_disposition, mission_id)`
  - `cards (id, name, card_type[primary|secondary], subtype, when_drawn jsonb, actions jsonb, awards jsonb, text)`
    — `awards`/`when_drawn`/`actions` stored as JSONB and **interpreted** by the scoring
    evaluator (Option B), not just rendered.
  - `deployment_patterns (id, name, objectives jsonb, territories jsonb, zones jsonb)` — new,
    needed by the board model to place objectives and derive roles.
  - `detachments` gains `detachment_points int`, `force_dispositions text[]`.
- Per-player game state JSONB (`active_secondaries`, etc.) carries over but with the new
  card shape; drop gambit/challenger/adapt columns. Add per-player `force_disposition`,
  `mission_id`, `secondary_mode`.
- **Board state (Option B)**: persist the `objectives[]` (control + tags) and per-turn
  control snapshots in the game-state JSONB (full-state broadcast model already serialises
  `GameState`, so this rides along — no separate table needed).
- Bump game-state persistence schema; since games are wiped, no data migration needed.

### Phase 3 — Seeding & import (`backend/internal/seed`, `cmd/seed`, admin import)
- Rewrite seeders to read 40kdc-data JSON: `SeedFactions`, `SeedDetachments`,
  `SeedStratagems` (new shape: `category/type/cp_cost/phases/player_turn/timing`),
  plus new `SeedForceDispositions`, `SeedMissions`, `SeedMissionMatchups`, `SeedCards`,
  `SeedDeploymentPatterns`.
- Add a **conformance check** to seeding/CI: assert every `when.type` and `per` value in the
  seeded cards is implemented by the scoring evaluator — so an upstream data refresh that
  introduces a new predicate fails loudly instead of silently mis-scoring.
- Update `admin/internal/seed` import paths and the admin import handler
  (`admin_import_handler.go`) to accept the new JSON formats.

### Phase 4 — API + generated types (`shared/`, `backend/internal/handler`)
- Update huma input/output structs (`handler/types.go`) and `models.go` for the new
  entities (force dispositions, mission matchups, cards) and removed ones (gambits,
  challenger, twists).
- Update faction/mission handlers; remove gambit/challenger/twist endpoints; add
  force-disposition + matchup endpoints.
- Regenerate `shared/openapi.json` + `shared/api.generated.ts` (golden files).

### Phase 5 — Frontend (`frontend/src`)
- **Setup flow** (`pages/Setup`, `components/setup`): replace mission picker with a
  **Force Disposition picker**; show each player the *resolved asymmetric mission* + its
  VP caps. Detachment picker shows `detachment_points`.
- **Game screen** (`components/game`): VP counter respects per-round (15) + per-game (45)
  caps; remove gambit/challenger UI; rework the secondary panel for draw-2-keep mechanic
  and the kanban deck management; surface Start/End-of-turn steps in the phase tracker.
- **Board view (new, Option B)** — render the deployment map (territories/zones + the 5
  objectives) and let players tap each objective to set control (none/P1/P2). Show derived
  roles. This is the primary new UI surface. On trigger moments, present the
  **structured scoring prompts** (`scoring_prompt`) for Layer-2 awards — auto-generated text
  + a count/confirm input — and auto-applied Layer-1 scores as log entries the player can undo.
  A simple top-down SVG board is sufficient (positions are 0–60 × 0–44); no drag/measure.
- Remove gambit/challenger/twist/adapt-or-die components, queries, mutations, types.
- Update Zustand store, query hooks, mocks (`src/mocks`), and screenshots.
- Update tests via the `frontend-test` skill (browser-mode Vitest).

### Phase 6 — Admin (`admin/src`)
- Replace entity pages: drop `gambits/`, `challenger-cards/`, `mission-rules/`; rework
  `missions/`, `secondaries/` (now unified `cards/`), add `force-dispositions/`,
  `mission-matchups/`. Update `api/admin.ts`, `DataTable`, `ImportDialog` for new shapes.

### Phase 7 — Docs & cleanup
- Rewrite `docs/*.md` (game-overview, turn-structure, scoring, secondary-objectives,
  special-mechanics) for 11e. Update `ARCHITECTURE.md` (tech-stack line says "10th Edition";
  data-source section). Update `CLAUDE.md`. Remove the PDF + temp artifacts from the repo
  root (move PDF into `docs/reference/` or gitignore — it's 32 MB).
- The existing `.claude/skills/game-issue` workflow assumes docs↔code sync; re-validate.

---

## Suggested sequencing / PRs
1. **PR A** — Phase 0 vendoring + Phase 2 schema reset + Phase 3 seeders (data foundation).
2. **PR B** — Phase 1 engine rework + tests (the core risk; gambit/challenger/twist removal,
   caps, turn steps, disposition→mission, secondary deck).
3. **PR C** — Phase 4 API + regenerated types.
4. **PR D** — Phase 5 frontend.
5. **PR E** — Phase 6 admin + Phase 7 docs.

Each PR keeps the build green; B is gated on A; C–E follow B.

## Biggest risks / things to watch
- **`engine_test.go` (3.8k lines)** is deeply 10e-coupled — expect substantial rewrite, not patch.
- **Provisional data**: force-disposition text, the combined-cap question, and stratagem
  `type` categories are unconfirmed upstream. Pin to `launch` where available; make caps
  data-driven so a re-seed fixes them without code changes.
- **Asymmetric missions** touch setup, persistence, both players' scoring views, and history
  — the most cross-cutting behavioural change.
- **Option B scoring evaluator is the new centre of gravity** (was the engine's simplest
  part). Mitigations: the predicate/`per` set is a *closed, enumerable* list (≈14 + ≈22) —
  build it table-driven with a golden test per card award; gate seeding/CI on a conformance
  check so an upstream data refresh can't introduce an unhandled predicate silently. Keep
  manual `score_vp` as the escape hatch for anything the evaluator can't yet resolve.
- **Scope discipline**: hold the Layer-1 / Layer-2 line — model objective *control* (not OC
  values or model positions), and player-confirm off-board facts. Do **not** drift into
  rosters/measurement/combat in v1.

## Resolved decisions (2026-06-13, round 2)
1. **Vendoring — snapshot + CI auto-update.** Copy a pinned `data/core` snapshot into the
   repo (no submodule; Railway builds need no extra auth). Add a scheduled **GitHub Action**
   that checks `wn-mitch/40kdc-data` `main` for a newer commit/dataslate, refreshes the
   snapshot, runs the seeders + tests, and opens a PR. A `data_source` version file records
   the pinned commit so the diff is reviewable. (Auto-merge gated on green CI, optional.)
3. **Paint VP — kept.** The 10-VP painting bonus stays as an app-local concept
   (`MaxVPPaint = 10`); it is not part of 40kdc-data and is unaffected.
4. **Secondary mode — Fixed vs Tactical choice persists.** Confirmed: 11e keeps the
   player-level Fixed/Tactical choice. **Tactical** = draw **2 cards per turn, keep all
   unscored cards, no maximum hand size** (no discard-on-fail). **Fixed** = chosen for the
   whole game, scores the (usually lower) `fixed`-mode awards. This matches the
   `mode: fixed|tactical` flag on `secondary-card.awards[]`. The deck is ~18 secondary cards.

## Confirmed caps
Primary and Secondary each cap at **45 VP / game** and **15 VP / battle round**, tracked
**separately** (not a single combined 45). Engine enforces both per category.

## Scoring approach — DECIDED: Option B (board-state model + DSL auto-scoring)
The app models objectives + control (Layer 1, auto-scored) and uses structured
DSL-generated prompts for off-board facts (Layer 2). See "Scope decision: Option B" above.

## DECIDED — Layer-2 = player-confirm (2026-06-13)
2b. Layer-2 off-board predicates use **structured player-confirm prompts**. **Army-list /
   roster modelling (`units.json`) is deferred** to a later epic — v1 never looks up unit
   profiles; the player supplies the count/boolean when prompted. All scope questions now
   resolved; implementation begins with PR A.

## Note / risk: Gambits
WarCom 11e coverage references a **Gambit** mechanic ("when to take a Gambit"), but the
current 40kdc-data dataset models **no gambit entity** and its 11e migration doc omits them.
We are removing the 10e gambit system. If 11e gambits turn out to be in-scope, they'd be a
follow-up once the data lands upstream — flag, don't block.

---

# PR D — current state & brief (handoff, 2026-06-14)

**Status:** PR A (#48, data), PR B (#50, engine+scoring), PR C (#51, reference API +
10e-table drop) are all merged/green. Backend is fully 11e end-to-end. The frontend +
admin are **compile-level only** — they type-check and the app builds, but the game UI
still renders 10e-shaped placeholders and the new 11e mechanics have **no real UI yet**.
PR D builds that UX. PR E (admin management of 11e reference entities + docs) follows.

## What PR D must build (the 11e UX), and the contract for each

The live game state arrives via WebSocket as a hand-written type in
`frontend/src/types/game.ts` (NOT the OpenAPI type). Reference data comes via REST
(`frontend/src/api/*.ts` + `hooks/queries/*`). Actions are dispatched over WS as
`{type, data}` — the engine's action set is in `backend/internal/game/actions.go`.

1. **Setup — force-disposition picker.** Each player picks a disposition
   (`GET /api/force-dispositions`) via the `select_force_disposition {disposition,
   dispositionName}` action. The engine resolves each player's asymmetric mission once both
   have chosen (mission-matchup matrix, `GET /api/mission-matchups` + `GET /api/missions`).
   Also: `select_side {side}` (attacker/defender) and the existing `select_first_turn_player`,
   `select_secondary_mode {mode}`, `set_paint_score`, `set_ready`. The setup page
   (`GameSetupPage.tsx`) currently has placeholder mission/secondary logic to replace.
2. **Board / objective-control view.** Render the deployment pattern
   (`GET /api/deployment-patterns` → objectives + territory/zone polygons, board is 0–60 × 0–44).
   Player state carries `board` (`scoring.Board`: objectives with role/control/tags).
   Actions: `set_objective_control {objectiveIndex, player}` and
   `set_objective_tag {objectiveIndex, tag, add}`. A simple top-down SVG is sufficient.
3. **Secondary deck (draw-2-keep).** `secondaryDeck`/`secondaryHand`/`secondaryScored` on
   player state; `draw_secondaries` action (tactical mode, Command phase, draws 2 keeps all).
   `GET /api/secondary-cards` lists the deck. Render card `name` + `text`.
4. **Scoring prompts.** The engine fires scoring at end-of-command / end-of-turn /
   end-of-battle: Layer-1 awards auto-apply (a `card_scored` event); Layer-2 awards emit a
   `score_prompt` event and land in `player.pendingScorePrompts`. The player answers with
   `confirm_award {promptId, count}`. Build the prompt UI from `pendingScorePrompts`.
5. **Turn stages.** `currentPhase` now includes `start_of_turn` and `end_of_turn` bookends
   (see PHASE_ORDER/PHASE_LABELS in `types/game.ts`) — the phase tracker should show them.

## Known gaps / decisions to respect
- **Fixed-mode secondaries have no selection UI yet.** The engine supports fixed vs tactical,
  but only tactical (draw-2-keep) is wired; fixed-mode players' deck isn't dealt to hand at
  game start. PR D should add the fixed selection flow (or the engine `startGame` should deal
  the chosen fixed set to hand). Flagged in PR B.
- `MissionScoring.tsx` / `ScoringPrompt.tsx` + the local `ScoringAction` type in
  `types/mission.ts` are a **kept UI-only scaffold** for primary scoring (not fed by the API).
  Either wire them to the new `score_prompt`/`card_scored` flow or remove them.
- DB tables are named `stratagems_11e` and `primary_missions` (transitional names kept to
  avoid churn; not renamed). Other 11e tables: `force_dispositions`, `mission_matchups`,
  `cards`, `deployment_patterns`.
- Objective roles (home/central/expansion) derive geometrically in `internal/game/scoring/board.go`.
- Admin manages factions/detachments + read-only stratagems only; 11e reference entities are
  seeded from 40kdc-data, not admin-managed yet (PR E).

## Local dev gotchas (see also project memory)
- `vp` needs node on PATH: `export PATH="$HOME/.local/share/mise/installs/node/24.14.1/bin:$HOME/.local/share/mise/installs/vp/0.1.19/bin:$PATH"` then `vp check .` / `vp test` from `frontend/`.
- **Browser tests (`vp test`) hang in local warmup** — rely on CI for them; `vp check` is the
  reliable local gate. CI's `vp test` is authoritative (it caught a stale test `vp check` missed).
- **Don't leave `dist/` build output around** — the pre-commit `admin-check`/`frontend-check`
  (`vp check`) will flag it; `rm -rf admin/dist frontend/dist` before committing.
- Backend tests need Docker (testcontainers); Docker works locally. Regenerate types with
  `make generate-types` (deterministic) whenever Go API types change, or CI's
  "Generated Types Up-to-Date" fails.
- Local Postgres for manual runs: brew `postgresql@18`; throwaway cluster pattern used during
  this migration was port 5439. `make dev-stack` (Docker) is the designed path but the Go
  service image builds can fail on this machine's Docker IPv6 networking — run Postgres in
  Docker + Go/Vite on the host instead.
