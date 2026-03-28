package game

const (
	MaxRounds      = 5
	MaxVPPrimary   = 50
	MaxVPSecondary = 40
	MaxVPGambit    = 12
	MaxVPPaint     = 10
	MaxVPTotal     = 100

	// CP gained at start of command phase (rounds 2-5)
	CPPerCommandPhase = 1
)

// ShouldGainCP returns true if CP should be auto-gained this round.
// Players gain 1 CP at the start of their command phase from round 2 onward.
func ShouldGainCP(round int) bool {
	return round >= 2
}

func NextPhase(current Phase) (Phase, bool) {
	for i, p := range PhaseOrder {
		if p == current {
			if i+1 < len(PhaseOrder) {
				return PhaseOrder[i+1], false
			}
			// End of phases for this player's turn
			return PhaseCommand, true
		}
	}
	return PhaseCommand, true
}

func ClampVP(value, max int) int {
	if value < 0 {
		return 0
	}
	if value > max {
		return max
	}
	return value
}
