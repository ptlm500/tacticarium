package game

import (
	"context"
	"testing"

	"github.com/peter/tacticarium/backend/internal/game/scoring"
)

// --- helpers ---

// testBoard is a 5-objective board: 1 central, 2 home (one per side), 2 expansion.
func testBoard() scoring.Board {
	return scoring.Board{
		PlayerSides: [2]scoring.Side{scoring.SideDefender, scoring.SideAttacker},
		Objectives: []scoring.Objective{
			{Index: 0, Role: scoring.RoleCentral},
			{Index: 1, Role: scoring.RoleHome, HomeSide: scoring.SideDefender, TerritorySide: scoring.SideDefender},
			{Index: 2, Role: scoring.RoleHome, HomeSide: scoring.SideAttacker, TerritorySide: scoring.SideAttacker},
			{Index: 3, Role: scoring.RoleExpansion},
			{Index: 4, Role: scoring.RoleExpansion},
		},
	}
}

// primaryControlCard scores 3 VP per controlled objective at the end of the
// owner's turn (a Layer-1, board-derivable award).
func primaryControlCard() scoring.Card {
	three := 3
	return scoring.Card{
		ID: "m", Name: "Test Primary", CardType: "primary",
		Awards: []scoring.Award{{
			Trigger: scoring.Trigger{Timing: "end-of-turn", PlayerTurn: "your-turn"},
			VPPer:   &three, Per: "controlled-objective",
		}},
	}
}

func apply(t *testing.T, e *Engine, a GameAction) []GameEvent {
	t.Helper()
	events, err := e.Apply(context.Background(), a)
	if err != nil {
		t.Fatalf("apply %s: %v", a.Type, err)
	}
	return events
}

// setupActiveGame drives a two-player game from setup to active.
func setupActiveGame(t *testing.T, primary scoring.Card, deck []SecondaryCard) *Engine {
	t.Helper()
	state := &GameState{
		GameID: "g", InviteCode: "c", Status: StatusSetup,
		Players: [2]*PlayerState{{UserID: "u1", PlayerNumber: 1}, {UserID: "u2", PlayerNumber: 2}},
	}
	e := NewEngine(state)
	e.SetMissionResolver(func(_, _ string) (ResolvedMission, bool) {
		return ResolvedMission{ID: "m", Name: "Test Primary", GameCap: 45, RoundCap: 15, PrimaryCard: primary}, true
	})
	e.SetBoardBuilder(func(s1, s2 scoring.Side) (scoring.Board, error) {
		b := testBoard()
		b.PlayerSides = [2]scoring.Side{s1, s2}
		return b, nil
	})
	e.SetSecondaryDeckBuilder(func(string) []SecondaryCard { return deck })

	apply(t, e, GameAction{Type: ActionSelectSide, PlayerNumber: 1, Data: map[string]any{"side": "defender"}})
	apply(t, e, GameAction{Type: ActionSelectFirstTurnPlayer, PlayerNumber: 1, Data: map[string]any{"firstTurnPlayer": 1}})
	apply(t, e, GameAction{Type: ActionSelectForceDisposition, PlayerNumber: 1, Data: map[string]any{"disposition": "take-and-hold"}})
	apply(t, e, GameAction{Type: ActionSelectForceDisposition, PlayerNumber: 2, Data: map[string]any{"disposition": "disruption"}})
	apply(t, e, GameAction{Type: ActionSelectSecondaryMode, PlayerNumber: 1, Data: map[string]any{"mode": "tactical"}})
	apply(t, e, GameAction{Type: ActionSelectSecondaryMode, PlayerNumber: 2, Data: map[string]any{"mode": "tactical"}})
	apply(t, e, GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}})
	apply(t, e, GameAction{Type: ActionSetReady, PlayerNumber: 2, Data: map[string]any{"ready": true}})
	return e
}

func advance(t *testing.T, e *Engine) {
	t.Helper()
	apply(t, e, GameAction{Type: ActionAdvancePhase, PlayerNumber: e.state.ActivePlayer})
}

// --- tests ---

func TestSetupToActive(t *testing.T) {
	e := setupActiveGame(t, primaryControlCard(), nil)
	s := e.State()
	if s.Status != StatusActive {
		t.Fatalf("status = %q, want active", s.Status)
	}
	if s.CurrentRound != 1 || s.CurrentTurn != 1 || s.CurrentPhase != PhaseCommand {
		t.Errorf("start = R%d T%d %s, want R1 T1 command", s.CurrentRound, s.CurrentTurn, s.CurrentPhase)
	}
	if s.ActivePlayer != 1 {
		t.Errorf("active player = %d, want 1", s.ActivePlayer)
	}
	// Both players gained 1 CP entering Command.
	for _, p := range s.Players {
		if p.CP != 1 {
			t.Errorf("player %d CP = %d, want 1", p.PlayerNumber, p.CP)
		}
		if p.MissionID != "m" {
			t.Errorf("player %d mission not resolved", p.PlayerNumber)
		}
	}
	// Sides assigned (player 1 chose defender).
	if s.Players[0].Side != scoring.SideDefender || s.Players[1].Side != scoring.SideAttacker {
		t.Errorf("sides = %q/%q, want defender/attacker", s.Players[0].Side, s.Players[1].Side)
	}
	if len(s.Board.Objectives) != 5 {
		t.Errorf("board not built: %d objectives", len(s.Board.Objectives))
	}
}

func TestTurnStageProgression(t *testing.T) {
	e := setupActiveGame(t, primaryControlCard(), nil)
	// Command -> Movement -> Shooting -> Charge -> Fight -> EndOfTurn
	wantStages := []Phase{PhaseMovement, PhaseShooting, PhaseCharge, PhaseFight, PhaseEndOfTurn}
	for _, want := range wantStages {
		advance(t, e)
		if e.state.CurrentPhase != want {
			t.Fatalf("after advance, phase = %q, want %q", e.state.CurrentPhase, want)
		}
	}
	// Advancing past EndOfTurn switches to player 2's turn at StartOfTurn.
	advance(t, e)
	if e.state.ActivePlayer != 2 || e.state.CurrentTurn != 2 || e.state.CurrentPhase != PhaseStartOfTurn {
		t.Fatalf("after turn end: active=%d turn=%d phase=%q, want 2/2/start_of_turn",
			e.state.ActivePlayer, e.state.CurrentTurn, e.state.CurrentPhase)
	}
	// StartOfTurn -> Command grants CP to both players (now 2 each).
	advance(t, e)
	if e.state.CurrentPhase != PhaseCommand {
		t.Fatalf("phase = %q, want command", e.state.CurrentPhase)
	}
	for _, p := range e.state.Players {
		if p.CP != 2 {
			t.Errorf("player %d CP = %d, want 2 after second command", p.PlayerNumber, p.CP)
		}
	}
}

func TestRoundAdvancesAfterBothTurns(t *testing.T) {
	e := setupActiveGame(t, primaryControlCard(), nil)
	// Run player 1's full turn (command..end_of_turn = 6 advances incl. switch).
	for i := 0; i < 6; i++ {
		advance(t, e)
	}
	// Now player 2, round 1, start_of_turn. Run to end and switch.
	for e.state.CurrentPhase != PhaseStartOfTurn || e.state.CurrentRound != 2 {
		advance(t, e)
		if e.state.Status != StatusActive {
			break
		}
	}
	if e.state.CurrentRound != 2 || e.state.ActivePlayer != 1 {
		t.Fatalf("after both turns: round=%d active=%d, want round 2 active 1",
			e.state.CurrentRound, e.state.ActivePlayer)
	}
}

func TestPrimaryAutoScoreAtEndOfTurn(t *testing.T) {
	e := setupActiveGame(t, primaryControlCard(), nil)
	// Player 1 controls 3 objectives.
	for _, idx := range []int{0, 1, 3} {
		apply(t, e, GameAction{
			Type: ActionSetObjectiveControl, PlayerNumber: 1,
			Data: map[string]any{"objectiveIndex": idx, "player": 1},
		})
	}
	// Advance to end of player 1's turn.
	for e.state.CurrentPhase != PhaseEndOfTurn {
		advance(t, e)
	}
	if got := e.state.Players[0].VPPrimary; got != 9 {
		t.Errorf("primary VP = %d, want 9 (3 objectives x 3)", got)
	}
	// Opponent scored nothing (controls no objectives, and it's not their turn).
	if got := e.state.Players[1].VPPrimary; got != 0 {
		t.Errorf("opponent primary VP = %d, want 0", got)
	}
}

func TestPerRoundCapEnforced(t *testing.T) {
	// Award scores 30 in one shot — above the 15/round cap.
	thirty := 30
	card := scoring.Card{
		ID: "m", CardType: "primary",
		Awards: []scoring.Award{{
			Trigger: scoring.Trigger{Timing: "end-of-turn", PlayerTurn: "your-turn"},
			VP:      &thirty,
		}},
	}
	e := setupActiveGame(t, card, nil)
	for e.state.CurrentPhase != PhaseEndOfTurn {
		advance(t, e)
	}
	if got := e.state.Players[0].VPPrimary; got != 15 {
		t.Errorf("primary VP = %d, want 15 (per-round cap)", got)
	}
}

func TestDrawSecondariesKeepsCards(t *testing.T) {
	deck := []SecondaryCard{
		{Card: scoring.Card{ID: "s1", Name: "S1"}},
		{Card: scoring.Card{ID: "s2", Name: "S2"}},
		{Card: scoring.Card{ID: "s3", Name: "S3"}},
	}
	e := setupActiveGame(t, primaryControlCard(), deck)
	apply(t, e, GameAction{Type: ActionDrawSecondaries, PlayerNumber: 1})
	p := e.state.Players[0]
	if len(p.SecondaryHand) != 2 {
		t.Fatalf("hand size = %d, want 2", len(p.SecondaryHand))
	}
	if len(p.SecondaryDeck) != 1 {
		t.Errorf("deck size = %d, want 1", len(p.SecondaryDeck))
	}
}

// setupFixedGameState builds a setup-phase engine with side/disposition/first
// player chosen, player 1 on fixed mode and player 2 on tactical, leaving the
// fixed selection + ready-up to the caller.
func setupFixedGameState(t *testing.T, deck []SecondaryCard) *Engine {
	t.Helper()
	state := &GameState{
		GameID: "g", InviteCode: "c", Status: StatusSetup,
		Players: [2]*PlayerState{{UserID: "u1", PlayerNumber: 1}, {UserID: "u2", PlayerNumber: 2}},
	}
	e := NewEngine(state)
	e.SetMissionResolver(func(_, _ string) (ResolvedMission, bool) {
		return ResolvedMission{ID: "m", Name: "M", GameCap: 45, RoundCap: 15, PrimaryCard: primaryControlCard()}, true
	})
	e.SetBoardBuilder(func(s1, s2 scoring.Side) (scoring.Board, error) {
		b := testBoard()
		b.PlayerSides = [2]scoring.Side{s1, s2}
		return b, nil
	})
	e.SetSecondaryDeckBuilder(func(string) []SecondaryCard { return deck })

	apply(t, e, GameAction{Type: ActionSelectSide, PlayerNumber: 1, Data: map[string]any{"side": "defender"}})
	apply(t, e, GameAction{Type: ActionSelectFirstTurnPlayer, PlayerNumber: 1, Data: map[string]any{"firstTurnPlayer": 1}})
	apply(t, e, GameAction{Type: ActionSelectForceDisposition, PlayerNumber: 1, Data: map[string]any{"disposition": "take-and-hold"}})
	apply(t, e, GameAction{Type: ActionSelectForceDisposition, PlayerNumber: 2, Data: map[string]any{"disposition": "disruption"}})
	apply(t, e, GameAction{Type: ActionSelectSecondaryMode, PlayerNumber: 1, Data: map[string]any{"mode": "fixed"}})
	apply(t, e, GameAction{Type: ActionSelectSecondaryMode, PlayerNumber: 2, Data: map[string]any{"mode": "tactical"}})
	return e
}

func threeCardDeck() []SecondaryCard {
	return []SecondaryCard{
		{Card: scoring.Card{ID: "s1", Name: "S1"}},
		{Card: scoring.Card{ID: "s2", Name: "S2"}},
		{Card: scoring.Card{ID: "s3", Name: "S3"}},
	}
}

func TestFixedSecondariesDealtToHandAtStart(t *testing.T) {
	e := setupFixedGameState(t, threeCardDeck())
	apply(t, e, GameAction{Type: ActionSelectFixedSecondaries, PlayerNumber: 1, Data: map[string]any{"secondaryIds": []string{"s1", "s3"}}})
	apply(t, e, GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}})
	apply(t, e, GameAction{Type: ActionSetReady, PlayerNumber: 2, Data: map[string]any{"ready": true}})

	if e.state.Status != StatusActive {
		t.Fatalf("status = %q, want active", e.state.Status)
	}
	p1 := e.state.Players[0]
	if len(p1.SecondaryHand) != 2 || p1.SecondaryHand[0].ID != "s1" || p1.SecondaryHand[1].ID != "s3" {
		t.Fatalf("fixed hand ids = %v, want [s1 s3]", cardIDs(p1.SecondaryHand))
	}
	if len(p1.SecondaryDeck) != 0 {
		t.Errorf("fixed player deck = %d, want 0", len(p1.SecondaryDeck))
	}
	p2 := e.state.Players[1]
	if len(p2.SecondaryDeck) != 3 {
		t.Errorf("tactical player deck = %d, want 3", len(p2.SecondaryDeck))
	}
	if len(p2.SecondaryHand) != 0 {
		t.Errorf("tactical player hand = %d, want 0", len(p2.SecondaryHand))
	}
}

func TestFixedPlayerCannotReadyWithoutSelection(t *testing.T) {
	e := setupFixedGameState(t, threeCardDeck())
	_, err := e.Apply(context.Background(), GameAction{Type: ActionSetReady, PlayerNumber: 1, Data: map[string]any{"ready": true}})
	if err == nil {
		t.Fatal("expected error readying fixed player with no secondaries chosen")
	}
	if e.state.Players[0].Ready {
		t.Error("fixed player marked ready despite no selection")
	}
}

func TestSelectFixedSecondariesValidation(t *testing.T) {
	t.Run("rejects more than the fixed count", func(t *testing.T) {
		e := setupFixedGameState(t, threeCardDeck())
		_, err := e.Apply(context.Background(), GameAction{Type: ActionSelectFixedSecondaries, PlayerNumber: 1, Data: map[string]any{"secondaryIds": []string{"s1", "s2", "s3"}}})
		if err == nil {
			t.Fatalf("expected error choosing more than %d", FixedSecondaryCount)
		}
	})

	t.Run("rejects duplicate ids", func(t *testing.T) {
		e := setupFixedGameState(t, threeCardDeck())
		_, err := e.Apply(context.Background(), GameAction{Type: ActionSelectFixedSecondaries, PlayerNumber: 1, Data: map[string]any{"secondaryIds": []string{"s1", "s1"}}})
		if err == nil {
			t.Fatal("expected error choosing duplicate ids")
		}
	})

	t.Run("rejects selection by a tactical player", func(t *testing.T) {
		e := setupFixedGameState(t, threeCardDeck())
		_, err := e.Apply(context.Background(), GameAction{Type: ActionSelectFixedSecondaries, PlayerNumber: 2, Data: map[string]any{"secondaryIds": []string{"s1"}}})
		if err == nil {
			t.Fatal("expected error: tactical player selecting fixed secondaries")
		}
	})

	t.Run("clears the selection when switching to tactical", func(t *testing.T) {
		e := setupFixedGameState(t, threeCardDeck())
		apply(t, e, GameAction{Type: ActionSelectFixedSecondaries, PlayerNumber: 1, Data: map[string]any{"secondaryIds": []string{"s1", "s2"}}})
		apply(t, e, GameAction{Type: ActionSelectSecondaryMode, PlayerNumber: 1, Data: map[string]any{"mode": "tactical"}})
		if len(e.state.Players[0].FixedSecondaryIDs) != 0 {
			t.Errorf("fixed ids = %v, want cleared after switching to tactical", e.state.Players[0].FixedSecondaryIDs)
		}
	})
}

func cardIDs(cards []SecondaryCard) []string {
	out := make([]string, len(cards))
	for i, c := range cards {
		out[i] = c.ID
	}
	return out
}

func TestConfirmLayer2Award(t *testing.T) {
	// A Layer-2 award (units-destroyed) needs a confirmed count.
	two := 2
	card := scoring.Card{
		ID: "m", CardType: "primary",
		Awards: []scoring.Award{{
			Trigger: scoring.Trigger{Timing: "end-of-turn", PlayerTurn: "your-turn"},
			VPPer:   &two, Per: "enemy-unit-destroyed-this-turn",
		}},
	}
	e := setupActiveGame(t, card, nil)
	for e.state.CurrentPhase != PhaseEndOfTurn {
		advance(t, e)
	}
	p := e.state.Players[0]
	if len(p.PendingScorePrompts) != 1 {
		t.Fatalf("expected 1 pending prompt, got %d", len(p.PendingScorePrompts))
	}
	if p.VPPrimary != 0 {
		t.Errorf("no VP should be scored before confirmation, got %d", p.VPPrimary)
	}
	promptID := p.PendingScorePrompts[0].ID
	apply(t, e, GameAction{
		Type: ActionConfirmAward, PlayerNumber: 1,
		Data: map[string]any{"promptId": promptID, "count": 3},
	})
	if got := e.state.Players[0].VPPrimary; got != 6 {
		t.Errorf("after confirming 3 kills: VP = %d, want 6 (3 x 2)", got)
	}
	if len(e.state.Players[0].PendingScorePrompts) != 0 {
		t.Errorf("prompt should be cleared after confirmation")
	}
}

func TestCPRevertRevokesGain(t *testing.T) {
	e := setupActiveGame(t, primaryControlCard(), nil)
	// Advance to player 2's first Command (CP becomes 2 each), then revert it.
	for e.state.CurrentTurn != 2 || e.state.CurrentPhase != PhaseCommand {
		advance(t, e)
	}
	if e.state.Players[0].CP != 2 {
		t.Fatalf("CP = %d, want 2 before revert", e.state.Players[0].CP)
	}
	apply(t, e, GameAction{Type: ActionRevertPhase, PlayerNumber: e.state.ActivePlayer})
	if e.state.CurrentPhase != PhaseStartOfTurn {
		t.Fatalf("phase = %q, want start_of_turn after revert", e.state.CurrentPhase)
	}
	if e.state.Players[0].CP != 1 {
		t.Errorf("CP = %d, want 1 after reverting the Command CP gain", e.state.Players[0].CP)
	}
}

func TestConcede(t *testing.T) {
	e := setupActiveGame(t, primaryControlCard(), nil)
	apply(t, e, GameAction{Type: ActionConcede, PlayerNumber: 1})
	s := e.State()
	if s.Status != StatusCompleted {
		t.Errorf("status = %q, want completed", s.Status)
	}
	if s.WinnerID != "u2" {
		t.Errorf("winner = %q, want u2 (opponent of conceder)", s.WinnerID)
	}
}

func TestStratagemCPSpend(t *testing.T) {
	e := setupActiveGame(t, primaryControlCard(), nil)
	apply(t, e, GameAction{
		Type: ActionUseStratagem, PlayerNumber: 1,
		Data: map[string]any{"stratagemId": "x", "stratagemName": "X", "cpCost": 1},
	})
	if e.state.Players[0].CP != 0 {
		t.Errorf("CP = %d, want 0 after spending 1", e.state.Players[0].CP)
	}
	// Insufficient CP is rejected.
	_, err := e.Apply(context.Background(), GameAction{
		Type: ActionUseStratagem, PlayerNumber: 1,
		Data: map[string]any{"stratagemId": "y", "cpCost": 5},
	})
	if err == nil {
		t.Errorf("expected error spending more CP than available")
	}
}

func TestNextStageBookends(t *testing.T) {
	if got, ended := nextStage(PhaseEndOfTurn); !ended || got != PhaseStartOfTurn {
		t.Errorf("nextStage(end_of_turn) = (%q,%v), want (start_of_turn,true)", got, ended)
	}
	if got, ended := nextStage(PhaseCommand); ended || got != PhaseMovement {
		t.Errorf("nextStage(command) = (%q,%v), want (movement,false)", got, ended)
	}
}
