package game

import (
	"fmt"

	"github.com/peter/tacticarium/backend/internal/game/scoring"
)

// --- Force disposition + mission resolution ---

func (e *Engine) applySelectForceDisposition(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select force disposition during setup")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	disposition := strFromData(action.Data, "disposition")
	if disposition == "" {
		return nil, fmt.Errorf("disposition is required")
	}
	player.ForceDisposition = disposition
	player.ForceDispositionName = strFromData(action.Data, "dispositionName")
	e.resetReady()

	events := []GameEvent{{
		Type:         EventForceDispositionSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"disposition": disposition, "dispositionName": player.ForceDispositionName},
	}}
	events = append(events, e.resolveMissions()...)
	return events, nil
}

// resolveMissions resolves each player's asymmetric primary mission once both
// players have chosen a force disposition.
func (e *Engine) resolveMissions() []GameEvent {
	if e.missionResolver == nil {
		return nil
	}
	p1, p2 := e.state.Players[0], e.state.Players[1]
	if p1 == nil || p2 == nil || p1.ForceDisposition == "" || p2.ForceDisposition == "" {
		return nil
	}

	var events []GameEvent
	for _, pair := range [][2]*PlayerState{{p1, p2}, {p2, p1}} {
		p, opp := pair[0], pair[1]
		rm, ok := e.missionResolver(p.ForceDisposition, opp.ForceDisposition)
		if !ok {
			continue
		}
		p.MissionID = rm.ID
		p.MissionName = rm.Name
		p.PrimaryCard = rm.PrimaryCard
		if rm.GameCap > 0 {
			e.state.VPPerGameCap = rm.GameCap
		}
		if rm.RoundCap > 0 {
			e.state.VPPerRoundCap = rm.RoundCap
		}
		events = append(events, GameEvent{
			Type:         EventMissionResolved,
			PlayerNumber: p.PlayerNumber,
			Data:         map[string]any{"missionId": rm.ID, "missionName": rm.Name},
		})
	}
	return events
}

// --- Board control ---

func (e *Engine) applySetObjectiveControl(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	idx := intFromData(action.Data, "objectiveIndex")
	controller := intFromData(action.Data, "player") // 0 = none, 1, 2
	if controller < 0 || controller > 2 {
		return nil, fmt.Errorf("invalid controlling player")
	}
	obj := e.objective(idx)
	if obj == nil {
		return nil, fmt.Errorf("unknown objective %d", idx)
	}
	obj.ControlledBy = controller
	return []GameEvent{{
		Type:         EventObjectiveControlChanged,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"objectiveIndex": idx, "controlledBy": controller},
	}}, nil
}

func (e *Engine) applySetObjectiveTag(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	idx := intFromData(action.Data, "objectiveIndex")
	tag := strFromData(action.Data, "tag")
	if tag == "" {
		return nil, fmt.Errorf("tag is required")
	}
	add, _ := action.Data["add"].(bool)
	obj := e.objective(idx)
	if obj == nil {
		return nil, fmt.Errorf("unknown objective %d", idx)
	}
	if add {
		obj.AddTag(tag)
	} else {
		removeTag(obj, tag)
	}
	return []GameEvent{{
		Type:         EventObjectiveTagged,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"objectiveIndex": idx, "tag": tag, "add": add},
	}}, nil
}

func (e *Engine) objective(index int) *scoring.Objective {
	for i := range e.state.Board.Objectives {
		if e.state.Board.Objectives[i].Index == index {
			return &e.state.Board.Objectives[i]
		}
	}
	return nil
}

func removeTag(o *scoring.Objective, tag string) {
	out := o.Tags[:0]
	for _, t := range o.Tags {
		if t != tag {
			out = append(out, t)
		}
	}
	o.Tags = out
}

// --- Secondary deck ---

func (e *Engine) applyDrawSecondaries(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	if e.state.CurrentPhase != PhaseCommand {
		return nil, fmt.Errorf("can only draw secondaries during the Command phase")
	}
	if action.PlayerNumber != e.state.ActivePlayer {
		return nil, fmt.Errorf("only the active player can draw secondaries")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("only tactical players draw secondaries")
	}

	// Draw two cards, keeping any already in hand (11e: no maximum hand size).
	var events []GameEvent
	for i := 0; i < 2 && len(player.SecondaryDeck) > 0; i++ {
		card := player.SecondaryDeck[0]
		player.SecondaryDeck = player.SecondaryDeck[1:]
		player.SecondaryHand = append(player.SecondaryHand, card)
		events = append(events, GameEvent{
			Type:         EventSecondaryDrawn,
			PlayerNumber: action.PlayerNumber,
			Round:        e.state.CurrentRound,
			Phase:        e.state.CurrentPhase,
			Data:         map[string]any{"cardId": card.ID, "cardName": card.Name},
		})
	}
	return events, nil
}

// --- Scoring (evaluator integration) ---

// fireScoring evaluates every player's active cards at a trigger moment,
// auto-applying Layer-1 VP and raising prompts for Layer-2 awards.
func (e *Engine) fireScoring(timing string, phase Phase) []GameEvent {
	var events []GameEvent
	for _, p := range e.state.Players {
		if p == nil {
			continue
		}
		ctx := e.scoreContext(p, timing, phase)

		// Primary mission card.
		if len(p.PrimaryCard.Awards) > 0 {
			events = append(events, e.scoreCard(p, &p.PrimaryCard, "primary", ctx)...)
		}
		// Secondary hand.
		for i := range p.SecondaryHand {
			card := &p.SecondaryHand[i].Card
			events = append(events, e.scoreCard(p, card, "secondary", ctx)...)
		}
	}
	return events
}

func (e *Engine) scoreContext(p *PlayerState, timing string, phase Phase) scoring.Context {
	playerTurn := "opponent-turn"
	if p.PlayerNumber == e.state.ActivePlayer {
		playerTurn = "your-turn"
	}
	return scoring.Context{
		Round:              e.state.CurrentRound,
		Phase:              string(phase),
		Timing:             timing,
		PlayerTurn:         playerTurn,
		Mode:               p.SecondaryMode,
		Player:             p.PlayerNumber,
		Board:              &e.state.Board,
		StartOfTurnControl: e.state.StartOfTurnControl,
	}
}

// scoreCard evaluates one card: auto-applies Layer-1 VP and emits a prompt for
// each Layer-2 award needing confirmation.
func (e *Engine) scoreCard(p *PlayerState, card *scoring.Card, category string, ctx scoring.Context) []GameEvent {
	results, total := scoring.EvaluateCard(card, ctx)
	var events []GameEvent

	if total > 0 {
		applied := e.applyCategoryVP(p, category, total)
		if applied > 0 {
			events = append(events, GameEvent{
				Type:         EventCardScored,
				PlayerNumber: p.PlayerNumber,
				Round:        e.state.CurrentRound,
				Phase:        e.state.CurrentPhase,
				Data: map[string]any{
					"cardId": card.ID, "cardName": card.Name,
					"category": category, "appliedDelta": applied, "newTotal": p.TotalVP(),
				},
			})
		}
	}

	for _, r := range results {
		if !r.NeedsConfirm {
			continue
		}
		e.promptSeq++
		prompt := ScorePrompt{
			ID:         fmt.Sprintf("prompt-%d", e.promptSeq),
			Category:   category,
			CardID:     card.ID,
			CardName:   card.Name,
			AwardIndex: r.Index,
			Round:      e.state.CurrentRound,
			Text:       card.Text,
		}
		p.PendingScorePrompts = append(p.PendingScorePrompts, prompt)
		events = append(events, GameEvent{
			Type:         EventScorePrompt,
			PlayerNumber: p.PlayerNumber,
			Round:        e.state.CurrentRound,
			Phase:        e.state.CurrentPhase,
			Data: map[string]any{
				"promptId": prompt.ID, "category": category,
				"cardId": card.ID, "cardName": card.Name, "text": card.Text,
			},
		})
	}
	return events
}

func (e *Engine) applyCategoryVP(p *PlayerState, category string, want int) int {
	if category == "primary" {
		return e.scorePrimary(p, want)
	}
	return e.scoreSecondary(p, want)
}

// applyConfirmAward resolves a Layer-2 prompt with the player-supplied count and
// applies the resulting VP.
func (e *Engine) applyConfirmAward(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}
	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}
	promptID := strFromData(action.Data, "promptId")
	count := intFromData(action.Data, "count")

	idx := -1
	for i := range player.PendingScorePrompts {
		if player.PendingScorePrompts[i].ID == promptID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("unknown scoring prompt %q", promptID)
	}
	prompt := player.PendingScorePrompts[idx]

	card := e.cardForPrompt(player, prompt)
	if card == nil || prompt.AwardIndex >= len(card.Awards) {
		return nil, fmt.Errorf("scoring prompt references a missing card/award")
	}

	ctx := e.scoreContext(player, "", "")
	ctx.Round = prompt.Round
	ctx.Confirmed = map[int]int{prompt.AwardIndex: count}
	res := scoring.EvaluateAward(&card.Awards[prompt.AwardIndex], prompt.AwardIndex, ctx)
	applied := e.applyCategoryVP(player, prompt.Category, res.VP)

	// Remove the resolved prompt.
	player.PendingScorePrompts = append(player.PendingScorePrompts[:idx], player.PendingScorePrompts[idx+1:]...)

	return []GameEvent{{
		Type:         EventAwardConfirmed,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"promptId": promptID, "category": prompt.Category, "count": count,
			"appliedDelta": applied, "newTotal": player.TotalVP(),
		},
	}}, nil
}

func (e *Engine) cardForPrompt(p *PlayerState, prompt ScorePrompt) *scoring.Card {
	if prompt.Category == "primary" {
		return &p.PrimaryCard
	}
	for i := range p.SecondaryHand {
		if p.SecondaryHand[i].ID == prompt.CardID {
			return &p.SecondaryHand[i].Card
		}
	}
	return nil
}
