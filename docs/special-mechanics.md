# Special Mechanics

This document covers stratagems, CP management, gambits, challenger cards, the Adapt or Die twist, and game-ending actions.

## Stratagems

Stratagems are powerful one-off abilities that cost Command Points (CP) to activate. They are determined by the player's faction and detachment. The app tracks CP expenditure and logs stratagem usage, but does not enforce which stratagems are valid in a given phase — that is left to the players.

- Action: `use_stratagem` with `{stratagemId, stratagemName, cpCost}`
- Validates that the player has sufficient CP before deducting
- **Both players** can use stratagems at any time (not just the active player) — this supports reactive stratagems used during the opponent's turn

## Command Points (CP)

### Automatic CP Gain

Both players gain **1 CP** at the start of each battle round during the Command Phase. This happens automatically:
- Round 1: during the setup → active transition (`engine.go:275-287`)
- Rounds 2-5: when the round advances after the second player's Fight phase (`engine.go:317-331`)

Constant: `CPPerCommandPhase = 1` (`rules.go:12`)

### Manual CP Adjustment

Players can manually adjust their CP (e.g., for abilities that grant bonus CP).

- Action: `adjust_cp` with `{delta}`
- `delta` can be positive (gain) or negative (spend)
- CP cannot go below 0

### CP Gain Limits

Players are limited to gaining **at most 1 additional CP per battle round** beyond the automatic Command Phase gain. This is tracked via `CPGainedThisRound`:

- Positive `adjust_cp` calls increment `CPGainedThisRound`
- Discarding a tactical secondary (non-free) also increments it
- The counter resets to 0 at the start of each new round
- If `CPGainedThisRound >= 1`, further positive CP adjustments are rejected

This mirrors the Warhammer 40K 10th Edition rule that limits bonus CP generation to prevent runaway resource accumulation.

## Gambits

Gambits are high-risk declarations that a player can make from **round 3 onward**. A player declares a gambit card, and if they achieve its conditions, they score VP in the gambit category (max 12 VP).

### Declaring a Gambit

- Action: `declare_gambit` with `{gambitId}`
- **Restriction**: Round 3 or later only
- Sets the player's `gambitId` and records the `gambitDeclaredRound`
- One active gambit per player at a time

### Scoring Gambit VP

Gambit VP is scored via the general `score_vp` action with `{category: "gambit", delta}`, capped at 12 VP.

## Challenger Cards

Challenger cards are a catch-up mechanic for players who are significantly behind on VP. They allow the trailing player to score bonus VP by completing a challenge.

### Drawing a Challenger Card

- Action: `draw_challenger_card` with `{challengerCardId, challengerCardName}`
- **Restrictions**:
  - Command Phase only
  - The player must be **trailing by 6+ VP** (`opponent.TotalVP() - player.TotalVP() >= 6`)
- Sets `isChallenger = true` and records the card ID
- One active challenger card per player at a time

Constant: `ChallengerVPThreshold = 6` (`rules.go:15`)

### Scoring a Challenger Card

- Action: `score_challenger` with `{vpScored?}`
- Awards VP to the **gambit** category (shares the 12 VP cap with gambits)
- Default award: **3 VP** if `vpScored` is not specified or is <= 0
- Clears the active challenger card after scoring

Constant: `ChallengerCardVP = 3` (`rules.go:16`)

## Adapt or Die

Adapt or Die is a **twist** (mission modifier, ID: `mission-rule-chapter-approved-2025-26-adapt-or-die`) that allows players to swap secondary objectives during the game. It only works if this specific twist was selected during setup.

### Behaviour

- Action: `adapt_or_die`
- **Fixed mode**: Swap one active secondary for a completely new one. Limited to **1 use** per game.
  - Data: `{discardSecondaryId, newSecondary: {...}}`
- **Tactical mode**: Draw one extra card from the deck, then shuffle one active secondary back into the deck. Limited to **2 uses** per game.
  - Data: `{shuffleBackSecondaryId}`
  - The shuffled-back card goes to the end of the deck
- Tracked via `adaptOrDieUses` counter on each player

## Concede and Abandon

See [Scoring — Win Conditions](scoring.md#win-conditions) for details on how concession and mutual abandonment work.
