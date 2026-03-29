# Rules Fixes Specification

This specification addresses divergences between the turn-tracker implementation and the Warhammer 40,000 10th Edition rules (Chapter Approved 2025-26).

---
## 1. Turn Structure — Fix: 2 Player Turns Per Battle Round

### Problem
`FirstTurnPlayer` in `GameState` is never set during setup — it defaults to `0`. This breaks the round advancement logic in `applyAdvancePhase()`:

```go
if e.state.ActivePlayer != e.state.FirstTurnPlayer {
    e.state.CurrentRound++
}
```

Since `FirstTurnPlayer` is `0` and `ActivePlayer` is always `1` or `2`, this condition is **always true**, so the round increments after every player turn. Result: 5 player turns total instead of 10 (2 per round).

Additionally, the UI does not clearly distinguish "battle round" from "player turn".

### Rules
- 5 battle rounds per game.
- Each battle round contains 2 player turns (first player, then second player).
- Each player turn has 5 phases: Command → Movement → Shooting → Charge → Fight.
- Total: 10 player turns across the game.

### Changes

**`backend/internal/game/engine.go` — `applySetReady()` (game start):**
- When both players are ready and the game starts, set `FirstTurnPlayer` to the chosen first player (defaulting to `1` if not explicitly chosen):
  ```go
  if e.state.FirstTurnPlayer == 0 {
      e.state.FirstTurnPlayer = 1
  }
  e.state.ActivePlayer = e.state.FirstTurnPlayer
  ```
- This ensures the round advancement check works: `ActivePlayer != FirstTurnPlayer` is only true when the **second** player finishes their turn.

**`backend/internal/game/state.go` — `GameState`:**
- Add field: `CurrentTurn int `json:"currentTurn"`` — tracks which player turn within the round (1 = first player’s turn, 2 = second player’s turn).
- Set to `1` at game start and when a new round begins; set to `2` when the second player’s turn starts.

**`backend/internal/game/engine.go` — `applyAdvancePhase()`:**
- When a turn ends and the first player just finished: set `CurrentTurn = 2`.
- When a turn ends and the second player just finished: set `CurrentTurn = 1`, increment round.

**`frontend/src/types/game.ts`:**
- Add `currentTurn` to the `GameState` type.

**Frontend UI improvements:**
- Turn banner should read: `"Battle Round {round} — {username}’s Turn — {Phase} Phase"` (or `"Your Turn"` / `"Opponent’s Turn"`).
- `RoundIndicator` should show both the round (1-5) and turn (1-2) context, e.g., `"Round 1 · Turn 1 of 2"`.

**Future consideration:** Add a setup action or UI for choosing which player goes first (e.g., roll-off). For now, default to player 1.

## 2. CP Gain Timing — Fix: All Rounds, Both Players Simultaneously

### Problem
- `ShouldGainCP()` returns `round >= 2`, skipping round 1.
- Only the **active player** gains CP when their command phase starts.

### Rules
Both players gain 1 CP at the start of each Command Phase (rounds 1-5). The Command Phase occurs once per battle round (not once per player turn).

### Changes

**`backend/internal/game/rules.go`:**
- Change `ShouldGainCP()` to always return `true` (or remove it entirely).

**`backend/internal/game/engine.go` — `applyAdvancePhase()`:**
- CP gain should happen when the **first player's** command phase begins (i.e., at the start of a new battle round), and grant 1 CP to **both** players simultaneously.
- Currently CP is granted inside the `turnEnded` block when switching players. Restructure so:
  - When a new battle round starts (round increments or game starts), both players get +1 CP.
  - When the second player's turn starts (mid-round player switch), no CP is granted.

**`backend/internal/game/engine.go` — `applySetReady()` (game start):**
- When the game transitions to `StatusActive` at round 1, grant both players 1 CP immediately.
- Emit `EventCPGain` events for both players.

---

## 3. CP Gain Cap — Max 1 Additional CP Per Battle Round

### Problem
No limit on how many extra CP a player can gain per round from game mechanics (tactical discard, etc.).

### Rules
Beyond the automatic Command Phase gain, players can gain at most 1 additional CP per battle round.

### Changes

**`backend/internal/game/state.go` — `PlayerState`:**
- Add field: `CPGainedThisRound int `json:"cpGainedThisRound"`` — tracks non-automatic CP gained this battle round.

**`backend/internal/game/engine.go` — `applyAdvancePhase()`:**
- When a new battle round starts, reset `CPGainedThisRound = 0` for both players.

**`backend/internal/game/engine_missions.go` — `applyDiscardSecondary()`:**
- Before granting the 1 CP for end-of-turn discard, check `player.CPGainedThisRound < 1`.
- If the cap is already reached, do not grant CP (still allow the discard).
- If granted, increment `player.CPGainedThisRound`.

**`backend/internal/game/engine.go` — `applyAdjustCP()`:**
- Manual adjustments (`adjust_cp`) are not exempt from this cap

**Any other future CP-granting mechanics** should check and increment `CPGainedThisRound`.

---

## 4. Scoring Timing — Add `scoringTiming` Field to Mission Data

### Problem
Scoring is fully manual with no timing information. Players are never prompted about when to score.

### Rules
Different missions score at different points in the game flow:
- **Most primaries:** End of active player's Command Phase (from BR2). Second player in BR5 scores at end of their turn.
- **Purge the Foe:** End of each battle round (both players simultaneously).
- **Terraform (bonus VP):** End of each player's turn (from BR2).
- **All secondaries:** End of your turn.

### Changes

**New type — `ScoringTiming`:**

Define a string enum with the following values:
- `"end_of_command_phase"` — Score at end of your Command Phase (BR2+). Second player in BR5 scores at end of turn.
- `"end_of_battle_round"` — Score at end of each battle round (both players simultaneously).
- `"end_of_turn"` — Score at end of your turn.

**`backend/internal/models/models.go` — `Mission`:**
- Add field: `ScoringTiming string `json:"scoringTiming"``

**`frontend/src/types/mission.ts` — `Mission`:**
- Add field: `scoringTiming: string`

**Database migration (new):**
- Add `scoring_timing TEXT NOT NULL DEFAULT 'end_of_command_phase'` to `missions` table.

**`backend/internal/seed/missions.go`:**
- Add scoring timing to each mission. Based on the rules:

| Mission | Scoring Timing |
|---------|---------------|
| Take and Hold | `end_of_command_phase` |
| Scorched Earth | `end_of_command_phase` |
| Purge the Foe | `end_of_battle_round` |
| The Ritual | `end_of_command_phase` |
| Supply Drop | `end_of_command_phase` |
| Burden of Trust | `end_of_command_phase` |
| Terraform | `end_of_command_phase` (control VP); bonus terraform VP is `end_of_turn` — see note below |
| Unexploded Ordnance | `end_of_command_phase` |
| Linchpin | `end_of_command_phase` |
| Hidden Supplies | `end_of_command_phase` |

**Note on Terraform:** Terraform has a dual timing (control VP at end of command phase, terraform bonus at end of turn). The simplest approach is to tag the mission as `end_of_command_phase` for the primary scoring prompt, and add the per-scoring-action timing as a separate field on `ScoringAction` if needed. Alternatively, add a `scoringTiming` field to `ScoringAction` itself so that the "Terraformed marker" rule can be tagged as `end_of_turn` while the others are `end_of_command_phase`. **Recommended: add `scoringTiming` to `ScoringAction` as well**, defaulting to the mission-level timing when not set.

**`backend/internal/seed/missions.go` — `scoringAction` struct:**
- Add field: `ScoringTiming string `json:"scoringTiming,omitempty"``
- Set this on the Terraform "Terraformed marker" rule to `"end_of_turn"`.

**Secondary missions** all score at `end_of_turn` — this does not need to be stored per-secondary since it is universal.

---

## 5. Phase-Transition Scoring Prompts (Frontend)

### Problem
Players are never prompted to score missions at the appropriate time.

### Design
When a player advances the phase past a scoring window, show a **confirmation prompt** asking if they've scored. This is a frontend-only concern — the backend remains permissive (scoring is still allowed at any time for flexibility/corrections).

### Changes

**`frontend/src/pages/GamePage.tsx` — `handleAdvancePhase()`:**

Before sending `advance_phase`, check if the current phase/timing is a scoring window and show a prompt:

**Primary mission scoring prompt triggers:**

| `scoringTiming` | Prompt when... |
|----------------|----------------|
| `end_of_command_phase` | Active player is about to advance **out of Command Phase** (BR2+). Also prompt second player advancing out of Fight Phase in BR5. |
| `end_of_battle_round` | Second player is about to advance out of Fight Phase (end of round). Prompt both — but since only the active player clicks, prompt the active player and include a note about the opponent. |
| `end_of_turn` | Active player is about to advance out of Fight Phase (end of their turn). |

**Primary prompt robustness:**
- If `scoringTiming` is empty/undefined (pre-migration or non-CA2025 packs), default to `end_of_command_phase` behaviour when the mission has scoring rules.

**Secondary mission scoring prompt triggers:**
- **Tactical mode:** Prompt when advancing out of Fight Phase (end of turn) — show active secondaries with Achieve/Discard buttons.
- **Fixed mode:** Prompt when advancing out of Fight Phase (end of turn) — show each fixed secondary with its name, description, and a VP number input + "Score" button. The `maxVp` on a fixed secondary is a per-game cap, NOT a per-turn score — scoring is variable per turn (e.g., Assassination scores 3 or 4 VP per character destroyed).
- **Tactical secondary draw prompt:** Prompt when advancing out of Command Phase (tactical mode, <2 active secondaries and deck not empty): "Have you drawn your tactical secondary missions?"

**Prompt UI:**
- Use a modal/dialog (not `window.confirm`) with:
  - Title: "Scoring Reminder"
  - Body: Context-specific message (e.g., "Score your primary mission — [Mission Name]" with quick-score buttons)
  - Actions: "I've scored" (proceeds with advance) / "Let me score first" (cancels advance)
- The prompt should show the relevant scoring actions (primary quick-score buttons, secondary cards) inline so the player can score directly from the prompt without dismissing it.

**New component: `frontend/src/components/game/ScoringPrompt.tsx`:**
- Renders context-specific reminders based on a list of `ScoringPromptItem` entries.
- Supports item kinds: `primary`, `end_of_round_primary`, `secondary` (tactical), `fixed_secondary`, `tactical_draw`.

**State management:**
- Add `scoringPromptItems: ScoringPromptItem[] | null` to local component state in `GamePage.tsx`.
- When the player clicks "Advance Phase", check what prompts are needed. If any, set `scoringPromptItems` and show the modal. If none, send the action immediately.

---

## 6. Tactical Secondary Draw — Restrict to Command Phase

### Problem
`draw_secondary` can be called at any time.

### Rules
Tactical missions are drawn during the Command Phase.

### Changes

**`backend/internal/game/engine_missions.go` — `applyDrawSecondary()`:**
- Add validation: `if e.state.CurrentPhase != PhaseCommand { return error }`.
- Also validate the caller is the active player: `if action.PlayerNumber != e.state.ActivePlayer { return error }`.

---

## 7. Challenger Card Timing — Restrict to Start of Battle Round

### Problem
Challenger cards can be drawn at any time when trailing by 6+ VP.

### Rules
Challenger cards are drawn at the start of a battle round when a player is trailing by 6+ VP.

### Changes

**`backend/internal/game/engine_missions.go` — `applyDrawChallengerCard()`:**
- Add validation: `if e.state.CurrentPhase != PhaseCommand { return error }`.
- The check for VP deficit remains as-is.

**Frontend:**
- Only show the challenger card banner during the Command Phase.

---

## 8. New Orders Timing — Restrict to Command Phase

### Problem
`new_orders` can be called at any time.

### Rules
The New Orders stratagem is used at the end of the Command Phase.

### Changes

**`backend/internal/game/engine_missions.go` — `applyNewOrders()`:**
- Add validation: `if e.state.CurrentPhase != PhaseCommand { return error }`.

---

## 9. Secondary Scoring Options — Multiple VP Criteria Per Secondary

### Problem
Secondaries currently have a single `maxVp` field. Fixed secondaries show only an "Achieve +maxVp" button, and tactical secondaries likewise. But most secondaries have **multiple scoring criteria** with different VP values — e.g., Assassination awards 4VP for a W4+ character destroyed and 3VP for a W<4 character. Some secondaries also have **different criteria depending on mode** (fixed vs tactical).

### Rules
Each secondary mission card lists specific scoring conditions, each with its own VP reward. Some conditions are mode-specific:
- **Fixed mode:** Scored each time a condition is met (cumulative), up to a per-game cap (`maxVp`).
- **Tactical mode:** Typically scored once at end of turn (achieve), with the highest applicable condition.

### Data Model

**New type — `ScoringOption`:**
```go
type ScoringOption struct {
    Label string `json:"label"`
    VP    int    `json:"vp"`
    Mode  string `json:"mode,omitempty"` // "fixed", "tactical", or "" (both)
}
```

- `Label`: Short description of the condition (e.g., "W4+ CHARACTER destroyed").
- `VP`: Victory points awarded when this condition is met.
- `Mode`: If `"fixed"`, only shown in fixed mode. If `"tactical"`, only shown in tactical mode. If empty/omitted, shown in both modes.

### Changes

**`backend/internal/models/models.go` — `Secondary`:**
- Add field: `ScoringOptions []ScoringOption `json:"scoringOptions"``

**`backend/internal/game/state.go` — `ActiveSecondary`:**
- Add field: `ScoringOptions []ScoringOption `json:"scoringOptions"``

**`backend/internal/db/migrations/006_secondary_scoring_options.sql`:**
```sql
-- +goose Up
ALTER TABLE secondaries ADD COLUMN IF NOT EXISTS scoring_options JSONB NOT NULL DEFAULT '[]';
```

**`backend/internal/seed/missions.go`:**
- Add `secondaryScoringOptions()` map (similar to `missionScoringRules()`) returning `map[string][]ScoringOption`.
- Update secondary INSERT to include `scoring_options` column.
- Populate scoring options for all 19 Chapter Approved 2025-26 secondaries:

| Secondary | Scoring Options |
|-----------|----------------|
| A Tempting Target | 5VP (both) |
| Area Denial | 2VP (within 3", no enemy within 3"); 5VP (within 3", no enemy within 6") |
| Assassination | 4VP W4+ char (fixed); 3VP W<4 char (fixed); 5VP 1+ chars destroyed (tactical) |
| Behind Enemy Lines | 3VP (1 unit in enemy zone); 4VP (2+ units in enemy zone) |
| Bring It Down | 2VP per MONSTER/VEHICLE (fixed); +2VP if W15+ (fixed); +2VP if W20+ (fixed); 4VP 1+ destroyed (tactical) |
| Cleanse | 2VP (1 objective cleansed); 4VP (2+ cleansed, fixed); 5VP (2+ cleansed, tactical) |
| Cull the Horde | 5VP per 13+ model unit (fixed); 5VP 1+ destroyed (tactical) |
| Defend Stronghold | 3VP (both) |
| Display of Might | 4VP (both) |
| Engage on All Fronts | 1VP (2 quarters); 2VP (3 quarters); 4VP (4 quarters) |
| Establish Locus | 2VP (locus near centre); 4VP (locus in enemy zone) |
| Extend Battle Lines | 2VP (NML only); 4VP (deployment zone + NML) |
| Marked for Death | 5VP (alpha target destroyed); 2VP (gamma target only) |
| No Prisoners | 2VP per Bodyguard/non-char (fixed); 2VP per unit (tactical) |
| Overwhelming Force | 3VP per unit destroyed near objective (both, up to 5VP) |
| Recover Assets | 3VP (2 units); 5VP (3 units) |
| Sabotage | 3VP (not in enemy zone); 6VP (in enemy zone) |
| Secure No Man's Land | 2VP (1 NML objective); 5VP (2+ NML objectives) |
| Storm Hostile Objective | 4VP (both) |

**`backend/internal/handler/mission_handler.go` — `ListSecondaries()`:**
- Update SELECT query to include `scoring_options`.
- Scan `scoring_options` from JSONB into `[]ScoringOption`.

**`backend/internal/game/engine_missions.go` — `applyAchieveSecondary()`:**
- Validate that `vpScored` matches one of the secondary's `scoringOptions[].vp` values (filtered by the player's current mode). If no match, return an error.
- Note: scoring options are stored on the `ActiveSecondary` in player state, so no DB lookup needed.

**`frontend/src/types/mission.ts`:**
- Add `ScoringOption` interface: `{ label: string; vp: number; mode?: string }`.
- Add `scoringOptions: ScoringOption[]` to `Secondary`.

**`frontend/src/types/game.ts` — `ActiveSecondary`:**
- Add `scoringOptions: ScoringOption[]`.

**`frontend/src/components/game/SecondaryPanel.tsx`:**
- For tactical secondaries: replace single "Achieve +maxVp" button with one button per scoring option (filtered to `mode !== 'fixed'`).
- For fixed secondaries: replace single "Score +maxVp" button with one button per scoring option (filtered to `mode !== 'tactical'`).

**`frontend/src/components/game/ScoringPrompt.tsx`:**
- `SecondaryReminder`: render scoring option buttons per active secondary instead of a single achieve button.
- `FixedSecondaryReminder`: render scoring option buttons per fixed secondary instead of a free-text VP input.

---

## Summary of Files to Modify

### Backend
| File | Changes |
|------|---------|
| `game/state.go` | Add `CurrentTurn` to `GameState`; add `CPGainedThisRound` to `PlayerState`; add `ScoringOptions` to `ActiveSecondary` |
| `game/rules.go` | Fix `ShouldGainCP` to include round 1 |
| `game/engine.go` | Fix `FirstTurnPlayer` initialization; track `CurrentTurn`; grant CP to both players at round start; reset CP cap per round; grant CP at game start |
| `game/engine_missions.go` | Enforce CP cap on discard; restrict draw/challenger/new-orders to command phase; validate `vpScored` against scoring options in `applyAchieveSecondary` |
| `models/models.go` | Add `ScoringTiming` to `Mission`; add `ScoringTiming` to `ScoringAction`; add `ScoringOption` type; add `ScoringOptions` to `Secondary` |
| `seed/missions.go` | Add `ScoringTiming` to `scoringAction` struct; populate timing for all missions; add `secondaryScoringOptions()` map; populate scoring options for all secondaries |
| `handler/mission_handler.go` | Update `ListSecondaries` to SELECT and scan `scoring_options` |
| `db/migrations/005_*.sql` | Add `scoring_timing` column to `missions` table |
| `db/migrations/006_*.sql` | Add `scoring_options` column to `secondaries` table |

### Frontend
| File | Changes |
|------|---------|
| `types/game.ts` | Add `currentTurn` to `GameState` type; add `ScoringOption` to `ActiveSecondary` |
| `types/mission.ts` | Add `scoringTiming` to `Mission` and `ScoringAction` types; add `ScoringOption` type; add `scoringOptions` to `Secondary` |
| `pages/GamePage.tsx` | Update turn banner to show round + turn; add scoring prompt logic before phase advance |
| `components/game/RoundIndicator.tsx` | Show round and turn context (e.g., "Round 1 · Turn 1 of 2") |
| `components/game/ScoringPrompt.tsx` | Scoring reminder modal; render per-option buttons for tactical and fixed secondaries |
| `components/game/SecondaryPanel.tsx` | Render per-option achieve/score buttons instead of single maxVp button |
| Challenger card UI | Only show during Command Phase |

### Migration
- Migration `005_scoring_timing.sql` adding `scoring_timing` column to `missions` table.
- Migration `006_secondary_scoring_options.sql` adding `scoring_options` JSONB column to `secondaries` table.
