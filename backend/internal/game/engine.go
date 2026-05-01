package game

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("tacticarium/game")

// StratagemInfo is the canonical (DB-sourced) representation of a stratagem
// used by the engine to validate stratagem usage.
type StratagemInfo struct {
	Name   string
	CPCost int
}

// StratagemLookup resolves a stratagem ID to its canonical info. If nil,
// the engine falls back to trusting the client-supplied name/cost (tests).
type StratagemLookup func(id string) (*StratagemInfo, error)

type Engine struct {
	state           *GameState
	stratagemLookup StratagemLookup
}

func NewEngine(state *GameState) *Engine {
	return &Engine{state: state}
}

// SetStratagemLookup wires up a canonical stratagem source (typically the DB).
// When set, the engine uses DB values for the stratagem's original CP cost and
// name, and treats the client-supplied cpCost as the (possibly overridden)
// amount the player is spending.
func (e *Engine) SetStratagemLookup(fn StratagemLookup) {
	e.stratagemLookup = fn
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

func (e *Engine) Apply(ctx context.Context, action GameAction) ([]GameEvent, error) {
	_, span := tracer.Start(ctx, "game.Apply")
	span.SetAttributes(
		attribute.String("game.action_type", string(action.Type)),
		attribute.Int("game.player_number", action.PlayerNumber),
		attribute.String("game.phase", string(e.state.CurrentPhase)),
		attribute.Int("game.round", e.state.CurrentRound),
	)
	defer span.End()

	events, err := e.applyAction(action)
	if err != nil {
		span.RecordError(err)
	}
	return events, err
}

func (e *Engine) applyAction(action GameAction) ([]GameEvent, error) {
	switch action.Type {
	case ActionSelectFaction:
		return e.applySelectFaction(action)
	case ActionSelectDetachment:
		return e.applySelectDetachment(action)
	case ActionSelectFirstTurnPlayer:
		return e.applySelectFirstTurnPlayer(action)
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
	case ActionRevertPhase:
		return e.applyRevertPhase(action)
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
	case ActionReshuffleSecondary:
		return e.applyReshuffleSecondary(action)
	case ActionMoveSecondary:
		return e.applyMoveSecondary(action)
	case ActionDrawChallengerCard:
		return e.applyDrawChallengerCard(action)
	case ActionScoreChallenger:
		return e.applyScoreChallenger(action)
	case ActionAdaptOrDie:
		return e.applyAdaptOrDie(action)
	case ActionRequestAbandon:
		return e.applyRequestAbandon(action)
	case ActionRespondAbandon:
		return e.applyRespondAbandon(action)
	case ActionUndoPrimaryScore:
		return e.applyUndoPrimaryScore(action)
	case ActionAdjustVPManual:
		return e.applyAdjustVPManual(action)
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

func (e *Engine) applySelectFirstTurnPlayer(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select first turn player during setup")
	}

	// NOTE: the data field is named `firstTurnPlayer` rather than `playerNumber`
	// because the WS client handler strips `playerNumber` from incoming action
	// data (to prevent clients spoofing which player they are) — see
	// backend/internal/ws/client.go.
	firstTurnPlayer := intFromData(action.Data, "firstTurnPlayer")
	if firstTurnPlayer != 1 && firstTurnPlayer != 2 {
		return nil, fmt.Errorf("first turn player must be 1 or 2")
	}

	e.state.FirstTurnPlayer = firstTurnPlayer

	// Reset readiness when the first turn player changes
	for _, p := range e.state.Players {
		if p != nil {
			p.Ready = false
		}
	}

	return []GameEvent{{
		Type:         EventFirstTurnPlayerSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"firstTurnPlayer": firstTurnPlayer},
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
	if ready && e.state.FirstTurnPlayer == 0 {
		return nil, fmt.Errorf("first turn player must be selected before readying up")
	}
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
		e.state.ActivePlayer = e.state.FirstTurnPlayer

		events = append(events, GameEvent{
			Type: EventGameStart,
			Data: map[string]any{"round": 1, "firstPlayer": e.state.ActivePlayer},
		})

		// Both players gain 1 CP at the start of the first Command Phase
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

	// Stratagems are tracked once per phase; clear the per-player list whenever
	// the phase changes so the repeat-use confirmation resets each phase.
	for _, p := range e.state.Players {
		if p != nil {
			p.StratagemsUsedThisPhase = nil
		}
	}

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

			// Reset the per-round additional CP cap at the start of each new battle round
			for _, p := range e.state.Players {
				if p != nil {
					p.CPGainedThisRound = 0
				}
			}
		} else {
			// First player just finished — second player's turn begins
			e.state.CurrentTurn = 2
		}
		e.state.ActivePlayer = otherPlayer
		e.state.CurrentPhase = PhaseCommand

		// Both players gain 1 CP at the start of each Command Phase
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

func (e *Engine) applyRevertPhase(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	if action.PlayerNumber != e.state.ActivePlayer {
		return nil, fmt.Errorf("only the active player can revert the phase")
	}
	if e.state.CurrentRound == 1 && e.state.CurrentTurn == 1 && e.state.CurrentPhase == PhaseCommand {
		return nil, fmt.Errorf("cannot revert before the start of the game")
	}

	oldPhase := e.state.CurrentPhase

	// Mirror advance_phase: stratagem "used this phase" lists reset on any
	// phase change so the repeat-use confirmation resets for the reverted phase.
	for _, p := range e.state.Players {
		if p != nil {
			p.StratagemsUsedThisPhase = nil
		}
	}

	crossedTurnBoundary := oldPhase == PhaseCommand

	if !crossedTurnBoundary {
		e.state.CurrentPhase = PrevPhase(oldPhase)
	} else if e.state.CurrentTurn == 2 {
		// Rolling back to the first player's Fight phase, same round.
		e.state.CurrentTurn = 1
		e.state.ActivePlayer = e.state.FirstTurnPlayer
		e.state.CurrentPhase = PhaseFight
	} else {
		// currentTurn == 1, rolling back into the previous round's second turn.
		e.state.CurrentRound--
		e.state.CurrentTurn = 2
		e.state.ActivePlayer = 3 - e.state.FirstTurnPlayer
		e.state.CurrentPhase = PhaseFight
	}

	events := []GameEvent{{
		Type:         EventPhaseRevert,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"from": string(oldPhase), "to": string(e.state.CurrentPhase)},
	}}

	if crossedTurnBoundary {
		// Revoke the 1 CP each player auto-gained when entering the Command
		// phase we just rolled out of. Clamp at 0 — if a player already spent
		// the CP, we don't push them negative; their stratagem use stands.
		for _, p := range e.state.Players {
			if p == nil {
				continue
			}
			newCP := p.CP - CPPerCommandPhase
			if newCP < 0 {
				newCP = 0
			}
			delta := newCP - p.CP
			p.CP = newCP
			events = append(events, GameEvent{
				Type:         EventCPAdjust,
				PlayerNumber: p.PlayerNumber,
				Round:        e.state.CurrentRound,
				Phase:        e.state.CurrentPhase,
				Data:         map[string]any{"delta": delta, "newTotal": newCP, "reason": "phase_revert"},
			})
		}
	}

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
	force, _ := action.Data["force"].(bool)
	// Positive adjustments are subject to the per-round CP gain cap unless the
	// client explicitly opts out via force=true (player has confirmed override).
	if delta > 0 && player.CPGainedThisRound >= 1 && !force {
		return nil, fmt.Errorf("cannot gain more than 1 additional CP per battle round")
	}
	newCP := player.CP + delta
	if newCP < 0 {
		return nil, fmt.Errorf("insufficient CP")
	}

	player.CP = newCP
	if delta > 0 {
		player.CPGainedThisRound++
	}

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

	var (
		eventType   EventType
		oldVP       int
		newVP       int
		scoringSlot string
	)
	switch category {
	case "primary":
		scoringSlot, _ = action.Data["scoringSlot"].(string)
		if !IsValidPrimaryScoringSlot(scoringSlot) {
			return nil, fmt.Errorf("primary score requires a valid scoringSlot (end_of_command_phase, end_of_battle_round, end_of_turn)")
		}
		if _, used := player.VPPrimaryScoredSlots[e.state.CurrentRound][scoringSlot]; used {
			return nil, fmt.Errorf("primary already scored for slot %q in round %d", scoringSlot, e.state.CurrentRound)
		}
		oldVP = player.VPPrimary
		newVP = ClampVP(oldVP+delta, MaxVPPrimary)
		player.VPPrimary = newVP
		if player.VPPrimaryScoredSlots == nil {
			player.VPPrimaryScoredSlots = map[int]map[string]int{}
		}
		if player.VPPrimaryScoredSlots[e.state.CurrentRound] == nil {
			player.VPPrimaryScoredSlots[e.state.CurrentRound] = map[string]int{}
		}
		player.VPPrimaryScoredSlots[e.state.CurrentRound][scoringSlot] = newVP - oldVP
		eventType = EventVPPrimaryScore
	case "secondary":
		oldVP = player.VPSecondary
		newVP = ClampVP(oldVP+delta, MaxVPSecondary)
		player.VPSecondary = newVP
		eventType = EventVPSecondaryScore
	case "gambit":
		oldVP = player.VPGambit
		newVP = ClampVP(oldVP+delta, MaxVPGambit)
		player.VPGambit = newVP
		eventType = EventVPGambitScore
	default:
		return nil, fmt.Errorf("invalid VP category: %s", category)
	}

	data := map[string]any{
		"category":     category,
		"delta":        delta,
		"appliedDelta": newVP - oldVP,
		"newTotal":     player.TotalVP(),
	}
	if scoringSlot != "" {
		data["scoringSlot"] = scoringSlot
	}
	if label, _ := action.Data["scoringRuleLabel"].(string); label != "" {
		data["scoringRuleLabel"] = label
	}

	return []GameEvent{{
		Type:         eventType,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         data,
	}}, nil
}

func (e *Engine) applyUndoPrimaryScore(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	round := intFromData(action.Data, "round")
	scoringSlot, _ := action.Data["scoringSlot"].(string)
	if !IsValidPrimaryScoringSlot(scoringSlot) {
		return nil, fmt.Errorf("invalid scoringSlot")
	}
	if round <= 0 {
		return nil, fmt.Errorf("invalid round")
	}

	slots, ok := player.VPPrimaryScoredSlots[round]
	if !ok {
		return nil, fmt.Errorf("no primary score recorded for round %d", round)
	}
	appliedDelta, ok := slots[scoringSlot]
	if !ok {
		return nil, fmt.Errorf("no primary score recorded for slot %q in round %d", scoringSlot, round)
	}

	player.VPPrimary = ClampVP(player.VPPrimary-appliedDelta, MaxVPPrimary)
	delete(slots, scoringSlot)
	if len(slots) == 0 {
		delete(player.VPPrimaryScoredSlots, round)
	}

	return []GameEvent{{
		Type:         EventVPPrimaryScoreReverted,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"revertedRound": round,
			"scoringSlot":   scoringSlot,
			"revertedDelta": appliedDelta,
			"newTotal":      player.TotalVP(),
		},
	}}, nil
}

func (e *Engine) applyAdjustVPManual(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	category, _ := action.Data["category"].(string)
	delta := intFromData(action.Data, "delta")

	var oldVP, newVP int
	switch category {
	case "primary":
		oldVP = player.VPPrimary
		newVP = ClampVP(oldVP+delta, MaxVPPrimary)
		player.VPPrimary = newVP
	case "secondary":
		oldVP = player.VPSecondary
		newVP = ClampVP(oldVP+delta, MaxVPSecondary)
		player.VPSecondary = newVP
	case "gambit":
		oldVP = player.VPGambit
		newVP = ClampVP(oldVP+delta, MaxVPGambit)
		player.VPGambit = newVP
	default:
		return nil, fmt.Errorf("invalid VP category: %s", category)
	}

	return []GameEvent{{
		Type:         EventVPManualAdjust,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"category":     category,
			"delta":        delta,
			"appliedDelta": newVP - oldVP,
			"newTotal":     player.TotalVP(),
		},
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
	cpSpent := intFromData(action.Data, "cpCost")

	originalCpCost := cpSpent
	if e.stratagemLookup != nil {
		info, err := e.stratagemLookup(stratagemID)
		if err != nil {
			return nil, fmt.Errorf("stratagem not found: %w", err)
		}
		stratagemName = info.Name
		originalCpCost = info.CPCost
	}

	if cpSpent < 0 {
		return nil, fmt.Errorf("cp cost cannot be negative")
	}
	if player.CP < cpSpent {
		return nil, fmt.Errorf("insufficient CP: have %d, need %d", player.CP, cpSpent)
	}

	player.CP -= cpSpent

	alreadyUsedThisPhase := false
	for _, id := range player.StratagemsUsedThisPhase {
		if id == stratagemID {
			alreadyUsedThisPhase = true
			break
		}
	}
	if !alreadyUsedThisPhase {
		player.StratagemsUsedThisPhase = append(player.StratagemsUsedThisPhase, stratagemID)
	}

	return []GameEvent{{
		Type:         EventStratagemUsed,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"stratagemId":    stratagemID,
			"stratagemName":  stratagemName,
			"cpSpent":        cpSpent,
			"originalCpCost": originalCpCost,
			"cpRemaining":    player.CP,
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
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only set paint score during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	score := intFromData(action.Data, "score")
	player.VPPaint = ClampVP(score, MaxVPPaint)
	player.Ready = false

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

func (e *Engine) applyRequestAbandon(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	if e.state.AbandonRequestedBy != nil {
		return nil, fmt.Errorf("an abandon request is already pending")
	}

	e.state.AbandonRequestedBy = &action.PlayerNumber

	return []GameEvent{{
		Type:         EventAbandonRequested,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
	}}, nil
}

func (e *Engine) applyRespondAbandon(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	if e.state.AbandonRequestedBy == nil {
		return nil, fmt.Errorf("no abandon request is pending")
	}
	if *e.state.AbandonRequestedBy == action.PlayerNumber {
		return nil, fmt.Errorf("cannot respond to your own abandon request")
	}

	accept, _ := action.Data["accept"].(bool)

	if !accept {
		e.state.AbandonRequestedBy = nil
		return []GameEvent{{
			Type:         EventAbandonRejected,
			PlayerNumber: action.PlayerNumber,
			Round:        e.state.CurrentRound,
			Phase:        e.state.CurrentPhase,
		}}, nil
	}

	e.state.AbandonRequestedBy = nil
	e.state.Status = StatusAbandoned
	now := time.Now()
	e.state.CompletedAt = &now

	return []GameEvent{{
		Type: EventGameEnd,
		Data: map[string]any{"reason": "abandoned"},
	}}, nil
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
