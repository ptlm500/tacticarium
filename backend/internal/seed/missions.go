package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type missionEntry struct {
	ID   string `json:"id"`
	Lore string `json:"lore"`
	Body string `json:"body"`
}

// slugToName converts a slug like "take-and-hold" to "Take and Hold"
func slugToName(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func SeedMissions(ctx context.Context, pool *pgxpool.Pool, filePath string) (stats MissionSeedStats, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return stats, fmt.Errorf("reading missions file: %w", err)
	}

	var entries []missionEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return stats, fmt.Errorf("parsing missions JSON: %w", err)
	}

	// Define mission pack prefixes and their pack IDs
	type packConfig struct {
		packID   string
		packName string
	}

	packs := map[string]packConfig{
		"chapter-approved-2025-26": {packID: "chapter-approved-2025-26", packName: "Chapter Approved 2025-26"},
		"pariah-nexus":             {packID: "pariah-nexus", packName: "Pariah Nexus"},
		"leviathan":               {packID: "leviathan", packName: "Leviathan"},
	}

	// Ensure mission packs exist
	for _, pc := range packs {
		_, err := pool.Exec(ctx,
			`INSERT INTO mission_packs (id, name) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET name = $2`,
			pc.packID, pc.packName)
		if err != nil {
			return stats, fmt.Errorf("upserting mission pack %s: %w", pc.packID, err)
		}
	}

	scoringRules := missionScoringRules()

	for _, entry := range entries {
		// Skip generic setup rules and entries without a pack prefix
		if strings.HasPrefix(entry.ID, "setup-rule-") {
			continue
		}

		packID, entryType, slug := classifyEntry(entry.ID)
		if packID == "" || entryType == "" {
			continue
		}

		name := slugToName(slug)

		switch entryType {
		case "mission":
			rulesJSON := []byte("[]")
			if rules, ok := scoringRules[entry.ID]; ok {
				rulesJSON, _ = json.Marshal(rules)
			}
			_, err := pool.Exec(ctx,
				`INSERT INTO missions (id, mission_pack_id, name, lore, description, scoring_rules)
				 VALUES ($1, $2, $3, $4, $5, $6)
				 ON CONFLICT (id) DO UPDATE SET name = $3, lore = $4, description = $5, scoring_rules = $6`,
				entry.ID, packID, name, entry.Lore, entry.Body, rulesJSON)
			if err != nil {
				return stats, fmt.Errorf("inserting mission %s: %w", entry.ID, err)
			}
			stats.Missions++

		case "mission-rule":
			_, err := pool.Exec(ctx,
				`INSERT INTO mission_rules (id, mission_pack_id, name, lore, description)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO UPDATE SET name = $3, lore = $4, description = $5`,
				entry.ID, packID, name, entry.Lore, entry.Body)
			if err != nil {
				return stats, fmt.Errorf("inserting mission rule %s: %w", entry.ID, err)
			}
			stats.MissionRules++

		case "secondary":
			isFixed := strings.Contains(entry.Body, "FIXED")
			maxVP := 5
			if isFixed {
				maxVP = 20
			}
			_, err := pool.Exec(ctx,
				`INSERT INTO secondaries (id, mission_pack_id, name, lore, description, max_vp, is_fixed)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)
				 ON CONFLICT (id) DO UPDATE SET name = $3, lore = $4, description = $5, max_vp = $6, is_fixed = $7`,
				entry.ID, packID, name, entry.Lore, entry.Body, maxVP, isFixed)
			if err != nil {
				return stats, fmt.Errorf("inserting secondary %s: %w", entry.ID, err)
			}
			stats.Secondaries++

		case "challenger-card":
			_, err := pool.Exec(ctx,
				`INSERT INTO challenger_cards (id, mission_pack_id, name, lore, description)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO UPDATE SET name = $3, lore = $4, description = $5`,
				entry.ID, packID, name, entry.Lore, entry.Body)
			if err != nil {
				return stats, fmt.Errorf("inserting challenger card %s: %w", entry.ID, err)
			}
			stats.ChallengerCards++

		case "gambit":
			vpValue := 30
			_, err := pool.Exec(ctx,
				`INSERT INTO gambits (id, mission_pack_id, name, description, vp_value)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO UPDATE SET name = $3, description = $4, vp_value = $5`,
				entry.ID, packID, name, entry.Body, vpValue)
			if err != nil {
				return stats, fmt.Errorf("inserting gambit %s: %w", entry.ID, err)
			}
			stats.Gambits++

		case "secret-mission":
			// Store secret missions as challenger cards for now (Pariah Nexus specific)
			_, err := pool.Exec(ctx,
				`INSERT INTO challenger_cards (id, mission_pack_id, name, lore, description)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO UPDATE SET name = $3, lore = $4, description = $5`,
				entry.ID, packID, name, entry.Lore, entry.Body)
			if err != nil {
				return stats, fmt.Errorf("inserting secret mission %s: %w", entry.ID, err)
			}
			stats.ChallengerCards++
		}
	}

	return stats, nil
}

// classifyEntry parses an entry ID into (packID, entryType, slug).
// Returns empty strings if the entry doesn't match any known pattern.
func classifyEntry(id string) (packID, entryType, slug string) {
	// Order matters: check longer prefixes first

	// Challenger cards (only in Chapter Approved)
	if p := "challenger-card-mission-chapter-approved-2025-26-"; strings.HasPrefix(id, p) {
		return "chapter-approved-2025-26", "challenger-card", id[len(p):]
	}

	// Secondary missions
	for _, pack := range []struct{ prefix, packID string }{
		{"secondary-mission-chapter-approved-2025-26-", "chapter-approved-2025-26"},
		{"secondary-mission-pariah-nexus-", "pariah-nexus"},
	} {
		if strings.HasPrefix(id, pack.prefix) {
			return pack.packID, "secondary", id[len(pack.prefix):]
		}
	}
	// Leviathan secondaries have no pack prefix
	if strings.HasPrefix(id, "secondary-mission-") {
		slug := id[len("secondary-mission-"):]
		// Skip if it matches a pack-specific prefix we already handled
		if !strings.HasPrefix(slug, "chapter-approved-") && !strings.HasPrefix(slug, "pariah-nexus-") {
			return "leviathan", "secondary", slug
		}
	}

	// Mission rules / twists
	for _, pack := range []struct{ prefix, packID string }{
		{"mission-rule-chapter-approved-2025-26-", "chapter-approved-2025-26"},
		{"mission-rule-pariah-nexus-", "pariah-nexus"},
		{"mission-rule-leviathan-", "leviathan"},
	} {
		if strings.HasPrefix(id, pack.prefix) {
			return pack.packID, "mission-rule", id[len(pack.prefix):]
		}
	}

	// Primary missions
	for _, pack := range []struct{ prefix, packID string }{
		{"mission-chapter-approved-2025-26-", "chapter-approved-2025-26"},
		{"mission-pariah-nexus-", "pariah-nexus"},
		{"mission-leviathan-", "leviathan"},
	} {
		if strings.HasPrefix(id, pack.prefix) {
			return pack.packID, "mission", id[len(pack.prefix):]
		}
	}

	// Gambits (shared, no pack prefix — assign to leviathan)
	if strings.HasPrefix(id, "gambit-") {
		return "leviathan", "gambit", id[len("gambit-"):]
	}

	// Secret missions (Pariah Nexus)
	if strings.HasPrefix(id, "secret-mission-pariah-nexus-") {
		return "pariah-nexus", "secret-mission", id[len("secret-mission-pariah-nexus-"):]
	}

	return "", "", ""
}

type MissionSeedStats struct {
	Missions        int
	MissionRules    int
	Secondaries     int
	ChallengerCards int
	Gambits         int
}

type scoringAction struct {
	Label       string `json:"label"`
	VP          int    `json:"vp"`
	MinRound    int    `json:"minRound,omitempty"`
	Description string `json:"description,omitempty"`
}

func missionScoringRules() map[string][]scoringAction {
	return map[string][]scoringAction{
		"mission-chapter-approved-2025-26-take-and-hold": {
			{Label: "1 Objective", VP: 5, MinRound: 2},
			{Label: "2 Objectives", VP: 10, MinRound: 2},
			{Label: "3 Objectives", VP: 15, MinRound: 2},
		},
		"mission-chapter-approved-2025-26-scorched-earth": {
			{Label: "Control 1 Objective", VP: 5, MinRound: 2},
			{Label: "Control 2 Objectives", VP: 10, MinRound: 2},
			{Label: "Burn (No Man's Land)", VP: 5, MinRound: 2},
			{Label: "Burn (Enemy Zone)", VP: 10, MinRound: 2},
		},
		"mission-chapter-approved-2025-26-purge-the-foe": {
			{Label: "Destroyed 1+ enemy unit", VP: 4},
			{Label: "Destroyed more than lost", VP: 4, MinRound: 2},
			{Label: "Control 1+ objective", VP: 4, MinRound: 2},
			{Label: "Control more objectives", VP: 4, MinRound: 2},
		},
		"mission-chapter-approved-2025-26-the-ritual": {
			{Label: "1 NML Objective", VP: 5, MinRound: 2},
			{Label: "2 NML Objectives", VP: 10, MinRound: 2},
			{Label: "3 NML Objectives", VP: 15, MinRound: 2},
		},
		"mission-chapter-approved-2025-26-supply-drop": {
			{Label: "1 NML Obj (R2-3)", VP: 5, MinRound: 2, Description: "Rounds 2-3"},
			{Label: "2 NML Obj (R2-3)", VP: 10, MinRound: 2, Description: "Rounds 2-3"},
			{Label: "1 NML Obj (R4)", VP: 8, MinRound: 4, Description: "Round 4"},
			{Label: "1 NML Obj (R5)", VP: 15, MinRound: 5, Description: "Round 5"},
		},
		"mission-chapter-approved-2025-26-burden-of-trust": {
			{Label: "1 Obj outside deploy zone", VP: 4, MinRound: 2},
			{Label: "2 Obj outside deploy zone", VP: 8, MinRound: 2},
			{Label: "Opponent guards (per unit)", VP: 2, MinRound: 2, Description: "Opponent scores this"},
		},
		"mission-chapter-approved-2025-26-terraform": {
			{Label: "1 Objective controlled", VP: 4, MinRound: 2},
			{Label: "2 Objectives controlled", VP: 8, MinRound: 2},
			{Label: "3 Objectives controlled", VP: 12, MinRound: 2},
			{Label: "Terraformed marker", VP: 1, MinRound: 2},
		},
		"mission-chapter-approved-2025-26-unexploded-ordinance": {
			{Label: "Hazard in enemy zone", VP: 8, MinRound: 2},
			{Label: "Hazard within 6\"", VP: 5, MinRound: 2},
			{Label: "Hazard within 12\"", VP: 2, MinRound: 2},
		},
		"mission-chapter-approved-2025-26-linchpin": {
			{Label: "Centre + 1 other", VP: 8, MinRound: 2, Description: "Control centre + 1 other objective"},
			{Label: "Centre + 2 others", VP: 13, MinRound: 2, Description: "Control centre + 2 other objectives"},
			{Label: "Centre + 3 others", VP: 15, MinRound: 2, Description: "Control centre + all others (capped at 15)"},
			{Label: "No centre (per obj)", VP: 3, MinRound: 2, Description: "Don't control centre: 3VP per objective"},
		},
		"mission-chapter-approved-2025-26-hidden-supplies": {
			{Label: "1 Obj outside zone", VP: 5, MinRound: 2},
			{Label: "2 Obj outside zone", VP: 10, MinRound: 2, Description: "Cumulative: 5+5"},
			{Label: "2 Obj + more than opp", VP: 15, MinRound: 2, Description: "Cumulative: 5+5+5"},
		},
	}
}
