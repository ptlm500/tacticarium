// Package scoring evaluates 40kdc-data mission/secondary card `awards` against
// the tracked game/board state (Option B). This file declares the closed
// vocabulary of DSL constructs the evaluator is expected to handle.
//
// The conformance test (conformance_test.go) asserts that every construct
// appearing in the vendored card data is present in these sets. When an
// upstream data refresh introduces a new predicate, `per` counter, draw
// operation, or action effect, the test fails loudly — signalling that the
// evaluator (and these sets) must be extended before the data is shipped, so a
// new construct can never silently mis-score.
//
// The evaluator implementations live alongside this file (added with the engine
// rework); this file is the source-of-truth manifest they are checked against.
package scoring

// KnownWhenTypes is the set of award `when` predicate types the evaluator
// handles. These appear in `awards[].when` (including nested and/or/not
// operands). Sourced from the 11th-edition launch dataslate.
var KnownWhenTypes = map[string]bool{
	"controls-objective":           true,
	"objective-majority":           true,
	"objective-has-tag":            true,
	"new-objective-controlled":     true,
	"territory-control":            true,
	"engagement-fronts":            true,
	"operation-markers":            true,
	"units-destroyed":              true,
	"units-destroyed-comparison":   true,
	"destroyed-while-on-objective": true,
	"destroyed-in-tagged-terrain":  true,
	"action-completed":             true,
	"unit-has-tag":                 true,
}

// KnownPerKinds is the set of `per` descriptors the evaluator can count when an
// award scores `vp_per`. Each is a kebab-case noun describing the thing being
// counted.
var KnownPerKinds = map[string]bool{
	"controlled-objective":                                                    true,
	"controlled-non-home-objective":                                           true,
	"controlled-objective-in-enemy-territory":                                 true,
	"objective-newly-controlled-this-turn":                                    true,
	"decoyed-objective":                                                       true,
	"decoyed-objective-in-enemy-territory":                                    true,
	"objective-guarded-by-your-army":                                          true,
	"operation-marker-within-range-of-a-controlled-central-objective":         true,
	"terrain-area-trapped-this-turn":                                          true,
	"terrain-area-trapped-this-turn-that-is-an-objective":                     true,
	"extract-intelligence-action-completed-this-turn":                         true,
	"friendly-unit-that-committed-sabotage-this-turn":                         true,
	"sabotaging-unit-within-range-of-an-objective-in-enemy-territory":         true,
	"friendly-unit-wholly-within-opponent-deployment-zone":                    true,
	"beacon-unit-on-battlefield-not-in-own-deployment-zone":                   true,
	"beacon-unit-on-battlefield-not-in-own-territory":                         true,
	"enemy-unit-destroyed-this-turn":                                          true,
	"enemy-unit-destroyed-that-started-the-turn-within-range-of-an-objective": true,
	"enemy-unit-of-13-or-more-starting-strength-destroyed-this-turn":          true,
	"enemy-model-with-wounds-10-or-more-destroyed-this-turn":                  true,
	"enemy-character-model-destroyed-this-turn":                               true,
	"enemy-character-model-with-wounds-4-or-more-destroyed-this-turn":         true,
}

// KnownWhenDrawnOps is the set of deck operations a card may trigger on draw.
// Mirrors secondary-card.schema.json#/properties/when_drawn/operation.
var KnownWhenDrawnOps = map[string]bool{
	"reshuffle":  true,
	"replace":    true,
	"redraw":     true,
	"draw-extra": true,
	"swap":       true,
}

// KnownActionEffectTypes is the set of card-action effect types the tracker
// recognises (the marker-placing subset of the Ability DSL effect language).
var KnownActionEffectTypes = map[string]bool{
	"objective-tag":    true,
	"terrain-area-tag": true,
	"unit-tag":         true,
}
