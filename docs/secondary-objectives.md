# Secondary Objectives

Secondary objectives are player-chosen scoring conditions that award VP throughout the game. Each player independently selects a **mode** (fixed or tactical) that determines how they interact with secondaries.

Secondary VP is capped at **40 points** total across all secondaries.

## Modes

### Fixed Mode

In fixed mode, the player selects **exactly 2 secondary objectives** during setup. These remain active for the entire game — there is no drawing, discarding, or deck management.

**Setup:**
1. Select mode: `select_secondary_mode` with `{mode: "fixed"}`
2. Choose 2 secondaries: `set_fixed_secondaries` with `{secondaries: [...]}`

**During gameplay**, the player can only achieve their 2 fixed secondaries. There is no deck.

### Tactical Mode

In tactical mode, the player builds a **deck** of secondary objective cards during setup, then draws from it during play. This creates more variety and decision-making but requires more active management.

**Setup:**
1. Select mode: `select_secondary_mode` with `{mode: "tactical"}`
2. Build deck: `init_tactical_deck` with `{deck: [...]}`

**During gameplay**, the player draws cards from their deck into active slots, then achieves or discards them over the course of the game.

## Tactical Mode Gameplay

### Drawing Secondaries

- Action: `draw_secondary`
- **Restrictions**: Command Phase only, active player only, tactical mode only
- Draws cards from the top of the deck until the player has **2 active secondaries**
- If the deck is empty, no cards are drawn
- Implementation: `engine_missions.go:158-200`

### Achieving a Secondary

When a player completes a secondary objective's conditions on the tabletop, they mark it as achieved.

- Action: `achieve_secondary` with `{secondaryId, vpScored}`
- Moves the card from active to the **achieved pile**
- Awards the specified VP (added to `vpSecondary`, clamped to max 40)
- `vpScored` is validated against the secondary's defined `scoringOptions` — each option has a label, VP value, and optional mode filter
- Works in both fixed and tactical modes
- Implementation: `engine_missions.go:202-264`

### Discarding a Secondary

Tactical mode players can discard an active secondary they don't want to pursue.

- Action: `discard_secondary` with `{secondaryId, free?}`
- Moves the card from active to the **discarded pile**
- **CP reward**: If `free` is false (or not set) and the current round is less than 5, the player gains **1 CP** — but only if they haven't already gained additional CP this round (`CPGainedThisRound < 1`)
- Tactical mode only
- Implementation: `engine_missions.go:266-323`

### New Orders

A Command Phase ability that lets a tactical mode player swap an active secondary for a new one from the deck. Costs CP.

- Action: `new_orders` with `{discardSecondaryId}`
- **Restrictions**: Command Phase only, tactical mode only, **once per Command phase per player**
- **CP cost**: 1 (may be modified by twists in future — see `twists.go:9-12`)
- Discards the specified active secondary and draws a replacement from the deck
- The discarded card goes to the discarded pile (no CP reward from this discard)
- The per-phase limit is tracked via `NewOrdersUsedThisPhase` on `PlayerState` and
  reset on every `advance_phase` / `revert_phase`, mirroring the stratagem
  once-per-phase limit
- Implementation: `engine_missions.go:325-395`

## Secondary Objective Data Model

Each active secondary card has:

```
ActiveSecondary {
  id              — Unique identifier
  name            — Display name
  description     — What the objective requires
  isFixed         — Whether this is a fixed-mode secondary
  maxVp           — Maximum VP this secondary can award
  scoringOptions  — Array of {label, vp, mode} scoring criteria
  drawRestriction — Optional {round, mode} "When Drawn" rule
  scoringTiming   — When the owner is prompted to score (see below)
}
```

## Scoring Timing

Most secondaries score at the end of their owner's own turn — the engine
prompts the active player when they advance out of the Fight phase. A handful
of cards (e.g. *Sabotage*, *Defend Stronghold*, the Pariah Nexus *Recover
Assets* variant) score at the end of the **opponent's** turn instead. The
`scoringTiming` field controls when the owner is prompted:

| Value | When the prompt fires |
|---|---|
| `end_of_own_turn` (default) | When the owner advances out of their own Fight phase. |
| `end_of_opponent_turn` | Two prompts fire together: a read-only **reminder block** in the active player's blocking advance modal listing the opponent's pending cards, and a reactive scoring modal on the opponent's (owner's) client. The active player can't click Continue without seeing the reminder. |

Timing is resolved at prompt time by looking up the secondary's ID against the
`useSecondaries` query — this means admin edits to `scoringTiming` take effect
immediately for in-flight games. The embedded `scoringTiming` on the
`ActiveSecondary` card is the fallback when source data is unavailable.

The reactive prompt is non-blocking on the owner's side — dismissing it does
not freeze the game, since `achieve_secondary` works at any time anyway. The
*active player's* modal is what gates the advance; the reminder block in that
modal is what enforces the wait. Cards with mixed timing (scoring options that
resolve at different moments) are not modelled; one card, one timing.

The timing is seeded per-mission-pack in
`backend/internal/seed/missions.go` (`secondaryScoringTimings`) and editable
per-card via the admin UI. New cards default to `end_of_own_turn` unless
explicitly tagged.

Scoring options can be mode-filtered: an option with `mode: "fixed"` only applies in fixed mode, `mode: "tactical"` only in tactical mode, and empty/omitted applies in both.

## Draw Restrictions (Tactical Mode)

Some secondaries carry a "When Drawn" rule that triggers when the card is drawn
during a specific battle round — e.g. *Defend Stronghold*, which cannot be
achieved in the first battle round and must be shuffled back if drawn then.

The optional `drawRestriction` field on a secondary carries two subfields:

| Field | Description |
|---|---|
| `round` | The battle round on which the restriction triggers (typically `1`) |
| `mode`  | `"mandatory"` or `"optional"` |

### Mandatory

When a card with a mandatory restriction is drawn during its target round, the
engine automatically shuffles it back into the deck at a random position and
draws the next card. This repeats until either a non-restricted card is drawn
or the deck is exhausted of drawable cards. A `secondary_reshuffled` event is
emitted (with `reason: "mandatory"`) for each auto-reshuffle.

Mandatory restrictions apply uniformly across every engine draw:
- `draw_secondary` (Command Phase top-up)
- `new_orders` (the replacement draw)
- `adapt_or_die` (the extra draw in tactical mode)

### Optional

An optional restriction lets the player, but does not require them to, shuffle
the drawn card back during its target round. The card is dealt into active
play like a normal draw; the player can then choose to reshuffle it via the
`reshuffle_secondary` action while the round matches.

- Action: `reshuffle_secondary` with `{secondaryId}`
- **Restrictions**: Game active; tactical mode; card is in active secondaries;
  card has `drawRestriction.mode === "optional"` with
  `drawRestriction.round === currentRound`.
- The card is shuffled back into the deck at a random position, a replacement
  is drawn (which also applies mandatory rules), and `secondary_reshuffled` +
  `secondary_drawn` events are emitted.
- Once the round advances past `drawRestriction.round`, the option is no
  longer available — the "When Drawn" moment has passed.
- Implementation: `engine_missions.go` (`applyReshuffleSecondary`)

## Manual Management (Escape Hatch)

Tactical-mode players can opt to manage their secondaries by hand — for
example, when playing with a physical deck of cards alongside the app. The
frontend exposes a **"Manage manually"** toggle in the Secondary Missions
panel. While the toggle is on:

- Normal controls (Achieve buttons, Draw, Discard, Discard +1CP, New Orders,
  Shuffle Back) are hidden.
- Every card in every pile (deck, active, discarded, achieved) is rendered
  with manual move buttons.

The underlying action is `move_secondary`, which moves a card between any two
piles with **no phase, active-player, or CP restrictions** and is independent
of the deck order. Moves to the achieved pile may include a VP score; the
engine does not validate the supplied VP against the card's `scoringOptions`.

- Action: `move_secondary` with `{secondaryId, fromPile, toPile, vpScored?}`
- **Restrictions**: Game active; tactical mode; `fromPile` and `toPile` are
  one of `deck` / `active` / `achieved` / `discarded`; the two piles must
  differ; the card must currently be in `fromPile`.
- **VP**: If `vpScored` is non-zero, `vpSecondary` is adjusted by that amount
  (clamped to `0..MaxVPSecondary`). Negative values revoke VP — useful when
  moving a card *out* of the achieved pile.
- Emits `secondary_moved` with `{secondaryId, secondaryName, fromPile, toPile,
  vpDelta, vpSecondary}`.
- Implementation: `engine_missions.go` (`applyMoveSecondary`)

This is an escape hatch only — players using the digital deck normally should
prefer `draw_secondary` / `achieve_secondary` / `discard_secondary` /
`new_orders` / `reshuffle_secondary`, which carry the proper rules
(restriction handling, CP economy, scoring-option validation).

## Card Piles (Tactical Mode)

A tactical mode player's secondaries are organised into 4 piles:

| Pile | Field | Description |
|---|---|---|
| **Deck** | `tacticalDeck` | Undrawn cards, drawn from the front |
| **Active** | `activeSecondaries` | Currently in play (max 2) |
| **Achieved** | `achievedSecondaries` | Completed objectives (VP already scored) |
| **Discarded** | `discardedSecondaries` | Objectives the player chose not to pursue |

Cards flow: **Deck → Active → Achieved or Discarded**
