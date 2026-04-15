# Game Setup

This document covers how games are created, joined, and configured before play begins.

## Creating a Game

A game is created via `POST /api/games`. The creating player is automatically assigned as **Player 1**. The server generates a **6-character alphanumeric invite code** that the creator shares with their opponent out-of-band (e.g., via Discord or text).

The game starts in `setup` status with default state:
- Round: 0, Phase: `setup`
- Both player slots initialised (Player 1 filled, Player 2 empty)

## Joining a Game

Player 2 joins via `POST /api/games/join/{code}` using the invite code. They are automatically assigned as **Player 2**. After joining, both players connect via WebSocket (`GET /ws/game/{gameId}?token={jwt}`) to receive real-time state updates.

## Setup Configuration

During setup, players configure the following options. **Any configuration change resets both players' ready status**, requiring them to re-confirm before the game can start.

### Faction Selection

Each player selects their army faction (e.g., Space Marines, Necrons). Changing faction also clears the player's detachment selection.

- Action: `select_faction` with `{factionId, factionName}`

### Detachment Selection

Each player selects a detachment within their chosen faction. Detachments determine available stratagems during gameplay.

- Action: `select_detachment` with `{detachmentId, detachmentName}`

### Mission Selection

Either player selects the mission from a mission pack (e.g., Chapter Approved 2025-26). The mission defines the primary objective scoring conditions.

There are two levels of mission configuration:
1. **Mission pack + mission** — via `select_mission` (sets pack, mission ID, and name)
2. **Primary mission** — via `select_primary_mission` (sets just the primary mission within a pack)

- Action: `select_mission` with `{missionPackId, missionId, missionName}`
- Action: `select_primary_mission` with `{missionId, missionName}`

### Twist Selection

An optional mission modifier (twist) can be selected. Twists add special rules — for example, the "Adapt or Die" twist allows players to swap secondary objectives mid-game (see [Special Mechanics](special-mechanics.md#adapt-or-die)).

- Action: `select_twist` with `{twistId, twistName}`

### First Turn Player

Either player can set which player takes the first turn each battle round (see `firstTurnPlayer` in [Turn Structure](turn-structure.md#turns)). There is **no default** — the first turn player must be explicitly chosen before either player can ready up. Either player may set or change the value; the last write wins.

- Action: `select_first_turn_player` with `{playerNumber}` — must be `1` or `2`
- Changing this resets both players' ready status

### Secondary Objective Mode

Each player independently chooses how they will handle secondary objectives: **fixed** or **tactical** mode. See [Secondary Objectives](secondary-objectives.md) for full details.

- Action: `select_secondary_mode` with `{mode}` — must be `"fixed"` or `"tactical"`
- Changing mode clears any previous secondary selections

### Secondary Objective Selection (Setup)

Depending on the chosen mode:

- **Fixed mode**: Select exactly 2 secondary objectives that remain active for the entire game.
  - Action: `set_fixed_secondaries` with `{secondaries: [...]}` (array of 2)

- **Tactical mode**: Build a deck of secondary objective cards to draw from during gameplay.
  - Action: `init_tactical_deck` with `{deck: [...]}`
  - Deck must contain at least 1 card

### Ready Up

Once both players are satisfied with their configuration, each sets their ready status. Readying up (`ready: true`) is rejected unless a first turn player has been selected.

- Action: `set_ready` with `{ready: true/false}`

When **both players are ready**, the game automatically transitions:
1. Status changes to `active`
2. Round set to 1, turn to 1, phase to `command`
3. Active player set to the previously chosen `firstTurnPlayer`
4. Both players gain **1 CP** (first Command Phase CP gain)
5. `game_start` event is emitted
