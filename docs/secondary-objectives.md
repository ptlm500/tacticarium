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
- **Restrictions**: Command Phase only, tactical mode only
- **CP cost**: 1 (may be modified by twists in future — see `twists.go:9-12`)
- Discards the specified active secondary and draws a replacement from the deck
- The discarded card goes to the discarded pile (no CP reward from this discard)
- Implementation: `engine_missions.go:325-395`

## Secondary Objective Data Model

Each active secondary card has:

```
ActiveSecondary {
  id             — Unique identifier
  name           — Display name
  description    — What the objective requires
  isFixed        — Whether this is a fixed-mode secondary
  maxVp          — Maximum VP this secondary can award
  scoringOptions — Array of {label, vp, mode} scoring criteria
}
```

Scoring options can be mode-filtered: an option with `mode: "fixed"` only applies in fixed mode, `mode: "tactical"` only in tactical mode, and empty/omitted applies in both.

## Card Piles (Tactical Mode)

A tactical mode player's secondaries are organised into 4 piles:

| Pile | Field | Description |
|---|---|---|
| **Deck** | `tacticalDeck` | Undrawn cards, drawn from the front |
| **Active** | `activeSecondaries` | Currently in play (max 2) |
| **Achieved** | `achievedSecondaries` | Completed objectives (VP already scored) |
| **Discarded** | `discardedSecondaries` | Objectives the player chose not to pursue |

Cards flow: **Deck → Active → Achieved or Discarded**
