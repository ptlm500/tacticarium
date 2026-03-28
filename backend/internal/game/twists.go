package game

const (
	TwistAdaptOrDie = "mission-rule-chapter-approved-2025-26-adapt-or-die"
)

// newOrdersCPCost returns the CP cost of the New Orders stratagem,
// which may be modified by the active twist.
func (e *Engine) newOrdersCPCost() int {
	// Future: Vox Static (Leviathan) makes it cost 2CP
	return 1
}

// canUseAdaptOrDie checks whether the player can use the Adapt or Die twist ability.
func (e *Engine) canUseAdaptOrDie(player *PlayerState) bool {
	if player == nil {
		return false
	}
	if e.state.TwistID != TwistAdaptOrDie {
		return false
	}
	if player.SecondaryMode == "fixed" {
		return player.AdaptOrDieUses < 1
	}
	return player.AdaptOrDieUses < 2
}
