package scoring

// otherSide returns the opposing side.
func otherSide(s Side) Side {
	if s == SideAttacker {
		return SideDefender
	}
	return SideAttacker
}

// evalCondition evaluates a Layer-1 (board-derivable) condition tree. Layer-2
// leaves should never reach here (awardLayer routes those to confirmation); if
// one does, it evaluates to false.
func evalCondition(c *Condition, ctx Context) bool {
	if c == nil {
		return true
	}
	if c.Operator != "" {
		switch c.Operator {
		case "and":
			for i := range c.Operands {
				if !evalCondition(&c.Operands[i], ctx) {
					return false
				}
			}
			return true
		case "or":
			for i := range c.Operands {
				if evalCondition(&c.Operands[i], ctx) {
					return true
				}
			}
			return false
		case "not":
			if len(c.Operands) == 1 {
				return !evalCondition(&c.Operands[0], ctx)
			}
		}
		return false
	}

	res := evalLeaf(c, ctx)
	if c.Negated {
		return !res
	}
	return res
}

func evalLeaf(c *Condition, ctx Context) bool {
	if ctx.Board == nil {
		return false
	}
	switch c.Type {
	case "controls-objective":
		min := 1
		if m, ok := c.intParam("count_min"); ok {
			min = m
		}
		return countMatchingControlled(c, ctx) >= min

	case "objective-majority":
		mine, opp := 0, 0
		for i := range ctx.Board.Objectives {
			switch ctx.Board.Objectives[i].ControlledBy {
			case ctx.Player:
				mine++
			case opponent(ctx.Player):
				opp++
			}
		}
		return mine > opp

	case "objective-has-tag":
		tag, _ := c.strParam("tag")
		n := 0
		for i := range ctx.Board.Objectives {
			o := &ctx.Board.Objectives[i]
			if tag != "" && !o.HasTag(tag) {
				continue
			}
			if !matchObjectiveName(c, o, ctx) {
				continue
			}
			n++
		}
		if min, ok := c.intParam("count_min"); ok && n < min {
			return false
		}
		if max, ok := c.intParam("count_max"); ok && n > max {
			return false
		}
		_, hasMin := c.intParam("count_min")
		_, hasMax := c.intParam("count_max")
		if !hasMin && !hasMax {
			return n > 0
		}
		return true

	case "new-objective-controlled":
		min := 1
		if m, ok := c.intParam("count_min"); ok {
			min = m
		}
		return newlyControlled(ctx) >= min
	}
	return false
}

// boardCount resolves a Layer-1 `per` counter to an instance count.
func boardCount(per string, ctx Context) int {
	if ctx.Board == nil {
		return 0
	}
	side := ctx.Board.SideOf(ctx.Player)
	n := 0
	for i := range ctx.Board.Objectives {
		o := &ctx.Board.Objectives[i]
		controlled := o.ControlledBy == ctx.Player
		switch per {
		case "controlled-objective":
			if controlled {
				n++
			}
		case "controlled-non-home-objective":
			if controlled && o.Role != RoleHome {
				n++
			}
		case "controlled-objective-in-enemy-territory":
			if controlled && o.TerritorySide == otherSide(side) {
				n++
			}
		case "objective-newly-controlled-this-turn":
			if controlled && ctx.StartOfTurnControl[o.Index] != ctx.Player {
				n++
			}
		case "decoyed-objective":
			if o.HasTag("decoyed") {
				n++
			}
		case "decoyed-objective-in-enemy-territory":
			if o.HasTag("decoyed") && o.TerritorySide == otherSide(side) {
				n++
			}
		}
	}
	return n
}

// countMatchingControlled counts objectives the player controls that match a
// controls-objective filter.
func countMatchingControlled(c *Condition, ctx Context) int {
	n := 0
	for i := range ctx.Board.Objectives {
		o := &ctx.Board.Objectives[i]
		if o.ControlledBy != ctx.Player {
			continue
		}
		if !matchControlsFilter(c, o, ctx) {
			continue
		}
		n++
	}
	return n
}

// matchControlsFilter applies the controls-objective filter parameters to an
// objective (objective name, objective_role, scope, exclude).
func matchControlsFilter(c *Condition, o *Objective, ctx Context) bool {
	if !matchObjectiveName(c, o, ctx) {
		return false
	}
	if role, ok := c.strParam("objective_role"); ok && string(o.Role) != role {
		return false
	}
	if scope, ok := c.strParam("scope"); ok {
		if scope == "no-mans-land" && o.TerritorySide != "" {
			return false
		}
	}
	if excl, ok := c.strParam("exclude"); ok {
		if excl == "home" && o.Role == RoleHome {
			return false
		}
	}
	return true
}

// matchObjectiveName resolves the `objective` parameter (your-home,
// opponent-home, a tag name like tempting-target) against an objective. Absent
// parameter matches any objective.
func matchObjectiveName(c *Condition, o *Objective, ctx Context) bool {
	name, ok := c.strParam("objective")
	if !ok {
		return true
	}
	side := ctx.Board.SideOf(ctx.Player)
	switch name {
	case "your-home":
		return o.Role == RoleHome && o.HomeSide == side
	case "opponent-home":
		return o.Role == RoleHome && o.HomeSide == otherSide(side)
	default:
		// A named objective marker (e.g. "tempting-target") tracked as a tag.
		return o.HasTag(name)
	}
}

func newlyControlled(ctx Context) int {
	n := 0
	for i := range ctx.Board.Objectives {
		o := &ctx.Board.Objectives[i]
		if o.ControlledBy == ctx.Player && ctx.StartOfTurnControl[o.Index] != ctx.Player {
			n++
		}
	}
	return n
}

func opponent(player int) int {
	if player == 1 {
		return 2
	}
	if player == 2 {
		return 1
	}
	return 0
}

// --- VP caps ---

// Caps bounds VP scored in a category. 11th edition caps each of Primary and
// Secondary at 45 VP per game and 15 VP per battle round.
type Caps struct {
	PerRound int
	PerGame  int
}

// DefaultCaps are the launch-dataslate caps.
var DefaultCaps = Caps{PerRound: 15, PerGame: 45}

// Clamp returns how much of `want` may actually be scored given the amount
// already scored this round and this game in the same category.
func (c Caps) Clamp(want, scoredThisRound, scoredThisGame int) int {
	if want <= 0 {
		return 0
	}
	roundRoom := c.PerRound - scoredThisRound
	gameRoom := c.PerGame - scoredThisGame
	room := roundRoom
	if gameRoom < room {
		room = gameRoom
	}
	if room < 0 {
		room = 0
	}
	if want > room {
		return room
	}
	return want
}
