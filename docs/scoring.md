# Scoring

This document covers the victory point (VP) system, scoring categories, and how winners are determined.

## VP Categories

Victory Points are the currency for winning. There are 4 categories, each with its own cap:

| Category | Max VP | How Earned |
|---|---|---|
| **Primary** | 50 | Mission-specific objectives (e.g., holding objective markers) |
| **Secondary** | 40 | Player-chosen secondary objectives (see [Secondary Objectives](secondary-objectives.md)) |
| **Gambit** | 12 | Gambit declarations (round 3+) and challenger card completions |
| **Paint** | 10 | Hobby/painting bonus — awarded for having a painted army |
| **Total** | **100** | Sum of all 4 categories (hard cap) |

Constants defined in `rules.go:3-9`. There is also a combined cap of 90 for Primary + Secondary + Challenger (`MaxVPCombined`), though this is currently only referenced as a constant.

## Scoring VP

VP can be scored by **either player at any time** during an active game — it is not restricted to the active player.

### Primary, Secondary, and Gambit VP

- Action: `score_vp` with `{category, delta}`
- `category`: `"primary"`, `"secondary"`, or `"gambit"`
- `delta`: positive or negative integer (allows corrections)
- Values are clamped between 0 and the category maximum

### Paint VP

- Action: `set_paint_score` with `{score}`
- Sets the paint VP directly (not a delta) — clamped to 0-10
- Can be set during setup or active game (no status restriction)

### Total VP Calculation

```go
func (p *PlayerState) TotalVP() int {
    return p.VPPrimary + p.VPSecondary + p.VPGambit + p.VPPaint
}
```

## Win Conditions

The game can end in three ways:

### 1. Natural Completion (Round 5 Ends)

When the second player finishes the Fight phase of round 5, `advance_phase` triggers `endGame()`.

- The player with the **higher total VP wins**
- If both players have equal total VP, it is a **tie** (no winner is recorded — `WinnerID` remains empty)

### 2. Concession

Either player can concede at any time during an active game.

- Action: `concede`
- The **opponent immediately wins**, regardless of VP totals
- Game status set to `completed`

### 3. Mutual Abandonment

Players can agree to abandon the game without a winner.

- Player A sends `request_abandon`
- Player B responds with `respond_abandon` (`{accept: true/false}`)
- If accepted: game status set to `abandoned`, **no winner** recorded
- If rejected: the request is cleared and the game continues
- Only one abandon request can be pending at a time
- A player cannot respond to their own request

## State Fields

| Field | Type | Description |
|---|---|---|
| `vpPrimary` | int (0-50) | Player's primary objective VP |
| `vpSecondary` | int (0-40) | Player's secondary objective VP |
| `vpGambit` | int (0-12) | Player's gambit VP |
| `vpPaint` | int (0-10) | Player's painting bonus VP |
| `winnerId` | string | User ID of the winner (empty for tie or abandoned) |
| `completedAt` | timestamp | When the game ended |
