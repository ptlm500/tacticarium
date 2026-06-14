package game

type ActionType string

const (
	// Setup
	ActionSelectFaction          ActionType = "select_faction"
	ActionSelectDetachments      ActionType = "select_detachments"
	ActionSelectSide             ActionType = "select_side"
	ActionSelectFirstTurnPlayer  ActionType = "select_first_turn_player"
	ActionSelectForceDisposition ActionType = "select_force_disposition"
	ActionSelectSecondaryMode    ActionType = "select_secondary_mode"
	ActionSelectFixedSecondaries ActionType = "select_fixed_secondaries"
	ActionSetPaintScore          ActionType = "set_paint_score"
	ActionSetReady               ActionType = "set_ready"

	// Turn flow
	ActionAdvancePhase ActionType = "advance_phase"
	ActionRevertPhase  ActionType = "revert_phase"

	// Resources & stratagems
	ActionAdjustCP     ActionType = "adjust_cp"
	ActionUseStratagem ActionType = "use_stratagem"

	// Board
	ActionSetObjectiveControl ActionType = "set_objective_control"
	ActionSetObjectiveTag     ActionType = "set_objective_tag"

	// Secondary deck (draw 2 / keep)
	ActionDrawSecondaries ActionType = "draw_secondaries"

	// Scoring
	ActionConfirmAward   ActionType = "confirm_award"
	ActionScoreVP        ActionType = "score_vp"         // manual escape hatch
	ActionAdjustVPManual ActionType = "adjust_vp_manual" // manual correction

	// Game end
	ActionConcede        ActionType = "concede"
	ActionRequestAbandon ActionType = "request_abandon"
	ActionRespondAbandon ActionType = "respond_abandon"
)

type GameAction struct {
	Type         ActionType     `json:"type"`
	PlayerNumber int            `json:"playerNumber"`
	Data         map[string]any `json:"data,omitempty"`
}

type EventType string

const (
	EventPhaseAdvance             EventType = "phase_advance"
	EventPhaseRevert              EventType = "phase_revert"
	EventCPGain                   EventType = "cp_gain"
	EventCPAdjust                 EventType = "cp_adjust"
	EventVPPrimaryScore           EventType = "vp_primary_score"
	EventVPSecondaryScore         EventType = "vp_secondary_score"
	EventVPManualAdjust           EventType = "vp_manual_adjust"
	EventStratagemUsed            EventType = "stratagem_used"
	EventFactionSelected          EventType = "faction_selected"
	EventDetachmentsSelected      EventType = "detachments_selected"
	EventSideSelected             EventType = "side_selected"
	EventFirstTurnPlayerSelected  EventType = "first_turn_player_selected"
	EventForceDispositionSelected EventType = "force_disposition_selected"
	EventMissionResolved          EventType = "mission_resolved"
	EventSecondaryModeSelected    EventType = "secondary_mode_selected"
	EventFixedSecondariesSelected EventType = "fixed_secondaries_selected"
	EventPlayerReady              EventType = "player_ready"
	EventGameStart                EventType = "game_start"
	EventGameEnd                  EventType = "game_end"
	EventObjectiveControlChanged  EventType = "objective_control_changed"
	EventObjectiveTagged          EventType = "objective_tagged"
	EventSecondaryDrawn           EventType = "secondary_drawn"
	EventCardScored               EventType = "card_scored"
	EventScorePrompt              EventType = "score_prompt"
	EventAwardConfirmed           EventType = "award_confirmed"
	EventPlayerConcede            EventType = "player_concede"
	EventAbandonRequested         EventType = "abandon_requested"
	EventAbandonRejected          EventType = "abandon_rejected"
)

type GameEvent struct {
	// ID is the persisted game_events.id, assigned during PersistGameState.
	ID           int64          `json:"id,omitempty"`
	Type         EventType      `json:"eventType"`
	PlayerNumber int            `json:"playerNumber,omitempty"`
	Round        int            `json:"round,omitempty"`
	Phase        Phase          `json:"phase,omitempty"`
	Data         map[string]any `json:"data,omitempty"`
}
