package game

import (
	"context"
	"fmt"
	"testing"
)

// --- Helpers ---

func newTestState() *GameState {
	return &GameState{
		GameID: "test-game",
		Status: StatusSetup,
		Players: [2]*PlayerState{
			{UserID: "user1", Username: "Player1", PlayerNumber: 1, CP: 0},
			{UserID: "user2", Username: "Player2", PlayerNumber: 2, CP: 0},
		},
	}
}

func newActiveTestState() *GameState {
	state := newTestState()
	state.Status = StatusActive
	state.CurrentRound = 1
	state.CurrentTurn = 1
	state.CurrentPhase = PhaseCommand
	state.ActivePlayer = 1
	state.FirstTurnPlayer = 1
	return state
}

func makeSecondary(id, name string) map[string]any {
	return map[string]any{
		"id":          id,
		"name":        name,
		"description": "desc",
		"isFixed":     false,
		"maxVp":       5,
	}
}

func makeActiveSecondary(id, name string) ActiveSecondary {
	return ActiveSecondary{ID: id, Name: name, Description: "desc", MaxVP: 5}
}

func makeDeck(count int) []ActiveSecondary {
	deck := make([]ActiveSecondary, count)
	for i := range deck {
		deck[i] = ActiveSecondary{
			ID:   fmt.Sprintf("sec-%d", i+1),
			Name: fmt.Sprintf("Secondary %d", i+1),
			MaxVP: 5,
		}
	}
	return deck
}

// --- Setup Actions ---

func TestSelectPrimaryMission(t *testing.T) {
	e := NewEngine(newTestState())

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectPrimaryMission,
		PlayerNumber: 1,
		Data:         map[string]any{"missionPackId": "pack1", "missionId": "m1", "missionName": "Take and Hold"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventPrimaryMissionSelected {
		t.Fatal("expected primary_mission_selected event")
	}
	if e.state.MissionPackID != "pack1" || e.state.MissionID != "m1" || e.state.MissionName != "Take and Hold" {
		t.Fatal("mission not set on state")
	}
}

func TestSelectPrimaryMission_RequiresSetup(t *testing.T) {
	e := NewEngine(newActiveTestState())
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectPrimaryMission,
		PlayerNumber: 1,
		Data:         map[string]any{"missionId": "m1", "missionName": "M"},
	})
	if err == nil {
		t.Fatal("expected error when not in setup")
	}
}

func TestSelectPrimaryMission_RequiresMissionId(t *testing.T) {
	e := NewEngine(newTestState())
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectPrimaryMission,
		PlayerNumber: 1,
		Data:         map[string]any{"missionName": "M"},
	})
	if err == nil {
		t.Fatal("expected error for missing missionId")
	}
}

func TestSelectPrimaryMission_ResetsReadiness(t *testing.T) {
	state := newTestState()
	state.Players[0].Ready = true
	state.Players[1].Ready = true
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{
		Type:         ActionSelectPrimaryMission,
		PlayerNumber: 1,
		Data:         map[string]any{"missionId": "m1", "missionName": "M"},
	})
	if state.Players[0].Ready || state.Players[1].Ready {
		t.Fatal("expected readiness to be reset")
	}
}

func TestSelectTwist(t *testing.T) {
	e := NewEngine(newTestState())
	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectTwist,
		PlayerNumber: 1,
		Data:         map[string]any{"twistId": "t1", "twistName": "Adapt or Die"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventTwistSelected {
		t.Fatal("expected twist_selected event")
	}
	if e.state.TwistID != "t1" {
		t.Fatal("twist not set on state")
	}
}

func TestSelectTwist_RequiresSetup(t *testing.T) {
	e := NewEngine(newActiveTestState())
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectTwist,
		PlayerNumber: 1,
		Data:         map[string]any{"twistId": "t1", "twistName": "T"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSelectSecondaryMode(t *testing.T) {
	for _, mode := range []string{"fixed", "tactical"} {
		t.Run(mode, func(t *testing.T) {
			e := NewEngine(newTestState())
			events, err := e.Apply(context.Background(), GameAction{
				Type:         ActionSelectSecondaryMode,
				PlayerNumber: 1,
				Data:         map[string]any{"mode": mode},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 1 || events[0].Type != EventSecondaryModeSelected {
				t.Fatal("expected secondary_mode_selected event")
			}
			if e.state.Players[0].SecondaryMode != mode {
				t.Fatalf("expected mode %s, got %s", mode, e.state.Players[0].SecondaryMode)
			}
		})
	}
}

func TestSelectSecondaryMode_InvalidMode(t *testing.T) {
	e := NewEngine(newTestState())
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectSecondaryMode,
		PlayerNumber: 1,
		Data:         map[string]any{"mode": "invalid"},
	})
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestSelectSecondaryMode_ClearsPreviousSelections(t *testing.T) {
	state := newTestState()
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	state.Players[0].TacticalDeck = []ActiveSecondary{makeActiveSecondary("s2", "S2")}
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{
		Type:         ActionSelectSecondaryMode,
		PlayerNumber: 1,
		Data:         map[string]any{"mode": "tactical"},
	})
	if state.Players[0].ActiveSecondaries != nil {
		t.Fatal("expected active secondaries to be cleared")
	}
	if state.Players[0].TacticalDeck != nil {
		t.Fatal("expected tactical deck to be cleared")
	}
}

func TestSetFixedSecondaries(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "fixed"
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetFixedSecondaries,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaries": []any{
				makeSecondary("s1", "Sec 1"),
				makeSecondary("s2", "Sec 2"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatal("expected 1 event")
	}
	if len(state.Players[0].ActiveSecondaries) != 2 {
		t.Fatal("expected 2 active secondaries")
	}
}

func TestSetFixedSecondaries_RequiresFixedMode(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "tactical"
	e := NewEngine(state)
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetFixedSecondaries,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaries": []any{makeSecondary("s1", "S1"), makeSecondary("s2", "S2")},
		},
	})
	if err == nil {
		t.Fatal("expected error for non-fixed mode")
	}
}

func TestSetFixedSecondaries_RequiresExactly2(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "fixed"
	e := NewEngine(state)
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetFixedSecondaries,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaries": []any{makeSecondary("s1", "S1")},
		},
	})
	if err == nil {
		t.Fatal("expected error for != 2 secondaries")
	}
}

func TestInitTacticalDeck(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "tactical"
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionInitTacticalDeck,
		PlayerNumber: 1,
		Data: map[string]any{
			"deck": []any{
				makeSecondary("s1", "S1"),
				makeSecondary("s2", "S2"),
				makeSecondary("s3", "S3"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Players[0].TacticalDeck) != 3 {
		t.Fatal("expected deck of 3")
	}
}

func TestInitTacticalDeck_RequiresTacticalMode(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "fixed"
	e := NewEngine(state)
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionInitTacticalDeck,
		PlayerNumber: 1,
		Data:         map[string]any{"deck": []any{makeSecondary("s1", "S1")}},
	})
	if err == nil {
		t.Fatal("expected error for non-tactical mode")
	}
}

func TestInitTacticalDeck_EmptyDeckError(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "tactical"
	e := NewEngine(state)
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionInitTacticalDeck,
		PlayerNumber: 1,
		Data:         map[string]any{"deck": []any{}},
	})
	if err == nil {
		t.Fatal("expected error for empty deck")
	}
}

// --- Gameplay Actions ---

func TestDrawSecondary(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].TacticalDeck = makeDeck(5)
	state.Players[0].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 draw events, got %d", len(events))
	}
	if len(state.Players[0].ActiveSecondaries) != 2 {
		t.Fatal("expected 2 active secondaries after draw")
	}
	if len(state.Players[0].TacticalDeck) != 3 {
		t.Fatal("expected 3 remaining in deck")
	}
}

func TestDrawSecondary_AlreadyHas2(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "S1"),
		makeActiveSecondary("s2", "S2"),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err == nil {
		t.Fatal("expected error when already at 2 active")
	}
}

func TestDrawSecondary_DrawsOnlyNeeded(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].TacticalDeck = makeDeck(5)
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("existing", "E")}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 draw event, got %d", len(events))
	}
	if len(state.Players[0].ActiveSecondaries) != 2 {
		t.Fatal("expected 2 active secondaries")
	}
}

func TestDrawSecondary_EmptyDeck(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].TacticalDeck = []ActiveSecondary{}
	state.Players[0].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatal("expected no events from empty deck draw")
	}
}

func TestDrawSecondary_RequiresTacticalMode(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "fixed"
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err == nil {
		t.Fatal("expected error for non-tactical mode")
	}
}

func TestAchieveSecondary(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
		makeActiveSecondary("s2", "Sec 2"),
	}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventSecondaryAchieved {
		t.Fatal("expected secondary_achieved event")
	}
	if len(state.Players[0].ActiveSecondaries) != 1 {
		t.Fatal("expected 1 remaining active")
	}
	if len(state.Players[0].AchievedSecondaries) != 1 {
		t.Fatal("expected 1 achieved")
	}
	if state.Players[0].VPSecondary != 5 {
		t.Fatalf("expected 5 VP secondary, got %d", state.Players[0].VPSecondary)
	}
}

func TestAchieveSecondary_NotFound(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "nonexistent", "vpScored": 5},
	})
	if err == nil {
		t.Fatal("expected error for missing secondary")
	}
}

func TestAchieveSecondary_VPCapped(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].VPSecondary = 38
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 5},
	})
	if state.Players[0].VPSecondary != MaxVPSecondary {
		t.Fatalf("expected VP capped at %d, got %d", MaxVPSecondary, state.Players[0].VPSecondary)
	}
}

func TestDiscardSecondary(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 3
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 0
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
		makeActiveSecondary("s2", "Sec 2"),
	}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventSecondaryDiscarded {
		t.Fatal("expected secondary_discarded event")
	}
	if len(state.Players[0].ActiveSecondaries) != 1 {
		t.Fatal("expected 1 remaining active")
	}
	if len(state.Players[0].DiscardedSecondaries) != 1 {
		t.Fatal("expected 1 discarded")
	}
	if state.Players[0].CP != 1 {
		t.Fatal("expected +1 CP from discard (not round 5)")
	}
}

func TestDiscardSecondary_FreeNoCPGain(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 3
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 0
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "free": true},
	})
	if state.Players[0].CP != 0 {
		t.Fatal("expected no CP gain from free discard")
	}
}

func TestDiscardSecondary_Round5NoCPGain(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 5
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 0
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	})
	if state.Players[0].CP != 0 {
		t.Fatal("expected no CP gain in round 5")
	}
}

func TestDiscardSecondary_RequiresTactical(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "fixed"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	})
	if err == nil {
		t.Fatal("expected error for fixed mode")
	}
}

func TestNewOrders(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 2
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
		makeActiveSecondary("s2", "Sec 2"),
	}
	state.Players[0].TacticalDeck = makeDeck(3)
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionNewOrders,
		PlayerNumber: 1,
		Data:         map[string]any{"discardSecondaryId": "s1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventNewOrdersUsed {
		t.Fatal("expected new_orders_used event")
	}
	if state.Players[0].CP != 1 {
		t.Fatalf("expected CP=1 after spending 1, got %d", state.Players[0].CP)
	}
	if len(state.Players[0].ActiveSecondaries) != 2 {
		t.Fatal("expected still 2 active (discarded one, drew one)")
	}
	if len(state.Players[0].DiscardedSecondaries) != 1 {
		t.Fatal("expected 1 discarded")
	}
	if len(state.Players[0].TacticalDeck) != 2 {
		t.Fatal("expected 2 remaining in deck")
	}
}

func TestNewOrders_InsufficientCP(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 0
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionNewOrders,
		PlayerNumber: 1,
		Data:         map[string]any{"discardSecondaryId": "s1"},
	})
	if err == nil {
		t.Fatal("expected error for insufficient CP")
	}
}

func TestNewOrders_EmptyDeck(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 2
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "S1"),
		makeActiveSecondary("s2", "S2"),
	}
	state.Players[0].TacticalDeck = []ActiveSecondary{}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionNewOrders,
		PlayerNumber: 1,
		Data:         map[string]any{"discardSecondaryId": "s1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should still succeed, just no replacement drawn
	if len(state.Players[0].ActiveSecondaries) != 1 {
		t.Fatal("expected 1 active (discarded, no draw)")
	}
	if events[0].Data["drawnId"] != nil {
		t.Fatal("expected no drawnId when deck empty")
	}
}

// --- Challenger Cards ---

func TestDrawChallengerCard(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].VPPrimary = 0  // player 1 trailing
	state.Players[1].VPPrimary = 10 // player 2 leading by 10
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawChallengerCard,
		PlayerNumber: 1,
		Data:         map[string]any{"challengerCardId": "cc1", "challengerCardName": "Challenge A"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventChallengerActivated {
		t.Fatal("expected challenger_activated event")
	}
	if !state.Players[0].IsChallenger {
		t.Fatal("expected IsChallenger = true")
	}
	if state.Players[0].ChallengerCardID != "cc1" {
		t.Fatal("expected challenger card ID set")
	}
}

func TestDrawChallengerCard_NotTrailingEnough(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].VPPrimary = 5
	state.Players[1].VPPrimary = 10 // only trailing by 5, need 6
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawChallengerCard,
		PlayerNumber: 1,
		Data:         map[string]any{"challengerCardId": "cc1", "challengerCardName": "C"},
	})
	if err == nil {
		t.Fatal("expected error when not trailing by 6+")
	}
}

func TestDrawChallengerCard_ExactlyAtThreshold(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].VPPrimary = 0
	state.Players[1].VPPrimary = 6 // exactly 6 VP difference
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawChallengerCard,
		PlayerNumber: 1,
		Data:         map[string]any{"challengerCardId": "cc1", "challengerCardName": "C"},
	})
	if err != nil {
		t.Fatal("expected success at exactly 6 VP deficit")
	}
}

func TestScoreChallenger(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].IsChallenger = true
	state.Players[0].ChallengerCardID = "cc1"
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreChallenger,
		PlayerNumber: 1,
		Data:         map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventChallengerScored {
		t.Fatal("expected challenger_scored event")
	}
	if state.Players[0].VPGambit != ChallengerCardVP {
		t.Fatalf("expected %d VP gambit, got %d", ChallengerCardVP, state.Players[0].VPGambit)
	}
	if state.Players[0].ChallengerCardID != "" {
		t.Fatal("expected challenger card cleared")
	}
}

func TestScoreChallenger_NoActiveCard(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].IsChallenger = false
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreChallenger,
		PlayerNumber: 1,
		Data:         map[string]any{},
	})
	if err == nil {
		t.Fatal("expected error for no active challenger card")
	}
}

func TestScoreChallenger_CustomVP(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].IsChallenger = true
	state.Players[0].ChallengerCardID = "cc1"
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{
		Type:         ActionScoreChallenger,
		PlayerNumber: 1,
		Data:         map[string]any{"vpScored": 5},
	})
	if state.Players[0].VPGambit != 5 {
		t.Fatalf("expected 5 VP gambit, got %d", state.Players[0].VPGambit)
	}
}

// --- Adapt or Die ---

func TestAdaptOrDie_Fixed(t *testing.T) {
	state := newActiveTestState()
	state.TwistID = TwistAdaptOrDie
	state.Players[0].SecondaryMode = "fixed"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
		makeActiveSecondary("s2", "Sec 2"),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdaptOrDie,
		PlayerNumber: 1,
		Data: map[string]any{
			"discardSecondaryId": "s1",
			"newSecondary":       makeSecondary("s3", "Sec 3"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].ActiveSecondaries[0].ID != "s3" {
		t.Fatal("expected s1 to be swapped for s3")
	}
	if state.Players[0].AdaptOrDieUses != 1 {
		t.Fatal("expected 1 use")
	}
}

func TestAdaptOrDie_Fixed_OnlyOnce(t *testing.T) {
	state := newActiveTestState()
	state.TwistID = TwistAdaptOrDie
	state.Players[0].SecondaryMode = "fixed"
	state.Players[0].AdaptOrDieUses = 1
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "S1"),
		makeActiveSecondary("s2", "S2"),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdaptOrDie,
		PlayerNumber: 1,
		Data: map[string]any{
			"discardSecondaryId": "s1",
			"newSecondary":       makeSecondary("s3", "S3"),
		},
	})
	if err == nil {
		t.Fatal("expected error: fixed mode only allows 1 use")
	}
}

func TestAdaptOrDie_Tactical(t *testing.T) {
	state := newActiveTestState()
	state.TwistID = TwistAdaptOrDie
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
	}
	state.Players[0].TacticalDeck = makeDeck(3)
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdaptOrDie,
		PlayerNumber: 1,
		Data:         map[string]any{"shuffleBackSecondaryId": "s1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Drew 1 from deck (sec-1), shuffled back s1 → deck should have 2 original + s1 = 3
	if len(state.Players[0].TacticalDeck) != 3 {
		t.Fatalf("expected deck size 3, got %d", len(state.Players[0].TacticalDeck))
	}
	// Active should have the drawn card (sec-1) but not s1
	if len(state.Players[0].ActiveSecondaries) != 1 {
		t.Fatalf("expected 1 active, got %d", len(state.Players[0].ActiveSecondaries))
	}
	if state.Players[0].ActiveSecondaries[0].ID != "sec-1" {
		t.Fatalf("expected sec-1 active, got %s", state.Players[0].ActiveSecondaries[0].ID)
	}
	if state.Players[0].AdaptOrDieUses != 1 {
		t.Fatal("expected 1 use")
	}
}

func TestAdaptOrDie_Tactical_TwiceAllowed(t *testing.T) {
	state := newActiveTestState()
	state.TwistID = TwistAdaptOrDie
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].AdaptOrDieUses = 1 // already used once
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	state.Players[0].TacticalDeck = makeDeck(3)
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdaptOrDie,
		PlayerNumber: 1,
		Data:         map[string]any{"shuffleBackSecondaryId": "s1"},
	})
	if err != nil {
		t.Fatal("tactical mode should allow 2 uses")
	}
	if state.Players[0].AdaptOrDieUses != 2 {
		t.Fatal("expected 2 uses")
	}
}

func TestAdaptOrDie_Tactical_ThirdUseDenied(t *testing.T) {
	state := newActiveTestState()
	state.TwistID = TwistAdaptOrDie
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].AdaptOrDieUses = 2
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	state.Players[0].TacticalDeck = makeDeck(3)
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdaptOrDie,
		PlayerNumber: 1,
		Data:         map[string]any{"shuffleBackSecondaryId": "s1"},
	})
	if err == nil {
		t.Fatal("expected error: tactical mode only allows 2 uses")
	}
}

func TestAdaptOrDie_WrongTwist(t *testing.T) {
	state := newActiveTestState()
	state.TwistID = "some-other-twist"
	state.Players[0].SecondaryMode = "fixed"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdaptOrDie,
		PlayerNumber: 1,
		Data: map[string]any{
			"discardSecondaryId": "s1",
			"newSecondary":       makeSecondary("s3", "S3"),
		},
	})
	if err == nil {
		t.Fatal("expected error when twist is not adapt or die")
	}
}

// --- Existing Engine Actions ---

func TestAdvancePhase(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdvancePhase,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventPhaseAdvance {
		t.Fatal("expected phase_advance event")
	}
	if state.CurrentPhase != PhaseMovement {
		t.Fatalf("expected movement phase, got %s", state.CurrentPhase)
	}
}

func TestAdvancePhase_TurnEnd_FirstPlayer(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseFight // last phase
	state.Players[0].CP = 0
	state.Players[1].CP = 0
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdvancePhase,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should switch to player 2, same round
	if state.ActivePlayer != 2 {
		t.Fatalf("expected active player 2, got %d", state.ActivePlayer)
	}
	if state.CurrentRound != 1 {
		t.Fatalf("expected still round 1, got %d", state.CurrentRound)
	}
	if state.CurrentTurn != 2 {
		t.Fatalf("expected turn 2, got %d", state.CurrentTurn)
	}
	if state.CurrentPhase != PhaseCommand {
		t.Fatal("expected command phase")
	}
	// Both players gain 1 CP at the start of each Command Phase (including turn 2)
	if state.Players[0].CP != 1 {
		t.Fatalf("expected player 1 CP=1, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 1 {
		t.Fatalf("expected player 2 CP=1, got %d", state.Players[1].CP)
	}
	cpEvents := 0
	for _, ev := range events {
		if ev.Type == EventCPGain {
			cpEvents++
		}
	}
	if cpEvents != 2 {
		t.Fatalf("expected 2 CP gain events, got %d", cpEvents)
	}
}

func TestAdvancePhase_RoundEnd_CPGain(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.CurrentTurn = 2
	state.CurrentPhase = PhaseFight
	state.ActivePlayer = 2 // second player finishing = round advance
	state.FirstTurnPlayer = 1
	state.Players[0].CP = 0
	state.Players[1].CP = 0
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdvancePhase,
		PlayerNumber: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Round should advance to 2, player 1's turn
	if state.CurrentRound != 2 {
		t.Fatalf("expected round 2, got %d", state.CurrentRound)
	}
	if state.CurrentTurn != 1 {
		t.Fatalf("expected turn 1, got %d", state.CurrentTurn)
	}
	if state.ActivePlayer != 1 {
		t.Fatal("expected player 1 active")
	}
	// Both players should gain 1 CP at start of new battle round
	if state.Players[0].CP != 1 {
		t.Fatalf("expected player 1 CP=1, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 1 {
		t.Fatalf("expected player 2 CP=1, got %d", state.Players[1].CP)
	}
	// Should have 2 CP gain events (one per player)
	cpEvents := 0
	for _, ev := range events {
		if ev.Type == EventCPGain {
			cpEvents++
		}
	}
	if cpEvents != 2 {
		t.Fatalf("expected 2 CP gain events, got %d", cpEvents)
	}
}

func TestAdvancePhase_GameEnd(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 5
	state.CurrentPhase = PhaseFight
	state.ActivePlayer = 2
	state.FirstTurnPlayer = 1
	state.Players[0].VPPrimary = 30
	state.Players[1].VPPrimary = 20
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdvancePhase,
		PlayerNumber: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != StatusCompleted {
		t.Fatal("expected game completed")
	}
	hasEndEvent := false
	for _, ev := range events {
		if ev.Type == EventGameEnd {
			hasEndEvent = true
		}
	}
	if !hasEndEvent {
		t.Fatal("expected game_end event")
	}
	if state.WinnerID != "user1" {
		t.Fatalf("expected user1 as winner, got %s", state.WinnerID)
	}
}

func TestAdvancePhase_OnlyActivePlayer(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdvancePhase,
		PlayerNumber: 2, // not the active player
	})
	if err == nil {
		t.Fatal("expected error for non-active player")
	}
}

func TestAdjustCP(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": -2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].CP != 3 {
		t.Fatalf("expected 3 CP, got %d", state.Players[0].CP)
	}
}

func TestAdjustCP_CannotGoNegative(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 1
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": -5},
	})
	if err == nil {
		t.Fatal("expected error for negative CP")
	}
}

func TestScoreVP(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	tests := []struct {
		category string
		delta    int
		checkFn  func() int
	}{
		{"primary", 10, func() int { return state.Players[0].VPPrimary }},
		{"secondary", 5, func() int { return state.Players[0].VPSecondary }},
		{"gambit", 3, func() int { return state.Players[0].VPGambit }},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			_, err := e.Apply(context.Background(), GameAction{
				Type:         ActionScoreVP,
				PlayerNumber: 1,
				Data:         map[string]any{"category": tt.category, "delta": tt.delta},
			})
			if err != nil {
				t.Fatal(err)
			}
			if tt.checkFn() != tt.delta {
				t.Fatalf("expected %d VP, got %d", tt.delta, tt.checkFn())
			}
		})
	}
}

func TestScoreVP_InvalidCategory(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "invalid", "delta": 5},
	})
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestConcede(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionConcede,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != StatusCompleted {
		t.Fatal("expected completed")
	}
	if state.WinnerID != "user2" {
		t.Fatal("expected player 2 as winner")
	}
	hasEnd := false
	for _, ev := range events {
		if ev.Type == EventGameEnd {
			hasEnd = true
		}
	}
	if !hasEnd {
		t.Fatal("expected game_end event")
	}
}

func TestDeclareGambit(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 3
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDeclareGambit,
		PlayerNumber: 1,
		Data:         map[string]any{"gambitId": "g1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].GambitID != "g1" {
		t.Fatal("expected gambit set")
	}
}

func TestDeclareGambit_TooEarly(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 2
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDeclareGambit,
		PlayerNumber: 1,
		Data:         map[string]any{"gambitId": "g1"},
	})
	if err == nil {
		t.Fatal("expected error before round 3")
	}
}

func TestSetReady_StartGame(t *testing.T) {
	state := newTestState()
	state.Players[0].Ready = false
	state.Players[1].Ready = false
	e := NewEngine(state)

	// Player 1 readies up
	e.Apply(context.Background(), GameAction{
		Type:         ActionSetReady,
		PlayerNumber: 1,
		Data:         map[string]any{"ready": true},
	})
	if state.Status != StatusSetup {
		t.Fatal("game should still be in setup")
	}

	// Player 2 readies up
	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetReady,
		PlayerNumber: 2,
		Data:         map[string]any{"ready": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != StatusActive {
		t.Fatal("expected game to be active after both ready")
	}
	if state.FirstTurnPlayer != 1 {
		t.Fatalf("expected FirstTurnPlayer=1, got %d", state.FirstTurnPlayer)
	}
	if state.CurrentTurn != 1 {
		t.Fatalf("expected CurrentTurn=1, got %d", state.CurrentTurn)
	}
	hasStart := false
	for _, ev := range events {
		if ev.Type == EventGameStart {
			hasStart = true
		}
	}
	if !hasStart {
		t.Fatal("expected game_start event")
	}
	// Both players should gain 1 CP at game start
	if state.Players[0].CP != 1 {
		t.Fatalf("expected player 1 CP=1 at game start, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 1 {
		t.Fatalf("expected player 2 CP=1 at game start, got %d", state.Players[1].CP)
	}
	// Should have 2 CP gain events
	cpEvents := 0
	for _, ev := range events {
		if ev.Type == EventCPGain {
			cpEvents++
		}
	}
	if cpEvents != 2 {
		t.Fatalf("expected 2 CP gain events at game start, got %d", cpEvents)
	}
}

func TestUseStratagem(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 3
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUseStratagem,
		PlayerNumber: 1,
		Data:         map[string]any{"stratagemId": "str1", "stratagemName": "Insane Bravery", "cpCost": 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != EventStratagemUsed {
		t.Fatal("expected stratagem_used event")
	}
	if state.Players[0].CP != 1 {
		t.Fatalf("expected 1 CP remaining, got %d", state.Players[0].CP)
	}
}

func TestUseStratagem_InsufficientCP(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 0
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUseStratagem,
		PlayerNumber: 1,
		Data:         map[string]any{"stratagemId": "str1", "stratagemName": "S", "cpCost": 1},
	})
	if err == nil {
		t.Fatal("expected error for insufficient CP")
	}
}

// --- Turn Structure Tests ---

func TestFullBattleRound_TwoPlayerTurns(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	// Advance player 1 through all 5 phases
	for _, expectedNext := range []Phase{PhaseMovement, PhaseShooting, PhaseCharge, PhaseFight} {
		_, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1})
		if err != nil {
			t.Fatal(err)
		}
		if state.CurrentPhase != expectedNext {
			t.Fatalf("expected %s, got %s", expectedNext, state.CurrentPhase)
		}
		if state.ActivePlayer != 1 {
			t.Fatal("player 1 should still be active mid-turn")
		}
	}

	// Player 1 advances past Fight → switches to player 2's turn
	_, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1})
	if err != nil {
		t.Fatal(err)
	}
	if state.ActivePlayer != 2 {
		t.Fatalf("expected player 2 active, got %d", state.ActivePlayer)
	}
	if state.CurrentRound != 1 {
		t.Fatalf("should still be round 1, got %d", state.CurrentRound)
	}
	if state.CurrentTurn != 2 {
		t.Fatalf("expected turn 2, got %d", state.CurrentTurn)
	}
	if state.CurrentPhase != PhaseCommand {
		t.Fatalf("expected command phase, got %s", state.CurrentPhase)
	}

	// Advance player 2 through all 5 phases
	for range 4 {
		_, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 2})
		if err != nil {
			t.Fatal(err)
		}
	}
	if state.CurrentPhase != PhaseFight {
		t.Fatalf("expected fight phase, got %s", state.CurrentPhase)
	}

	// Player 2 advances past Fight → round ends, round 2 begins
	_, err = e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 2})
	if err != nil {
		t.Fatal(err)
	}
	if state.CurrentRound != 2 {
		t.Fatalf("expected round 2, got %d", state.CurrentRound)
	}
	if state.CurrentTurn != 1 {
		t.Fatalf("expected turn 1, got %d", state.CurrentTurn)
	}
	if state.ActivePlayer != 1 {
		t.Fatalf("expected player 1 active, got %d", state.ActivePlayer)
	}
}

func TestFullGame_10PlayerTurns(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	turnCount := 0
	for state.Status == StatusActive {
		activePlayer := state.ActivePlayer
		// Advance through all 5 phases
		for range 5 {
			_, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: activePlayer})
			if err != nil {
				t.Fatalf("round %d turn %d: %v", state.CurrentRound, state.CurrentTurn, err)
			}
		}
		turnCount++
	}

	if turnCount != 10 {
		t.Fatalf("expected 10 player turns, got %d", turnCount)
	}
	if state.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s", state.Status)
	}
}

func TestCPGain_BothPlayersAtEachCommandPhase(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 0
	state.Players[1].CP = 0
	e := NewEngine(state)

	// Player 1: 5 phases → turn 2 command phase, +1 CP each
	for range 5 {
		e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1})
	}
	if state.Players[0].CP != 1 {
		t.Fatalf("expected player 1 CP=1 after turn 2 start, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 1 {
		t.Fatalf("expected player 2 CP=1 after turn 2 start, got %d", state.Players[1].CP)
	}

	// Player 2: 5 phases → round 2 turn 1 command phase, +1 CP each
	for range 5 {
		e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 2})
	}
	if state.Players[0].CP != 2 {
		t.Fatalf("expected player 1 CP=2 after round 2 start, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 2 {
		t.Fatalf("expected player 2 CP=2 after round 2 start, got %d", state.Players[1].CP)
	}
}

func TestCPGain_BothPlayersGainCPOnTurnSwitch(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	state.Players[1].CP = 3
	e := NewEngine(state)

	// Player 1 finishes turn (5 phases) → switches to player 2's command phase
	for range 5 {
		e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1})
	}

	// Both players should gain 1 CP at the start of player 2's command phase
	if state.Players[0].CP != 6 {
		t.Fatalf("expected player 1 CP=6, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 4 {
		t.Fatalf("expected player 2 CP=4, got %d", state.Players[1].CP)
	}
}

func TestSetReady_FirstTurnPlayerDefaultsTo1(t *testing.T) {
	state := newTestState()
	state.FirstTurnPlayer = 0 // not set
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}})
	e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 2, Data: map[string]any{"ready": true}})

	if state.FirstTurnPlayer != 1 {
		t.Fatalf("expected FirstTurnPlayer=1, got %d", state.FirstTurnPlayer)
	}
	if state.ActivePlayer != 1 {
		t.Fatalf("expected ActivePlayer=1, got %d", state.ActivePlayer)
	}
}

func TestSetReady_PresetFirstTurnPlayer(t *testing.T) {
	state := newTestState()
	state.FirstTurnPlayer = 2 // explicitly set to player 2
	e := NewEngine(state)

	e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}})
	e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 2, Data: map[string]any{"ready": true}})

	if state.FirstTurnPlayer != 2 {
		t.Fatalf("expected FirstTurnPlayer=2, got %d", state.FirstTurnPlayer)
	}
	if state.ActivePlayer != 2 {
		t.Fatalf("expected ActivePlayer=2, got %d", state.ActivePlayer)
	}
}

func TestCPGain_AccumulatesAcrossRounds(t *testing.T) {
	state := newTestState()
	e := NewEngine(state)

	// Start game via set_ready (grants 1 CP each for first command phase)
	e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}})
	e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 2, Data: map[string]any{"ready": true}})

	if state.Players[0].CP != 1 || state.Players[1].CP != 1 {
		t.Fatalf("expected 1 CP each after game start, got %d and %d", state.Players[0].CP, state.Players[1].CP)
	}

	// Play through rounds 1-3 (3 full battle rounds = 6 player turns)
	for round := 0; round < 3; round++ {
		// Player 1 turn
		for range 5 {
			e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: state.ActivePlayer})
		}
		// Player 2 turn
		for range 5 {
			e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: state.ActivePlayer})
		}
	}

	// CP gains (1 per command phase, 2 command phases per round):
	// Round 1 turn 1 start (game start): +1 each = 1
	// Round 1 turn 2 start:              +1 each = 2
	// Round 2 turn 1 start:              +1 each = 3
	// Round 2 turn 2 start:              +1 each = 4
	// Round 3 turn 1 start:              +1 each = 5
	// Round 3 turn 2 start:              +1 each = 6
	// Round 4 turn 1 start:              +1 each = 7
	// (We played 3 full rounds, so we're now at start of round 4)
	if state.Players[0].CP != 7 {
		t.Fatalf("expected player 1 CP=7, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 7 {
		t.Fatalf("expected player 2 CP=7, got %d", state.Players[1].CP)
	}
}

// --- CP Gain Cap Tests ---

func TestDiscardSecondary_CPCappedAt1PerRound(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 3
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 0
	state.Players[0].CPGainedThisRound = 0
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
		makeActiveSecondary("s2", "Sec 2"),
	}
	e := NewEngine(state)

	// First discard: should gain 1 CP
	e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	})
	if state.Players[0].CP != 1 {
		t.Fatalf("expected 1 CP after first discard, got %d", state.Players[0].CP)
	}
	if state.Players[0].CPGainedThisRound != 1 {
		t.Fatalf("expected CPGainedThisRound=1, got %d", state.Players[0].CPGainedThisRound)
	}

	// Second discard: should NOT gain CP (cap reached)
	e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s2"},
	})
	if state.Players[0].CP != 1 {
		t.Fatalf("expected still 1 CP after second discard (capped), got %d", state.Players[0].CP)
	}
}

func TestDiscardSecondary_CPCapResetsNextRound(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 2
	state.CurrentTurn = 1
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 0
	state.Players[0].CPGainedThisRound = 1 // already gained 1 this round
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
	}
	state.Players[0].TacticalDeck = makeDeck(5)
	e := NewEngine(state)

	// Discard in round 2 should not gain CP (cap already reached)
	e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	})
	if state.Players[0].CP != 0 {
		t.Fatalf("expected 0 CP (cap hit), got %d", state.Players[0].CP)
	}

	// Play through rest of round 2 to trigger round 3
	// Player 1 finishes their turn (already past some phases, set to fight)
	// Turn switch → +1 CP each (turn 2 command phase)
	state.CurrentPhase = PhaseFight
	e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1})
	// Player 2's full turn → round advance → +1 CP each (round 3 command phase)
	for range 5 {
		e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 2})
	}

	// Now in round 3 — CPGainedThisRound should be reset
	if state.CurrentRound != 3 {
		t.Fatalf("expected round 3, got %d", state.CurrentRound)
	}
	if state.Players[0].CPGainedThisRound != 0 {
		t.Fatalf("expected CPGainedThisRound reset to 0, got %d", state.Players[0].CPGainedThisRound)
	}
}

func TestDiscardSecondary_FreeDiscardDoesNotCountTowardCap(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 3
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 0
	state.Players[0].CPGainedThisRound = 0
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
		makeActiveSecondary("s2", "Sec 2"),
	}
	e := NewEngine(state)

	// Free discard: no CP, should not affect cap
	e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "free": true},
	})
	if state.Players[0].CPGainedThisRound != 0 {
		t.Fatalf("free discard should not affect CPGainedThisRound, got %d", state.Players[0].CPGainedThisRound)
	}

	// Normal discard after free: should still gain CP
	e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s2"},
	})
	if state.Players[0].CP != 1 {
		t.Fatalf("expected 1 CP after normal discard, got %d", state.Players[0].CP)
	}
}

func TestAdjustCP_PositiveSubjectToCap(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 0
	state.Players[0].CPGainedThisRound = 1 // cap already reached
	e := NewEngine(state)

	// Positive adjust should be blocked by cap
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": 1},
	})
	if err == nil {
		t.Fatal("expected error when CP gain cap reached")
	}
	if state.Players[0].CP != 0 {
		t.Fatalf("expected CP unchanged at 0, got %d", state.Players[0].CP)
	}
}

func TestAdjustCP_NegativeNotAffectedByCap(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	state.Players[0].CPGainedThisRound = 1 // cap reached
	e := NewEngine(state)

	// Negative adjust (spending) should always work
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": -2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].CP != 3 {
		t.Fatalf("expected CP=3, got %d", state.Players[0].CP)
	}
}

func TestAdjustCP_PositiveCountsTowardCap(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 0
	state.Players[0].CPGainedThisRound = 0
	e := NewEngine(state)

	// First positive adjust succeeds
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].CPGainedThisRound != 1 {
		t.Fatalf("expected CPGainedThisRound=1, got %d", state.Players[0].CPGainedThisRound)
	}

	// Second positive adjust blocked
	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": 1},
	})
	if err == nil {
		t.Fatal("expected error on second positive adjust")
	}
}

func TestCPGainedThisRound_DoesNotResetOnTurnSwitch(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	state.Players[0].CPGainedThisRound = 1 // already gained bonus CP this round
	state.CurrentPhase = PhaseFight
	e := NewEngine(state)

	// Player 1 finishes turn → switches to player 2's command phase
	e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1})

	// CPGainedThisRound should NOT reset on turn switch (only on round advance)
	if state.Players[0].CPGainedThisRound != 1 {
		t.Fatalf("expected CPGainedThisRound=1 (not reset on turn switch), got %d", state.Players[0].CPGainedThisRound)
	}
}

// --- Rules helpers ---

func TestClampVP(t *testing.T) {
	if ClampVP(-5, 50) != 0 {
		t.Fatal("negative should clamp to 0")
	}
	if ClampVP(60, 50) != 50 {
		t.Fatal("over max should clamp to max")
	}
	if ClampVP(25, 50) != 25 {
		t.Fatal("in-range should stay")
	}
}

func TestNextPhase(t *testing.T) {
	next, ended := NextPhase(PhaseCommand)
	if next != PhaseMovement || ended {
		t.Fatal("command → movement")
	}
	next, ended = NextPhase(PhaseFight)
	if next != PhaseCommand || !ended {
		t.Fatal("fight → command (turn ended)")
	}
}

func TestShouldGainCP(t *testing.T) {
	if !ShouldGainCP(1) {
		t.Fatal("should gain CP in round 1")
	}
	if !ShouldGainCP(2) {
		t.Fatal("should gain CP in round 2")
	}
	if !ShouldGainCP(5) {
		t.Fatal("should gain CP in round 5")
	}
}

// --- Phase Restriction Tests ---

func TestDrawSecondary_RequiresCommandPhase(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseMovement
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].TacticalDeck = makeDeck(5)
	state.Players[0].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err == nil {
		t.Fatal("expected error when drawing outside Command Phase")
	}
}

func TestDrawSecondary_RequiresActivePlayer(t *testing.T) {
	state := newActiveTestState()
	state.ActivePlayer = 1
	state.Players[1].SecondaryMode = "tactical"
	state.Players[1].TacticalDeck = makeDeck(5)
	state.Players[1].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 2,
	})
	if err == nil {
		t.Fatal("expected error when non-active player tries to draw")
	}
}

func TestNewOrders_RequiresCommandPhase(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseShooting
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 2
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	state.Players[0].TacticalDeck = makeDeck(3)
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionNewOrders,
		PlayerNumber: 1,
		Data:         map[string]any{"discardSecondaryId": "s1"},
	})
	if err == nil {
		t.Fatal("expected error when using New Orders outside Command Phase")
	}
}

func TestDrawChallengerCard_RequiresCommandPhase(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseFight
	state.Players[0].VPPrimary = 0
	state.Players[1].VPPrimary = 10
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawChallengerCard,
		PlayerNumber: 1,
		Data:         map[string]any{"challengerCardId": "cc1", "challengerCardName": "Test"},
	})
	if err == nil {
		t.Fatal("expected error when drawing challenger card outside Command Phase")
	}
}

// --- Scoring Options Validation ---

func makeActiveSecondaryWithOptions(id, name string, opts []SecondaryScoringOption) ActiveSecondary {
	return ActiveSecondary{
		ID:             id,
		Name:           name,
		Description:    "desc",
		MaxVP:          20,
		ScoringOptions: opts,
	}
}

func TestAchieveSecondary_ValidScoringOption(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondaryWithOptions("s1", "Behind Enemy Lines", []SecondaryScoringOption{
			{Label: "1 unit in enemy zone", VP: 3},
			{Label: "2+ units in enemy zone", VP: 4},
		}),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 4},
	})
	if err != nil {
		t.Fatalf("expected valid scoring option to be accepted, got: %v", err)
	}
	if state.Players[0].VPSecondary != 4 {
		t.Fatalf("expected 4 VP, got %d", state.Players[0].VPSecondary)
	}
}

func TestAchieveSecondary_InvalidScoringOption(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondaryWithOptions("s1", "Behind Enemy Lines", []SecondaryScoringOption{
			{Label: "1 unit in enemy zone", VP: 3},
			{Label: "2+ units in enemy zone", VP: 4},
		}),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 5},
	})
	if err == nil {
		t.Fatal("expected error for invalid VP score")
	}
}

func TestAchieveSecondary_ModeFiltering(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondaryWithOptions("s1", "Assassination", []SecondaryScoringOption{
			{Label: "W4+ char", VP: 4, Mode: "fixed"},
			{Label: "W<4 char", VP: 3, Mode: "fixed"},
			{Label: "1+ chars destroyed", VP: 5, Mode: "tactical"},
		}),
	}
	e := NewEngine(state)

	// VP=5 should work (tactical option)
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 5},
	})
	if err != nil {
		t.Fatalf("expected tactical VP=5 to be accepted, got: %v", err)
	}
}

func TestAchieveSecondary_ModeFiltering_RejectsWrongMode(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondaryWithOptions("s1", "Assassination", []SecondaryScoringOption{
			{Label: "W4+ char", VP: 4, Mode: "fixed"},
			{Label: "W<4 char", VP: 3, Mode: "fixed"},
			{Label: "1+ chars destroyed", VP: 5, Mode: "tactical"},
		}),
	}
	e := NewEngine(state)

	// VP=4 should be rejected (fixed-only option, player is in tactical mode)
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 4},
	})
	if err == nil {
		t.Fatal("expected error: VP=4 is a fixed-only option but player is in tactical mode")
	}
}

func TestAchieveSecondary_NoOptionsSkipsValidation(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Legacy Secondary"),
	}
	e := NewEngine(state)

	// With no scoring options, any VP should be accepted
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 7},
	})
	if err != nil {
		t.Fatalf("expected no validation when scoring options are empty, got: %v", err)
	}
}
