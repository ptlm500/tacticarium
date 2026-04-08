package game

const (
	MaxRounds      = 5
	MaxVPPrimary   = 50
	MaxVPSecondary = 40
	MaxVPGambit    = 12
	MaxVPPaint     = 10
	MaxVPTotal     = 100

	// CP gained by both players at the start of each Command Phase (every player turn)
	CPPerCommandPhase = 1

	// Challenger card constants
	ChallengerVPThreshold = 6  // Must trail by this many VP
	ChallengerCardVP      = 3  // Default VP for completing a challenger card mission
	MaxVPCombined         = 90 // Primary + Secondary + Challenger combined cap
)

// ShouldGainCP returns true if CP should be auto-gained this command phase.
// Both players gain 1 CP at the start of every Command Phase (each player turn).
func ShouldGainCP(round int) bool {
	return round >= 1
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
