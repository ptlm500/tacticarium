package scoring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

// card mirrors the subset of the 40kdc-data secondary-card shape the evaluator
// reasons about. Both primary mission cards and secondary cards live in
// secondary-cards.json.
type card struct {
	ID        string `json:"id"`
	WhenDrawn *struct {
		Operation string `json:"operation"`
	} `json:"when_drawn"`
	Actions []struct {
		Effect *struct {
			Type string `json:"type"`
		} `json:"effect"`
	} `json:"actions"`
	Awards []struct {
		Per  string          `json:"per"`
		When json.RawMessage `json:"when"`
	} `json:"awards"`
}

// vendoredCardsPath locates data/40kdc/secondary-cards.json relative to this
// source file, independent of the test's working directory.
func vendoredCardsPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// scoring -> game -> internal -> backend -> repo root (4 levels up).
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	return filepath.Join(root, "data", "40kdc", "secondary-cards.json")
}

// collectWhenTypes walks a `when` predicate node, gathering every `type` value
// (including those nested under and/or/not operands).
func collectWhenTypes(raw json.RawMessage, out map[string]bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return
	}
	var node struct {
		Type     string            `json:"type"`
		Operands []json.RawMessage `json:"operands"`
	}
	if err := json.Unmarshal(raw, &node); err != nil {
		return
	}
	if node.Type != "" {
		out[node.Type] = true
	}
	for _, op := range node.Operands {
		collectWhenTypes(op, out)
	}
}

// TestCardDataConformance asserts every DSL construct used by the vendored card
// data is declared in the evaluator's known-vocabulary sets. A failure means an
// upstream data refresh introduced a construct the evaluator does not yet
// handle — extend the relevant Known* set (and its implementation) before
// shipping the data.
func TestCardDataConformance(t *testing.T) {
	data, err := os.ReadFile(vendoredCardsPath(t))
	if err != nil {
		t.Fatalf("reading vendored cards: %v", err)
	}
	var cards []card
	if err := json.Unmarshal(data, &cards); err != nil {
		t.Fatalf("parsing vendored cards: %v", err)
	}
	if len(cards) == 0 {
		t.Fatal("no cards found in vendored data")
	}

	whenTypes := map[string]bool{}
	pers := map[string]bool{}
	drawOps := map[string]bool{}
	effects := map[string]bool{}

	for _, c := range cards {
		if c.WhenDrawn != nil && c.WhenDrawn.Operation != "" {
			drawOps[c.WhenDrawn.Operation] = true
		}
		for _, a := range c.Actions {
			if a.Effect != nil && a.Effect.Type != "" {
				effects[a.Effect.Type] = true
			}
		}
		for _, aw := range c.Awards {
			if aw.Per != "" {
				pers[aw.Per] = true
			}
			collectWhenTypes(aw.When, whenTypes)
		}
	}

	assertCovered(t, "when.type", whenTypes, KnownWhenTypes)
	assertCovered(t, "per", pers, KnownPerKinds)
	assertCovered(t, "when_drawn.operation", drawOps, KnownWhenDrawnOps)
	assertCovered(t, "action.effect.type", effects, KnownActionEffectTypes)
}

func assertCovered(t *testing.T, label string, found, known map[string]bool) {
	t.Helper()
	var missing []string
	for v := range found {
		if !known[v] {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("vendored card data uses %d %s value(s) not handled by the scoring evaluator: %v\n"+
			"Add them to the Known* set in predicates.go and implement their evaluation before shipping this data.",
			len(missing), label, missing)
	}
}
