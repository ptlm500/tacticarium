package game

type ActionType string

const (
	ActionAdvancePhase     ActionType = "advance_phase"
	ActionAdjustCP         ActionType = "adjust_cp"
	ActionScoreVP          ActionType = "score_vp"
	ActionUseStratagem     ActionType = "use_stratagem"
	ActionSelectFaction    ActionType = "select_faction"
	ActionSelectDetachment ActionType = "select_detachment"
	ActionSelectMission    ActionType = "select_mission"
	ActionSelectSecondary  ActionType = "select_secondary"
	ActionRemoveSecondary  ActionType = "remove_secondary"
	ActionDeclareGambit    ActionType = "declare_gambit"
	ActionSetReady         ActionType = "set_ready"
	ActionConcede              ActionType = "concede"
	ActionSetPaintScore        ActionType = "set_paint_score"
	ActionSelectPrimaryMission ActionType = "select_primary_mission"
	ActionSelectTwist          ActionType = "select_twist"
	ActionSelectSecondaryMode  ActionType = "select_secondary_mode"
	ActionSetFixedSecondaries  ActionType = "set_fixed_secondaries"
	ActionInitTacticalDeck     ActionType = "init_tactical_deck"
	ActionDrawSecondary        ActionType = "draw_secondary"
	ActionAchieveSecondary     ActionType = "achieve_secondary"
	ActionDiscardSecondary     ActionType = "discard_secondary"
	ActionNewOrders            ActionType = "new_orders"
	ActionDrawChallengerCard   ActionType = "draw_challenger_card"
	ActionScoreChallenger      ActionType = "score_challenger"
	ActionAdaptOrDie           ActionType = "adapt_or_die"
	ActionRequestAbandon       ActionType = "request_abandon"
	ActionRespondAbandon       ActionType = "respond_abandon"
)

type GameAction struct {
	Type         ActionType     `json:"type"`
	PlayerNumber int            `json:"playerNumber"`
	Data         map[string]any `json:"data,omitempty"`
}

type EventType string

const (
	EventPhaseAdvance      EventType = "phase_advance"
	EventCPGain            EventType = "cp_gain"
	EventCPSpend           EventType = "cp_spend"
	EventCPAdjust          EventType = "cp_adjust"
	EventVPPrimaryScore    EventType = "vp_primary_score"
	EventVPSecondaryScore  EventType = "vp_secondary_score"
	EventVPGambitScore     EventType = "vp_gambit_score"
	EventStratagemUsed     EventType = "stratagem_used"
	EventSecondarySelected EventType = "secondary_selected"
	EventGambitDeclared    EventType = "gambit_declared"
	EventPlayerConcede     EventType = "player_concede"
	EventGameStart         EventType = "game_start"
	EventGameEnd           EventType = "game_end"
	EventFactionSelected   EventType = "faction_selected"
	EventMissionSelected        EventType = "mission_selected"
	EventPlayerReady            EventType = "player_ready"
	EventPrimaryMissionSelected EventType = "primary_mission_selected"
	EventTwistSelected          EventType = "twist_selected"
	EventSecondaryModeSelected  EventType = "secondary_mode_selected"
	EventSecondaryDrawn         EventType = "secondary_drawn"
	EventSecondaryAchieved      EventType = "secondary_achieved"
	EventSecondaryDiscarded     EventType = "secondary_discarded"
	EventNewOrdersUsed          EventType = "new_orders_used"
	EventChallengerActivated    EventType = "challenger_activated"
	EventChallengerScored       EventType = "challenger_scored"
	EventAbandonRequested       EventType = "abandon_requested"
	EventAbandonRejected        EventType = "abandon_rejected"
)

type GameEvent struct {
	Type         EventType      `json:"eventType"`
	PlayerNumber int            `json:"playerNumber,omitempty"`
	Round        int            `json:"round,omitempty"`
	Phase        Phase          `json:"phase,omitempty"`
	Data         map[string]any `json:"data,omitempty"`
}
