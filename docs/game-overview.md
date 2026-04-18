# Game Overview

Tacticarium is a real-time, multiplayer turn tracker for **Warhammer 40,000 (10th Edition)** tabletop games. It allows two players on separate devices to collaboratively track game state — rounds, phases, command points, victory points, stratagems, and secondary objectives — during a match.

The app does not enforce all Warhammer 40K rules (players still resolve combat, movement, etc. on the physical tabletop). Instead, it acts as a **digital scoreboard and phase tracker**, keeping both players in sync via WebSocket.

## Supported Game Modes

Only the **Core** game mode is currently supported. Alternate modes that ship with the same data (e.g., Boarding Actions) are excluded from player-facing endpoints. Content is tagged with a `game_mode` column on the `stratagems` and `detachments` tables; the player API filters on `game_mode = 'core'`. Admins can still see and edit boarding-actions content through the admin UI. If support for additional modes is added later, the filter on the player endpoints would be relaxed / made configurable.

## Warhammer 40K Concepts (Brief Primer)

For developers unfamiliar with Warhammer 40K:

- **Factions**: Each player fields an army from one of ~27 factions (e.g., Space Marines, Orks, Aeldari). Faction choice determines available stratagems and detachments.
- **Detachments**: Sub-faction variants within a faction that grant different special rules and stratagems.
- **Command Points (CP)**: A resource spent to activate **stratagems** — powerful one-off abilities used during specific phases. Players gain CP automatically each round.
- **Stratagems**: Tactical abilities costing CP. Can be used by the active player on their turn or reactively by the opponent during specific phases.
- **Victory Points (VP)**: The scoring currency. Earned through primary objectives (mission-based), secondary objectives (player-chosen), gambits, and painting bonuses. Highest VP at end of game wins.
- **Missions**: Define the battlefield scenario, primary objective scoring conditions, and available secondary objectives. Drawn from a **mission pack** (e.g., Chapter Approved 2025-26).
- **Twists**: Optional mission modifiers that add special rules to a game (e.g., Adapt or Die).

## Game Lifecycle

A game moves through four statuses:

| Status | Description |
|---|---|
| `setup` | Players configure factions, detachments, mission, and secondaries. Both must ready up to start. |
| `active` | The game is in progress. Players advance through 5 battle rounds of 5 phases each. |
| `completed` | The game ended naturally (round 5 finished) or by concession. A winner is determined. |
| `abandoned` | Both players mutually agreed to abandon the game. No winner is recorded. |

```
setup ──[both ready]──> active ──[round 5 ends / concede]──> completed
                          │
                          └──[mutual abandon]──> abandoned
```

## Related Documentation

- [Game Setup](game-setup.md) — Creating, joining, and configuring a game
- [Turn Structure](turn-structure.md) — Rounds, turns, and phases
- [Scoring](scoring.md) — Victory points, categories, and win conditions
- [Secondary Objectives](secondary-objectives.md) — Fixed and tactical secondary systems
- [Special Mechanics](special-mechanics.md) — Stratagems, gambits, challenger cards, and more
