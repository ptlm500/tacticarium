package game

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/peter/tacticarium/backend/internal/game/scoring"
)

var tracer = otel.Tracer("tacticarium/game")

// StratagemInfo is the canonical (DB-sourced) representation of a stratagem.
type StratagemInfo struct {
	Name   string
	CPCost int
}

// StratagemLookup resolves a stratagem ID to its canonical info. If nil, the
// engine trusts the client-supplied name/cost (tests).
type StratagemLookup func(id string) (*StratagemInfo, error)

// ResolvedMission is the primary mission a force-disposition matchup yields for
// a player, plus its VP caps and scoring card.
type ResolvedMission struct {
	ID          string
	Name        string
	GameCap     int
	RoundCap    int
	PrimaryCard scoring.Card
}

// MissionResolver maps (own disposition, opponent disposition) to the resolved
// primary mission for that player. Injected by the handler (DB-backed).
type MissionResolver func(disposition, opponentDisposition string) (ResolvedMission, bool)

// BoardBuilder builds the board for the game's chosen deployment pattern and the
// players' sides. Injected by the handler (DB-backed).
type BoardBuilder func(player1Side, player2Side scoring.Side) (scoring.Board, error)

// SecondaryDeckBuilder builds a player's secondary deck for a scoring mode.
type SecondaryDeckBuilder func(mode string) []SecondaryCard

type Engine struct {
	state                *GameState
	stratagemLookup      StratagemLookup
	missionResolver      MissionResolver
	boardBuilder         BoardBuilder
	secondaryDeckBuilder SecondaryDeckBuilder
	promptSeq            int
}

func NewEngine(state *GameState) *Engine {
	return &Engine{state: state}
}

func (e *Engine) SetStratagemLookup(fn StratagemLookup)           { e.stratagemLookup = fn }
func (e *Engine) SetMissionResolver(fn MissionResolver)           { e.missionResolver = fn }
func (e *Engine) SetBoardBuilder(fn BoardBuilder)                 { e.boardBuilder = fn }
func (e *Engine) SetSecondaryDeckBuilder(fn SecondaryDeckBuilder) { e.secondaryDeckBuilder = fn }

func (e *Engine) State() GameState { return *e.state }

// AddPlayer adds a player to an empty slot (used on late join).
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
	case ActionSelectSide:
		return e.applySelectSide(action)
	case ActionSelectFirstTurnPlayer:
		return e.applySelectFirstTurnPlayer(action)
	case ActionSelectForceDisposition:
		return e.applySelectForceDisposition(action)
	case ActionSelectSecondaryMode:
		return e.applySelectSecondaryMode(action)
	case ActionSelectFixedSecondaries:
		return e.applySelectFixedSecondaries(action)
	case ActionSetPaintScore:
		return e.applySetPaintScore(action)
	case ActionSetReady:
		return e.applySetReady(action)
	case ActionAdvancePhase:
		return e.applyAdvancePhase(action)
	case ActionRevertPhase:
		return e.applyRevertPhase(action)
	case ActionAdjustCP:
		return e.applyAdjustCP(action)
	case ActionUseStratagem:
		return e.applyUseStratagem(action)
	case ActionSetObjectiveControl:
		return e.applySetObjectiveControl(action)
	case ActionSetObjectiveTag:
		return e.applySetObjectiveTag(action)
	case ActionDrawSecondaries:
		return e.applyDrawSecondaries(action)
	case ActionConfirmAward:
		return e.applyConfirmAward(action)
	case ActionScoreVP:
		return e.applyScoreVP(action)
	case ActionAdjustVPManual:
		return e.applyAdjustVPManual(action)
	case ActionConcede:
		return e.applyConcede(action)
	case ActionRequestAbandon:
		return e.applyRequestAbandon(action)
	case ActionRespondAbandon:
		return e.applyRespondAbandon(action)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// --- Setup ---

func (e *Engine) applySelectFaction(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select faction during setup")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	player.FactionID = strFromData(action.Data, "factionId")
	player.FactionName = strFromData(action.Data, "factionName")
	player.DetachmentID = ""
	player.DetachmentName = ""
	player.Ready = false
	return []GameEvent{{
		Type:         EventFactionSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"factionId": player.FactionID, "factionName": player.FactionName},
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
	player.DetachmentID = strFromData(action.Data, "detachmentId")
	player.DetachmentName = strFromData(action.Data, "detachmentName")
	player.Ready = false
	return []GameEvent{{
		Type:         EventFactionSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"detachmentId": player.DetachmentID, "detachmentName": player.DetachmentName},
	}}, nil
}

func (e *Engine) applySelectSide(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select side during setup")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	side := scoring.Side(strFromData(action.Data, "side"))
	if side != scoring.SideAttacker && side != scoring.SideDefender {
		return nil, fmt.Errorf("side must be attacker or defender")
	}
	player.Side = side
	// The opponent takes the other side.
	if opp := e.state.GetPlayer(opponentNumber(action.PlayerNumber)); opp != nil {
		opp.Side = otherSide(side)
	}
	e.resetReady()
	return []GameEvent{{
		Type:         EventSideSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"side": string(side)},
	}}, nil
}

func (e *Engine) applySelectFirstTurnPlayer(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select first turn player during setup")
	}
	firstTurnPlayer := intFromData(action.Data, "firstTurnPlayer")
	if firstTurnPlayer != 1 && firstTurnPlayer != 2 {
		return nil, fmt.Errorf("first turn player must be 1 or 2")
	}
	e.state.FirstTurnPlayer = firstTurnPlayer
	e.resetReady()
	return []GameEvent{{
		Type:         EventFirstTurnPlayerSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"firstTurnPlayer": firstTurnPlayer},
	}}, nil
}

func (e *Engine) applySelectSecondaryMode(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select secondary mode during setup")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	mode := strFromData(action.Data, "mode")
	if mode != "fixed" && mode != "tactical" {
		return nil, fmt.Errorf("secondary mode must be fixed or tactical")
	}
	player.SecondaryMode = mode
	// A tactical player has no pre-chosen set; drop any stale fixed selection.
	if mode != "fixed" {
		player.FixedSecondaryIDs = nil
	}
	player.Ready = false
	return []GameEvent{{
		Type:         EventSecondaryModeSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"mode": mode},
	}}, nil
}

// applySelectFixedSecondaries records the card ids a fixed-mode player will play
// for the whole game. The set is dealt to their hand at game start.
func (e *Engine) applySelectFixedSecondaries(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select fixed secondaries during setup")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	if player.SecondaryMode != "fixed" {
		return nil, fmt.Errorf("only fixed-mode players select fixed secondaries")
	}
	ids := stringsFromData(action.Data, "secondaryIds")
	if len(ids) > FixedSecondaryCount {
		return nil, fmt.Errorf("choose at most %d fixed secondaries", FixedSecondaryCount)
	}
	// Reject duplicates so a player can't pad the count with one card.
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			return nil, fmt.Errorf("fixed secondaries must be distinct, non-empty ids")
		}
		seen[id] = true
	}
	player.FixedSecondaryIDs = ids
	player.Ready = false
	return []GameEvent{{
		Type:         EventFixedSecondariesSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"secondaryIds": ids},
	}}, nil
}

func (e *Engine) applySetPaintScore(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only set paint score during setup")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	player.VPPaint = ClampVP(intFromData(action.Data, "score"), MaxVPPaint)
	player.Ready = false
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
	if ready && player.SecondaryMode == "fixed" && len(player.FixedSecondaryIDs) != FixedSecondaryCount {
		return nil, fmt.Errorf("fixed-mode players must choose %d secondaries before readying up", FixedSecondaryCount)
	}
	player.Ready = ready

	events := []GameEvent{{
		Type:         EventPlayerReady,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"ready": ready},
	}}

	if e.bothReady() {
		events = append(events, e.startGame()...)
	}
	return events, nil
}

func (e *Engine) startGame() []GameEvent {
	e.state.Status = StatusActive
	e.state.CurrentRound = 1
	e.state.CurrentTurn = 1
	e.state.CurrentPhase = PhaseCommand
	e.state.ActivePlayer = e.state.FirstTurnPlayer

	// Build the board from the players' sides.
	if e.boardBuilder != nil {
		if board, err := e.boardBuilder(e.sideOf(1), e.sideOf(2)); err == nil {
			e.state.Board = board
		}
	}
	// Build secondary decks per player mode. Tactical players draw from the full
	// pool each turn; fixed players are dealt their chosen set as their hand and
	// hold no deck.
	if e.secondaryDeckBuilder != nil {
		for _, p := range e.state.Players {
			if p == nil {
				continue
			}
			pool := e.secondaryDeckBuilder(p.SecondaryMode)
			if p.SecondaryMode == "fixed" {
				p.SecondaryHand = dealFixedSecondaries(pool, p.FixedSecondaryIDs)
				p.SecondaryDeck = nil
			} else {
				p.SecondaryDeck = pool
			}
		}
	}
	e.snapshotControl()

	events := []GameEvent{{
		Type: EventGameStart,
		Data: map[string]any{"round": 1, "firstPlayer": e.state.ActivePlayer},
	}}
	events = append(events, e.gainCommandCP()...)
	return events
}

// --- Turn flow ---

func (e *Engine) applyAdvancePhase(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	if action.PlayerNumber != e.state.ActivePlayer {
		return nil, fmt.Errorf("only the active player can advance the phase")
	}

	oldPhase := e.state.CurrentPhase
	e.clearPhaseStratagems()

	var events []GameEvent

	// Fire end-of-command-phase scoring when leaving the Command stage.
	if oldPhase == PhaseCommand {
		events = append(events, e.fireScoring("end-of-phase", PhaseCommand)...)
	}

	next, turnEnded := nextStage(oldPhase)

	if !turnEnded {
		e.state.CurrentPhase = next
		// Entering Command grants Core CP.
		if next == PhaseCommand {
			events = append(events, e.gainCommandCP()...)
		}
		// Entering End-of-Turn fires end-of-turn scoring.
		if next == PhaseEndOfTurn {
			events = append(events, e.fireScoring("end-of-turn", "")...)
		}
		events = append(events, e.phaseAdvanceEvent(action, oldPhase)...)
		return events, nil
	}

	// Turn ended (advancing past End-of-Turn): switch turn/round.
	if e.state.ActivePlayer != e.state.FirstTurnPlayer {
		// Second player finished — advance the battle round.
		e.state.CurrentRound++
		if e.state.CurrentRound > MaxRounds {
			events = append(events, e.fireScoring("end-of-battle", "")...)
			return e.endGame(events)
		}
		e.state.CurrentTurn = 1
		e.resetRound()
	} else {
		e.state.CurrentTurn = 2
	}
	e.state.ActivePlayer = opponentNumber(e.state.ActivePlayer)
	e.state.CurrentPhase = PhaseStartOfTurn
	e.snapshotControl()

	events = append(events, e.phaseAdvanceEvent(action, oldPhase)...)
	return events, nil
}

func (e *Engine) phaseAdvanceEvent(action GameAction, from Phase) []GameEvent {
	return []GameEvent{{
		Type:         EventPhaseAdvance,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"from": string(from), "to": string(e.state.CurrentPhase)},
	}}
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
	e.clearPhaseStratagems()

	crossedTurnBoundary := oldPhase == firstStage() || (oldPhase == PhaseCommand && e.state.CurrentTurn == 1 && e.state.CurrentRound == 1)
	_ = crossedTurnBoundary

	var events []GameEvent
	// Reverting out of the first stage crosses the turn boundary.
	if oldPhase == firstStage() {
		if e.state.CurrentTurn == 2 {
			e.state.CurrentTurn = 1
			e.state.ActivePlayer = e.state.FirstTurnPlayer
		} else {
			e.state.CurrentRound--
			e.state.CurrentTurn = 2
			e.state.ActivePlayer = opponentNumber(e.state.FirstTurnPlayer)
		}
		e.state.CurrentPhase = lastStage()
	} else {
		// Reverting CP grant when stepping back out of Command.
		if oldPhase == PhaseCommand {
			events = append(events, e.revokeCommandCP()...)
		}
		e.state.CurrentPhase = prevStage(oldPhase)
	}

	events = append(events, GameEvent{
		Type:         EventPhaseRevert,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"from": string(oldPhase), "to": string(e.state.CurrentPhase)},
	})
	return events, nil
}

// gainCommandCP grants 1 Core CP to both players (entering a Command phase).
func (e *Engine) gainCommandCP() []GameEvent {
	var events []GameEvent
	for _, p := range e.state.Players {
		if p == nil {
			continue
		}
		p.CP += CPPerCommandPhase
		events = append(events, GameEvent{
			Type:         EventCPGain,
			PlayerNumber: p.PlayerNumber,
			Round:        e.state.CurrentRound,
			Phase:        PhaseCommand,
			Data:         map[string]any{"amount": CPPerCommandPhase, "newTotal": p.CP},
		})
	}
	return events
}

// revokeCommandCP reverses the Core CP grant, clamped at 0.
func (e *Engine) revokeCommandCP() []GameEvent {
	var events []GameEvent
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
	return events
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
	if delta > 0 && player.CPGainedThisRound >= 1 && !force {
		return nil, fmt.Errorf("cannot gain more than 1 additional CP per battle round")
	}
	newCP := player.CP + delta
	if newCP < 0 {
		return nil, fmt.Errorf("not enough CP")
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

func (e *Engine) applyUseStratagem(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	stratagemID := strFromData(action.Data, "stratagemId")
	stratagemName := strFromData(action.Data, "stratagemName")
	cpSpent := intFromData(action.Data, "cpCost")
	originalCpCost := cpSpent
	if e.stratagemLookup != nil {
		info, err := e.stratagemLookup(stratagemID)
		if err != nil {
			return nil, fmt.Errorf("stratagem not found")
		}
		stratagemName = info.Name
		originalCpCost = info.CPCost
	}
	if cpSpent < 0 {
		return nil, fmt.Errorf("CP cost cannot be negative")
	}
	if player.CP < cpSpent {
		return nil, fmt.Errorf("not enough CP — you have %d, need %d", player.CP, cpSpent)
	}
	player.CP -= cpSpent

	alreadyUsed := false
	for _, id := range player.StratagemsUsedThisPhase {
		if id == stratagemID {
			alreadyUsed = true
			break
		}
	}
	if !alreadyUsed {
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

// applyScoreVP is the manual scoring escape hatch (primary/secondary).
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

	var applied int
	var eventType EventType
	switch category {
	case "primary":
		applied = e.scorePrimary(player, delta)
		eventType = EventVPPrimaryScore
	case "secondary":
		applied = e.scoreSecondary(player, delta)
		eventType = EventVPSecondaryScore
	default:
		return nil, fmt.Errorf("unknown VP category: %s", category)
	}
	return []GameEvent{{
		Type:         eventType,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"category": category, "delta": delta, "appliedDelta": applied, "newTotal": player.TotalVP(),
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
	gameCap := e.state.gameCap()
	var oldVP, newVP int
	switch category {
	case "primary":
		oldVP = player.VPPrimary
		newVP = ClampVP(oldVP+delta, gameCap)
		player.VPPrimary = newVP
	case "secondary":
		oldVP = player.VPSecondary
		newVP = ClampVP(oldVP+delta, gameCap)
		player.VPSecondary = newVP
	case "paint":
		oldVP = player.VPPaint
		newVP = ClampVP(oldVP+delta, MaxVPPaint)
		player.VPPaint = newVP
	default:
		return nil, fmt.Errorf("unknown VP category: %s", category)
	}
	return []GameEvent{{
		Type:         EventVPManualAdjust,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"category": category, "delta": delta, "appliedDelta": newVP - oldVP, "newTotal": player.TotalVP(),
		},
	}}, nil
}

// --- Game end ---

func (e *Engine) applyConcede(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	winner := e.state.GetPlayer(opponentNumber(action.PlayerNumber))
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

func (e *Engine) endGame(events []GameEvent) ([]GameEvent, error) {
	e.state.Status = StatusCompleted
	now := time.Now()
	e.state.CompletedAt = &now
	p1, p2 := e.state.Players[0], e.state.Players[1]
	if p1 != nil && p2 != nil {
		if p1.TotalVP() > p2.TotalVP() {
			e.state.WinnerID = p1.UserID
		} else if p2.TotalVP() > p1.TotalVP() {
			e.state.WinnerID = p2.UserID
		}
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

// --- Shared helpers ---

func (e *Engine) resetReady() {
	for _, p := range e.state.Players {
		if p != nil {
			p.Ready = false
		}
	}
}

func (e *Engine) bothReady() bool {
	return e.state.Players[0] != nil && e.state.Players[1] != nil &&
		e.state.Players[0].Ready && e.state.Players[1].Ready
}

func (e *Engine) clearPhaseStratagems() {
	for _, p := range e.state.Players {
		if p != nil {
			p.StratagemsUsedThisPhase = nil
		}
	}
}

func (e *Engine) resetRound() {
	for _, p := range e.state.Players {
		if p != nil {
			p.CPGainedThisRound = 0
			p.PrimaryScoredThisRound = 0
			p.SecondaryScoredThisRound = 0
		}
	}
}

func (e *Engine) sideOf(playerNumber int) scoring.Side {
	if p := e.state.GetPlayer(playerNumber); p != nil {
		return p.Side
	}
	return ""
}

// snapshotControl records each objective's controller at the start of the turn.
func (e *Engine) snapshotControl() {
	m := make(map[int]int, len(e.state.Board.Objectives))
	for _, o := range e.state.Board.Objectives {
		m[o.Index] = o.ControlledBy
	}
	e.state.StartOfTurnControl = m
}

// scorePrimary / scoreSecondary apply VP for a category, clamped by both the
// per-round and per-game caps, returning the applied amount.
func (e *Engine) scorePrimary(p *PlayerState, want int) int {
	caps := scoring.Caps{PerRound: e.state.roundCap(), PerGame: e.state.gameCap()}
	applied := caps.Clamp(want, p.PrimaryScoredThisRound, p.VPPrimary)
	p.VPPrimary += applied
	p.PrimaryScoredThisRound += applied
	return applied
}

func (e *Engine) scoreSecondary(p *PlayerState, want int) int {
	caps := scoring.Caps{PerRound: e.state.roundCap(), PerGame: e.state.gameCap()}
	applied := caps.Clamp(want, p.SecondaryScoredThisRound, p.VPSecondary)
	p.VPSecondary += applied
	p.SecondaryScoredThisRound += applied
	return applied
}

func opponentNumber(playerNumber int) int {
	if playerNumber == 1 {
		return 2
	}
	if playerNumber == 2 {
		return 1
	}
	return 0
}

func otherSide(s scoring.Side) scoring.Side {
	if s == scoring.SideAttacker {
		return scoring.SideDefender
	}
	return scoring.SideAttacker
}

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

// stringsFromData reads a string slice from JSON-decoded action data, where a
// list arrives as []any of strings (or, in tests, a []string directly).
func stringsFromData(data map[string]any, key string) []string {
	switch v := data[key].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
