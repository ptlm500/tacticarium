package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type factionJSON struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	ParentFactionID *string     `json:"parent_faction_id"`
	FactionRuleID   *string     `json:"faction_rule_id"`
	GameVersion     gameVersion `json:"game_version"`
}

// SeedFactions upserts every faction in a 40kdc-data factions.json array
// (one such file per faction directory).
func SeedFactions(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	var factions []factionJSON
	if err := readJSON(filePath, &factions); err != nil {
		return 0, err
	}

	count := 0
	for _, f := range factions {
		if f.ID == "" {
			continue
		}
		_, err := pool.Exec(ctx,
			`INSERT INTO factions (id, name, parent_faction_id, faction_rule_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (id) DO UPDATE
			   SET name = $2, parent_faction_id = $3, faction_rule_id = $4`,
			f.ID, f.Name, f.ParentFactionID, f.FactionRuleID)
		if err != nil {
			return count, fmt.Errorf("inserting faction %s: %w", f.ID, err)
		}
		count++
	}
	return count, nil
}
