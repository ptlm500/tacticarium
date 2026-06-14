package scoring

import (
	"encoding/json"
	"math"
)

// Side is a deployment-pattern player label. A game assigns each player (1/2) a
// side; the geometry (territories, deployment zones, objective roles) is defined
// in terms of these labels.
type Side string

const (
	SideAttacker Side = "attacker"
	SideDefender Side = "defender"
)

// ObjectiveRole is the geometric role of an objective, derived from the
// deployment pattern: a home objective sits in a player's deployment zone, the
// central objective is the one nearest the board centre, and the remaining
// no-man's-land objectives are expansion objectives.
type ObjectiveRole string

const (
	RoleHome      ObjectiveRole = "home"
	RoleCentral   ObjectiveRole = "central"
	RoleExpansion ObjectiveRole = "expansion"
)

// Point is a position on the battlefield, in inches.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Objective is a resolved board objective: its position, geometrically-derived
// role, and the mutable in-game state players maintain (who controls it and any
// card-placed tags).
type Objective struct {
	Index int           `json:"index"`
	Point Point         `json:"point"`
	Role  ObjectiveRole `json:"role"`
	// HomeSide is the side whose deployment zone a home objective sits in; empty
	// for non-home objectives.
	HomeSide Side `json:"homeSide,omitempty"`
	// TerritorySide is the side whose territory the objective sits in; empty when
	// it lies in no-man's-land (on or across the territory boundary).
	TerritorySide Side `json:"territorySide,omitempty"`
	// ControlledBy is the player number (1/2) controlling the objective, or 0 if
	// uncontrolled.
	ControlledBy int `json:"controlledBy"`
	// Tags are transient markers placed by card actions (e.g. "decoyed",
	// "consecrated", "tempting-target").
	Tags []string `json:"tags,omitempty"`
}

// HasTag reports whether the objective carries tag.
func (o *Objective) HasTag(tag string) bool {
	for _, t := range o.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds tag if not already present.
func (o *Objective) AddTag(tag string) {
	if !o.HasTag(tag) {
		o.Tags = append(o.Tags, tag)
	}
}

// Board is the per-game battlefield state: the resolved objectives plus the
// mapping from player number to deployment-pattern side. It lives on the game
// state and is persisted/broadcast.
type Board struct {
	DeploymentPatternID string      `json:"deploymentPatternId"`
	Objectives          []Objective `json:"objectives"`
	// PlayerSides[i] is the side of player (i+1).
	PlayerSides [2]Side `json:"playerSides"`
}

// SideOf returns the deployment-pattern side of the given player number (1/2).
func (b *Board) SideOf(player int) Side {
	if player == 1 || player == 2 {
		return b.PlayerSides[player-1]
	}
	return ""
}

// --- Deployment pattern geometry (parsed from 40kdc-data) ---

type shapeJSON struct {
	Type   string  `json:"type"`
	Points []Point `json:"points"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type regionJSON struct {
	Player   Side      `json:"player"`
	Position Point     `json:"position"`
	Shape    shapeJSON `json:"shape"`
}

// DeploymentPattern is the parsed geometry of a deployment pattern.
type DeploymentPattern struct {
	ID          string       `json:"id"`
	Objectives  []Point      `json:"objectives"`
	Zones       []regionJSON `json:"zones"`
	Territories []regionJSON `json:"territories"`
}

// ParseDeploymentPattern unmarshals a deployment pattern from its JSON form
// (the shape used in data/40kdc/deployment-patterns.json and the
// deployment_patterns table columns).
func ParseDeploymentPattern(id string, objectives, zones, territories json.RawMessage) (DeploymentPattern, error) {
	dp := DeploymentPattern{ID: id}
	if len(objectives) > 0 {
		if err := json.Unmarshal(objectives, &dp.Objectives); err != nil {
			return dp, err
		}
	}
	if len(zones) > 0 {
		if err := json.Unmarshal(zones, &dp.Zones); err != nil {
			return dp, err
		}
	}
	if len(territories) > 0 {
		if err := json.Unmarshal(territories, &dp.Territories); err != nil {
			return dp, err
		}
	}
	return dp, nil
}

// polygon returns the region's absolute polygon vertices, resolving both the
// polygon and rectangle shape forms and applying the region's position offset.
func (r regionJSON) polygon() []Point {
	off := r.Position
	switch r.Shape.Type {
	case "rectangle":
		w, h := r.Shape.Width, r.Shape.Height
		return []Point{
			{X: off.X, Y: off.Y},
			{X: off.X + w, Y: off.Y},
			{X: off.X + w, Y: off.Y + h},
			{X: off.X, Y: off.Y + h},
		}
	default: // polygon
		pts := make([]Point, len(r.Shape.Points))
		for i, p := range r.Shape.Points {
			pts[i] = Point{X: off.X + p.X, Y: off.Y + p.Y}
		}
		return pts
	}
}

// BuildBoard resolves a deployment pattern into a Board for a game, deriving
// each objective's role and assigning each player their side.
func BuildBoard(dp DeploymentPattern, player1Side, player2Side Side) Board {
	b := Board{
		DeploymentPatternID: dp.ID,
		PlayerSides:         [2]Side{player1Side, player2Side},
		Objectives:          make([]Objective, len(dp.Objectives)),
	}

	zones := make([]struct {
		side Side
		poly []Point
	}, len(dp.Zones))
	for i, z := range dp.Zones {
		zones[i].side = z.Player
		zones[i].poly = z.polygon()
	}
	territories := make([]struct {
		side Side
		poly []Point
	}, len(dp.Territories))
	for i, t := range dp.Territories {
		territories[i].side = t.Player
		territories[i].poly = t.polygon()
	}

	center := boardCenter(dp)

	// First pass: home (in a deployment zone) and territory membership.
	centralIdx, centralDist := -1, math.MaxFloat64
	for i, p := range dp.Objectives {
		obj := Objective{Index: i, Point: p}
		for _, z := range zones {
			if pointInPolygon(p, z.poly) {
				obj.Role = RoleHome
				obj.HomeSide = z.side
				break
			}
		}
		for _, t := range territories {
			if pointInPolygon(p, t.poly) {
				obj.TerritorySide = t.side
				break
			}
		}
		b.Objectives[i] = obj

		if obj.Role != RoleHome {
			if d := dist2(p, center); d < centralDist {
				centralDist, centralIdx = d, i
			}
		}
	}

	// Second pass: the nearest-to-centre non-home objective is central; the rest
	// of the non-home objectives are expansion.
	for i := range b.Objectives {
		if b.Objectives[i].Role == RoleHome {
			continue
		}
		if i == centralIdx {
			b.Objectives[i].Role = RoleCentral
		} else {
			b.Objectives[i].Role = RoleExpansion
		}
	}

	return b
}

func boardCenter(dp DeploymentPattern) Point {
	maxX, maxY := 0.0, 0.0
	consider := func(p Point) {
		maxX = math.Max(maxX, p.X)
		maxY = math.Max(maxY, p.Y)
	}
	for _, p := range dp.Objectives {
		consider(p)
	}
	for _, regs := range [][]regionJSON{dp.Zones, dp.Territories} {
		for _, r := range regs {
			for _, p := range r.polygon() {
				consider(p)
			}
		}
	}
	return Point{X: maxX / 2, Y: maxY / 2}
}

func dist2(a, b Point) float64 {
	dx, dy := a.X-b.X, a.Y-b.Y
	return dx*dx + dy*dy
}

// pointInPolygon reports whether p lies inside the polygon (ray-casting; points
// on an edge are treated as inside-enough for objective placement).
func pointInPolygon(p Point, poly []Point) bool {
	if len(poly) < 3 {
		return false
	}
	inside := false
	j := len(poly) - 1
	for i := 0; i < len(poly); i++ {
		pi, pj := poly[i], poly[j]
		if (pi.Y > p.Y) != (pj.Y > p.Y) {
			xCross := (pj.X-pi.X)*(p.Y-pi.Y)/(pj.Y-pi.Y) + pi.X
			if p.X < xCross {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}
