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
	ActionConcede          ActionType = "concede"
	ActionSetPaintScore    ActionType = "set_paint_score"
)

type GameAction struct {
	Type         ActionType     `json:"type"`
	PlayerNumber int            `json:"playerNumber"`
	Data         map[string]any `json:"data,omitempty"`
}

type EventType string

const (
	EventPhaseAdvance     EventType = "phase_advance"
	EventCPGain           EventType = "cp_gain"
	EventCPSpend          EventType = "cp_spend"
	EventCPAdjust         EventType = "cp_adjust"
	EventVPPrimaryScore   EventType = "vp_primary_score"
	EventVPSecondaryScore EventType = "vp_secondary_score"
	EventVPGambitScore    EventType = "vp_gambit_score"
	EventStratagemUsed    EventType = "stratagem_used"
	EventSecondarySelected EventType = "secondary_selected"
	EventGambitDeclared   EventType = "gambit_declared"
	EventPlayerConcede    EventType = "player_concede"
	EventGameStart        EventType = "game_start"
	EventGameEnd          EventType = "game_end"
	EventFactionSelected  EventType = "faction_selected"
	EventMissionSelected  EventType = "mission_selected"
	EventPlayerReady      EventType = "player_ready"
)

type GameEvent struct {
	Type         EventType      `json:"eventType"`
	PlayerNumber int            `json:"playerNumber,omitempty"`
	Round        int            `json:"round,omitempty"`
	Phase        Phase          `json:"phase,omitempty"`
	Data         map[string]any `json:"data,omitempty"`
}
