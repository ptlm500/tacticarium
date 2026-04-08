# Turn Structure

This document describes how the game progresses through rounds, turns, and phases.

## Overview

A Warhammer 40K game consists of **5 battle rounds**. Each round has **2 player turns** (one per player). Each player turn has **5 sequential phases**. The game tracks which round, turn, and phase is active, and which player has control.

## Phases

Each player turn progresses through these 5 phases in order:

| # | Phase | Description |
|---|---|---|
| 1 | **Command** | Gain CP, draw tactical secondaries, use New Orders. Administrative start-of-turn. |
| 2 | **Movement** | Active player moves their units on the tabletop. |
| 3 | **Shooting** | Active player's units fire ranged weapons. |
| 4 | **Charge** | Active player declares and resolves charges into melee. |
| 5 | **Fight** | Melee combat is resolved. Last phase of a player's turn. |

Phase order is defined in `state.go:16-18` as `PhaseOrder`.

## Round and Turn Flow

```
Battle Round 1
  ├── Player A's Turn (turn 1)
  │   Command → Movement → Shooting → Charge → Fight
  │
  └── Player B's Turn (turn 2)
      Command → Movement → Shooting → Charge → Fight

Battle Round 2
  ├── Player A's Turn (turn 1)
  │   ...
  └── Player B's Turn (turn 2)
      ...

... (rounds 3-4) ...

Battle Round 5
  ├── Player A's Turn (turn 1)
  │   ...
  └── Player B's Turn (turn 2)
      Command → Movement → Shooting → Charge → Fight → GAME ENDS
```

"Player A" is whichever player has the first turn (`firstTurnPlayer` in game state). This defaults to Player 1 but can be configured.

## Phase Advancement

Only the **active player** can advance the phase via the `advance_phase` action.

The `NextPhase()` function (`rules.go:26-37`) determines what happens:

- **Within a turn**: Advances to the next phase in sequence (e.g., Command → Movement).
- **End of first player's turn** (after Fight phase): Switches active player, resets phase to Command, sets `currentTurn = 2`.
- **End of second player's turn** (after Fight phase): Advances to the next battle round, resets phase to Command, sets `currentTurn = 1`, switches active player back.
- **After round 5, turn 2, Fight phase**: The game ends (see [Scoring — Win Conditions](scoring.md#win-conditions)).

## Active Player

The `activePlayer` field tracks who currently has control (1 or 2). It toggles via `3 - activePlayer` when turns switch.

Key constraint: only the active player can call `advance_phase`. However, **both players can perform other actions** during any phase — such as scoring VP, using reactive stratagems, or adjusting CP. The active player restriction only applies to advancing the game forward.

## Command Points (CP) Gain

At the start of each battle round (when the Command Phase begins for the first player's turn), **both players automatically gain 1 CP**. This happens:

- At game start (round 1) — during the `set_ready` transition (`engine.go:275-287`)
- At the start of rounds 2-5 — during `advance_phase` when the round increments (`engine.go:317-331`)

The `ShouldGainCP()` function (`rules.go:22-24`) returns true for all rounds >= 1.

Additionally, both players' `CPGainedThisRound` counters are reset to 0 at the start of each new round, allowing for additional CP gains during that round (see [Special Mechanics — CP Gain Limits](special-mechanics.md#cp-gain-limits)).

## State Fields

| Field | Type | Description |
|---|---|---|
| `currentRound` | int (1-5) | Current battle round |
| `currentTurn` | int (1-2) | Which player's turn within the round |
| `currentPhase` | Phase | Current phase (command, movement, shooting, charge, fight) |
| `activePlayer` | int (1-2) | Which player number currently has control |
| `firstTurnPlayer` | int (1-2) | Which player goes first each round |
