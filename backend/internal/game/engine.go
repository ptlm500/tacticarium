package game

import (
	"fmt"
	"time"
)

type Engine struct {
	state *GameState
}

func NewEngine(state *GameState) *Engine {
	return &Engine{state: state}
}

func (e *Engine) State() GameState {
	return *e.state
}

// AddPlayer adds a player to the state if the slot is empty.
// Used when a player joins after the room was already created.
func (e *Engine) AddPlayer(player *PlayerState) {
	idx := player.PlayerNumber - 1
	if idx >= 0 && idx < 2 && e.state.Players[idx] == nil {
		e.state.Players[idx] = player
	}
}

func (e *Engine) Apply(action GameAction) ([]GameEvent, error) {
	switch action.Type {
	case ActionSelectFaction:
		return e.applySelectFaction(action)
	case ActionSelectDetachment:
		return e.applySelectDetachment(action)
	case ActionSelectMission:
		return e.applySelectMission(action)
	case ActionSelectSecondary:
		return e.applySelectSecondary(action)
	case ActionRemoveSecondary:
		return e.applyRemoveSecondary(action)
	case ActionSetReady:
		return e.applySetReady(action)
	case ActionAdvancePhase:
		return e.applyAdvancePhase(action)
	case ActionAdjustCP:
		return e.applyAdjustCP(action)
	case ActionScoreVP:
		return e.applyScoreVP(action)
	case ActionUseStratagem:
		return e.applyUseStratagem(action)
	case ActionDeclareGambit:
		return e.applyDeclareGambit(action)
	case ActionConcede:
		return e.applyConcede(action)
	case ActionSetPaintScore:
		return e.applySetPaintScore(action)
	case ActionSelectPrimaryMission:
		return e.applySelectPrimaryMission(action)
	case ActionSelectTwist:
		return e.applySelectTwist(action)
	case ActionSelectSecondaryMode:
		return e.applySelectSecondaryMode(action)
	case ActionSetFixedSecondaries:
		return e.applySetFixedSecondaries(action)
	case ActionInitTacticalDeck:
		return e.applyInitTacticalDeck(action)
	case ActionDrawSecondary:
		return e.applyDrawSecondary(action)
	case ActionAchieveSecondary:
		return e.applyAchieveSecondary(action)
	case ActionDiscardSecondary:
		return e.applyDiscardSecondary(action)
	case ActionNewOrders:
		return e.applyNewOrders(action)
	case ActionDrawChallengerCard:
		return e.applyDrawChallengerCard(action)
	case ActionScoreChallenger:
		return e.applyScoreChallenger(action)
	case ActionAdaptOrDie:
		return e.applyAdaptOrDie(action)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (e *Engine) applySelectFaction(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select faction during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	factionID, _ := action.Data["factionId"].(string)
	factionName, _ := action.Data["factionName"].(string)

	player.FactionID = factionID
	player.FactionName = factionName
	player.DetachmentID = ""
	player.DetachmentName = ""
	player.Ready = false

	return []GameEvent{{
		Type:         EventFactionSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"factionId": factionID, "factionName": factionName},
	}}, nil
}

func (e *Engine) applySelectDetachment(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select detachment during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	detachmentID, _ := action.Data["detachmentId"].(string)
	detachmentName, _ := action.Data["detachmentName"].(string)

	player.DetachmentID = detachmentID
	player.DetachmentName = detachmentName
	player.Ready = false

	return []GameEvent{{
		Type:         EventFactionSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"detachmentId": detachmentID, "detachmentName": detachmentName},
	}}, nil
}

func (e *Engine) applySelectMission(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select mission during setup")
	}

	missionPackID, _ := action.Data["missionPackId"].(string)
	missionID, _ := action.Data["missionId"].(string)
	missionName, _ := action.Data["missionName"].(string)

	e.state.MissionPackID = missionPackID
	e.state.MissionID = missionID
	e.state.MissionName = missionName

	// Reset readiness when mission changes
	for _, p := range e.state.Players {
		if p != nil {
			p.Ready = false
		}
	}

	return []GameEvent{{
		Type:         EventMissionSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"missionPackId": missionPackID, "missionId": missionID, "missionName": missionName},
	}}, nil
}

func (e *Engine) applySelectSecondary(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select secondaries during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	secondary := SecondaryObjective{
		ID:          fmt.Sprintf("sec_%d_%d", action.PlayerNumber, len(player.Secondaries)),
		SecondaryID: strFromData(action.Data, "secondaryId"),
		CustomName:  strFromData(action.Data, "customName"),
		CustomMaxVP: intFromData(action.Data, "customMaxVp"),
	}

	player.Secondaries = append(player.Secondaries, secondary)
	player.Ready = false

	return []GameEvent{{
		Type:         EventSecondarySelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"secondary": secondary},
	}}, nil
}

func (e *Engine) applyRemoveSecondary(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only modify secondaries during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	secondaryID, _ := action.Data["secondaryId"].(string)
	for i, s := range player.Secondaries {
		if s.ID == secondaryID {
			player.Secondaries = append(player.Secondaries[:i], player.Secondaries[i+1:]...)
			player.Ready = false
			break
		}
	}

	return nil, nil
}

func (e *Engine) applySetReady(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only ready up during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	ready, _ := action.Data["ready"].(bool)
	player.Ready = ready

	events := []GameEvent{{
		Type:         EventPlayerReady,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"ready": ready},
	}}

	// Check if both players are ready to start the game
	if e.state.Players[0] != nil && e.state.Players[1] != nil &&
		e.state.Players[0].Ready && e.state.Players[1].Ready {
		e.state.Status = StatusActive
		e.state.CurrentRound = 1
		e.state.CurrentTurn = 1
		e.state.CurrentPhase = PhaseCommand
		if e.state.FirstTurnPlayer == 0 {
			e.state.FirstTurnPlayer = 1
		}
		e.state.ActivePlayer = e.state.FirstTurnPlayer

		events = append(events, GameEvent{
			Type: EventGameStart,
			Data: map[string]any{"round": 1, "firstPlayer": e.state.ActivePlayer},
		})

		// Both players gain 1 CP at the start of battle round 1
		for _, p := range e.state.Players {
			if p != nil {
				p.CP += CPPerCommandPhase
				events = append(events, GameEvent{
					Type:         EventCPGain,
					PlayerNumber: p.PlayerNumber,
					Round:        1,
					Phase:        PhaseCommand,
					Data:         map[string]any{"amount": CPPerCommandPhase, "newTotal": p.CP},
				})
			}
		}
	}

	return events, nil
}

func (e *Engine) applyAdvancePhase(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	if action.PlayerNumber != e.state.ActivePlayer {
		return nil, fmt.Errorf("only the active player can advance the phase")
	}

	oldPhase := e.state.CurrentPhase
	nextPhase, turnEnded := NextPhase(e.state.CurrentPhase)

	var events []GameEvent

	if turnEnded {
		// Switch active player
		otherPlayer := 3 - e.state.ActivePlayer
		if e.state.ActivePlayer != e.state.FirstTurnPlayer {
			// Second player just finished — advance to next battle round
			e.state.CurrentRound++
			if e.state.CurrentRound > MaxRounds {
				return e.endGame(events)
			}
			e.state.CurrentTurn = 1

			// Both players gain 1 CP at the start of each new battle round
			for _, p := range e.state.Players {
				if p != nil {
					p.CP += CPPerCommandPhase
					events = append(events, GameEvent{
						Type:         EventCPGain,
						PlayerNumber: p.PlayerNumber,
						Round:        e.state.CurrentRound,
						Phase:        PhaseCommand,
						Data:         map[string]any{"amount": CPPerCommandPhase, "newTotal": p.CP},
					})
				}
			}
		} else {
			// First player just finished — second player's turn begins
			e.state.CurrentTurn = 2
		}
		e.state.ActivePlayer = otherPlayer
		e.state.CurrentPhase = PhaseCommand
	} else {
		e.state.CurrentPhase = nextPhase
	}

	events = append(events, GameEvent{
		Type:         EventPhaseAdvance,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"from": string(oldPhase), "to": string(e.state.CurrentPhase)},
	})

	return events, nil
}

func (e *Engine) applyAdjustCP(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	delta := intFromData(action.Data, "delta")
	newCP := player.CP + delta
	if newCP < 0 {
		return nil, fmt.Errorf("insufficient CP")
	}

	player.CP = newCP

	return []GameEvent{{
		Type:         EventCPAdjust,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"delta": delta, "newTotal": newCP},
	}}, nil
}

func (e *Engine) applyScoreVP(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	category, _ := action.Data["category"].(string)
	delta := intFromData(action.Data, "delta")

	var eventType EventType
	switch category {
	case "primary":
		player.VPPrimary = ClampVP(player.VPPrimary+delta, MaxVPPrimary)
		eventType = EventVPPrimaryScore
	case "secondary":
		player.VPSecondary = ClampVP(player.VPSecondary+delta, MaxVPSecondary)
		eventType = EventVPSecondaryScore
	case "gambit":
		player.VPGambit = ClampVP(player.VPGambit+delta, MaxVPGambit)
		eventType = EventVPGambitScore
	default:
		return nil, fmt.Errorf("invalid VP category: %s", category)
	}

	return []GameEvent{{
		Type:         eventType,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"category": category, "delta": delta, "newTotal": player.TotalVP()},
	}}, nil
}

func (e *Engine) applyUseStratagem(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	stratagemID, _ := action.Data["stratagemId"].(string)
	stratagemName, _ := action.Data["stratagemName"].(string)
	cpCost := intFromData(action.Data, "cpCost")

	if player.CP < cpCost {
		return nil, fmt.Errorf("insufficient CP: have %d, need %d", player.CP, cpCost)
	}

	player.CP -= cpCost

	return []GameEvent{{
		Type:         EventStratagemUsed,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"stratagemId":   stratagemID,
			"stratagemName": stratagemName,
			"cpSpent":       cpCost,
			"cpRemaining":   player.CP,
		},
	}}, nil
}

func (e *Engine) applyDeclareGambit(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	if e.state.CurrentRound < 3 {
		return nil, fmt.Errorf("gambits can only be declared from round 3 onward")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	gambitID, _ := action.Data["gambitId"].(string)
	player.GambitID = gambitID
	player.GambitDeclaredRound = e.state.CurrentRound

	return []GameEvent{{
		Type:         EventGambitDeclared,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"gambitId": gambitID},
	}}, nil
}

func (e *Engine) applyConcede(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	winnerNumber := 3 - action.PlayerNumber
	winner := e.state.GetPlayer(winnerNumber)

	events := []GameEvent{{
		Type:         EventPlayerConcede,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
	}}

	e.state.Status = StatusCompleted
	now := time.Now()
	e.state.CompletedAt = &now
	if winner != nil {
		e.state.WinnerID = winner.UserID
	}

	events = append(events, GameEvent{
		Type: EventGameEnd,
		Data: map[string]any{"reason": "concede", "winnerId": e.state.WinnerID},
	})

	return events, nil
}

func (e *Engine) applySetPaintScore(action GameAction) ([]GameEvent, error) {
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	score := intFromData(action.Data, "score")
	player.VPPaint = ClampVP(score, MaxVPPaint)

	return nil, nil
}

func (e *Engine) endGame(events []GameEvent) ([]GameEvent, error) {
	e.state.Status = StatusCompleted
	now := time.Now()
	e.state.CompletedAt = &now

	// Determine winner by total VP
	p1 := e.state.Players[0]
	p2 := e.state.Players[1]
	if p1 != nil && p2 != nil {
		if p1.TotalVP() > p2.TotalVP() {
			e.state.WinnerID = p1.UserID
		} else if p2.TotalVP() > p1.TotalVP() {
			e.state.WinnerID = p2.UserID
		}
		// Tie: no winner
	}

	events = append(events, GameEvent{
		Type: EventGameEnd,
		Data: map[string]any{"reason": "rounds_complete", "winnerId": e.state.WinnerID},
	})

	return events, nil
}

// Helpers to extract typed values from action data
func strFromData(data map[string]any, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

func intFromData(data map[string]any, key string) int {
	switch v := data[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}
