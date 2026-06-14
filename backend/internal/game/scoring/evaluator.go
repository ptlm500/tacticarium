package scoring

// Layer classifies whether an award can be scored automatically from the board
// model (Layer 1) or needs a player-confirmed input for facts off the board
// (Layer 2 — unit destruction, positions, markers, etc.).
type Layer int

const (
	LayerBoard   Layer = iota // auto-scored from board state
	LayerConfirm              // needs a player-confirmed count/boolean
)

// Context carries the game state an award is evaluated against.
type Context struct {
	Round      int
	Phase      string // "command", ...
	Timing     string // "end-of-turn", "end-of-phase", "end-of-battle"
	PlayerTurn string // "your-turn" / "opponent-turn", relative to the card owner
	Mode       string // "fixed" / "tactical" (the owner's secondary approach)
	Player     int    // the card owner (1/2)
	Board      *Board
	// StartOfTurnControl maps objective index -> controlling player at the start
	// of the current turn, for "newly controlled this turn" predicates.
	StartOfTurnControl map[int]int
	// Confirmed holds player-supplied Layer-2 inputs, keyed by award index: an
	// instance count for vp_per awards, or 0/1 for a flat-vp award gated on a
	// Layer-2 condition.
	Confirmed map[int]int
}

// AwardResult is the outcome of evaluating a single award.
type AwardResult struct {
	Index        int
	VP           int
	NeedsConfirm bool // true when a Layer-2 input is required but absent
	Layer        Layer
}

// boardWhenTypes are the `when` predicate types derivable from board state.
var boardWhenTypes = map[string]bool{
	"controls-objective":       true,
	"objective-majority":       true,
	"objective-has-tag":        true,
	"new-objective-controlled": true,
}

// boardPerKinds are the `per` counters derivable from board state.
var boardPerKinds = map[string]bool{
	"controlled-objective":                    true,
	"controlled-non-home-objective":           true,
	"controlled-objective-in-enemy-territory": true,
	"objective-newly-controlled-this-turn":    true,
	"decoyed-objective":                       true,
	"decoyed-objective-in-enemy-territory":    true,
}

// conditionLayer returns LayerConfirm if any leaf in the condition tree is not
// board-derivable.
func conditionLayer(c *Condition) Layer {
	if c == nil {
		return LayerBoard
	}
	if c.Operator != "" {
		for i := range c.Operands {
			if conditionLayer(&c.Operands[i]) == LayerConfirm {
				return LayerConfirm
			}
		}
		return LayerBoard
	}
	if boardWhenTypes[c.Type] {
		return LayerBoard
	}
	return LayerConfirm
}

// awardLayer returns the layer an award is scored at.
func awardLayer(a *Award) Layer {
	if conditionLayer(a.When) == LayerConfirm {
		return LayerConfirm
	}
	if a.Per != "" && !boardPerKinds[a.Per] {
		return LayerConfirm
	}
	return LayerBoard
}

// TriggerFires reports whether an award's trigger matches the current moment.
func TriggerFires(t Trigger, ctx Context) bool {
	if t.Timing != "" && t.Timing != ctx.Timing {
		return false
	}
	if t.Phase != "" && t.Phase != ctx.Phase {
		return false
	}
	if t.PlayerTurn != "" && t.PlayerTurn != "either" && t.PlayerTurn != ctx.PlayerTurn {
		return false
	}
	if t.BattleRound != nil {
		if t.BattleRound.Min != nil && ctx.Round < *t.BattleRound.Min {
			return false
		}
		if t.BattleRound.Max != nil && ctx.Round > *t.BattleRound.Max {
			return false
		}
	}
	return true
}

// EvaluateAward evaluates a single award at index idx. If the award is Layer-2
// and no confirmed input is present, it returns NeedsConfirm. Mode-tagged awards
// that don't match the owner's mode score nothing.
func EvaluateAward(a *Award, idx int, ctx Context) AwardResult {
	res := AwardResult{Index: idx, Layer: awardLayer(a)}

	if a.Mode != "" && a.Mode != ctx.Mode {
		return res // wrong scoring track for this player
	}

	if res.Layer == LayerConfirm {
		count, ok := ctx.Confirmed[idx]
		if !ok {
			res.NeedsConfirm = true
			return res
		}
		res.VP = awardVP(a, count)
		return res
	}

	// Layer 1: evaluate from the board.
	if a.When != nil && !evalCondition(a.When, ctx) {
		return res
	}
	if a.VPPer != nil {
		res.VP = awardVP(a, boardCount(a.Per, ctx))
	} else if a.VP != nil {
		res.VP = *a.VP
	}
	return res
}

// awardVP computes the VP for an award given a resolved instance count (used for
// vp_per; flat vp ignores count beyond the >0 gate handled by the caller).
func awardVP(a *Award, count int) int {
	if a.VPPer != nil {
		if a.PerMax != nil && count > *a.PerMax {
			count = *a.PerMax
		}
		vp := *a.VPPer * count
		if a.VPMax != nil && vp > *a.VPMax {
			vp = *a.VPMax
		}
		return vp
	}
	if a.VP != nil {
		// Flat award gated on a confirmed condition: count==0 means the condition
		// did not hold.
		if count <= 0 {
			return 0
		}
		return *a.VP
	}
	return 0
}

// EvaluateCard evaluates every award on a card whose trigger fires now, applies
// exclusive_group (only the highest-scoring award in a group counts), and
// returns the per-award results and the summed VP. Results needing confirmation
// are returned with NeedsConfirm=true and excluded from the total.
func EvaluateCard(card *Card, ctx Context) (results []AwardResult, total int) {
	groupBest := map[string]int{} // exclusive_group -> best VP
	for i := range card.Awards {
		a := &card.Awards[i]
		if !TriggerFires(a.Trigger, ctx) {
			continue
		}
		r := EvaluateAward(a, i, ctx)
		results = append(results, r)
		if r.NeedsConfirm {
			continue
		}
		if a.ExclusiveGroup != "" {
			if r.VP > groupBest[a.ExclusiveGroup] {
				groupBest[a.ExclusiveGroup] = r.VP
			}
			continue
		}
		total += r.VP
	}
	for _, best := range groupBest {
		total += best
	}
	return results, total
}
