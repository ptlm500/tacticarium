package scoring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// loadCard reads a single card from the vendored secondary-cards.json by id.
func loadCard(t *testing.T, id string) Card {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	data, err := os.ReadFile(filepath.Join(root, "data", "40kdc", "secondary-cards.json"))
	if err != nil {
		t.Fatalf("reading cards: %v", err)
	}
	var raw []struct {
		ID       string          `json:"id"`
		Name     string          `json:"name"`
		CardType string          `json:"card_type"`
		Text     string          `json:"text"`
		Awards   json.RawMessage `json:"awards"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing cards: %v", err)
	}
	for _, r := range raw {
		if r.ID == id {
			c, err := ParseCard(r.ID, r.Name, r.CardType, r.Text, r.Awards)
			if err != nil {
				t.Fatalf("parse card %s: %v", id, err)
			}
			return c
		}
	}
	t.Fatalf("card %s not found", id)
	return Card{}
}

// tippingPointBoard builds a tipping-point board (p1=defender, p2=attacker) and
// applies the given control map (objective index -> player).
func tippingPointBoard(t *testing.T, control map[int]int) *Board {
	t.Helper()
	var tp *deploymentPatternFile
	for _, f := range loadDeploymentPatterns(t) {
		if f.ID == "tipping-point" {
			ff := f
			tp = &ff
		}
	}
	if tp == nil {
		t.Skip("tipping-point not present")
	}
	b := buildFromFile(t, *tp, SideDefender, SideAttacker)
	for i := range b.Objectives {
		if p, ok := control[b.Objectives[i].Index]; ok {
			b.Objectives[i].ControlledBy = p
		}
	}
	return &b
}

func objIndexByPoint(b *Board, x, y float64) int {
	for i := range b.Objectives {
		if b.Objectives[i].Point.X == x && b.Objectives[i].Point.Y == y {
			return b.Objectives[i].Index
		}
	}
	return -1
}

// Battlefield Dominance (primary) is entirely board-derivable. Verify it
// auto-scores correctly at end of a BR3 Command phase.
func TestBattlefieldDominancePrimary(t *testing.T) {
	// Build with no control first to learn indices, then set control.
	base := tippingPointBoard(t, nil)
	home := objIndexByPoint(base, 14, 34)    // defender home (player 1)
	central := objIndexByPoint(base, 30, 22) // central
	exp := objIndexByPoint(base, 22, 8)      // expansion
	board := tippingPointBoard(t, map[int]int{home: 1, central: 1, exp: 1})

	card := loadCard(t, "battlefield-dominance")
	ctx := Context{
		Round: 3, Phase: "command", Timing: "end-of-phase",
		PlayerTurn: "your-turn", Player: 1, Board: board,
	}
	_, total := EvaluateCard(&card, ctx)
	// Award 2: 3 VP per controlled objective = 3*3 = 9.
	// Award 3: +2 VP per controlled non-home objective (central, expansion = 2),
	//          gated on controlling your-home (true) = 2*2 = 4.
	if total != 13 {
		t.Errorf("BR3 end-of-command total = %d, want 13", total)
	}

	// At BR1 end-of-turn only the objective-majority award fires (1 controls 3,
	// 2 controls 0) -> 2 VP.
	ctx2 := Context{Round: 1, Timing: "end-of-turn", PlayerTurn: "your-turn", Player: 1, Board: board}
	_, total2 := EvaluateCard(&card, ctx2)
	if total2 != 2 {
		t.Errorf("BR1 end-of-turn total = %d, want 2", total2)
	}
}

// Assassination is a Layer-2 card with fixed/tactical tracks. Fixed needs a
// confirmed count; tactical scores a flat amount and ignores the fixed awards.
func TestAssassinationLayers(t *testing.T) {
	card := loadCard(t, "assassination")
	board := tippingPointBoard(t, nil)

	// Fixed mode, no confirmation yet -> the per-character award needs confirm.
	fixed := Context{Timing: "end-of-turn", PlayerTurn: "either", Mode: "fixed", Player: 1, Board: board}
	results, total := EvaluateCard(&card, fixed)
	if total != 0 {
		t.Errorf("fixed/unconfirmed total = %d, want 0", total)
	}
	needs := false
	for _, r := range results {
		if r.NeedsConfirm {
			needs = true
		}
	}
	if !needs {
		t.Error("expected a Layer-2 award to need confirmation in fixed mode")
	}

	// Fixed mode, confirm 2 characters destroyed -> 3 VP each = 6 (plus any
	// cumulative wounds-4+ award, which needs its own confirm -> 0 here).
	idx := -1
	for i := range card.Awards {
		if card.Awards[i].Per == "enemy-character-model-destroyed-this-turn" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("assassination missing expected fixed award")
	}
	fixed.Confirmed = map[int]int{idx: 2}
	_, total = EvaluateCard(&card, fixed)
	if total != 6 {
		t.Errorf("fixed/confirmed total = %d, want 6", total)
	}

	// Tactical mode scores the flat 5 and ignores the fixed-mode awards.
	tac := Context{Timing: "end-of-turn", PlayerTurn: "either", Mode: "tactical", Player: 1, Board: board}
	_, total = EvaluateCard(&card, tac)
	if total != 5 {
		t.Errorf("tactical total = %d, want 5", total)
	}
}

// Exclusive groups score only the highest tier, not the sum.
func TestExclusiveGroup(t *testing.T) {
	five, three := 5, 3
	card := Card{Awards: []Award{
		{Trigger: Trigger{Timing: "end-of-turn"}, VP: &three, ExclusiveGroup: "g"},
		{Trigger: Trigger{Timing: "end-of-turn"}, VP: &five, ExclusiveGroup: "g"},
	}}
	_, total := EvaluateCard(&card, Context{Timing: "end-of-turn"})
	if total != 5 {
		t.Errorf("exclusive group total = %d, want 5 (highest only)", total)
	}
}

func TestTriggerFires(t *testing.T) {
	two := 2
	tr := Trigger{Timing: "end-of-phase", Phase: "command", PlayerTurn: "your-turn", BattleRound: &RoundWindow{Min: &two}}
	if !TriggerFires(tr, Context{Timing: "end-of-phase", Phase: "command", PlayerTurn: "your-turn", Round: 3}) {
		t.Error("trigger should fire at BR3 end-of-command on your turn")
	}
	if TriggerFires(tr, Context{Timing: "end-of-phase", Phase: "command", PlayerTurn: "your-turn", Round: 1}) {
		t.Error("trigger should not fire before its min battle round")
	}
	if TriggerFires(tr, Context{Timing: "end-of-turn", PlayerTurn: "your-turn", Round: 3}) {
		t.Error("trigger should not fire at the wrong timing")
	}
}

func TestCapsClamp(t *testing.T) {
	c := DefaultCaps
	if got := c.Clamp(10, 8, 8); got != 7 {
		t.Errorf("clamp limited by per-round room: got %d, want 7", got)
	}
	if got := c.Clamp(10, 0, 44); got != 1 {
		t.Errorf("clamp limited by per-game room: got %d, want 1", got)
	}
	if got := c.Clamp(10, 0, 0); got != 10 {
		t.Errorf("clamp with room: got %d, want 10", got)
	}
	if got := c.Clamp(5, 15, 20); got != 0 {
		t.Errorf("clamp at round cap: got %d, want 0", got)
	}
}
