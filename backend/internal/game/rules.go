package game

const (
	MaxRounds = 5

	// 11th edition VP caps. Primary and Secondary are each capped per game and
	// per battle round; the per-game/per-round values are also carried on the
	// mission card and may override these defaults.
	DefaultVPPerGameCap  = 45
	DefaultVPPerRoundCap = 15
	MaxVPPaint           = 10

	// MaxVPTotal is the hard ceiling on a player's total score: primary (45) +
	// secondary (45) + paint (10).
	MaxVPTotal = DefaultVPPerGameCap*2 + MaxVPPaint

	// CP gained by both players when entering each Command phase (every turn).
	CPPerCommandPhase = 1

	// FixedSecondaryCount is how many secondary cards a fixed-mode player chooses
	// before the game; the chosen set is their hand for the whole battle. (11e is
	// pre-launch-provisional here; this matches the established fixed count and is
	// a single knob to change if the dataslate confirms otherwise.)
	FixedSecondaryCount = 2
)

// TurnStages is the ordered sequence of stages within a single player turn. 11th
// edition wraps the five phases in explicit Start-of-Turn and End-of-Turn steps,
// which anchor mission scoring timings.
var TurnStages = []Phase{
	PhaseStartOfTurn,
	PhaseCommand,
	PhaseMovement,
	PhaseShooting,
	PhaseCharge,
	PhaseFight,
	PhaseEndOfTurn,
}

// nextStage returns the stage following current within a turn and whether the
// turn has ended (i.e. current was the last stage).
func nextStage(current Phase) (next Phase, turnEnded bool) {
	for i, s := range TurnStages {
		if s == current {
			if i+1 < len(TurnStages) {
				return TurnStages[i+1], false
			}
			return TurnStages[0], true
		}
	}
	return TurnStages[0], true
}

// prevStage returns the stage preceding current within a turn. It is only
// meaningful when current is not the first stage; callers handle the cross-turn
// rollback separately.
func prevStage(current Phase) Phase {
	for i, s := range TurnStages {
		if s == current && i > 0 {
			return TurnStages[i-1]
		}
	}
	return TurnStages[0]
}

// firstStage / lastStage name the turn bookends.
func firstStage() Phase { return TurnStages[0] }
func lastStage() Phase  { return TurnStages[len(TurnStages)-1] }

// ClampVP clamps value to [0, max].
func ClampVP(value, max int) int {
	if value < 0 {
		return 0
	}
	if value > max {
		return max
	}
	return value
}
