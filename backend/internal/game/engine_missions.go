package game

import (
	"fmt"
	"math/rand/v2"
)

// --- Setup Actions ---

func (e *Engine) applySelectPrimaryMission(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select primary mission during setup")
	}

	missionPackID := strFromData(action.Data, "missionPackId")
	missionID := strFromData(action.Data, "missionId")
	missionName := strFromData(action.Data, "missionName")

	if missionID == "" {
		return nil, fmt.Errorf("missionId is required")
	}

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
		Type:         EventPrimaryMissionSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"missionPackId": missionPackID, "missionId": missionID, "missionName": missionName},
	}}, nil
}

func (e *Engine) applySelectTwist(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select twist during setup")
	}

	twistID := strFromData(action.Data, "twistId")
	twistName := strFromData(action.Data, "twistName")

	if twistID == "" {
		return nil, fmt.Errorf("twistId is required")
	}

	e.state.TwistID = twistID
	e.state.TwistName = twistName

	// Reset readiness when twist changes
	for _, p := range e.state.Players {
		if p != nil {
			p.Ready = false
		}
	}

	return []GameEvent{{
		Type:         EventTwistSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"twistId": twistID, "twistName": twistName},
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
		return nil, fmt.Errorf("mode must be 'fixed' or 'tactical'")
	}

	player.SecondaryMode = mode
	player.Ready = false
	// Clear any previous selections when mode changes
	player.ActiveSecondaries = nil
	player.TacticalDeck = nil

	return []GameEvent{{
		Type:         EventSecondaryModeSelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"mode": mode},
	}}, nil
}

func (e *Engine) applySetFixedSecondaries(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only set fixed secondaries during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "fixed" {
		return nil, fmt.Errorf("player must be in fixed secondary mode")
	}

	secondaries, err := activeSecondariesFromData(action.Data, "secondaries")
	if err != nil {
		return nil, fmt.Errorf("invalid secondaries data: %w", err)
	}

	if len(secondaries) != 2 {
		return nil, fmt.Errorf("must select exactly 2 fixed secondaries")
	}

	player.ActiveSecondaries = secondaries
	player.Ready = false

	return []GameEvent{{
		Type:         EventSecondarySelected,
		PlayerNumber: action.PlayerNumber,
		Data:         map[string]any{"secondaries": secondaries},
	}}, nil
}

func (e *Engine) applyInitTacticalDeck(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only initialize tactical deck during setup")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("player must be in tactical secondary mode")
	}

	deck, err := activeSecondariesFromData(action.Data, "deck")
	if err != nil {
		return nil, fmt.Errorf("invalid deck data: %w", err)
	}

	if len(deck) == 0 {
		return nil, fmt.Errorf("deck cannot be empty")
	}

	player.TacticalDeck = deck
	player.ActiveSecondaries = nil
	player.Ready = false

	return nil, nil
}

// --- Gameplay Actions ---

func (e *Engine) applyDrawSecondary(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	if e.state.CurrentPhase != PhaseCommand {
		return nil, fmt.Errorf("can only draw secondaries during the Command Phase")
	}

	if action.PlayerNumber != e.state.ActivePlayer {
		return nil, fmt.Errorf("only the active player can draw secondaries")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("can only draw secondaries in tactical mode")
	}

	if len(player.ActiveSecondaries) >= 2 {
		return nil, fmt.Errorf("already have 2 active secondaries")
	}

	var events []GameEvent
	for len(player.ActiveSecondaries) < 2 && len(player.TacticalDeck) > 0 {
		drawn, drawEvents := drawNextCard(player, e.state.CurrentRound, action.PlayerNumber, e.state.CurrentPhase)
		events = append(events, drawEvents...)
		if drawn == nil {
			break
		}
		player.ActiveSecondaries = append(player.ActiveSecondaries, *drawn)
		events = append(events, GameEvent{
			Type:         EventSecondaryDrawn,
			PlayerNumber: action.PlayerNumber,
			Round:        e.state.CurrentRound,
			Phase:        e.state.CurrentPhase,
			Data:         map[string]any{"secondaryId": drawn.ID, "secondaryName": drawn.Name},
		})
	}

	return events, nil
}

// drawNextCard pops the top card of the player's tactical deck and applies any
// mandatory "when drawn" restriction by shuffling the card back into the deck
// (at a random position) and drawing the next one. Returns the final drawn
// card (or nil if the deck contains no drawable cards) along with any
// reshuffle events emitted along the way.
func drawNextCard(player *PlayerState, round, playerNumber int, phase Phase) (*ActiveSecondary, []GameEvent) {
	var events []GameEvent
	for len(player.TacticalDeck) > 0 {
		card := player.TacticalDeck[0]
		player.TacticalDeck = player.TacticalDeck[1:]
		if isMandatoryReshuffle(card, round) {
			// Bail cleanly if the rest of the deck is all restricted — otherwise
			// we'd cycle forever.
			anyDrawable := false
			for _, c := range player.TacticalDeck {
				if !isMandatoryReshuffle(c, round) {
					anyDrawable = true
					break
				}
			}
			player.TacticalDeck = insertRandomlyIntoDeck(player.TacticalDeck, card)
			if !anyDrawable {
				return nil, events
			}
			events = append(events, GameEvent{
				Type:         EventSecondaryReshuffled,
				PlayerNumber: playerNumber,
				Round:        round,
				Phase:        phase,
				Data: map[string]any{
					"secondaryId":   card.ID,
					"secondaryName": card.Name,
					"reason":        "mandatory",
				},
			})
			continue
		}
		return &card, events
	}
	return nil, events
}

func isMandatoryReshuffle(card ActiveSecondary, round int) bool {
	return card.DrawRestriction != nil &&
		card.DrawRestriction.Round == round &&
		card.DrawRestriction.Mode == DrawRestrictionMandatory
}

// insertRandomlyIntoDeck returns the deck with card inserted at a random
// position (inclusive of start and end).
func insertRandomlyIntoDeck(deck []ActiveSecondary, card ActiveSecondary) []ActiveSecondary {
	pos := rand.IntN(len(deck) + 1)
	result := make([]ActiveSecondary, 0, len(deck)+1)
	result = append(result, deck[:pos]...)
	result = append(result, card)
	result = append(result, deck[pos:]...)
	return result
}

func (e *Engine) applyAchieveSecondary(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	secondaryID := strFromData(action.Data, "secondaryId")
	vpScored := intFromData(action.Data, "vpScored")

	idx := -1
	for i, s := range player.ActiveSecondaries {
		if s.ID == secondaryID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("secondary not found in active secondaries")
	}

	achieved := player.ActiveSecondaries[idx]

	// Validate vpScored against scoring options
	if len(achieved.ScoringOptions) > 0 && vpScored > 0 {
		valid := false
		for _, opt := range achieved.ScoringOptions {
			if opt.Mode != "" && opt.Mode != player.SecondaryMode {
				continue
			}
			if opt.VP == vpScored {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("invalid VP score %d: does not match any scoring option", vpScored)
		}
	}
	player.ActiveSecondaries = append(player.ActiveSecondaries[:idx], player.ActiveSecondaries[idx+1:]...)
	player.AchievedSecondaries = append(player.AchievedSecondaries, achieved)

	// Score VP
	if vpScored > 0 {
		player.VPSecondary = ClampVP(player.VPSecondary+vpScored, MaxVPSecondary)
	}

	return []GameEvent{{
		Type:         EventSecondaryAchieved,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"secondaryId":   secondaryID,
			"secondaryName": achieved.Name,
			"vpScored":      vpScored,
			"vpSecondary":   player.VPSecondary,
		},
	}}, nil
}

func (e *Engine) applyDiscardSecondary(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("can only discard secondaries in tactical mode")
	}

	secondaryID := strFromData(action.Data, "secondaryId")

	idx := -1
	for i, s := range player.ActiveSecondaries {
		if s.ID == secondaryID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("secondary not found in active secondaries")
	}

	discarded := player.ActiveSecondaries[idx]
	player.ActiveSecondaries = append(player.ActiveSecondaries[:idx], player.ActiveSecondaries[idx+1:]...)
	player.DiscardedSecondaries = append(player.DiscardedSecondaries, discarded)

	// End-of-turn discard grants 1CP (except round 5); free discard grants nothing.
	// Players can only gain a maximum of 1 additional CP per battle round beyond
	// the automatic Command Phase gain.
	free := false
	if v, ok := action.Data["free"].(bool); ok {
		free = v
	}
	cpGained := 0
	if !free && e.state.CurrentRound < MaxRounds && player.CPGainedThisRound < 1 {
		player.CP++
		player.CPGainedThisRound++
		cpGained = 1
	}

	return []GameEvent{{
		Type:         EventSecondaryDiscarded,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"secondaryId":   secondaryID,
			"secondaryName": discarded.Name,
			"cpGained":      cpGained,
			"free":          free,
		},
	}}, nil
}

func (e *Engine) applyNewOrders(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	if e.state.CurrentPhase != PhaseCommand {
		return nil, fmt.Errorf("can only use New Orders during the Command Phase")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("new orders only available in tactical mode")
	}

	if player.NewOrdersUsedThisPhase {
		return nil, fmt.Errorf("new orders can only be used once per Command phase")
	}

	cpCost := e.newOrdersCPCost()
	if player.CP < cpCost {
		return nil, fmt.Errorf("insufficient CP: have %d, need %d", player.CP, cpCost)
	}

	discardID := strFromData(action.Data, "discardSecondaryId")

	// Find and discard the specified secondary
	idx := -1
	for i, s := range player.ActiveSecondaries {
		if s.ID == discardID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("secondary not found in active secondaries")
	}

	discarded := player.ActiveSecondaries[idx]
	player.ActiveSecondaries = append(player.ActiveSecondaries[:idx], player.ActiveSecondaries[idx+1:]...)
	player.DiscardedSecondaries = append(player.DiscardedSecondaries, discarded)

	// Spend CP
	player.CP -= cpCost
	player.NewOrdersUsedThisPhase = true

	// Draw replacement from deck (applies mandatory reshuffle rules)
	drawn, drawEvents := drawNextCard(player, e.state.CurrentRound, action.PlayerNumber, e.state.CurrentPhase)
	if drawn != nil {
		player.ActiveSecondaries = append(player.ActiveSecondaries, *drawn)
	}

	data := map[string]any{
		"discardedId":   discardID,
		"discardedName": discarded.Name,
		"cpSpent":       cpCost,
	}
	if drawn != nil {
		data["drawnId"] = drawn.ID
		data["drawnName"] = drawn.Name
	}

	events := drawEvents
	events = append(events, GameEvent{
		Type:         EventNewOrdersUsed,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         data,
	})
	return events, nil
}

func (e *Engine) applyReshuffleSecondary(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("can only reshuffle secondaries in tactical mode")
	}

	secondaryID := strFromData(action.Data, "secondaryId")

	idx := -1
	for i, s := range player.ActiveSecondaries {
		if s.ID == secondaryID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("secondary not found in active secondaries")
	}

	card := player.ActiveSecondaries[idx]
	if card.DrawRestriction == nil ||
		card.DrawRestriction.Mode != DrawRestrictionOptional ||
		card.DrawRestriction.Round != e.state.CurrentRound {
		return nil, fmt.Errorf("secondary cannot be reshuffled: no optional draw restriction active this round")
	}

	player.ActiveSecondaries = append(player.ActiveSecondaries[:idx], player.ActiveSecondaries[idx+1:]...)
	player.TacticalDeck = insertRandomlyIntoDeck(player.TacticalDeck, card)

	events := []GameEvent{{
		Type:         EventSecondaryReshuffled,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"secondaryId":   card.ID,
			"secondaryName": card.Name,
			"reason":        "optional",
		},
	}}

	// Draw a replacement from the deck (applies mandatory reshuffle rules).
	drawn, drawEvents := drawNextCard(player, e.state.CurrentRound, action.PlayerNumber, e.state.CurrentPhase)
	events = append(events, drawEvents...)
	if drawn != nil {
		player.ActiveSecondaries = append(player.ActiveSecondaries, *drawn)
		events = append(events, GameEvent{
			Type:         EventSecondaryDrawn,
			PlayerNumber: action.PlayerNumber,
			Round:        e.state.CurrentRound,
			Phase:        e.state.CurrentPhase,
			Data:         map[string]any{"secondaryId": drawn.ID, "secondaryName": drawn.Name},
		})
	}

	return events, nil
}

// applyMoveSecondary is the manual escape hatch for tactical mode players who
// are tracking their secondaries with a physical deck. It moves a card between
// any two of the four piles (deck, active, achieved, discarded) with no phase,
// active-player, or CP restrictions, and optionally adjusts secondary VP by
// the supplied delta (no scoring-option validation).
func (e *Engine) applyMoveSecondary(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("can only manually move secondaries in tactical mode")
	}

	secondaryID := strFromData(action.Data, "secondaryId")
	fromPile := strFromData(action.Data, "fromPile")
	toPile := strFromData(action.Data, "toPile")
	vpDelta := intFromData(action.Data, "vpScored")

	if !isValidSecondaryPile(fromPile) || !isValidSecondaryPile(toPile) {
		return nil, fmt.Errorf("fromPile and toPile must be one of: deck, active, achieved, discarded")
	}
	if fromPile == toPile {
		return nil, fmt.Errorf("fromPile and toPile must differ")
	}

	card, ok := removeFromSecondaryPile(player, fromPile, secondaryID)
	if !ok {
		return nil, fmt.Errorf("secondary not found in %s pile", fromPile)
	}
	appendToSecondaryPile(player, toPile, card)

	appliedDelta := 0
	if vpDelta != 0 {
		oldVP := player.VPSecondary
		player.VPSecondary = ClampVP(oldVP+vpDelta, MaxVPSecondary)
		appliedDelta = player.VPSecondary - oldVP
	}

	return []GameEvent{{
		Type:         EventSecondaryMoved,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data: map[string]any{
			"secondaryId":   card.ID,
			"secondaryName": card.Name,
			"fromPile":      fromPile,
			"toPile":        toPile,
			"vpDelta":       appliedDelta,
			"vpSecondary":   player.VPSecondary,
		},
	}}, nil
}

func isValidSecondaryPile(name string) bool {
	switch name {
	case "deck", "active", "achieved", "discarded":
		return true
	}
	return false
}

func secondaryPilePtr(player *PlayerState, pile string) *[]ActiveSecondary {
	switch pile {
	case "deck":
		return &player.TacticalDeck
	case "active":
		return &player.ActiveSecondaries
	case "achieved":
		return &player.AchievedSecondaries
	case "discarded":
		return &player.DiscardedSecondaries
	}
	return nil
}

func removeFromSecondaryPile(player *PlayerState, pile, id string) (ActiveSecondary, bool) {
	p := secondaryPilePtr(player, pile)
	if p == nil {
		return ActiveSecondary{}, false
	}
	for i, s := range *p {
		if s.ID == id {
			card := (*p)[i]
			*p = append((*p)[:i], (*p)[i+1:]...)
			return card, true
		}
	}
	return ActiveSecondary{}, false
}

func appendToSecondaryPile(player *PlayerState, pile string, card ActiveSecondary) {
	p := secondaryPilePtr(player, pile)
	if p == nil {
		return
	}
	*p = append(*p, card)
}

func (e *Engine) applyDrawChallengerCard(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	if e.state.CurrentPhase != PhaseCommand {
		return nil, fmt.Errorf("can only draw challenger cards during the Command Phase")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	// Validate trailing by 6+ VP
	opponent := e.state.GetPlayer(3 - action.PlayerNumber)
	if opponent == nil {
		return nil, fmt.Errorf("opponent not found")
	}

	vpDiff := opponent.TotalVP() - player.TotalVP()
	if vpDiff < ChallengerVPThreshold {
		return nil, fmt.Errorf("must be trailing by %d+ VP to draw a challenger card (currently trailing by %d)", ChallengerVPThreshold, vpDiff)
	}

	cardID := strFromData(action.Data, "challengerCardId")
	cardName := strFromData(action.Data, "challengerCardName")

	player.IsChallenger = true
	player.ChallengerCardID = cardID

	return []GameEvent{{
		Type:         EventChallengerActivated,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"challengerCardId": cardID, "challengerCardName": cardName},
	}}, nil
}

func (e *Engine) applyScoreChallenger(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if !player.IsChallenger || player.ChallengerCardID == "" {
		return nil, fmt.Errorf("no active challenger card")
	}

	vpScored := intFromData(action.Data, "vpScored")
	if vpScored <= 0 {
		vpScored = ChallengerCardVP
	}

	player.VPGambit = ClampVP(player.VPGambit+vpScored, MaxVPGambit)
	cardID := player.ChallengerCardID
	player.ChallengerCardID = ""

	return []GameEvent{{
		Type:         EventChallengerScored,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         map[string]any{"challengerCardId": cardID, "vpScored": vpScored, "vpGambit": player.VPGambit},
	}}, nil
}

func (e *Engine) applyAdaptOrDie(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	if !e.canUseAdaptOrDie(e.state.GetPlayer(action.PlayerNumber)) {
		return nil, fmt.Errorf("adapt or die not available")
	}

	player := e.state.GetPlayer(action.PlayerNumber)

	var events []GameEvent

	if player.SecondaryMode == "fixed" {
		// Swap one fixed secondary for another
		discardID := strFromData(action.Data, "discardSecondaryId")
		newSecondary, err := singleActiveSecondaryFromData(action.Data, "newSecondary")
		if err != nil {
			return nil, fmt.Errorf("invalid new secondary: %w", err)
		}

		idx := -1
		for i, s := range player.ActiveSecondaries {
			if s.ID == discardID {
				idx = i
				break
			}
		}
		if idx == -1 {
			return nil, fmt.Errorf("secondary not found in active secondaries")
		}

		player.ActiveSecondaries[idx] = newSecondary
	} else {
		// Tactical: draw extra card, shuffle one back
		// The client sends which card to shuffle back
		shuffleBackID := strFromData(action.Data, "shuffleBackSecondaryId")

		if len(player.TacticalDeck) == 0 {
			return nil, fmt.Errorf("tactical deck is empty")
		}

		// Draw one extra — applies mandatory reshuffle rules.
		drawn, drawEvents := drawNextCard(player, e.state.CurrentRound, action.PlayerNumber, e.state.CurrentPhase)
		events = append(events, drawEvents...)
		if drawn == nil {
			return nil, fmt.Errorf("tactical deck is empty")
		}
		player.ActiveSecondaries = append(player.ActiveSecondaries, *drawn)

		// Shuffle one back (random position)
		idx := -1
		for i, s := range player.ActiveSecondaries {
			if s.ID == shuffleBackID {
				idx = i
				break
			}
		}
		if idx == -1 {
			return nil, fmt.Errorf("secondary to shuffle back not found")
		}
		toShuffle := player.ActiveSecondaries[idx]
		player.ActiveSecondaries = append(player.ActiveSecondaries[:idx], player.ActiveSecondaries[idx+1:]...)
		player.TacticalDeck = insertRandomlyIntoDeck(player.TacticalDeck, toShuffle)
	}

	player.AdaptOrDieUses++

	return events, nil
}

// --- Helpers ---

func activeSecondariesFromData(data map[string]any, key string) ([]ActiveSecondary, error) {
	raw, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("missing key: %s", key)
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("expected array for %s", key)
	}

	result := make([]ActiveSecondary, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected object in %s array", key)
		}
		s := ActiveSecondary{
			ID:              strFromMapAny(m, "id"),
			Name:            strFromMapAny(m, "name"),
			Description:     strFromMapAny(m, "description"),
			IsFixed:         boolFromMapAny(m, "isFixed"),
			MaxVP:           intFromMapAny(m, "maxVp"),
			ScoringOptions:  scoringOptionsFromMapAny(m, "scoringOptions"),
			DrawRestriction: drawRestrictionFromMapAny(m, "drawRestriction"),
			ScoringTiming:   strFromMapAny(m, "scoringTiming"),
		}
		result = append(result, s)
	}
	return result, nil
}

func singleActiveSecondaryFromData(data map[string]any, key string) (ActiveSecondary, error) {
	raw, ok := data[key]
	if !ok {
		return ActiveSecondary{}, fmt.Errorf("missing key: %s", key)
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return ActiveSecondary{}, fmt.Errorf("expected object for %s", key)
	}
	return ActiveSecondary{
		ID:              strFromMapAny(m, "id"),
		Name:            strFromMapAny(m, "name"),
		Description:     strFromMapAny(m, "description"),
		IsFixed:         boolFromMapAny(m, "isFixed"),
		MaxVP:           intFromMapAny(m, "maxVp"),
		ScoringOptions:  scoringOptionsFromMapAny(m, "scoringOptions"),
		DrawRestriction: drawRestrictionFromMapAny(m, "drawRestriction"),
		ScoringTiming:   strFromMapAny(m, "scoringTiming"),
	}, nil
}

func drawRestrictionFromMapAny(m map[string]any, key string) *SecondaryDrawRestriction {
	raw, ok := m[key]
	if !ok || raw == nil {
		return nil
	}
	om, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return &SecondaryDrawRestriction{
		Round: intFromMapAny(om, "round"),
		Mode:  strFromMapAny(om, "mode"),
	}
}

func strFromMapAny(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func boolFromMapAny(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func intFromMapAny(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func scoringOptionsFromMapAny(m map[string]any, key string) []SecondaryScoringOption {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	opts := make([]SecondaryScoringOption, 0, len(items))
	for _, item := range items {
		om, ok := item.(map[string]any)
		if !ok {
			continue
		}
		opts = append(opts, SecondaryScoringOption{
			Label: strFromMapAny(om, "label"),
			VP:    intFromMapAny(om, "vp"),
			Mode:  strFromMapAny(om, "mode"),
		})
	}
	return opts
}
