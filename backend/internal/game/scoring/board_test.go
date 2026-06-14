package scoring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type deploymentPatternFile struct {
	ID          string          `json:"id"`
	Objectives  json.RawMessage `json:"objectives"`
	Zones       json.RawMessage `json:"zones"`
	Territories json.RawMessage `json:"territories"`
}

func loadDeploymentPatterns(t *testing.T) []deploymentPatternFile {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	data, err := os.ReadFile(filepath.Join(root, "data", "40kdc", "deployment-patterns.json"))
	if err != nil {
		t.Fatalf("reading deployment patterns: %v", err)
	}
	var patterns []deploymentPatternFile
	if err := json.Unmarshal(data, &patterns); err != nil {
		t.Fatalf("parsing deployment patterns: %v", err)
	}
	return patterns
}

func buildFromFile(t *testing.T, f deploymentPatternFile, p1, p2 Side) Board {
	t.Helper()
	dp, err := ParseDeploymentPattern(f.ID, f.Objectives, f.Zones, f.Territories)
	if err != nil {
		t.Fatalf("parse %s: %v", f.ID, err)
	}
	return BuildBoard(dp, p1, p2)
}

// Every launch deployment pattern should resolve to the standard 5-objective
// layout: exactly one central, two home (one per side), and two expansion.
func TestRoleDerivationAcrossAllPatterns(t *testing.T) {
	for _, f := range loadDeploymentPatterns(t) {
		b := buildFromFile(t, f, SideDefender, SideAttacker)
		if len(b.Objectives) != 5 {
			t.Errorf("%s: expected 5 objectives, got %d", f.ID, len(b.Objectives))
		}
		var central, home, expansion int
		homeSides := map[Side]int{}
		for _, o := range b.Objectives {
			switch o.Role {
			case RoleCentral:
				central++
			case RoleHome:
				home++
				homeSides[o.HomeSide]++
			case RoleExpansion:
				expansion++
			default:
				t.Errorf("%s: objective %d has no role", f.ID, o.Index)
			}
		}
		if central != 1 {
			t.Errorf("%s: expected 1 central objective, got %d", f.ID, central)
		}
		if home != 2 {
			t.Errorf("%s: expected 2 home objectives, got %d", f.ID, home)
		}
		if homeSides[SideAttacker] != 1 || homeSides[SideDefender] != 1 {
			t.Errorf("%s: expected one home per side, got %v", f.ID, homeSides)
		}
		if expansion != 2 {
			t.Errorf("%s: expected 2 expansion objectives, got %d", f.ID, expansion)
		}
	}
}

// Spot-check tipping-point against hand-verified roles.
func TestTippingPointRoles(t *testing.T) {
	var tp *deploymentPatternFile
	for _, f := range loadDeploymentPatterns(t) {
		if f.ID == "tipping-point" {
			ff := f
			tp = &ff
			break
		}
	}
	if tp == nil {
		t.Skip("tipping-point pattern not present")
	}
	b := buildFromFile(t, *tp, SideDefender, SideAttacker)

	want := map[Point]struct {
		role ObjectiveRole
		home Side
	}{
		{30, 22}: {RoleCentral, ""},
		{22, 8}:  {RoleExpansion, ""},
		{14, 34}: {RoleHome, SideDefender},
		{38, 36}: {RoleExpansion, ""},
		{46, 10}: {RoleHome, SideAttacker},
	}
	for _, o := range b.Objectives {
		w, ok := want[o.Point]
		if !ok {
			t.Errorf("unexpected objective at %+v", o.Point)
			continue
		}
		if o.Role != w.role {
			t.Errorf("objective %+v: role = %q, want %q", o.Point, o.Role, w.role)
		}
		if o.HomeSide != w.home {
			t.Errorf("objective %+v: homeSide = %q, want %q", o.Point, o.HomeSide, w.home)
		}
	}
}

// Sides map to player numbers correctly.
func TestSideOf(t *testing.T) {
	b := Board{PlayerSides: [2]Side{SideDefender, SideAttacker}}
	if b.SideOf(1) != SideDefender {
		t.Errorf("player 1 side = %q, want defender", b.SideOf(1))
	}
	if b.SideOf(2) != SideAttacker {
		t.Errorf("player 2 side = %q, want attacker", b.SideOf(2))
	}
	if b.SideOf(3) != "" {
		t.Errorf("invalid player should have empty side")
	}
}

func TestObjectiveTags(t *testing.T) {
	o := Objective{}
	o.AddTag("decoyed")
	o.AddTag("decoyed") // idempotent
	if len(o.Tags) != 1 || !o.HasTag("decoyed") {
		t.Errorf("tag handling broken: %v", o.Tags)
	}
	if o.HasTag("consecrated") {
		t.Errorf("unexpected tag")
	}
}
