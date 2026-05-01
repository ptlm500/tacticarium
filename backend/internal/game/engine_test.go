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
			ID:    fmt.Sprintf("sec-%d", i+1),
			Name:  fmt.Sprintf("Secondary %d", i+1),
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

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectPrimaryMission,
		PlayerNumber: 1,
		Data:         map[string]any{"missionId": "m1", "missionName": "M"},
	}); err != nil {
		t.Fatal(err)
	}
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

func TestSelectFirstTurnPlayer(t *testing.T) {
	for _, pn := range []int{1, 2} {
		t.Run(fmt.Sprintf("player_%d", pn), func(t *testing.T) {
			state := newTestState()
			e := NewEngine(state)
			events, err := e.Apply(context.Background(), GameAction{
				Type:         ActionSelectFirstTurnPlayer,
				PlayerNumber: 1,
				Data:         map[string]any{"firstTurnPlayer": pn},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 1 || events[0].Type != EventFirstTurnPlayerSelected {
				t.Fatal("expected first_turn_player_selected event")
			}
			if state.FirstTurnPlayer != pn {
				t.Fatalf("expected FirstTurnPlayer=%d, got %d", pn, state.FirstTurnPlayer)
			}
		})
	}
}

func TestSelectFirstTurnPlayer_RejectsInvalid(t *testing.T) {
	for _, pn := range []int{0, 3, -1} {
		t.Run(fmt.Sprintf("value_%d", pn), func(t *testing.T) {
			e := NewEngine(newTestState())
			_, err := e.Apply(context.Background(), GameAction{
				Type:         ActionSelectFirstTurnPlayer,
				PlayerNumber: 1,
				Data:         map[string]any{"firstTurnPlayer": pn},
			})
			if err == nil {
				t.Fatalf("expected error for playerNumber=%d", pn)
			}
		})
	}
}

func TestSelectFirstTurnPlayer_RequiresSetup(t *testing.T) {
	e := NewEngine(newActiveTestState())
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectFirstTurnPlayer,
		PlayerNumber: 1,
		Data:         map[string]any{"firstTurnPlayer": 2},
	})
	if err == nil {
		t.Fatal("expected error when not in setup")
	}
}

func TestSelectFirstTurnPlayer_ResetsReadiness(t *testing.T) {
	state := newTestState()
	state.FirstTurnPlayer = 1
	state.Players[0].Ready = true
	state.Players[1].Ready = true
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectFirstTurnPlayer,
		PlayerNumber: 1,
		Data:         map[string]any{"firstTurnPlayer": 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].Ready || state.Players[1].Ready {
		t.Fatal("expected readiness to be reset after first turn player change")
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

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSelectSecondaryMode,
		PlayerNumber: 1,
		Data:         map[string]any{"mode": "tactical"},
	}); err != nil {
		t.Fatal(err)
	}
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

func TestInitTacticalDeck_PreservesScoringTiming(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "tactical"
	e := NewEngine(state)

	ownTurn := makeSecondary("s1", "S1")
	oppTurn := makeSecondary("s2", "S2")
	oppTurn["scoringTiming"] = "end_of_opponent_turn"

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionInitTacticalDeck,
		PlayerNumber: 1,
		Data: map[string]any{
			"deck": []any{ownTurn, oppTurn},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	deck := state.Players[0].TacticalDeck
	if len(deck) != 2 {
		t.Fatalf("expected deck of 2, got %d", len(deck))
	}
	if deck[0].ScoringTiming != "" {
		t.Errorf("card 1 should default to empty timing (= own turn), got %q", deck[0].ScoringTiming)
	}
	if deck[1].ScoringTiming != "end_of_opponent_turn" {
		t.Errorf("card 2 should carry end_of_opponent_turn, got %q", deck[1].ScoringTiming)
	}
}

func TestSetFixedSecondaries_PreservesScoringTiming(t *testing.T) {
	state := newTestState()
	state.Players[0].SecondaryMode = "fixed"
	e := NewEngine(state)

	a := makeSecondary("a", "A")
	a["isFixed"] = true
	a["scoringTiming"] = "end_of_opponent_turn"
	b := makeSecondary("b", "B")
	b["isFixed"] = true

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetFixedSecondaries,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaries": []any{a, b}},
	})
	if err != nil {
		t.Fatal(err)
	}
	active := state.Players[0].ActiveSecondaries
	if len(active) != 2 {
		t.Fatalf("expected 2 active secondaries, got %d", len(active))
	}
	if active[0].ScoringTiming != "end_of_opponent_turn" {
		t.Errorf("card a should carry end_of_opponent_turn, got %q", active[0].ScoringTiming)
	}
	if active[1].ScoringTiming != "" {
		t.Errorf("card b should default to empty timing, got %q", active[1].ScoringTiming)
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
	if got := state.Players[0].AchievedSecondaries[0].VPScored; got != 5 {
		t.Fatalf("expected achieved card VPScored=5, got %d", got)
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

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAchieveSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "vpScored": 5},
	}); err != nil {
		t.Fatal(err)
	}
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

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "free": true},
	}); err != nil {
		t.Fatal(err)
	}
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

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	}); err != nil {
		t.Fatal(err)
	}
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

func TestNewOrders_OncePerPhase(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 5
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("s1", "Sec 1"),
		makeActiveSecondary("s2", "Sec 2"),
	}
	state.Players[0].TacticalDeck = makeDeck(5)
	e := NewEngine(state)

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionNewOrders,
		PlayerNumber: 1,
		Data:         map[string]any{"discardSecondaryId": "s1"},
	}); err != nil {
		t.Fatalf("first new_orders: %v", err)
	}
	if !state.Players[0].NewOrdersUsedThisPhase {
		t.Fatal("expected NewOrdersUsedThisPhase to be true after use")
	}

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionNewOrders,
		PlayerNumber: 1,
		Data:         map[string]any{"discardSecondaryId": "s2"},
	})
	if err == nil {
		t.Fatal("expected second new_orders in same phase to be rejected")
	}
}

func TestAdvancePhase_ResetsNewOrdersUsedThisPhase(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].NewOrdersUsedThisPhase = true
	state.Players[1].NewOrdersUsedThisPhase = true
	e := NewEngine(state)

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdvancePhase,
		PlayerNumber: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if state.Players[0].NewOrdersUsedThisPhase || state.Players[1].NewOrdersUsedThisPhase {
		t.Fatal("expected NewOrdersUsedThisPhase to reset on advance_phase")
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

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreChallenger,
		PlayerNumber: 1,
		Data:         map[string]any{"vpScored": 5},
	}); err != nil {
		t.Fatal(err)
	}
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

// --- Draw Restrictions ---

func makeRestrictedSecondary(id, name string, round int, mode string) ActiveSecondary {
	s := makeActiveSecondary(id, name)
	s.DrawRestriction = &SecondaryDrawRestriction{Round: round, Mode: mode}
	return s
}

func TestDrawSecondary_MandatoryRestriction_Round1_Reshuffles(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "tactical"
	// Deck: [restricted, plain1, plain2]. On draw, restricted should be reshuffled
	// (mandatory round 1), then plain1 drawn, then plain2 drawn.
	state.Players[0].TacticalDeck = []ActiveSecondary{
		makeRestrictedSecondary("restricted", "Restricted", 1, DrawRestrictionMandatory),
		makeActiveSecondary("plain1", "P1"),
		makeActiveSecondary("plain2", "P2"),
	}
	state.Players[0].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	active := state.Players[0].ActiveSecondaries
	if len(active) != 2 {
		t.Fatalf("expected 2 active, got %d", len(active))
	}
	for _, s := range active {
		if s.ID == "restricted" {
			t.Fatal("restricted card should not be in active in round 1")
		}
	}
	// Deck should still contain the restricted card (shuffled back in).
	foundInDeck := false
	for _, s := range state.Players[0].TacticalDeck {
		if s.ID == "restricted" {
			foundInDeck = true
		}
	}
	if !foundInDeck {
		t.Fatal("expected restricted card to be reshuffled into deck")
	}

	// Expect at least one reshuffle event. Random insertion means the
	// restricted card can land back at the top and trigger additional
	// reshuffles before a drawable card comes up — that's valid behavior.
	reshuffles := 0
	for _, ev := range events {
		if ev.Type == EventSecondaryReshuffled {
			reshuffles++
		}
	}
	if reshuffles < 1 {
		t.Fatalf("expected at least 1 reshuffle event, got %d", reshuffles)
	}
}

func TestDrawSecondary_MandatoryRestriction_NotRound1_DrawsNormally(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 2
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].TacticalDeck = []ActiveSecondary{
		makeRestrictedSecondary("restricted", "R", 1, DrawRestrictionMandatory),
		makeActiveSecondary("plain1", "P1"),
	}
	state.Players[0].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Players[0].ActiveSecondaries) != 2 {
		t.Fatalf("expected 2 active, got %d", len(state.Players[0].ActiveSecondaries))
	}
	// Restriction does not trigger in round 2, so the "restricted" card
	// should be drawn as the first active.
	if state.Players[0].ActiveSecondaries[0].ID != "restricted" {
		t.Fatalf("expected restricted card to be drawn normally in round 2")
	}
}

func TestDrawSecondary_AllRestrictedDeck_BailsCleanly(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "tactical"
	// Deck of only mandatory-restricted cards — nothing drawable. The helper
	// should bail without infinite-looping, leaving cards in the deck.
	state.Players[0].TacticalDeck = []ActiveSecondary{
		makeRestrictedSecondary("r1", "R1", 1, DrawRestrictionMandatory),
		makeRestrictedSecondary("r2", "R2", 1, DrawRestrictionMandatory),
	}
	state.Players[0].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Players[0].ActiveSecondaries) != 0 {
		t.Fatalf("expected 0 active, got %d", len(state.Players[0].ActiveSecondaries))
	}
	if len(state.Players[0].TacticalDeck) != 2 {
		t.Fatalf("expected deck to retain 2 cards, got %d", len(state.Players[0].TacticalDeck))
	}
}

func TestDrawSecondary_MixedRestrictedDeck_DrawsNonRestricted(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "tactical"
	// One restricted + one plain. Regardless of random insertion position,
	// the non-restricted card must end up drawn.
	state.Players[0].TacticalDeck = []ActiveSecondary{
		makeRestrictedSecondary("r1", "R1", 1, DrawRestrictionMandatory),
		makeActiveSecondary("ok", "OK"),
	}
	state.Players[0].ActiveSecondaries = []ActiveSecondary{}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDrawSecondary,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range state.Players[0].ActiveSecondaries {
		if s.ID == "ok" {
			found = true
		}
		if s.ID == "r1" {
			t.Fatal("restricted card should not be active in round 1")
		}
	}
	if !found {
		t.Fatal("expected non-restricted card to be drawn")
	}
}

func TestReshuffleSecondary_OptionalRound1_Succeeds(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeRestrictedSecondary("opt", "Optional", 1, DrawRestrictionOptional),
	}
	// Use a larger deck so the reshuffled "opt" card is very unlikely to land
	// back on top and get immediately redrawn — we want to assert the common
	// case (drew a different card) while keeping the random shuffle real.
	state.Players[0].TacticalDeck = []ActiveSecondary{
		makeActiveSecondary("next1", "N1"),
		makeActiveSecondary("next2", "N2"),
		makeActiveSecondary("next3", "N3"),
		makeActiveSecondary("next4", "N4"),
		makeActiveSecondary("next5", "N5"),
	}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionReshuffleSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "opt"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Invariants: total card count preserved, "opt" is somewhere in the
	// deck+active set, and we still have exactly 1 active card (the drawn
	// replacement — which may be "opt" itself if the shuffle landed it on
	// top, but is overwhelmingly likely to be one of the other cards).
	active := state.Players[0].ActiveSecondaries
	deck := state.Players[0].TacticalDeck
	if len(active) != 1 {
		t.Fatalf("expected 1 active, got %d", len(active))
	}
	if len(active)+len(deck) != 6 {
		t.Fatalf("expected 6 total cards, got %d", len(active)+len(deck))
	}
	foundOpt := false
	for _, s := range deck {
		if s.ID == "opt" {
			foundOpt = true
		}
	}
	if !foundOpt && active[0].ID != "opt" {
		t.Fatal("expected opt card to be in deck or re-drawn into active")
	}

	// Should emit reshuffle + draw events.
	var sawReshuffle, sawDraw bool
	for _, ev := range events {
		if ev.Type == EventSecondaryReshuffled {
			sawReshuffle = true
		}
		if ev.Type == EventSecondaryDrawn {
			sawDraw = true
		}
	}
	if !sawReshuffle || !sawDraw {
		t.Fatalf("expected reshuffle+draw events, got %+v", events)
	}
}

func TestReshuffleSecondary_MandatoryRestriction_Rejected(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeRestrictedSecondary("m", "Mandatory", 1, DrawRestrictionMandatory),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionReshuffleSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "m"},
	})
	if err == nil {
		t.Fatal("expected error: cannot manually reshuffle a mandatory-restriction card")
	}
}

func TestReshuffleSecondary_WrongRound_Rejected(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 2
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeRestrictedSecondary("opt", "Optional", 1, DrawRestrictionOptional),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionReshuffleSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "opt"},
	})
	if err == nil {
		t.Fatal("expected error: restriction round has passed")
	}
}

func TestReshuffleSecondary_NoRestriction_Rejected(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("plain", "P")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionReshuffleSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "plain"},
	})
	if err == nil {
		t.Fatal("expected error: card has no draw restriction")
	}
}

func TestReshuffleSecondary_FixedMode_Rejected(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "fixed"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeRestrictedSecondary("opt", "Optional", 1, DrawRestrictionOptional),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionReshuffleSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "opt"},
	})
	if err == nil {
		t.Fatal("expected error: fixed mode cannot reshuffle")
	}
}

func TestNewOrders_AppliesMandatoryReshuffle(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].CP = 1
	state.Players[0].ActiveSecondaries = []ActiveSecondary{
		makeActiveSecondary("current", "C"),
	}
	state.Players[0].TacticalDeck = []ActiveSecondary{
		makeRestrictedSecondary("restricted", "R", 1, DrawRestrictionMandatory),
		makeActiveSecondary("plain", "P"),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionNewOrders,
		PlayerNumber: 1,
		Data:         map[string]any{"discardSecondaryId": "current"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Replacement should be "plain", not "restricted".
	active := state.Players[0].ActiveSecondaries
	if len(active) != 1 {
		t.Fatalf("expected 1 active, got %d", len(active))
	}
	if active[0].ID == "restricted" {
		t.Fatal("mandatory-restricted card should have been reshuffled during new_orders draw")
	}
}

func TestAdaptOrDie_Tactical_AppliesMandatoryReshuffleOnDraw(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.TwistID = TwistAdaptOrDie
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	state.Players[0].TacticalDeck = []ActiveSecondary{
		makeRestrictedSecondary("restricted", "R", 1, DrawRestrictionMandatory),
		makeActiveSecondary("plain", "P"),
	}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdaptOrDie,
		PlayerNumber: 1,
		Data:         map[string]any{"shuffleBackSecondaryId": "s1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Drawn should be "plain"; "restricted" goes back to deck (+s1 shuffled back = 2 in deck).
	active := state.Players[0].ActiveSecondaries
	if len(active) != 1 {
		t.Fatalf("expected 1 active, got %d", len(active))
	}
	if active[0].ID == "restricted" {
		t.Fatal("restricted card should not be active after mandatory reshuffle")
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

func TestRevertPhase_WithinTurn(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseMovement
	state.Players[0].CP = 3
	state.Players[1].CP = 4
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionRevertPhase,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.CurrentPhase != PhaseCommand {
		t.Fatalf("expected command phase, got %s", state.CurrentPhase)
	}
	if len(events) != 1 || events[0].Type != EventPhaseRevert {
		t.Fatalf("expected a single phase_revert event, got %+v", events)
	}
	if state.Players[0].CP != 3 || state.Players[1].CP != 4 {
		t.Fatal("within-turn revert must not touch CP")
	}
}

func TestRevertPhase_TurnBoundary_T2CommandBackToT1Fight(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.CurrentTurn = 2
	state.CurrentPhase = PhaseCommand
	state.ActivePlayer = 2
	state.FirstTurnPlayer = 1
	// Both players just gained 1 CP entering this Command phase.
	state.Players[0].CP = 2
	state.Players[1].CP = 2
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionRevertPhase,
		PlayerNumber: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.CurrentRound != 1 || state.CurrentTurn != 1 || state.CurrentPhase != PhaseFight {
		t.Fatalf("expected round=1 turn=1 fight, got round=%d turn=%d phase=%s",
			state.CurrentRound, state.CurrentTurn, state.CurrentPhase)
	}
	if state.ActivePlayer != 1 {
		t.Fatalf("expected active player 1, got %d", state.ActivePlayer)
	}
	if state.Players[0].CP != 1 || state.Players[1].CP != 1 {
		t.Fatalf("expected both players CP=1 after revoking auto-CP, got %d/%d",
			state.Players[0].CP, state.Players[1].CP)
	}

	if events[0].Type != EventPhaseRevert {
		t.Fatalf("expected first event to be phase_revert, got %s", events[0].Type)
	}
	cpAdjustCount := 0
	for _, ev := range events {
		if ev.Type == EventCPAdjust {
			cpAdjustCount++
			if ev.Data["reason"] != "phase_revert" {
				t.Fatalf("expected reason=phase_revert on cp_adjust event, got %v", ev.Data["reason"])
			}
			if ev.Data["delta"] != -1 {
				t.Fatalf("expected delta=-1, got %v", ev.Data["delta"])
			}
		}
	}
	if cpAdjustCount != 2 {
		t.Fatalf("expected 2 cp_adjust events, got %d", cpAdjustCount)
	}
}

func TestRevertPhase_RoundBoundary_R2T1CommandBackToR1T2Fight(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 2
	state.CurrentTurn = 1
	state.CurrentPhase = PhaseCommand
	state.ActivePlayer = 1
	state.FirstTurnPlayer = 1
	state.Players[0].CP = 3
	state.Players[1].CP = 5
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionRevertPhase,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.CurrentRound != 1 || state.CurrentTurn != 2 || state.CurrentPhase != PhaseFight {
		t.Fatalf("expected round=1 turn=2 fight, got round=%d turn=%d phase=%s",
			state.CurrentRound, state.CurrentTurn, state.CurrentPhase)
	}
	if state.ActivePlayer != 2 {
		t.Fatalf("expected active player 2 (non-first-turn player), got %d", state.ActivePlayer)
	}
	if state.Players[0].CP != 2 || state.Players[1].CP != 4 {
		t.Fatalf("expected CP=2/4, got %d/%d", state.Players[0].CP, state.Players[1].CP)
	}
}

func TestRevertPhase_ClampsCPToZero(t *testing.T) {
	state := newActiveTestState()
	state.CurrentRound = 1
	state.CurrentTurn = 2
	state.CurrentPhase = PhaseCommand
	state.ActivePlayer = 2
	state.FirstTurnPlayer = 1
	state.Players[0].CP = 0
	state.Players[1].CP = 0
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionRevertPhase,
		PlayerNumber: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].CP != 0 || state.Players[1].CP != 0 {
		t.Fatalf("expected CP clamped to 0, got %d/%d", state.Players[0].CP, state.Players[1].CP)
	}
	// A cp_adjust event with delta=0 should still be emitted (bookkeeping).
	var p1Delta any
	for _, ev := range events {
		if ev.Type == EventCPAdjust && ev.PlayerNumber == 1 {
			p1Delta = ev.Data["delta"]
		}
	}
	if p1Delta != 0 {
		t.Fatalf("expected delta=0 when clamped, got %v", p1Delta)
	}
}

func TestRevertPhase_BlockedAtGameStart(t *testing.T) {
	state := newActiveTestState() // round=1 turn=1 phase=command
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionRevertPhase,
		PlayerNumber: 1,
	})
	if err == nil {
		t.Fatal("expected error reverting before game start")
	}
}

func TestRevertPhase_OnlyActivePlayer(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseMovement
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionRevertPhase,
		PlayerNumber: 2,
	})
	if err == nil {
		t.Fatal("expected error for non-active player")
	}
}

func TestRevertPhase_ClearsStratagemsUsedThisPhase(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseShooting
	state.Players[0].StratagemsUsedThisPhase = []string{"strat-1", "strat-2"}
	state.Players[1].StratagemsUsedThisPhase = []string{"strat-3"}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionRevertPhase,
		PlayerNumber: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Players[0].StratagemsUsedThisPhase) != 0 || len(state.Players[1].StratagemsUsedThisPhase) != 0 {
		t.Fatal("expected stratagems-used-this-phase to be cleared on revert")
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
		data     map[string]any
		checkFn  func() int
	}{
		{"primary", 10, map[string]any{"category": "primary", "delta": 10, "scoringSlot": ScoringSlotEndOfCommandPhase}, func() int { return state.Players[0].VPPrimary }},
		{"secondary", 5, map[string]any{"category": "secondary", "delta": 5}, func() int { return state.Players[0].VPSecondary }},
		{"gambit", 3, map[string]any{"category": "gambit", "delta": 3}, func() int { return state.Players[0].VPGambit }},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			_, err := e.Apply(context.Background(), GameAction{
				Type:         ActionScoreVP,
				PlayerNumber: 1,
				Data:         tt.data,
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

func TestScoreVP_Primary_RequiresValidSlot(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	// Missing slot
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5},
	})
	if err == nil {
		t.Fatal("expected error when scoringSlot is missing")
	}

	// Invalid slot
	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5, "scoringSlot": "bogus"},
	})
	if err == nil {
		t.Fatal("expected error for invalid scoringSlot")
	}
}

func TestScoreVP_Primary_RejectsDuplicateSlotSameRound(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err == nil {
		t.Fatal("expected duplicate-slot error")
	}

	if state.Players[0].VPPrimary != 5 {
		t.Fatalf("expected VPPrimary=5 (second score rejected), got %d", state.Players[0].VPPrimary)
	}
}

func TestScoreVP_Primary_AllowsDifferentSlotsSameRound(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	for _, slot := range []string{ScoringSlotEndOfCommandPhase, ScoringSlotEndOfTurn, ScoringSlotEndOfBattleRound} {
		_, err := e.Apply(context.Background(), GameAction{
			Type:         ActionScoreVP,
			PlayerNumber: 1,
			Data:         map[string]any{"category": "primary", "delta": 3, "scoringSlot": slot},
		})
		if err != nil {
			t.Fatalf("unexpected error for slot %s: %v", slot, err)
		}
	}

	if state.Players[0].VPPrimary != 9 {
		t.Fatalf("expected VPPrimary=9, got %d", state.Players[0].VPPrimary)
	}
}

func TestScoreVP_Primary_SlotTrackedPerRound(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Simulate round advance
	state.CurrentRound = 2

	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatalf("expected same slot to be reusable in a new round, got: %v", err)
	}

	if state.Players[0].VPPrimary != 10 {
		t.Fatalf("expected VPPrimary=10, got %d", state.Players[0].VPPrimary)
	}
}

func TestScoreVP_Primary_AppliedDeltaOnClamp(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].VPPrimary = 48
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 10, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}

	if state.Players[0].VPPrimary != MaxVPPrimary {
		t.Fatalf("expected VPPrimary clamped to %d, got %d", MaxVPPrimary, state.Players[0].VPPrimary)
	}
	applied, _ := events[0].Data["appliedDelta"].(int)
	if applied != 2 {
		t.Fatalf("expected appliedDelta=2 after clamp, got %v", events[0].Data["appliedDelta"])
	}
}

func TestScoreVP_Primary_IncludesScoringRuleLabel(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data: map[string]any{
			"category":         "primary",
			"delta":            5,
			"scoringSlot":      ScoringSlotEndOfCommandPhase,
			"scoringRuleLabel": "Hold the most",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	label, _ := events[0].Data["scoringRuleLabel"].(string)
	if label != "Hold the most" {
		t.Fatalf("expected scoringRuleLabel=%q in event data, got %v", "Hold the most", events[0].Data["scoringRuleLabel"])
	}
}

func TestScoreVP_Primary_OmitsScoringRuleLabelWhenAbsent(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := events[0].Data["scoringRuleLabel"]; ok {
		t.Fatalf("expected no scoringRuleLabel key when not provided, got %v", events[0].Data["scoringRuleLabel"])
	}
}

func TestUndoPrimaryScore_FreesSlotAndReverses(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 7, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUndoPrimaryScore,
		PlayerNumber: 1,
		Data:         map[string]any{"round": 1, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}

	if state.Players[0].VPPrimary != 0 {
		t.Fatalf("expected VPPrimary=0 after undo, got %d", state.Players[0].VPPrimary)
	}
	if events[0].Type != EventVPPrimaryScoreReverted {
		t.Fatalf("expected %s event, got %s", EventVPPrimaryScoreReverted, events[0].Type)
	}
	if _, exists := state.Players[0].VPPrimaryScoredSlots[1][ScoringSlotEndOfCommandPhase]; exists {
		t.Fatal("expected slot to be freed after undo")
	}

	// Can re-score the slot after undo
	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 4, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatalf("expected slot to be reusable after undo, got: %v", err)
	}
	if state.Players[0].VPPrimary != 4 {
		t.Fatalf("expected VPPrimary=4, got %d", state.Players[0].VPPrimary)
	}
}

func TestUndoPrimaryScore_PriorRound(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 6, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}

	state.CurrentRound = 3

	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionUndoPrimaryScore,
		PlayerNumber: 1,
		Data:         map[string]any{"round": 1, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatalf("expected undo of prior round to succeed, got: %v", err)
	}
	if state.Players[0].VPPrimary != 0 {
		t.Fatalf("expected VPPrimary=0 after undo, got %d", state.Players[0].VPPrimary)
	}
}

func TestUndoPrimaryScore_NoMatch(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUndoPrimaryScore,
		PlayerNumber: 1,
		Data:         map[string]any{"round": 1, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err == nil {
		t.Fatal("expected error when no matching score exists")
	}
}

func TestAdjustVPManual_BypassesSlots(t *testing.T) {
	state := newActiveTestState()
	e := NewEngine(state)

	// Use the rule-based slot
	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionScoreVP,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 5, "scoringSlot": ScoringSlotEndOfCommandPhase},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Manual adjustment should work regardless of slot usage
	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustVPManual,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 3},
	})
	if err != nil {
		t.Fatal(err)
	}

	if state.Players[0].VPPrimary != 8 {
		t.Fatalf("expected VPPrimary=8, got %d", state.Players[0].VPPrimary)
	}
	if events[0].Type != EventVPManualAdjust {
		t.Fatalf("expected %s event, got %s", EventVPManualAdjust, events[0].Type)
	}
	applied, _ := events[0].Data["appliedDelta"].(int)
	if applied != 3 {
		t.Fatalf("expected appliedDelta=3, got %v", events[0].Data["appliedDelta"])
	}
}

func TestAdjustVPManual_ClampsToMax(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].VPPrimary = 48
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustVPManual,
		PlayerNumber: 1,
		Data:         map[string]any{"category": "primary", "delta": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].VPPrimary != MaxVPPrimary {
		t.Fatalf("expected clamp to %d, got %d", MaxVPPrimary, state.Players[0].VPPrimary)
	}
	applied, _ := events[0].Data["appliedDelta"].(int)
	if applied != 2 {
		t.Fatalf("expected appliedDelta=2, got %v", events[0].Data["appliedDelta"])
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
	state.FirstTurnPlayer = 1 // must be selected before ready-up
	state.Players[0].Ready = false
	state.Players[1].Ready = false
	e := NewEngine(state)

	// Player 1 readies up
	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetReady,
		PlayerNumber: 1,
		Data:         map[string]any{"ready": true},
	}); err != nil {
		t.Fatal(err)
	}
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

func TestUseStratagem_LookupResolvesOriginalCost(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 3
	e := NewEngine(state)
	e.SetStratagemLookup(func(id string) (*StratagemInfo, error) {
		if id != "str1" {
			t.Fatalf("unexpected stratagem id: %s", id)
		}
		return &StratagemInfo{Name: "Canonical Name", CPCost: 2}, nil
	})

	// Client sends only the spent amount (override = 0, i.e. free).
	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUseStratagem,
		PlayerNumber: 1,
		Data:         map[string]any{"stratagemId": "str1", "stratagemName": "forged", "cpCost": 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	d := events[0].Data
	if d["stratagemName"] != "Canonical Name" {
		t.Errorf("expected server-sourced name, got %v", d["stratagemName"])
	}
	if d["originalCpCost"] != 2 {
		t.Errorf("expected originalCpCost=2, got %v", d["originalCpCost"])
	}
	if d["cpSpent"] != 0 {
		t.Errorf("expected cpSpent=0 (overridden), got %v", d["cpSpent"])
	}
	if state.Players[0].CP != 3 {
		t.Errorf("expected CP unchanged (free override), got %d", state.Players[0].CP)
	}
}

func TestUseStratagem_LookupAllowsRaisedCost(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	e := NewEngine(state)
	e.SetStratagemLookup(func(id string) (*StratagemInfo, error) {
		return &StratagemInfo{Name: "S", CPCost: 1}, nil
	})

	// Override raises cost above the default — some rules make stratagems pricier.
	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUseStratagem,
		PlayerNumber: 1,
		Data:         map[string]any{"stratagemId": "str1", "cpCost": 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	d := events[0].Data
	if d["cpSpent"] != 3 || d["originalCpCost"] != 1 {
		t.Errorf("expected cpSpent=3/original=1, got spent=%v original=%v", d["cpSpent"], d["originalCpCost"])
	}
	if state.Players[0].CP != 2 {
		t.Errorf("expected 2 CP remaining, got %d", state.Players[0].CP)
	}
}

func TestUseStratagem_TracksUsedThisPhase(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUseStratagem,
		PlayerNumber: 1,
		Data:         map[string]any{"stratagemId": "str-rr", "stratagemName": "Command Re-Roll", "cpCost": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := state.Players[0].StratagemsUsedThisPhase; len(got) != 1 || got[0] != "str-rr" {
		t.Fatalf("expected [str-rr], got %v", got)
	}
	if len(state.Players[1].StratagemsUsedThisPhase) != 0 {
		t.Errorf("opponent list should be untouched")
	}
}

func TestUseStratagem_RepeatUseAllowedAndNotDuplicated(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	e := NewEngine(state)

	for i := 0; i < 2; i++ {
		_, err := e.Apply(context.Background(), GameAction{
			Type:         ActionUseStratagem,
			PlayerNumber: 1,
			Data:         map[string]any{"stratagemId": "str-rr", "cpCost": 1},
		})
		if err != nil {
			t.Fatalf("use %d failed: %v", i+1, err)
		}
	}
	if state.Players[0].CP != 3 {
		t.Errorf("expected 3 CP remaining after two 1-CP uses, got %d", state.Players[0].CP)
	}
	got := state.Players[0].StratagemsUsedThisPhase
	if len(got) != 1 || got[0] != "str-rr" {
		t.Fatalf("expected single entry [str-rr], got %v", got)
	}
}

func TestUseStratagem_ClearedOnPhaseAdvance(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 3
	state.Players[1].CP = 3
	e := NewEngine(state)

	for _, pn := range []int{1, 2} {
		if _, err := e.Apply(context.Background(), GameAction{
			Type:         ActionUseStratagem,
			PlayerNumber: pn,
			Data:         map[string]any{"stratagemId": "str-rr", "cpCost": 1},
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Advance within the same turn (Command -> Movement)
	if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1}); err != nil {
		t.Fatal(err)
	}

	for i, p := range state.Players {
		if len(p.StratagemsUsedThisPhase) != 0 {
			t.Errorf("player %d list not cleared, got %v", i+1, p.StratagemsUsedThisPhase)
		}
	}
}

func TestUseStratagem_ClearedOnTurnEndPhaseAdvance(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseFight
	state.Players[0].CP = 3
	e := NewEngine(state)

	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUseStratagem,
		PlayerNumber: 1,
		Data:         map[string]any{"stratagemId": "str-rr", "cpCost": 1},
	}); err != nil {
		t.Fatal(err)
	}

	// Advance from Fight — ends player 1's turn, resets to Command phase for player 2.
	if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1}); err != nil {
		t.Fatal(err)
	}

	if len(state.Players[0].StratagemsUsedThisPhase) != 0 {
		t.Errorf("list not cleared on turn-end phase advance, got %v", state.Players[0].StratagemsUsedThisPhase)
	}
}

func TestUseStratagem_LookupUnknownID(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 3
	e := NewEngine(state)
	e.SetStratagemLookup(func(id string) (*StratagemInfo, error) {
		return nil, fmt.Errorf("not found")
	})

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionUseStratagem,
		PlayerNumber: 1,
		Data:         map[string]any{"stratagemId": "nope", "cpCost": 1},
	})
	if err == nil {
		t.Fatal("expected error when lookup fails")
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
		if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1}); err != nil {
			t.Fatal(err)
		}
	}
	if state.Players[0].CP != 1 {
		t.Fatalf("expected player 1 CP=1 after turn 2 start, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 1 {
		t.Fatalf("expected player 2 CP=1 after turn 2 start, got %d", state.Players[1].CP)
	}

	// Player 2: 5 phases → round 2 turn 1 command phase, +1 CP each
	for range 5 {
		if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 2}); err != nil {
			t.Fatal(err)
		}
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
		if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1}); err != nil {
			t.Fatal(err)
		}
	}

	// Both players should gain 1 CP at the start of player 2's command phase
	if state.Players[0].CP != 6 {
		t.Fatalf("expected player 1 CP=6, got %d", state.Players[0].CP)
	}
	if state.Players[1].CP != 4 {
		t.Fatalf("expected player 2 CP=4, got %d", state.Players[1].CP)
	}
}

func TestSetReady_RejectedWhenFirstTurnPlayerUnset(t *testing.T) {
	state := newTestState()
	state.FirstTurnPlayer = 0 // not selected yet
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetReady,
		PlayerNumber: 1,
		Data:         map[string]any{"ready": true},
	})
	if err == nil {
		t.Fatal("expected error when readying up with no first turn player set")
	}
	if state.Players[0].Ready {
		t.Fatal("player 1 should not be marked ready when action failed")
	}
	if state.Status != StatusSetup {
		t.Fatalf("expected status to remain setup, got %s", state.Status)
	}
}

func TestSetReady_UnreadyingAllowedWithoutFirstTurnPlayer(t *testing.T) {
	// Setting ready=false should always work, even if first turn player is unset.
	state := newTestState()
	state.FirstTurnPlayer = 0
	state.Players[0].Ready = true
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetReady,
		PlayerNumber: 1,
		Data:         map[string]any{"ready": false},
	})
	if err != nil {
		t.Fatalf("unexpected error unreadying: %v", err)
	}
	if state.Players[0].Ready {
		t.Fatal("player 1 should be unready")
	}
}

func TestSetReady_PresetFirstTurnPlayer(t *testing.T) {
	state := newTestState()
	state.FirstTurnPlayer = 2 // explicitly set to player 2
	e := NewEngine(state)

	if _, err := e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 2, Data: map[string]any{"ready": true}}); err != nil {
		t.Fatal(err)
	}

	if state.FirstTurnPlayer != 2 {
		t.Fatalf("expected FirstTurnPlayer=2, got %d", state.FirstTurnPlayer)
	}
	if state.ActivePlayer != 2 {
		t.Fatalf("expected ActivePlayer=2, got %d", state.ActivePlayer)
	}
}

func TestCPGain_AccumulatesAcrossRounds(t *testing.T) {
	state := newTestState()
	state.FirstTurnPlayer = 1 // must be selected before ready-up
	e := NewEngine(state)

	// Start game via set_ready (grants 1 CP each for first command phase)
	if _, err := e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 2, Data: map[string]any{"ready": true}}); err != nil {
		t.Fatal(err)
	}

	if state.Players[0].CP != 1 || state.Players[1].CP != 1 {
		t.Fatalf("expected 1 CP each after game start, got %d and %d", state.Players[0].CP, state.Players[1].CP)
	}

	// Play through rounds 1-3 (3 full battle rounds = 6 player turns)
	for round := 0; round < 3; round++ {
		// Player 1 turn
		for range 5 {
			if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: state.ActivePlayer}); err != nil {
				t.Fatal(err)
			}
		}
		// Player 2 turn
		for range 5 {
			if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: state.ActivePlayer}); err != nil {
				t.Fatal(err)
			}
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
	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	}); err != nil {
		t.Fatal(err)
	}
	if state.Players[0].CP != 1 {
		t.Fatalf("expected 1 CP after first discard, got %d", state.Players[0].CP)
	}
	if state.Players[0].CPGainedThisRound != 1 {
		t.Fatalf("expected CPGainedThisRound=1, got %d", state.Players[0].CPGainedThisRound)
	}

	// Second discard: should NOT gain CP (cap reached)
	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s2"},
	}); err != nil {
		t.Fatal(err)
	}
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
	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1"},
	}); err != nil {
		t.Fatal(err)
	}
	if state.Players[0].CP != 0 {
		t.Fatalf("expected 0 CP (cap hit), got %d", state.Players[0].CP)
	}

	// Play through rest of round 2 to trigger round 3
	// Player 1 finishes their turn (already past some phases, set to fight)
	// Turn switch → +1 CP each (turn 2 command phase)
	state.CurrentPhase = PhaseFight
	if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1}); err != nil {
		t.Fatal(err)
	}
	// Player 2's full turn → round advance → +1 CP each (round 3 command phase)
	for range 5 {
		if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 2}); err != nil {
			t.Fatal(err)
		}
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
	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s1", "free": true},
	}); err != nil {
		t.Fatal(err)
	}
	if state.Players[0].CPGainedThisRound != 0 {
		t.Fatalf("free discard should not affect CPGainedThisRound, got %d", state.Players[0].CPGainedThisRound)
	}

	// Normal discard after free: should still gain CP
	if _, err := e.Apply(context.Background(), GameAction{
		Type:         ActionDiscardSecondary,
		PlayerNumber: 1,
		Data:         map[string]any{"secondaryId": "s2"},
	}); err != nil {
		t.Fatal(err)
	}
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

func TestAdjustCP_PositiveCapBypassedByForce(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 4
	state.Players[0].CPGainedThisRound = 1 // cap already reached
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": 1, "force": true},
	})
	if err != nil {
		t.Fatalf("expected force=true to bypass cap, got error: %v", err)
	}
	if state.Players[0].CP != 5 {
		t.Fatalf("expected CP=5 after forced gain, got %d", state.Players[0].CP)
	}
	if state.Players[0].CPGainedThisRound != 2 {
		t.Fatalf("expected CPGainedThisRound=2 after forced gain, got %d", state.Players[0].CPGainedThisRound)
	}

	// Subsequent forced gain still works (each click is its own confirmation)
	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": 1, "force": true},
	})
	if err != nil {
		t.Fatalf("expected second forced gain to succeed, got error: %v", err)
	}
	if state.Players[0].CP != 6 {
		t.Fatalf("expected CP=6 after second forced gain, got %d", state.Players[0].CP)
	}

	// And without force the cap still rejects
	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionAdjustCP,
		PlayerNumber: 1,
		Data:         map[string]any{"delta": 1},
	})
	if err == nil {
		t.Fatal("expected unforced positive adjust to still be rejected")
	}
}

func TestCPGainedThisRound_DoesNotResetOnTurnSwitch(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].CP = 5
	state.Players[0].CPGainedThisRound = 1 // already gained bonus CP this round
	state.CurrentPhase = PhaseFight
	e := NewEngine(state)

	// Player 1 finishes turn → switches to player 2's command phase
	if _, err := e.Apply(context.Background(), GameAction{Type: ActionAdvancePhase, PlayerNumber: 1}); err != nil {
		t.Fatal(err)
	}

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

// --- Set Paint Score (Army Painted toggle) ---

func TestSetPaintScore_SetsValue(t *testing.T) {
	e := NewEngine(newTestState())

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetPaintScore,
		PlayerNumber: 1,
		Data:         map[string]any{"score": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.state.Players[0].VPPaint != 10 {
		t.Fatalf("expected VPPaint=10, got %d", e.state.Players[0].VPPaint)
	}

	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionSetPaintScore,
		PlayerNumber: 1,
		Data:         map[string]any{"score": 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.state.Players[0].VPPaint != 0 {
		t.Fatalf("expected VPPaint=0, got %d", e.state.Players[0].VPPaint)
	}
}

func TestSetPaintScore_Clamps(t *testing.T) {
	e := NewEngine(newTestState())

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetPaintScore,
		PlayerNumber: 1,
		Data:         map[string]any{"score": 99},
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.state.Players[0].VPPaint != MaxVPPaint {
		t.Fatalf("expected VPPaint clamped to %d, got %d", MaxVPPaint, e.state.Players[0].VPPaint)
	}

	_, err = e.Apply(context.Background(), GameAction{
		Type:         ActionSetPaintScore,
		PlayerNumber: 1,
		Data:         map[string]any{"score": -5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.state.Players[0].VPPaint != 0 {
		t.Fatalf("expected VPPaint clamped to 0, got %d", e.state.Players[0].VPPaint)
	}
}

func TestSetPaintScore_RejectedWhenActive(t *testing.T) {
	e := NewEngine(newActiveTestState())

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetPaintScore,
		PlayerNumber: 1,
		Data:         map[string]any{"score": 0},
	})
	if err == nil {
		t.Fatal("expected error when game is active")
	}
}

func TestSetPaintScore_ResetsReadiness(t *testing.T) {
	state := newTestState()
	state.Players[0].Ready = true
	state.Players[1].Ready = true
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionSetPaintScore,
		PlayerNumber: 1,
		Data:         map[string]any{"score": 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].Ready {
		t.Fatal("expected acting player's readiness to be reset")
	}
	if !state.Players[1].Ready {
		t.Fatal("expected other player's readiness to be unchanged")
	}
}

// --- Manual Move (escape hatch) ---

func TestMoveSecondary_ActiveToDiscarded(t *testing.T) {
	state := newActiveTestState()
	p := state.Players[0]
	p.SecondaryMode = "tactical"
	p.ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "discarded",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.ActiveSecondaries) != 0 {
		t.Fatalf("expected card removed from active, got %v", p.ActiveSecondaries)
	}
	if len(p.DiscardedSecondaries) != 1 || p.DiscardedSecondaries[0].ID != "s1" {
		t.Fatalf("expected card in discarded, got %v", p.DiscardedSecondaries)
	}
	if p.CP != 0 {
		t.Fatalf("expected no CP gain, got %d", p.CP)
	}
	if len(events) != 1 || events[0].Type != EventSecondaryMoved {
		t.Fatalf("expected one secondary_moved event, got %v", events)
	}
	if events[0].Data["fromPile"] != "active" || events[0].Data["toPile"] != "discarded" {
		t.Fatalf("event piles wrong: %v", events[0].Data)
	}
}

func TestMoveSecondary_ActiveToAchievedScoresVP(t *testing.T) {
	state := newActiveTestState()
	p := state.Players[0]
	p.SecondaryMode = "tactical"
	p.ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "achieved",
			"vpScored":    4,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.ActiveSecondaries) != 0 {
		t.Fatalf("expected card removed from active")
	}
	if len(p.AchievedSecondaries) != 1 {
		t.Fatalf("expected card in achieved")
	}
	if p.VPSecondary != 4 {
		t.Fatalf("expected VPSecondary=4, got %d", p.VPSecondary)
	}
}

func TestMoveSecondary_DeckToActiveBypassesPhaseAndPlayer(t *testing.T) {
	state := newActiveTestState()
	state.CurrentPhase = PhaseFight
	state.ActivePlayer = 1
	// Non-active player (player 2) moves a card from their own deck to active.
	p := state.Players[1]
	p.SecondaryMode = "tactical"
	p.TacticalDeck = []ActiveSecondary{makeActiveSecondary("s1", "S1"), makeActiveSecondary("s2", "S2")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 2,
		Data: map[string]any{
			"secondaryId": "s2",
			"fromPile":    "deck",
			"toPile":      "active",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.TacticalDeck) != 1 || p.TacticalDeck[0].ID != "s1" {
		t.Fatalf("expected only s1 left in deck, got %v", p.TacticalDeck)
	}
	if len(p.ActiveSecondaries) != 1 || p.ActiveSecondaries[0].ID != "s2" {
		t.Fatalf("expected s2 in active, got %v", p.ActiveSecondaries)
	}
}

func TestMoveSecondary_DiscardedToActive(t *testing.T) {
	state := newActiveTestState()
	p := state.Players[0]
	p.SecondaryMode = "tactical"
	p.DiscardedSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "discarded",
			"toPile":      "active",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.DiscardedSecondaries) != 0 {
		t.Fatal("expected card removed from discarded")
	}
	if len(p.ActiveSecondaries) != 1 {
		t.Fatal("expected card in active")
	}
}

func TestMoveSecondary_AchievedToActiveCanRevokeVP(t *testing.T) {
	state := newActiveTestState()
	p := state.Players[0]
	p.SecondaryMode = "tactical"
	p.AchievedSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	p.VPSecondary = 4
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "achieved",
			"toPile":      "active",
			"vpScored":    -4,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.VPSecondary != 0 {
		t.Fatalf("expected VPSecondary=0 after revoking, got %d", p.VPSecondary)
	}
	if len(p.ActiveSecondaries) != 1 {
		t.Fatal("expected card moved to active")
	}
}

func TestMoveSecondary_RejectedWhenNotTactical(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "fixed"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "discarded",
		},
	})
	if err == nil {
		t.Fatal("expected error in fixed mode")
	}
}

func TestMoveSecondary_RejectedWhenNotActive(t *testing.T) {
	state := newTestState() // Status = StatusSetup
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "discarded",
		},
	})
	if err == nil {
		t.Fatal("expected error when game not active")
	}
}

func TestMoveSecondary_RejectedWhenCardNotInPile(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].TacticalDeck = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "discarded",
		},
	})
	if err == nil {
		t.Fatal("expected error when card not in fromPile")
	}
}

func TestMoveSecondary_RejectedSamePile(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "active",
		},
	})
	if err == nil {
		t.Fatal("expected error when fromPile == toPile")
	}
}

func TestMoveSecondary_RejectedInvalidPile(t *testing.T) {
	state := newActiveTestState()
	state.Players[0].SecondaryMode = "tactical"
	state.Players[0].ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	e := NewEngine(state)

	_, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "graveyard",
		},
	})
	if err == nil {
		t.Fatal("expected error for unknown pile name")
	}
}

func TestMoveSecondary_VPClampedToMax(t *testing.T) {
	state := newActiveTestState()
	p := state.Players[0]
	p.SecondaryMode = "tactical"
	p.ActiveSecondaries = []ActiveSecondary{makeActiveSecondary("s1", "S1")}
	p.VPSecondary = MaxVPSecondary - 2
	e := NewEngine(state)

	events, err := e.Apply(context.Background(), GameAction{
		Type:         ActionMoveSecondary,
		PlayerNumber: 1,
		Data: map[string]any{
			"secondaryId": "s1",
			"fromPile":    "active",
			"toPile":      "achieved",
			"vpScored":    10,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.VPSecondary != MaxVPSecondary {
		t.Fatalf("expected VPSecondary clamped to %d, got %d", MaxVPSecondary, p.VPSecondary)
	}
	if d, _ := events[0].Data["vpDelta"].(int); d != 2 {
		t.Fatalf("expected applied vpDelta=2, got %v", events[0].Data["vpDelta"])
	}
}
