package seed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type stratagemJSON struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Category           string          `json:"category"`
	Type               *string         `json:"type"`
	DetachmentID       *string         `json:"detachment_id"`
	CPCost             int             `json:"cp_cost"`
	Phases             []string        `json:"phases"`
	PlayerTurn         *string         `json:"player_turn"`
	Timing             *string         `json:"timing"`
	TargetRestrictions json.RawMessage `json:"target_restrictions"`
	AbilityID          *string         `json:"ability_id"`
	GameVersion        gameVersion     `json:"game_version"`
}

// SeedStratagems upserts every stratagem in a 40kdc-data stratagems.json array
// (the top-level core file or a per-faction file). The stratagem's faction_id
// is resolved from its detachment (detachments must be seeded first); core
// stratagems have no detachment and therefore a null faction_id.
func SeedStratagems(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var stratagems []stratagemJSON
	if err := readJSON(filePath, &stratagems); err != nil {
		return 0, err
	}

	detachmentFactions, err := loadDetachmentFactions(ctx, pool)
	if err != nil {
		return 0, fmt.Errorf("loading detachment factions: %w", err)
	}

	count := 0
	for _, s := range stratagems {
		if s.ID == "" {
			continue
		}

		var factionID *string
		if s.DetachmentID != nil {
			if fid, ok := detachmentFactions[*s.DetachmentID]; ok {
				factionID = &fid
			}
		}

		phases := s.Phases
		if phases == nil {
			phases = []string{}
		}

		var targetRestrictions *string
		if len(s.TargetRestrictions) > 0 && string(s.TargetRestrictions) != "null" {
			tr := string(s.TargetRestrictions)
			targetRestrictions = &tr
		}

		// The same stratagem id recurs across detachments (sometimes with
		// differing attributes), so the surrogate primary key is detachment-scoped.
		scope := "core"
		if s.DetachmentID != nil {
			scope = *s.DetachmentID
		}
		pk := scope + "/" + s.ID

		_, err := pool.Exec(ctx,
			`INSERT INTO stratagems
			   (pk, id, faction_id, detachment_id, name, type, cp_cost, category, phases, player_turn, timing, target_restrictions, ability_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			 ON CONFLICT (pk) DO UPDATE SET
			   id = $2, faction_id = $3, detachment_id = $4, name = $5, type = $6, cp_cost = $7,
			   category = $8, phases = $9, player_turn = $10, timing = $11,
			   target_restrictions = $12, ability_id = $13`,
			pk, s.ID, factionID, s.DetachmentID, s.Name, s.Type, s.CPCost, s.Category,
			phases, s.PlayerTurn, s.Timing, targetRestrictions, s.AbilityID)
		if err != nil {
			return count, fmt.Errorf("inserting stratagem %s: %w", s.ID, err)
		}
		count++
	}
	return count, nil
}

func loadDetachmentFactions(ctx context.Context, pool *pgxpool.Pool) (map[string]string, error) {
	rows, err := pool.Query(ctx, `SELECT id, faction_id FROM detachments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var id, factionID string
		if err := rows.Scan(&id, &factionID); err != nil {
			return nil, err
		}
		out[id] = factionID
	}
	return out, rows.Err()
}
