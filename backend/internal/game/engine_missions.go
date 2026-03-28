package game

import "fmt"

// --- Setup Actions ---

func (e *Engine) applySelectPrimaryMission(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusSetup {
		return nil, fmt.Errorf("can only select primary mission during setup")
	}

	missionID := strFromData(action.Data, "missionId")
	missionName := strFromData(action.Data, "missionName")

	if missionID == "" {
		return nil, fmt.Errorf("missionId is required")
	}

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
		Data:         map[string]any{"missionId": missionID, "missionName": missionName},
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
		drawn := player.TacticalDeck[0]
		player.TacticalDeck = player.TacticalDeck[1:]
		player.ActiveSecondaries = append(player.ActiveSecondaries, drawn)

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

	// Grant 1CP if not round 5
	cpGained := 0
	if e.state.CurrentRound < MaxRounds {
		player.CP++
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
		},
	}}, nil
}

func (e *Engine) applyNewOrders(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
	}

	player := e.state.GetPlayer(action.PlayerNumber)
	if player == nil {
		return nil, fmt.Errorf("invalid player number")
	}

	if player.SecondaryMode != "tactical" {
		return nil, fmt.Errorf("new orders only available in tactical mode")
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

	// Draw replacement from deck
	var drawn *ActiveSecondary
	if len(player.TacticalDeck) > 0 {
		d := player.TacticalDeck[0]
		player.TacticalDeck = player.TacticalDeck[1:]
		player.ActiveSecondaries = append(player.ActiveSecondaries, d)
		drawn = &d
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

	return []GameEvent{{
		Type:         EventNewOrdersUsed,
		PlayerNumber: action.PlayerNumber,
		Round:        e.state.CurrentRound,
		Phase:        e.state.CurrentPhase,
		Data:         data,
	}}, nil
}

func (e *Engine) applyDrawChallengerCard(action GameAction) ([]GameEvent, error) {
	if e.state.Status != StatusActive {
		return nil, fmt.Errorf("game is not active")
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

		// Draw one extra
		if len(player.TacticalDeck) == 0 {
			return nil, fmt.Errorf("tactical deck is empty")
		}
		drawn := player.TacticalDeck[0]
		player.TacticalDeck = player.TacticalDeck[1:]
		player.ActiveSecondaries = append(player.ActiveSecondaries, drawn)

		// Shuffle one back
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
		// Add back to deck (at random position — for simplicity, add to end)
		player.TacticalDeck = append(player.TacticalDeck, toShuffle)
	}

	player.AdaptOrDieUses++

	return nil, nil
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
			ID:          strFromMapAny(m, "id"),
			Name:        strFromMapAny(m, "name"),
			Description: strFromMapAny(m, "description"),
			IsFixed:     boolFromMapAny(m, "isFixed"),
			MaxVP:       intFromMapAny(m, "maxVp"),
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
		ID:          strFromMapAny(m, "id"),
		Name:        strFromMapAny(m, "name"),
		Description: strFromMapAny(m, "description"),
		IsFixed:     boolFromMapAny(m, "isFixed"),
		MaxVP:       intFromMapAny(m, "maxVp"),
	}, nil
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
